package start

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/leansoftX/smartide-cli/cmd/dal"
	"github.com/leansoftX/smartide-cli/lib/common"
	"github.com/leansoftX/smartide-cli/lib/docker/compose"
	"github.com/leansoftX/smartide-cli/lib/tunnel"
	"gopkg.in/yaml.v2"
)

// 远程服务器执行 start
func ExecuteVmStartCmd(workspace dal.WorkspaceInfo, yamlExecuteFun func(yamlConfig dal.YamlFileConfig)) {
	common.SmartIDELog.Info(i18nInstance.VmStart.Info_starting)

	//0. 连接到远程主机
	msg := fmt.Sprintf(" %v@%v:%v ...", workspace.Remote.UserName, workspace.Remote.Addr, workspace.Remote.SSHPort)
	common.SmartIDELog.Info(i18nInstance.VmStart.Info_connect_remote + msg)
	var sshRemote common.SSHRemote
	err := sshRemote.Instance(workspace.Remote.Addr, workspace.Remote.SSHPort, workspace.Remote.UserName, workspace.Remote.Password)
	common.CheckError(err)

	//1. 检查远程主机是否有docker、docker-compose、git
	err = CheckRemoveEnv(sshRemote)
	common.CheckError(err)

	//2. git clone & checkout
	//2.1. 是否已 clone
	common.SmartIDELog.Info(i18nInstance.VmStart.Info_git_clone)
	isCloned := sshRemote.IsCloned(workspace.WorkingDirectoryPath)

	//2.2. git 操作
	if isCloned {
		common.SmartIDELog.Info(i18nInstance.VmStart.Info_git_cloned)
	} else {
		gitAction(sshRemote, workspace)

	}

	//3. 配置文件
	var ideBindingPort int
	var tempDockerCompose compose.DockerComposeYml
	ideYamlFilePath := common.FilePahtJoin(common.OS_Linux, workspace.WorkingDirectoryPath, workspace.ConfigFilePath) //fmt.Sprintf(`%v/.ide/.ide.yaml`, repoWorkspace)
	common.SmartIDELog.Info(fmt.Sprintf(i18nInstance.VmStart.Info_read_config, ideYamlFilePath))
	catCommand := fmt.Sprintf(`cat %v`, ideYamlFilePath)
	output, err := sshRemote.ExeSSHCommand(catCommand)
	if strings.Contains(output, "No such file or directory") {
		message := fmt.Sprintf(i18nInstance.Main.Err_file_not_exit2, workspace.ConfigFilePath)
		common.SmartIDELog.Error(message)
	}
	common.CheckError(err, output)
	yamlContent := output
	var yamlFileCongfig dal.YamlFileConfig
	yamlFileCongfig.GetConfigWithStr(yamlContent)
	yamlFileCongfig.SetWorkspace(workspace.WorkingDirectoryPath, workspace.ConfigFilePath)

	//3. docker-compose
	//3.1. 获取compose数据
	_, linkComposeFileContent := yamlFileCongfig.GetLocalLinkDockerComposeFile()
	hasChanged := workspace.ChangeConfig(yamlFileCongfig.ToYaml(), linkComposeFileContent) // 是否改变
	if hasChanged {                                                                        // 改变包括了初始化
		// log
		common.SmartIDELog.Info("工作区配置改变！")

		// 获取compose配置
		tempDockerCompose, ideBindingPort, _ = yamlFileCongfig.ConvertToDockerCompose(sshRemote, workspace.ProjectName, workspace.WorkingDirectoryPath, true)

		//
		workspaceExtend := dal.WorkspaceExtend{Ports: yamlFileCongfig.GetPortMappings()}
		workspace.Extend = workspaceExtend
	} else {
		yamlFileCongfig.SaveTempFilesForRemote(sshRemote, workspace.TempDockerCompose, workspace.ProjectName) // 先保存，确保临时文件存在
		common.CheckError(err)

		tempDockerCompose, ideBindingPort, _ = yamlFileCongfig.LoadDockerComposeFromTempFile(sshRemote, workspace.WorkingDirectoryPath, workspace.ProjectName)
	}
	//3.2. 保存 docker-compose、config 文件
	tempRemoteDockerComposeFilePath, _, err := yamlFileCongfig.SaveTempFilesForRemote(sshRemote, tempDockerCompose, workspace.ProjectName) // 保存docker-compose文件
	common.CheckError(err)

	//4. ai统计yaml
	yamlExecuteFun(yamlFileCongfig)

	//5. docker 容器
	//5.1. 对应容器是否运行
	isDockerComposeRunning := isRemoteDockerComposeRunning(sshRemote, yamlFileCongfig.GetServiceNames())

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
		//strDockerComposeContent := strings.ReplaceAll(string(bytesDockerComposeContent), "\"", "\\\"") // 文本中包含双引号
		common.CheckError(err, string(bytesDockerComposeContent))
		commandCreateDockerComposeFile := fmt.Sprintf("docker-compose -f %v --project-directory %v up -d",
			tempRemoteDockerComposeFilePath, workspace.WorkingDirectoryPath)
		err = sshRemote.ExecSSHCommandRealTimeFunc(commandCreateDockerComposeFile, func(output string) error {
			if strings.Contains(output, ":error") || strings.Contains(output, ":fatal") {
				common.SmartIDELog.Error(output)

			} else {
				common.SmartIDELog.ConsoleInLine(output)
				if strings.Contains(output, "Pulling") {
					fmt.Println()
				}
			}

			return nil
		})
		fmt.Println()
		common.CheckError(err, commandCreateDockerComposeFile)

	}

	//6. 当前主机绑定到远程端口
	var addrMapping map[string]string = map[string]string{}
	remotePortBindings := tempDockerCompose.GetPortBindings()
	unusedLocalPort4IdeBindingPort := ideBindingPort // 未使用的本地端口，与ide端口对应
	//6.1. 查找所有远程主机的端口
	for remoteBindingPort, containerPort := range remotePortBindings {
		remoteBindingPortInt, _ := strconv.Atoi(remoteBindingPort)
		unusedLocalPort := common.CheckAndGetAvailableLocalPort(remoteBindingPortInt, 100) // 得到一个未被占用的本地端口
		if remoteBindingPortInt == ideBindingPort && unusedLocalPort != ideBindingPort {
			unusedLocalPort4IdeBindingPort = unusedLocalPort
		}
		addrMapping["localhost:"+strconv.Itoa(unusedLocalPort)] = "localhost:" + remoteBindingPort

		// 日志
		unusedLocalPortStr := strconv.Itoa(unusedLocalPort)
		// 【注意】这里非常的绕！！！ 远程主机的docker-compose才保存了端口的label信息，所以只能使用远程主机的端口
		label := yamlFileCongfig.GetLabelWithPort(remoteBindingPortInt)
		if label != "" {
			unusedLocalPortStr += fmt.Sprintf("(%v)", label)
		}
		msg := fmt.Sprintf("localhost:%v -> %v:%v -> container:%v", unusedLocalPortStr, workspace.Remote.Addr, remoteBindingPort, containerPort)
		common.SmartIDELog.Info(msg)
	}
	//6.2. 执行绑定
	tunnel.TunnelMultiple(sshRemote.Connection, addrMapping)

	//7. 保存数据
	if hasChanged {
		remoteDockerComposeContainers, err := GetRemoteContainersWithServices(sshRemote, yamlFileCongfig.GetServiceNames())
		common.CheckError(err)
		//7.1. 补充数据
		devContainerName := getDevContainerName(remoteDockerComposeContainers, yamlFileCongfig.Workspace.DevContainer.ServiceName)
		workspace.Name = devContainerName
		workspace.TempDockerCompose = tempDockerCompose
		workspace.TempDockerComposeFilePath = tempRemoteDockerComposeFilePath
		//7.2. save
		workspaceId, err := dal.InsertOrUpdateWorkspace(workspace)
		common.CheckError(err)
		common.SmartIDELog.InfoF(i18nInstance.Start.Info_workspace_saved, workspaceId)

	}

	//8. 打开浏览器
	var url string
	//vscode启动时候默认打开文件夹处理
	if yamlFileCongfig.Workspace.DevContainer.IdeType == "vscode" {
		url = fmt.Sprintf("http://localhost:%v/?folder=vscode-remote://localhost:%v/home/project/%v", unusedLocalPort4IdeBindingPort, unusedLocalPort4IdeBindingPort, workspace.ProjectName)
	} else {
		url = fmt.Sprintf(`http://localhost:%v`, unusedLocalPort4IdeBindingPort)
	}
	common.SmartIDELog.Info(i18nInstance.VmStart.Info_warting_for_webide)
	go func(checkUrl string) {
		isUrlReady := false
		for !isUrlReady {
			resp, err := http.Get(checkUrl)
			if (err == nil) && (resp.StatusCode == 200) {
				isUrlReady = true
				common.OpenBrowser(checkUrl)
				common.SmartIDELog.InfoF(i18nInstance.VmStart.Info_open_brower, checkUrl)
			}
		}
	}(url)

	//99. 死循环进行驻守
	for {
		time.Sleep(500)
	}
}

