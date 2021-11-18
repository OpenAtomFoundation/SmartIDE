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
	"github.com/leansoftX/smartide-cli/cmd/lib"
	"github.com/leansoftX/smartide-cli/lib/common"
	"github.com/leansoftX/smartide-cli/lib/docker/compose"
	"github.com/leansoftX/smartide-cli/lib/tunnel"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

// 本地执行 start
func ExecuteStartCmd(workspace dal.WorkspaceInfo, endPostExecuteFun func(dockerContainerName string, docker common.Docker)) {

	//0. 检查本地环境
	err := CheckLocalEnv()
	common.CheckError(err)

	//1. 组织 docker compose 的文件内容
	var yamlFileCongfig lib.YamlFileConfig
	yamlFileCongfig.SetWorkspace(workspace.WorkingDirectoryPath, workspace.ConfigFilePath)
	yamlFileCongfig.GetConfig() // 读取配置

	var dockerCompose compose.DockerComposeYml
	var ideBindingPort, sshBindingPort int
	localWorkingDir := yamlFileCongfig.GetLocalWorkingDirectry()
	dockeComposeTempYamlFilePath := yamlFileCongfig.GetTempDockerComposeFilePath(localWorkingDir, workspace.ProjectName) // 获取临时docker-compose文件的路径

	//1.1. 校验 docker compose 文件对应的环境是否已经启动
	isDockerComposeRunning := false
	dockerComposeContainers := GetLocalContainersWithServices(dockeComposeTempYamlFilePath, yamlFileCongfig.GetServiceNames())
	if len(dockerComposeContainers) > 0 {
		common.SmartIDELog.Warning(i18nInstance.Start.Error.Docker_started)
		isDockerComposeRunning = true
	}

	//
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	common.CheckError(err)

	if isDockerComposeRunning { //
		dockerCompose, ideBindingPort, sshBindingPort = yamlFileCongfig.LoadDockerComposeFromTempFile(common.SSHRemote{}, "", workspace.ProjectName)
	} else { //TODO bug, 使用了新端口，但是因为没有执行如下步骤，导致无法使用
		dockerCompose, ideBindingPort, sshBindingPort = yamlFileCongfig.ConvertToDockerCompose(common.SSHRemote{}, "", true) // 转换为docker compose格式

		//1.2. 保存 docker-compose 、config 文件
		err := yamlFileCongfig.SaveTempFiles(dockerCompose, workspace.ProjectName) // 保存docker-compose文件
		common.CheckError(err)

		//1.3. print
		common.SmartIDELog.InfoF(i18nInstance.Start.Info.Info_docker_compose_filepath, dockeComposeTempYamlFilePath)
		common.SmartIDELog.InfoF(i18nInstance.Start.Info.Info_ssh_tunnel, sshBindingPort) // 提示用户ssh端口绑定到了本地的某个端口

		//2.1. 创建网络
		for network := range dockerCompose.Networks {
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
				common.SmartIDELog.InfoF(i18nInstance.Start.Info.Info_create_network, network)
			}
		}

		//2.2. 运行docker-compose命令
		// e.g. docker-compose -f /Users/jasonchen/.ide/docker-compose-product-service-dev.yaml --project-directory /Users/jasonchen/Project/boat-house/boat-house-backend/src/product-service/api up -d
		pwd, _ := os.Getwd()
		composeCmd := exec.Command("docker-compose", "-f", dockeComposeTempYamlFilePath, "--project-directory", pwd, "up", "-d")
		composeCmd.Stdout = os.Stdout
		composeCmd.Stderr = os.Stderr
		if composeCmdErr := composeCmd.Run(); composeCmdErr != nil {
			common.SmartIDELog.Fatal(composeCmdErr)
		}
	}

	// 启动的容器列表
	dockerComposeContainers = GetLocalContainersWithServices(dockeComposeTempYamlFilePath, yamlFileCongfig.GetServiceNames())
	var containerName string
	for _, container := range dockerComposeContainers {
		if container.ServiceName == yamlFileCongfig.Workspace.DevContainer.ServiceName {
			containerName = container.ContainerName
			break
		}
	}
	gitconfig := yamlFileCongfig.Workspace.DevContainer.Volumes.GitConfig
	lib.GitConfig(gitconfig, false, containerName, cli, compose.Service{}, common.SSHRemote{})
	docker := *common.NewDocker(cli)
	dockerContainerName := strings.ReplaceAll(containerName, "/", "")
	out, err := docker.Exec(context.Background(), strings.ReplaceAll(containerName, "/", ""), "/bin", []string{"chmod", "-R", "700", "/root"}, []string{})
	common.CheckError(err)
	common.SmartIDELog.Debug(out)

	//4. 保存到 workspace
	common.SmartIDELog.Info("缓存工作空间...")
	for _, dockerComposeContainer := range dockerComposeContainers {
		if dockerComposeContainer.ServiceName == yamlFileCongfig.Workspace.DevContainer.ServiceName {
			workspace.Name = dockerComposeContainer.ContainerName
			break
		}
	}
	workspaceId, err := dal.InsertOrUpdateWorkspace(workspace)
	common.CheckError(err)
	common.SmartIDELog.InfoF("数据保存成功，工作区ID: %v ", workspaceId)

	//执行函数内容
	endPostExecuteFun(dockerContainerName, docker)

	//3. 使用浏览器打开web ide
	common.SmartIDELog.Info(i18nInstance.Start.Info.Info_running_openbrower)
	// vscode启动时候默认打开文件夹处理
	var url string
	if yamlFileCongfig.Workspace.DevContainer.IdeType == "vscode" {
		url = fmt.Sprintf("http://localhost:%v/?folder=vscode-remote://localhost:%v/home/project", ideBindingPort, ideBindingPort)
	} else {
		url = fmt.Sprintf(`http://localhost:%v`, ideBindingPort)
	}
	common.SmartIDELog.Info(i18nInstance.Start.Info.Info_open_in_brower, url)
	isUrlReady := false
	for !isUrlReady {
		resp, err := http.Get(url)
		if (err == nil) && (resp.StatusCode == 200) {
			isUrlReady = true
			common.OpenBrowser(url)
		}
	}

	//99. 结束
	common.SmartIDELog.Info(i18nInstance.Start.Info.Info_end)

	// tunnel， 死循环，不结束
	for {
		tunnel.AutoTunnelMultiple(fmt.Sprintf("localhost:%v", sshBindingPort), "root", "root123", dockerCompose.GetLocalBindingPorts()) //TODO: 登录的用户名，密码要能够从配置文件中读取出来
		time.Sleep(time.Second * 10)
	}
}
