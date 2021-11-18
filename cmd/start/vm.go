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
	"github.com/leansoftX/smartide-cli/cmd/lib"
	"github.com/leansoftX/smartide-cli/lib/common"
	"github.com/leansoftX/smartide-cli/lib/docker/compose"
	"github.com/leansoftX/smartide-cli/lib/i18n"
	"github.com/leansoftX/smartide-cli/lib/tunnel"
	"gopkg.in/yaml.v2"
)

// 远程服务器执行 start
func ExecuteVmStartCmd(workspace dal.WorkspaceInfo) {
	common.SmartIDELog.Info(i18n.GetInstance().VmStart.Info.Info_starting)

	//0. 连接到远程主机
	msg := fmt.Sprintf(" %v@%v:%v ...", workspace.Remote.UserName, workspace.Remote.Addr, workspace.Remote.SSHPort)
	common.SmartIDELog.Info(i18n.GetInstance().VmStart.Info.Info_connect_remote + msg)
	var sshRemote common.SSHRemote
	err := sshRemote.Instance(workspace.Remote.Addr, workspace.Remote.SSHPort, workspace.Remote.UserName, workspace.Remote.Password)
	common.CheckError(err)

	//1. 检查远程主机是否有docker、docker-compose、git
	err = CheckRemoveEnv(sshRemote)
	common.CheckError(err)

	//2. 在远程主机上执行相应的命令
	//2.1. 执行git clone
	common.SmartIDELog.Info(i18n.GetInstance().VmStart.Info.Info_git_clone)
	output, err := sshRemote.GitClone(workspace.GitCloneRepoUrl, workspace.WorkingDirectoryPath)
	common.CheckError(err, output)

	//2.2. git checkout
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

	//2.3. git checkout & pull
	common.SmartIDELog.Info(i18n.GetInstance().VmStart.Info.Info_git_checkout_and_pull)
	gitPullCommand := fmt.Sprintf("cd %v && %v && git pull && cd ~", workspace.WorkingDirectoryPath, checkoutCommand)
	output, err = sshRemote.ExeSSHCommand(gitPullCommand)
	common.CheckError(err, output)

	//2.4. 读取配置.ide.yaml 并 转换为docker-compose
	var ideBindingPort int
	var dockerCompose compose.DockerComposeYml
	ideYamlFilePath := common.FilePahtJoin(common.OS_Linux, workspace.WorkingDirectoryPath, workspace.ConfigFilePath) //fmt.Sprintf(`%v/.ide/.ide.yaml`, repoWorkspace)
	common.SmartIDELog.Info(fmt.Sprintf(i18n.GetInstance().VmStart.Info.Info_read_config, ideYamlFilePath))
	catCommand := fmt.Sprintf(`cat %v`, ideYamlFilePath)
	output, err = sshRemote.ExeSSHCommand(catCommand)
	if strings.Contains(output, "No such file or directory") {
		message := fmt.Sprintf("配置文件（%v）不存在", workspace.ConfigFilePath)
		common.SmartIDELog.Error(message)
	}
	common.CheckError(err, output)
	yamlContent := output
	var yamlFileCongfig lib.YamlFileConfig
	yamlFileCongfig.GetConfigWithStr(yamlContent)
	yamlFileCongfig.SetWorkspace(workspace.WorkingDirectoryPath, workspace.ConfigFilePath)

	//2.5. 获取vm的容器列表
	isDockerComposeRunning := false
	remoteContainers, err := GetRemoteContainersWithServices(sshRemote, yamlFileCongfig.GetServiceNames())
	common.CheckError(err)
	if len(remoteContainers) > 0 {
		common.SmartIDELog.Info(i18nInstance.Start.Error.Docker_started)
		isDockerComposeRunning = true
	}

	//2.6. docker
	if isDockerComposeRunning { // 当 docker-compose 对应的容器运行时
		// dockerCompose, ideBindingPort, _ = yamlFileCongfig.ConvertToDockerCompose(sshRemote, filepath.Dir(ideYamlFilePath), false)
		dockerCompose, ideBindingPort, _ = yamlFileCongfig.LoadDockerComposeFromTempFile(sshRemote, workspace.WorkingDirectoryPath, workspace.ProjectName)
	} else {
		dockerCompose, ideBindingPort, _ = yamlFileCongfig.ConvertToDockerCompose(sshRemote, workspace.WorkingDirectoryPath, true)

		// 保存 docker-compose 、config 文件
		err := yamlFileCongfig.SaveTempFilesForRemote(sshRemote, dockerCompose, workspace.ProjectName) // 保存docker-compose文件
		common.CheckError(err)

		// 创建网络
		common.SmartIDELog.Info(i18n.GetInstance().VmStart.Info.Info_create_network)
		networkCreateCommand := ""
		for network := range dockerCompose.Networks {
			networkCreateCommand += "docker network create " + network + "\n "
		}
		sshRemote.ExeSSHCommand(networkCreateCommand)

		// 在远程vm上生成docker-compose文件，运行docker-compose up
		common.SmartIDELog.Info(i18n.GetInstance().VmStart.Info.Info_compose_up) // 提示文本：compose up
		bytesDockerComposeContent, err := yaml.Marshal(&dockerCompose)
		printServices(dockerCompose.Services)                                                          // 打印services
		strDockerComposeContent := strings.ReplaceAll(string(bytesDockerComposeContent), "\"", "\\\"") // 文本中包含双引号
		common.CheckError(err, string(bytesDockerComposeContent))
		commandCreateDockerComposeFile := fmt.Sprintf(`
mkdir -p ~/.ide
rm -rf ~/.ide/docker-compose-%v.yaml
echo "%v" >> ~/.ide/docker-compose-%v.yaml
docker-compose -f ~/.ide/docker-compose-%v.yaml --project-directory %v up -d
`, workspace.ProjectName, strDockerComposeContent, workspace.ProjectName, workspace.ProjectName, workspace.WorkingDirectoryPath)
		err = sshRemote.ExecSSHCommandRealTime(commandCreateDockerComposeFile)
		common.CheckError(err, commandCreateDockerComposeFile)

	}

	//3. 当前主机绑定到远程端口
	var addrMapping map[string]string = map[string]string{}
	remotePortBindings := dockerCompose.GetPortBindings()
	unusedLocalPort4IdeBindingPort := ideBindingPort // 未使用的本地端口，与ide端口对应
	// 查找所有远程主机的端口
	for bindingPort, containerPort := range remotePortBindings {
		portInt, _ := strconv.Atoi(bindingPort)
		unusedLocalPort := common.CheckAndGetAvailablePort(portInt, 100) // 得到一个未被占用的本地端口
		if portInt == ideBindingPort && unusedLocalPort != ideBindingPort {
			unusedLocalPort4IdeBindingPort = unusedLocalPort
		}
		addrMapping["localhost:"+strconv.Itoa(unusedLocalPort)] = "localhost:" + bindingPort
		msg := fmt.Sprintf("localhost:%v -> %v:%v -> container:%v", unusedLocalPort, workspace.Remote.Addr, bindingPort, containerPort)
		common.SmartIDELog.Info(msg)
	}
	// 执行绑定
	tunnel.TunnelMultiple(sshRemote.Connection, addrMapping)

	//4. 保存数据
	remoteContainers, err = GetRemoteContainersWithServices(sshRemote, yamlFileCongfig.GetServiceNames())
	common.CheckError(err)
	//4.1. workspace name
	for _, dockerComposeContainer := range remoteContainers {
		if dockerComposeContainer.ServiceName == yamlFileCongfig.Workspace.DevContainer.ServiceName {
			workspace.Name = dockerComposeContainer.ContainerName
			break
		}
	}
	/* 	//4.2. 配置文件路径
	   	workspace.ConfigFilePath = workspace.ConfigFilePath
	   	//4.3. 工作目录
	   	workspace.LocalWorkingDirectoryPath = workspace.LocalWorkingDirectoryPath */
	workspaceId, err := dal.InsertOrUpdateWorkspace(workspace)
	common.CheckError(err)
	common.SmartIDELog.InfoF("数据保存成功，工作区ID: %v ", workspaceId)

	//4. 打开浏览器
	var url string
	//vscode启动时候默认打开文件夹处理
	if yamlFileCongfig.Workspace.DevContainer.IdeType == "vscode" {
		url = fmt.Sprintf("http://localhost:%v/?folder=vscode-remote://localhost:%v/home/project", unusedLocalPort4IdeBindingPort, unusedLocalPort4IdeBindingPort)
	} else {
		url = fmt.Sprintf(`http://localhost:%v`, unusedLocalPort4IdeBindingPort)
	}
	common.SmartIDELog.Info(i18n.GetInstance().VmStart.Info.Info_warting_for_webide) //TODO: 国际化
	go func(checkUrl string) {
		isUrlReady := false
		for !isUrlReady {
			resp, err := http.Get(checkUrl)
			if (err == nil) && (resp.StatusCode == 200) {
				isUrlReady = true
				common.OpenBrowser(checkUrl)
				common.SmartIDELog.InfoF(i18n.GetInstance().VmStart.Info.Info_open_brower, checkUrl)
			}
		}
	}(url)

	//5. 死循环进行驻守
	for {
		time.Sleep(500)
	}
}

// 打印 service 列表
func printServices(services map[string]compose.Service) {
	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
	fmt.Fprintln(w, "service\timage\tports\t")
	for serviceName, service := range services {
		line := fmt.Sprintf("%v\t%v\t%v\t", serviceName, service.Image.Name+":"+service.Image.Tag, strings.Join(service.Ports, ";"))
		fmt.Fprintln(w, line)
	}
	w.Flush()
}
