/*
 * @Author: jason chen (jasonchen@leansoftx.com, http://smallidea.cnblogs.com)
 * @Description:
 * @Date: 2021-11
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-03-16 14:13:54
 */
package start

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	smartideServer "github.com/leansoftX/smartide-cli/cmd/server"
	"github.com/leansoftX/smartide-cli/internal/biz/config"
	"github.com/leansoftX/smartide-cli/internal/biz/workspace"
	"github.com/leansoftX/smartide-cli/internal/dal"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/leansoftX/smartide-cli/pkg/docker/compose"
	"github.com/leansoftX/smartide-cli/pkg/tunnel"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// 远程服务器执行 start 命令
func ExecuteVmStartCmd(workspaceInfo workspace.WorkspaceInfo, yamlExecuteFun func(yamlConfig config.SmartIdeConfig), cmd *cobra.Command) {
	common.SmartIDELog.Info(i18nInstance.VmStart.Info_starting)

	mode, _ := cmd.Flags().GetString("mode")
	isModeServer := strings.ToLower(mode) == "server"

	// 错误反馈
	serverFeedback := func(err error) {
		if !isModeServer {
			return
		}
		if err != nil {
			smartideServer.Feedback_Finish(smartideServer.FeedbackCommandEnum_Start, cmd, false, 0, workspace.WorkspaceInfo{}, err.Error())
		}
	}

	//0. 连接到远程主机
	msg := fmt.Sprintf(" %v@%v:%v ...", workspaceInfo.Remote.UserName, workspaceInfo.Remote.Addr, workspaceInfo.Remote.SSHPort)
	common.SmartIDELog.Info(i18nInstance.VmStart.Info_connect_remote + msg)
	sshRemote, err := common.NewSSHRemote(workspaceInfo.Remote.Addr, workspaceInfo.Remote.SSHPort, workspaceInfo.Remote.UserName, workspaceInfo.Remote.Password)
	common.CheckErrorFunc(err, serverFeedback)

	//1. 检查远程主机是否有docker、docker-compose、git
	err = CheckRemoveEnv(sshRemote)
	common.CheckErrorFunc(err, serverFeedback)

	//2. git clone & checkout
	//2.1. 是否已 clone
	common.SmartIDELog.Info(i18nInstance.VmStart.Info_git_clone)
	isCloned := sshRemote.IsCloned(workspaceInfo.WorkingDirectoryPath)

	//2.2. git 操作
	if isCloned {
		common.SmartIDELog.Info(i18nInstance.VmStart.Info_git_cloned)
	} else {
		err = gitAction(sshRemote, workspaceInfo)
		common.CheckErrorFunc(err, serverFeedback)
	}

	//3. 获取配置文件的内容
	var ideBindingPort int
	var tempDockerCompose compose.DockerComposeYml
	ideYamlFilePath := common.FilePahtJoin(common.OS_Linux, workspaceInfo.WorkingDirectoryPath, workspaceInfo.ConfigFilePath) //fmt.Sprintf(`%v/.ide/.ide.yaml`, repoWorkspace)
	common.SmartIDELog.Info(fmt.Sprintf(i18nInstance.VmStart.Info_read_config, ideYamlFilePath))
	if !sshRemote.IsExit(ideYamlFilePath) {
		message := fmt.Sprintf(i18nInstance.Main.Err_file_not_exit2, ideYamlFilePath)
		common.SmartIDELog.Error(message)
	}
	catCommand := fmt.Sprintf(`cat %v`, ideYamlFilePath)
	output, err := sshRemote.ExeSSHCommand(catCommand)
	common.CheckErrorFunc(err, serverFeedback)
	configYamlContent := output
	currentConfig := config.NewConfigRemote(workspaceInfo.WorkingDirectoryPath, workspaceInfo.ConfigFilePath, configYamlContent)

	//3. docker-compose
	//3.1. 获取 compose 数据
	_, linkComposeFileContent := currentConfig.GetRemoteLinkDockerComposeFile(&sshRemote)
	yamlStr, err := currentConfig.ToYaml()
	common.CheckErrorFunc(err, serverFeedback)
	hasChanged := workspaceInfo.ChangeConfig(yamlStr, linkComposeFileContent) // 是否改变
	if hasChanged {                                                           // 改变包括了初始化
		// log

		if workspaceInfo.ID != "" {
			common.SmartIDELog.Info(i18nInstance.Start.Info_workspace_changed)

		} else {
			common.SmartIDELog.Info(i18nInstance.Start.Info_workspace_create)

		}

		// 获取compose配置
		tempDockerCompose, ideBindingPort, _ = currentConfig.ConvertToDockerCompose(sshRemote, workspaceInfo.GetProjectDirctoryName(), workspaceInfo.WorkingDirectoryPath, true)
		workspaceInfo.TempDockerCompose = tempDockerCompose

		// 配置
		workspaceInfo.ConfigYaml = *currentConfig

		// 链接的 docker-compose 文件
		if workspaceInfo.ConfigYaml.IsLinkDockerComposeFile() {
			yaml.Unmarshal([]byte(linkComposeFileContent), workspaceInfo.LinkDockerCompose)
		}

		// 扩展信息
		workspaceExtend := workspace.WorkspaceExtend{Ports: currentConfig.GetPortMappings()}
		workspaceInfo.Extend = workspaceExtend

		// 保存 docker-compose、config 文件
		err = workspaceInfo.SaveTempFilesForRemote(sshRemote) // 保存 docker-compose 文件
		common.CheckError(err)

	} else {
		// 先保存，确保临时文件存在	且 是最新的
		err := workspaceInfo.SaveTempFilesForRemote(sshRemote)
		common.CheckErrorFunc(err, serverFeedback)

		// 从临时文件中加载docker-compose
		tempDockerCompose, ideBindingPort, _ = currentConfig.LoadDockerComposeFromTempFile(sshRemote, workspaceInfo.TempDockerComposeFilePath)
	}

	//3.2. 扩展信息
	workspaceInfo.Extend = workspaceInfo.GetWorkspaceExtend()

	//4. ai 统计yaml
	yamlExecuteFun(*currentConfig)

	//5. docker 容器
	//5.1. 对应容器是否运行
	isDockerComposeRunning, err := isRemoteDockerComposeRunning(sshRemote, currentConfig.GetServiceNames())
	common.CheckErrorFunc(err, serverFeedback)

	//5.2. docker
	if !isDockerComposeRunning || hasChanged { // 容器没有运行 或者 有改变，重新创建容器
		// 创建网络
		common.SmartIDELog.Info(i18nInstance.VmStart.Info_create_network)
		networkCreateCommand := ""
		for network := range tempDockerCompose.Networks {
			cmd := fmt.Sprintf("docker network ls|grep %v > /dev/null || docker network create %v\n", network, network) // 不存在才创建
			networkCreateCommand += cmd
		}
		sshRemote.ExeSSHCommand(networkCreateCommand)

		// 在远程vm上生成docker-compose文件，运行docker-compose up
		common.SmartIDELog.Info(i18nInstance.VmStart.Info_compose_up) // 提示文本：compose up
		bytesDockerComposeContent, err := yaml.Marshal(&tempDockerCompose)
		printServices(tempDockerCompose.Services) // 打印services
		common.CheckError(err, string(bytesDockerComposeContent))
		commandCreateDockerComposeFile := fmt.Sprintf("docker-compose -f %v --project-directory %v up -d",
			workspaceInfo.TempDockerComposeFilePath, workspaceInfo.WorkingDirectoryPath)
		fmt.Println() // 避免向前覆盖
		fun1 := func(output string) error {
			if strings.Contains(output, ":error") || strings.Contains(output, ":fatal") {
				common.SmartIDELog.Error(output)

			} else {
				common.SmartIDELog.ConsoleInLine(output)
				if strings.Contains(output, "Pulling") {
					fmt.Println()
				}
			}

			return nil
		}
		err = sshRemote.ExecSSHCommandRealTimeFunc(commandCreateDockerComposeFile, fun1)
		fmt.Println()
		common.CheckError(err, commandCreateDockerComposeFile)

	}

	//6. 当前主机绑定到远程端口
	common.SmartIDELog.Info(i18nInstance.VmStart.Info_tunnel_waiting) // log
	var addrMapping map[string]string = map[string]string{}
	remotePortBindings := tempDockerCompose.GetPortBindings()
	unusedLocalPort4IdeBindingPort := ideBindingPort // 未使用的本地端口，与ide端口对应
	//6.1. 查找所有远程主机的端口
	for remoteBindingPort, containerPort := range remotePortBindings {
		remoteBindingPortInt, _ := strconv.Atoi(remoteBindingPort)
		unusedLocalPort, err := common.CheckAndGetAvailableLocalPort(remoteBindingPortInt, 100) // 得到一个未被占用的本地端口
		if err != nil {
			common.SmartIDELog.Warning(err.Error())
		}
		if remoteBindingPortInt == ideBindingPort && unusedLocalPort != ideBindingPort {
			unusedLocalPort4IdeBindingPort = unusedLocalPort
		}
		addrMapping["localhost:"+strconv.Itoa(unusedLocalPort)] = "localhost:" + remoteBindingPort

		// 日志
		unusedLocalPortStr := strconv.Itoa(unusedLocalPort)
		// 【注意】这里非常的绕！！！ 远程主机的docker-compose才保存了端口的label信息，所以只能使用远程主机的端口
		containerPortInt, _ := strconv.Atoi(containerPort)
		label := currentConfig.GetLabelWithPort(remoteBindingPortInt, containerPortInt)
		if label != "" {
			unusedLocalPortStr += fmt.Sprintf("(%v)", label)
		}
		msg := fmt.Sprintf("localhost:%v -> %v:%v -> container:%v",
			unusedLocalPortStr, workspaceInfo.Remote.Addr, remoteBindingPort, containerPort)
		common.SmartIDELog.Info(msg)
	}
	//6.2. 执行绑定
	tunnel.TunnelMultiple(sshRemote.Connection, addrMapping)

	//7. 保存数据
	if hasChanged {
		common.SmartIDELog.InfoF(i18nInstance.Start.Info_workspace_saving) // log

		remoteDockerComposeContainers, err := GetRemoteContainersWithServices(sshRemote, currentConfig.GetServiceNames())
		common.CheckErrorFunc(err, serverFeedback)
		//7.1. 补充数据
		devContainerName := getDevContainerName(remoteDockerComposeContainers, currentConfig.Workspace.DevContainer.ServiceName)
		workspaceInfo.Name = devContainerName
		workspaceInfo.TempDockerCompose = tempDockerCompose
		//7.2. save
		workspaceId, err := dal.InsertOrUpdateWorkspace(workspaceInfo)
		common.CheckErrorFunc(err, serverFeedback)
		common.SmartIDELog.InfoF(i18nInstance.Start.Info_workspace_saved, workspaceId)

	}

	//8. 打开浏览器
	var url string
	//vscode启动时候默认打开文件夹处理
	common.SmartIDELog.Info(i18nInstance.VmStart.Info_warting_for_webide)
	switch strings.ToLower(currentConfig.Workspace.DevContainer.IdeType) {
	case "vscode":
		url = fmt.Sprintf("http://localhost:%v/?folder=vscode-remote://localhost:%v%v",
			unusedLocalPort4IdeBindingPort, unusedLocalPort4IdeBindingPort, workspaceInfo.GetContainerWorkingPathWithVolumes())
	case "jb-projector":
		url = fmt.Sprintf(`http://localhost:%v`, unusedLocalPort4IdeBindingPort)
	default:
		url = fmt.Sprintf(`http://localhost:%v`, unusedLocalPort4IdeBindingPort)
	}
	if isModeServer { // mode server 模式下，不打开浏览器
		common.SmartIDELog.InfoF(i18nInstance.VmStart.Info_open_brower, url)
	} else {
		// 检查url是否可以正常打开，可以正常访问代表容器运行正常
		isUrlReady := false
		for !isUrlReady {
			resp, err := http.Get(url)
			if (err == nil) && (resp.StatusCode == 200) {
				isUrlReady = true
				err = common.OpenBrowser(url)
				if err != nil {
					common.SmartIDELog.Importance(err.Error())
				}
				common.SmartIDELog.InfoF(i18nInstance.VmStart.Info_open_brower, url)
			} else {
				msg := fmt.Sprintf("%v 等待启动", url)
				common.SmartIDELog.Debug(msg)
			}
		}
	}

	//9. 反馈给smartide server
	if isModeServer {
		common.SmartIDELog.Info("feedback...")
		err = smartideServer.Feedback_Finish(smartideServer.FeedbackCommandEnum_Start, cmd, true, workspaceInfo.ConfigYaml.GetContainerWebIDEPort(), workspaceInfo, "")
		common.CheckError(err)
	}

	//99. 死循环进行驻守
	if isModeServer {
		common.SmartIDELog.Info("success")
		return
	}
	for {
		time.Sleep(500)
	}

}

