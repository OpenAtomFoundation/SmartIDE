package start

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/leansoftX/smartide-cli/cmd/dal"
	"github.com/leansoftX/smartide-cli/lib/common"
	"github.com/leansoftX/smartide-cli/lib/docker/compose"
	"github.com/leansoftX/smartide-cli/lib/tunnel"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

// 本地执行 start
func ExecuteStartCmd(workspace dal.WorkspaceInfo, endPostExecuteFun func(dockerContainerName string, docker common.Docker), yamlExecuteFun func(yamlConfig dal.YamlFileConfig)) {

	//0. 检查本地环境
	err := CheckLocalEnv()
	common.CheckError(err)

	//1. 变量
	//1.1. ctx & cli
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	common.CheckError(err)

	//1.2. 私有变量
	var tempDockerCompose compose.DockerComposeYml // 最后用于运行docker-compose命令的yaml文件
	var ideBindingPort, sshBindingPort int

	//1.3. 初始化配置文件对象
	var yamlFileCongfig dal.YamlFileConfig
	yamlFileCongfig.SetWorkspace(workspace.WorkingDirectoryPath, workspace.ConfigFilePath)
	yamlFileCongfig.GetLocalConfig() // 读取本地配置
	localWorkingDir := yamlFileCongfig.GetLocalWorkingDirectry()
	dockeComposeTempYamlFilePath := yamlFileCongfig.GetTempDockerComposeFilePath(localWorkingDir, workspace.ProjectName) // 获取临时docker-compose文件的路径

	//2. docker-compose
	//2.1. 获取compose数据
	_, linkComposeFileContent := yamlFileCongfig.GetLocalLinkDockerComposeFile()
	hasChanged := workspace.ChangeConfig(yamlFileCongfig.ToYaml(), linkComposeFileContent) // 是否改变
	if hasChanged {                                                                        // 改变包括了初始化
		// log
		common.SmartIDELog.Info("工作区配置改变！")

		// 获取compose配置
		tempDockerCompose, ideBindingPort, sshBindingPort = yamlFileCongfig.ConvertToDockerCompose(common.SSHRemote{}, workspace.ProjectName, "", true) // 转换为docker compose格式

		// 更新端口绑定列表，只在改变的时候才需要赋值
		workspaceExtend := dal.WorkspaceExtend{Ports: yamlFileCongfig.GetPortMappings()}
		workspace.Extend = workspaceExtend
	} else {
		yamlFileCongfig.SaveTempFiles(workspace.TempDockerCompose, workspace.ProjectName) // 先保存，确保临时文件存在
		common.CheckError(err)

		tempDockerCompose, ideBindingPort, sshBindingPort = yamlFileCongfig.LoadDockerComposeFromTempFile(common.SSHRemote{}, "", workspace.ProjectName)
	}

	//2.2. 保存 docker-compose 、config 文件到临时文件夹
	err = yamlFileCongfig.SaveTempFiles(tempDockerCompose, workspace.ProjectName) // 保存docker-compose文件
	common.CheckError(err)

	//3. 容器
	//3.1. 校验 docker-compose 文件对应的环境是否已经启动
	isDockerComposeRunning := isDockerComposeRunning(ctx, cli, yamlFileCongfig.GetServiceNames())

	//3.2. 运行容器
	if !isDockerComposeRunning || hasChanged { // 容器没有运行 或者 有改变，重新创建容器
		// print
		common.SmartIDELog.InfoF(i18nInstance.Start.Info_docker_compose_filepath, dockeComposeTempYamlFilePath)
		common.SmartIDELog.InfoF(i18nInstance.Start.Info_ssh_tunnel, sshBindingPort) // 提示用户ssh端口绑定到了本地的某个端口

		// 创建网络
		for network := range tempDockerCompose.Networks {
			networkList, _ := cli.NetworkList(ctx, types.NetworkListOptions{})
			isContain := false
			for _, item := range networkList {
				if item.Name == network {
					isContain = true
					break
				}
			}
			if !isContain {
				cli.NetworkCreate(ctx, network, types.NetworkCreate{})
				common.SmartIDELog.InfoF(i18nInstance.Start.Info_create_network, network)
			}
		}

		// 运行docker-compose命令
		// e.g. docker-compose -f /Users/jasonchen/.ide/docker-compose-product-service-dev.yaml --project-directory /Users/jasonchen/Project/boat-house/boat-house-backend/src/product-service/api up -d
		pwd, _ := os.Getwd()
		composeCmd := exec.Command("docker-compose", "-f", dockeComposeTempYamlFilePath, "--project-directory", pwd, "up", "-d")
		composeCmd.Stdout = os.Stdout
		composeCmd.Stderr = os.Stderr
		if composeCmdErr := composeCmd.Run(); composeCmdErr != nil {
			common.SmartIDELog.Fatal(composeCmdErr)
		}
	}

	//4. 获取启动的容器列表
	dockerComposeContainers := GetLocalContainersWithServices(ctx, cli, yamlFileCongfig.GetServiceNames())
	devContainerName := getDevContainerName(dockerComposeContainers, yamlFileCongfig.Workspace.DevContainer.ServiceName)
	gitconfig := yamlFileCongfig.Workspace.DevContainer.Volumes.GitConfig
	dal.GitConfig(gitconfig, false, devContainerName, cli, compose.Service{}, common.SSHRemote{})
	docker := *common.NewDocker(cli)
	dockerContainerName := strings.ReplaceAll(devContainerName, "/", "")
	out, err := docker.Exec(context.Background(), strings.ReplaceAll(devContainerName, "/", ""), "/usr/bin", []string{"sudo", "chmod", "-R", "700", "/root"}, []string{})
	common.CheckError(err)
	common.SmartIDELog.Debug(out)

	//5. 保存 workspace
	if hasChanged {
		common.SmartIDELog.Info(i18nInstance.Start.Info_workspace_saving)
		//5.1.
		workspace.Name = devContainerName
		workspace.TempDockerCompose = tempDockerCompose
		workspace.TempDockerComposeFilePath = dockeComposeTempYamlFilePath
		//5.2.
		workspaceId, err := dal.InsertOrUpdateWorkspace(workspace)
		common.CheckError(err)
		common.SmartIDELog.InfoF(i18nInstance.Start.Info_workspace_saved, workspaceId)
	}

	//6. 执行函数内容
	yamlExecuteFun(yamlFileCongfig)
	endPostExecuteFun(dockerContainerName, docker)

	//7. 使用浏览器打开web ide
	common.SmartIDELog.Info(i18nInstance.Start.Info_running_openbrower)
	// vscode启动时候默认打开文件夹处理
	var url string
	if yamlFileCongfig.Workspace.DevContainer.IdeType == "vscode" {
		url = fmt.Sprintf("http://localhost:%v/?folder=vscode-remote://localhost:%v/home/project/%v", ideBindingPort, ideBindingPort, workspace.ProjectName)
	} else {
		url = fmt.Sprintf(`http://localhost:%v`, ideBindingPort)
	}
	common.SmartIDELog.Info(i18nInstance.Start.Info_open_in_brower, url)
	isUrlReady := false
	for !isUrlReady {
		resp, err := http.Get(url)
		if (err == nil) && (resp.StatusCode == 200) {
			isUrlReady = true
			common.OpenBrowser(url)
		}
	}

	//99. 结束
	common.SmartIDELog.Info(i18nInstance.Start.Info_end)

	// tunnel， 死循环，不结束
	for {
		tunnel.AutoTunnelMultiple(fmt.Sprintf("localhost:%v", sshBindingPort), "root", "root123", tempDockerCompose.GetLocalBindingPorts(), yamlFileCongfig.GetPortLabelMap())
		time.Sleep(time.Second * 10)
	}
}

// docker-compose 对应的容器是否运行
func isDockerComposeRunning(ctx context.Context, cli *client.Client, serviceNames []string) bool {
	isDockerComposeRunning := false

	dockerComposeContainers := GetLocalContainersWithServices(ctx, cli, serviceNames)
	if len(dockerComposeContainers) > 0 {
		common.SmartIDELog.Info(i18nInstance.Start.Warn_docker_container_started)
		isDockerComposeRunning = true
	}

	return isDockerComposeRunning
}

// 获取容器的名称
func getDevContainerName(dockerComposeContainers []DockerComposeContainer, devServiceName string) string {
	for _, dockerComposeContainer := range dockerComposeContainers {
		if dockerComposeContainer.ServiceName == devServiceName {
			return dockerComposeContainer.ContainerName
		}
	}
	return ""
}
