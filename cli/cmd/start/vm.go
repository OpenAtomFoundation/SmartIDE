/*
 * @Author: jason chen (jasonchen@leansoftx.com, http://smallidea.cnblogs.com)
 * @Description:
 * @Date: 2021-11
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-09-14 14:42:49
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

	initExtended "github.com/leansoftX/smartide-cli/cmd/init"
	smartideServer "github.com/leansoftX/smartide-cli/cmd/server"
	"github.com/leansoftX/smartide-cli/internal/biz/config"
	"github.com/leansoftX/smartide-cli/internal/biz/workspace"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/leansoftX/smartide-cli/pkg/docker/compose"
	"github.com/leansoftX/smartide-cli/pkg/tunnel"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// 远程服务器执行 start 命令
func ExecuteVmStartCmd(workspaceInfo workspace.WorkspaceInfo, isUnforward bool,
	yamlExecuteFun func(yamlConfig config.SmartIdeConfig), cmd *cobra.Command, args []string, disableClone bool) {
	common.SmartIDELog.Info(i18nInstance.VmStart.Info_starting)

	mode, _ := cmd.Flags().GetString("mode")
	calbackAPI, _ := cmd.Flags().GetString("callback-api-address")
	userName, _ := cmd.Flags().GetString("serverusername")
	isModeServer := strings.ToLower(mode) == "server"
	isModePipeline := strings.ToLower(mode) == "pipeline"

	// 错误反馈
	serverFeedback := func(err error) {
		if !isModeServer {
			return
		}
		if err != nil {
			smartideServer.Feedback_Finish(smartideServer.FeedbackCommandEnum_Start, cmd, false, nil, workspaceInfo, err.Error(), "")
		}
	}

	//0. 连接到远程主机
	msg := fmt.Sprintf(" %v@%v:%v ...", workspaceInfo.Remote.UserName, workspaceInfo.Remote.Addr, workspaceInfo.Remote.SSHPort)
	common.SmartIDELog.Info(i18nInstance.VmStart.Info_connect_remote + msg)
	sshRemote, err := common.NewSSHRemote(workspaceInfo.Remote.Addr, workspaceInfo.Remote.SSHPort, workspaceInfo.Remote.UserName, workspaceInfo.Remote.Password)
	common.CheckErrorFunc(err, serverFeedback)

	//1. 检查远程主机是否有docker、docker-compose、git
	err = sshRemote.CheckRemoteEnv()
	common.CheckErrorFunc(err, serverFeedback)

	//2. git clone & checkout
	if !disableClone { // 是否禁止clone
		//2.1. 是否已 clone
		common.SmartIDELog.Info(i18nInstance.VmStart.Info_git_clone)
		isCloned := sshRemote.IsCloned(workspaceInfo.WorkingDirectoryPath)

		//2.2. git 操作
		if isCloned {
			common.SmartIDELog.Info(i18nInstance.VmStart.Info_git_cloned)
		} else {
			// 执行ssh-key 策略
			sshRemote.ExecSSHkeyPolicy(common.SmartIDELog.Ws_id, cmd)
			// Generate Authorizedkeys
			err = gitAction(sshRemote, workspaceInfo, cmd)
			common.CheckErrorFunc(err, serverFeedback)
		}
	}
	sshRemote.AddPublicKeyIntoAuthorizedkeys()

	//3. 获取配置文件的内容
	var ideBindingPort int
	var tempDockerCompose compose.DockerComposeYml
	ideYamlFilePath := common.FilePathJoin(common.OS_Linux, workspaceInfo.WorkingDirectoryPath, workspaceInfo.ConfigFileRelativePath) //fmt.Sprintf(`%v/.ide/.ide.yaml`, repoWorkspace)
	common.SmartIDELog.Info(fmt.Sprintf(i18nInstance.VmStart.Info_read_config, ideYamlFilePath))
	if !sshRemote.IsFileExist(ideYamlFilePath) {
		argsTemplateTypeName := ""
		argsTemplateSubTypeName := ""
		if len(args) > 0 {

			common.SmartIDELog.Info(i18nInstance.Init.Info_check_cmdtemplate)

			if cmd.Name() == "start" && len(cmd.Flags().Args()) == 2 {
				argsTemplateTypeName = args[1]
			}
			if cmd.Name() == "start" && len(cmd.Flags().Args()) == 1 {
				argsTemplateTypeName = args[0]
			}
			argsTemplateSubTypeName, err = cmd.Flags().GetString("type")
			if err != nil {
				return
			}
		}

		initExtended.GitCloneTemplateRepo4Remote(sshRemote, workspaceInfo.WorkingDirectoryPath, config.GlobalSmartIdeConfig.TemplateRepo, argsTemplateTypeName, argsTemplateSubTypeName)

	}
	currentConfig, err := config.NewRemoteConfig(&sshRemote,
		workspaceInfo.WorkingDirectoryPath, workspaceInfo.ConfigFileRelativePath)
	common.CheckError(err)

	// addonEnable()
	if workspaceInfo.Addon.IsEnable {
		workspaceInfo = AddonEnable(workspaceInfo)
		currentConfig.AddonWebTerminal(workspaceInfo.Name, workspaceInfo.WorkingDirectoryPath)
	}

	//3. docker-compose
	//3.1. 获取 compose 数据
	yamlStr, err := currentConfig.ToYaml()
	common.CheckErrorFunc(err, serverFeedback)
	linkComposeFileContent, err := currentConfig.Workspace.LinkCompose.ToYaml()
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
		tempDockerCompose, ideBindingPort, _ = currentConfig.ConvertToDockerCompose(sshRemote,
			workspaceInfo.GetProjectDirctoryName(), workspaceInfo.WorkingDirectoryPath, true, userName)
		workspaceInfo.TempDockerCompose = tempDockerCompose

		// 配置
		workspaceInfo.ConfigYaml = *currentConfig

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
		tempDockerCompose, ideBindingPort, _ =
			currentConfig.LoadDockerComposeFromTempFile(sshRemote, workspaceInfo.TempYamlFileAbsolutePath)
	}

	//3.2. 扩展信息
	workspaceInfo.Extend = workspaceInfo.GetWorkspaceExtend()

	//4. ai 统计yaml
	yamlExecuteFun(*currentConfig)

	//5. docker 容器
	//5.1. 对应容器是否运行
	isDockerComposeRunning, err := isRemoteDockerComposeRunning(sshRemote,
		workspaceInfo.WorkingDirectoryPath, currentConfig.GetServiceNames())
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
			workspaceInfo.TempYamlFileAbsolutePath, workspaceInfo.WorkingDirectoryPath)
		fmt.Println() // 避免向前覆盖
		fun1 := func(output string) error {
			output = strings.ToLower(output)
			if strings.Contains(output, ":error") || strings.Contains(output, ":fatal") {
				common.SmartIDELog.Error(output)

			} else {
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
	var addrMapping map[string]string = map[string]string{}
	unusedLocalPort4IdeBindingPort := ideBindingPort // 未使用的本地端口，与ide端口对应
	//6.1. 查找所有远程主机的端口
	for serviceName, service := range tempDockerCompose.Services {
		for _, portBinding := range service.Ports {
			ports := strings.Split(portBinding, ":")
			remoteBindingPort, containerPort := ports[0], ports[1]

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
			// 【注意】这里非常的绕！！！ 远程主机的docker-compose才保存了端口的label信息，所以只能使用远程主机的端口
			containerPortInt, _ := strconv.Atoi(containerPort)
			label := currentConfig.GetLabelWithPort(0, remoteBindingPortInt, containerPortInt)

			for i, port := range workspaceInfo.Extend.Ports {
				if port.HostPortDesc == label ||
					(port.ServiceName == serviceName && port.CurrentHostPort == remoteBindingPortInt && port.OriginHostPort == containerPortInt) {
					workspaceInfo.Extend.Ports[i].CurrentHostPort = remoteBindingPortInt
					workspaceInfo.Extend.Ports[i].OldClientPort = port.ClientPort
					workspaceInfo.Extend.Ports[i].ClientPort = unusedLocalPort
					break
				}
			}
		}
	}

	//7. 保存数据
	if hasChanged {
		common.SmartIDELog.InfoF(i18nInstance.Start.Info_workspace_saving) // log

		remoteDockerComposeContainers, err := GetRemoteContainersWithServices(sshRemote,
			workspaceInfo.WorkingDirectoryPath, currentConfig.GetServiceNames())
		common.CheckErrorFunc(err, serverFeedback)
		//7.1. 补充数据
		devContainerName := getDevContainerName(remoteDockerComposeContainers, currentConfig.Workspace.DevContainer.ServiceName)
		if workspaceInfo.Name == "" {
			workspaceInfo.Name = devContainerName
		}
		workspaceInfo.TempDockerCompose = tempDockerCompose
		//7.2. save
		if workspaceInfo.CliRunningEnv == workspace.CliRunningEnvEnum_Client {
			reloadWorkSpaceId(&workspaceInfo)
			common.SmartIDELog.InfoF(i18nInstance.Start.Info_workspace_saved, workspaceInfo.ID)

		} else {
			common.SmartIDELog.Importance(fmt.Sprintf("当前运行环境为 %v，工作区不需要缓存到本地！", workspaceInfo.CliRunningEnv))
		}

	}

	//calback external api
	if calbackAPI != "" {
		containerWebIDEPort := workspaceInfo.ConfigYaml.GetContainerWebIDEPort()
		err = smartideServer.Send_WorkspaceInfo(calbackAPI, smartideServer.FeedbackCommandEnum_Start, cmd, true, containerWebIDEPort, workspaceInfo)
		common.CheckError(err)

	}

	//7. 如果是不进行端口映射，直接退出
	if isUnforward {
		return
	}

	//7.1 如果mode=pipeline，也不需要端口映射，直接退出
	if isModePipeline {
		common.SmartIDELog.InfoF(i18nInstance.Start.Info_pipeline_mode_success)
		IDEAddress := fmt.Sprintf("http://%v:%v/?folder=vscode-remote://%v:%v%v", workspaceInfo.Remote.Addr, ideBindingPort, workspaceInfo.Remote.Addr, ideBindingPort, workspaceInfo.GetContainerWorkingPathWithVolumes())
		common.SmartIDELog.InfoF(IDEAddress)

		return
	}

	//7.2. ssh config file update
	workspaceInfo.UpdateSSHConfig()

	//8. 端口绑定
	common.SmartIDELog.Info(i18nInstance.VmStart.Info_tunnel_waiting) // log
	for _, item := range workspaceInfo.Extend.Ports {
		unusedLocalPortStr := strconv.Itoa(item.ClientPort)

		// 【注意】这里非常的绕！！！ 远程主机的docker-compose才保存了端口的label信息，所以只能使用远程主机的端口
		label := item.HostPortDesc
		if label == "" {
			label = currentConfig.GetLabelWithPort(0, item.CurrentHostPort, item.ContainerPort)
		}
		if label != "" {
			unusedLocalPortStr += fmt.Sprintf("(%v)", label)
		}

		// 检查是否包含在端口转发列表中
		if _, ok := addrMapping[fmt.Sprintf("localhost:%v", item.ClientPort)]; ok {
			msg := fmt.Sprintf("localhost:%v -> %v:%v -> container:%v",
				unusedLocalPortStr, workspaceInfo.Remote.Addr, item.CurrentHostPort, item.ContainerPort)
			common.SmartIDELog.Info(msg)
		}

	}
	//8.1. 执行绑定
	tunnel.TunnelMultiple(sshRemote.Connection, addrMapping) // 端口转发
	//8.2. 打开浏览器
	if currentConfig.Workspace.DevContainer.IdeType != config.IdeTypeEnum_SDKOnly {
		var url string
		//vscode启动时候默认打开文件夹处理
		common.SmartIDELog.Info(i18nInstance.VmStart.Info_warting_for_webide)
		switch currentConfig.Workspace.DevContainer.IdeType {
		case config.IdeTypeEnum_VsCode:
			url = fmt.Sprintf("http://localhost:%v/?folder=vscode-remote://localhost:%v%v",
				unusedLocalPort4IdeBindingPort, unusedLocalPort4IdeBindingPort, workspaceInfo.GetContainerWorkingPathWithVolumes())
		case config.IdeTypeEnum_JbProjector:
			url = fmt.Sprintf(`http://localhost:%v`, unusedLocalPort4IdeBindingPort)
		case config.IdeTypeEnum_Opensumi:
			url = fmt.Sprintf(`http://localhost:%v/?workspaceDir=/home/project`, unusedLocalPort4IdeBindingPort)
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
						common.SmartIDELog.ImportanceWithError(err)
					}
					common.SmartIDELog.InfoF(i18nInstance.VmStart.Info_open_brower, url)
				} else {
					msg := fmt.Sprintf("%v 等待启动", url)
					common.SmartIDELog.Debug(msg)
				}
			}
		}
	}

	//9. 反馈给smartide server
	if isModeServer {
		//获取容器id
		containerId := ""
		dcc, err := GetRemoteContainersWithServices(sshRemote,
			workspaceInfo.WorkingDirectoryPath, []string{workspaceInfo.ConfigYaml.Workspace.DevContainer.ServiceName})
		if err != nil {
			common.SmartIDELog.ImportanceWithError(err)
		}
		if len(dcc) == 0 {
			common.SmartIDELog.Error("没有查找到运行的容器！")
		}
		if containerId, err = sshRemote.ExeSSHCommand(fmt.Sprintf("docker ps  -f 'name=%s' -q", dcc[len(dcc)-1].ContainerName)); containerId != "" && err == nil {
			// smartide-agent install
			workspace.InstallSmartideAgent(sshRemote, containerId, cmd, workspaceInfo.ServerWorkSpace.ID)
		}

		common.SmartIDELog.Info("feedback...")
		containerWebIDEPort := workspaceInfo.ConfigYaml.GetContainerWebIDEPort()
		err = smartideServer.Feedback_Finish(smartideServer.FeedbackCommandEnum_Start, cmd, true, containerWebIDEPort, workspaceInfo, "", containerId)
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
func isRemoteDockerComposeRunning(sshRemote common.SSHRemote, workingDir string, serviceNames []string) (isDockerComposeRunning bool, err error) {
	isDockerComposeRunning = false

	remoteContainers, err := GetRemoteContainersWithServices(sshRemote, workingDir, serviceNames)
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
func gitAction(sshRemote common.SSHRemote, workspace workspace.WorkspaceInfo, cmd *cobra.Command) error {
	// 执行 git clone
	err := sshRemote.GitClone(workspace.GitCloneRepoUrl, workspace.WorkingDirectoryPath, common.SmartIDELog.Ws_id, cmd)
	//common.CheckErrorFunc(err, serverFeedback)
	if err != nil {
		return err
	}
	checkoutCommand := ""
	isSSHClone := strings.Index(workspace.GitCloneRepoUrl, "git@") == 0
	fflags := cmd.Flags()
	userName, _ := fflags.GetString("serverusername")
	GIT_SSH_COMMAND := fmt.Sprintf(`GIT_SSH_COMMAND='ssh -i ~/.ssh/id_rsa_%s_%s -o IdentitiesOnly=yes'`, userName, common.SmartIDELog.Ws_id)
	// git checkout
	if isSSHClone {
		checkoutCommand = fmt.Sprintf("%s git fetch ", GIT_SSH_COMMAND)
		if workspace.Branch != "" {
			checkoutCommand += fmt.Sprintf("&& %s git checkout ", GIT_SSH_COMMAND) + workspace.Branch
		} else { // 有可能当前目录所处的分支非主分支
			// 获取分支列表，确认主分支是master 还是 main
			command := fmt.Sprintf("cd %v && %s git branch", workspace.WorkingDirectoryPath, GIT_SSH_COMMAND)
			output, _ := sshRemote.ExeSSHCommand(command)
			branches := strings.Split(output, "\n")
			//isContainMaster := false
			for _, branch := range branches {
				if strings.Index(branch, "*") == 0 {
					checkoutCommand += fmt.Sprintf("&& %s git checkout ", GIT_SSH_COMMAND) + branch[2:]
				}

			}

		}
	} else {
		checkoutCommand = "git fetch "
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

	}

	// git checkout & pull
	common.SmartIDELog.Info(i18nInstance.VmStart.Info_git_checkout_and_pull)
	gitPullCommand := ""
	if isSSHClone {
		gitPullCommand = fmt.Sprintf("cd %v && %v && %s git pull && cd ~", workspace.WorkingDirectoryPath, checkoutCommand, GIT_SSH_COMMAND)

	} else {
		gitPullCommand = fmt.Sprintf("cd %v && %v && git pull && cd ~", workspace.WorkingDirectoryPath, checkoutCommand)

	}
	err = sshRemote.ExecSSHCommandRealTime(gitPullCommand)
	return err
}