// 打印 service 列表
func printServices(services map[string]compose.Service) {
	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
	fmt.Fprintln(w, "Service\tImage\tPorts\t")
	for serviceName, service := range services {
		line := fmt.Sprintf("%v\t%v\t%v\t", serviceName, service.Image, strings.Join(service.Ports, ";"))
		fmt.Fprintln(w, line)
	}
	w.Flush()
}

// docker-compose 对应的容器是否运行
func isRemoteDockerComposeRunning(sshRemote common.SSHRemote, serviceNames []string) (isDockerComposeRunning bool, err error) {
	isDockerComposeRunning = false

	remoteContainers, err := GetRemoteContainersWithServices(sshRemote, serviceNames)
	if err != nil {
		return isDockerComposeRunning, err
	}
	if len(remoteContainers) > 0 {
		common.SmartIDELog.Info(i18nInstance.Start.Warn_docker_container_started)
		isDockerComposeRunning = true
	}

	return isDockerComposeRunning, err
}

// git 相关操作
func gitAction(sshRemote common.SSHRemote, workspace workspace.WorkspaceInfo) error {
	// 执行 git clone
	err := sshRemote.GitClone(workspace.GitCloneRepoUrl, workspace.WorkingDirectoryPath)
	//common.CheckErrorFunc(err, serverFeedback)
	if err != nil {
		return err
	}

	// git checkout
	checkoutCommand := "git fetch "
	if workspace.Branch != "" {
		checkoutCommand += "&& git checkout " + workspace.Branch
	} else { // 有可能当前目录所处的分支非主分支
		// 获取分支列表，确认主分支是master 还是 main
		output, _ := sshRemote.ExeSSHCommand(fmt.Sprintf("cd %v && git branch", workspace.WorkingDirectoryPath))
		branches := strings.Split(output, "\n")
		//isContainMaster := false
		for _, branch := range branches {
			if strings.Index(branch, "*") == 0 {
				checkoutCommand += "&& git checkout " + branch[2:]
			}

		}

	}

	// git checkout & pull
	common.SmartIDELog.Info(i18nInstance.VmStart.Info_git_checkout_and_pull)
	gitPullCommand := fmt.Sprintf("cd %v && %v && git pull && cd ~", workspace.WorkingDirectoryPath, checkoutCommand)
	err = sshRemote.ExecSSHCommandRealTime(gitPullCommand)
	return err
}
