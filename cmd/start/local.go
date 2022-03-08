/*
 * @Author: jason chen (jasonchen@leansoftx.com, http://smallidea.cnblogs.com)
 * @Description:
 * @Date: 2021-11
 * @LastEditors: kenan
 * @LastEditTime: 2022-02-21 15:48:43
 */
package start

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/leansoftX/smartide-cli/internal/biz/config"
	"github.com/leansoftX/smartide-cli/internal/biz/workspace"
	"github.com/leansoftX/smartide-cli/internal/dal"
	"github.com/leansoftX/smartide-cli/internal/model"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/leansoftX/smartide-cli/pkg/docker/compose"
	"github.com/leansoftX/smartide-cli/pkg/kubectl"
	"github.com/leansoftX/smartide-cli/pkg/tunnel"
	"gopkg.in/yaml.v2"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

// 本地执行 start
func ExecuteStartCmd(workspaceInfo workspace.WorkspaceInfo,
	endPostExecuteFun func(dockerContainerName string, docker common.Docker),
	yamlExecuteFun func(yamlConfig config.SmartIdeConfig)) {

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
	currentConfig := config.NewConfig(workspaceInfo.WorkingDirectoryPath, workspaceInfo.ConfigFilePath, "")

	//2. docker-compose
	//2.1. 获取compose数据
	_, linkComposeFileContent := currentConfig.GetLocalLinkDockerComposeFile()
	configYamlStr, err := currentConfig.ToYaml()
	common.CheckError(err)
	hasChanged := workspaceInfo.ChangeConfig(configYamlStr, linkComposeFileContent) // 是否改变
	if hasChanged {                                                                 // 改变包括了初始化
		// log
		if workspaceInfo.ID != "" {
			common.SmartIDELog.Info(i18nInstance.Start.Info_workspace_changed)

		} else {
			common.SmartIDELog.Info(i18nInstance.Start.Info_workspace_create)
		}

		// 获取compose配置
		tempDockerCompose, ideBindingPort, sshBindingPort =
			currentConfig.ConvertToDockerCompose(common.SSHRemote{}, workspaceInfo.GetProjectDirctoryName(), "", true) // 转换为docker compose格式

		// 更新端口绑定列表，只在改变的时候才需要赋值
		workspaceInfo.Extend = workspace.WorkspaceExtend{Ports: currentConfig.GetPortMappings()}

		// 链接的docker-compose文件
		if workspaceInfo.ConfigYaml.IsLinkDockerComposeFile() {
			yaml.Unmarshal([]byte(linkComposeFileContent), workspaceInfo.LinkDockerCompose)
		}

		// 保存 docker-compose 、config 文件到临时文件夹
		workspaceInfo.ConfigYaml = *currentConfig
		workspaceInfo.TempDockerCompose = tempDockerCompose
		err = workspaceInfo.SaveTempFiles() // 保存docker-compose文件
		common.CheckError(err)
	} else {
		// 先保存，确保临时文件存在 以及是最新的
		err = workspaceInfo.SaveTempFiles()
		common.CheckError(err)

		tempDockerCompose, ideBindingPort, sshBindingPort = currentConfig.LoadDockerComposeFromTempFile(common.SSHRemote{}, workspaceInfo.TempDockerComposeFilePath)
	}
	//2.2. 扩展信息
	workspaceInfo.Extend = workspaceInfo.GetWorkspaceExtend()

	//3. 容器
	//3.1. 校验 docker-compose 文件对应的环境是否已经启动
	isDockerComposeRunning := isDockerComposeRunning(ctx, cli, currentConfig.GetServiceNames())

	//3.2. 运行容器
	if !isDockerComposeRunning || hasChanged { // 容器没有运行 或者 有改变，重新创建容器
		// print
		common.SmartIDELog.InfoF(i18nInstance.Start.Info_ssh_tunnel, sshBindingPort) // 提示用户ssh端口绑定到了本地的某个端口

		// 创建网络（docker-compose 创建的网络会增加文件夹名，导致无法匹配）
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
		// e.g. docker-compose -f {docker-compose文件路径} --project-directory {工作目录} up -d
		pwd, _ := os.Getwd()
		composeCmd := exec.Command("docker-compose", "-f", workspaceInfo.TempDockerComposeFilePath, "--project-directory", pwd, "up", "-d")
		composeCmd.Stdout = os.Stdout
		composeCmd.Stderr = os.Stderr
		if composeCmdErr := composeCmd.Run(); composeCmdErr != nil {
			common.CheckError(composeCmdErr)
		}
	}

	//4. 获取启动的容器列表
	dockerComposeContainers := GetLocalContainersWithServices(ctx, cli, currentConfig.GetServiceNames())
	devContainerName := getDevContainerName(dockerComposeContainers, currentConfig.Workspace.DevContainer.ServiceName)
	gitconfig := currentConfig.Workspace.DevContainer.Volumes.GitConfig
	config.GitConfig(gitconfig, false, devContainerName, cli, &compose.Service{}, common.SSHRemote{}, kubectl.ExecInPodRequest{})
	docker := *common.NewDocker(cli)
	dockerContainerName := strings.ReplaceAll(devContainerName, "/", "")
	config.LocalContainerGitSet(docker, dockerContainerName)

	//5. 保存 workspace
	if hasChanged {
		common.SmartIDELog.Info(i18nInstance.Start.Info_workspace_saving)
		//5.1.
		workspaceInfo.Name = devContainerName
		workspaceInfo.TempDockerCompose = tempDockerCompose
		//5.2.
		workspaceId, err := dal.InsertOrUpdateWorkspace(workspaceInfo)
		common.CheckError(err)
		common.SmartIDELog.InfoF(i18nInstance.Start.Info_workspace_saved, workspaceId)
	}

	//6. 执行函数内容
	yamlExecuteFun(*currentConfig)
	endPostExecuteFun(dockerContainerName, docker)

	//7. 使用浏览器打开web ide
	common.SmartIDELog.Info(i18nInstance.Start.Info_running_openbrower)
	// vscode启动时候默认打开文件夹处理
	var url string
	switch strings.ToLower(currentConfig.Workspace.DevContainer.IdeType) {
	case "vscode":
		url = fmt.Sprintf("http://localhost:%v/?folder=vscode-remote://localhost:%v%v",
			ideBindingPort, ideBindingPort, workspaceInfo.GetContainerWorkingPathWithVolumes())
	case "jb-projector":
		url = fmt.Sprintf(`http://localhost:%v`, ideBindingPort)
	default:
		url = fmt.Sprintf(`http://localhost:%v`, ideBindingPort)
	}

	common.SmartIDELog.Info(i18nInstance.VmStart.Info_warting_for_webide)
	isUrlReady := false
	for !isUrlReady {
		resp, err := http.Get(url)
		if (err == nil) && (resp.StatusCode == 200) {
			isUrlReady = true
			common.OpenBrowser(url)
			common.SmartIDELog.InfoF(i18nInstance.VmStart.Info_open_brower, url)
		} else {
			msg := fmt.Sprintf("%v 检测失败", url)
			common.SmartIDELog.Debug(msg)
		}
	}

	//99. 结束
	common.SmartIDELog.Info(i18nInstance.Start.Info_end)

	// tunnel
	sshPassword := workspaceInfo.TempDockerCompose.GetSSHPassword(currentConfig.Workspace.DevContainer.ServiceName)
	sshRemote, err := common.NewSSHRemote("localhost", sshBindingPort, model.CONST_DEV_CONTAINER_CUSTOM_USER, sshPassword)
	common.CheckError(err)
	options := tunnel.AutoTunnelMultipleOptions{}
	for _, portMap := range workspaceInfo.Extend.Ports {
		options.AppendPortMapping(tunnel.PortMapTypeEnum(portMap.PortMapType), portMap.OriginHostPort, portMap.CurrentHostPort,
			portMap.HostPortDesc, portMap.ContainerPort)
	}
	tunnel.AutoTunnel(sshRemote.Connection, options)

	// 死循环，不结束
	for {
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