// 打印 service 列表
func printServices(services map[string]compose.Service) {
	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
	fmt.Fprintln(w, "Service\tImage\tPorts\t")
	for serviceName, service := range services {
		line := fmt.Sprintf("%v\t%v\t%v\t", serviceName, service.Image.Name+":"+service.Image.Tag, strings.Join(service.Ports, ";"))
		fmt.Fprintln(w, line)
	}
	w.Flush()
}

// docker-compose 对应的容器是否运行
func isRemoteDockerComposeRunning(sshRemote common.SSHRemote, serviceNames []string) bool {
	isDockerComposeRunning := false

	remoteContainers, err := GetRemoteContainersWithServices(sshRemote, serviceNames)
	common.CheckError(err)
	if len(remoteContainers) > 0 {
		common.SmartIDELog.Info(i18nInstance.Start.Warn_docker_container_started)
		isDockerComposeRunning = true
	}

	return isDockerComposeRunning
}

func gitAction(sshRemote common.SSHRemote, workspace dal.WorkspaceInfo) {
	// 执行git clone
	err := sshRemote.GitClone(workspace.GitCloneRepoUrl, workspace.WorkingDirectoryPath)
	common.CheckError(err)

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
	output, err := sshRemote.ExeSSHCommand(gitPullCommand)
	common.CheckError(err, output)
}
