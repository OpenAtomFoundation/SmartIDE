/*
SmartIDE - CLI
Copyright (C) 2023 leansoftX.com

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package start

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/leansoftX/smartide-cli/internal/dal"
	"github.com/leansoftX/smartide-cli/internal/model"
	"github.com/spf13/cobra"

	initExtended "github.com/leansoftX/smartide-cli/cmd/init"
	"github.com/leansoftX/smartide-cli/internal/apk/appinsight"
	"github.com/leansoftX/smartide-cli/internal/biz/config"
	"github.com/leansoftX/smartide-cli/internal/biz/workspace"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/leansoftX/smartide-cli/pkg/docker/compose"
	"github.com/leansoftX/smartide-cli/pkg/k8s"
	"github.com/leansoftX/smartide-cli/pkg/tunnel"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

func saveDataAndReloadWorkSpaceId(workspace *workspace.WorkspaceInfo) {
	workspaceId, err := dal.InsertOrUpdateWorkspace(*workspace)
	workspace.ID = strconv.Itoa(int(workspaceId))
	common.CheckError(err)
}

// 本地执行 start
func ExecuteStartCmd(workspaceInfo workspace.WorkspaceInfo, isUnforward bool,
	endPostExecuteFun func(dockerContainerName string, docker common.Docker),
	yamlExecuteFun func(yamlConfig config.SmartIdeConfig, workspaceInfo workspace.WorkspaceInfo, cmdtype, userguid, workspaceid string), args []string, cmd *cobra.Command) (
	workspace.WorkspaceInfo, error) {

	//0. 检查本地环境
	err := common.CheckLocalEnv()
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
	if workspaceInfo.IsNil() {
		initExtended.InitLocalConfig(cmd, args)
	}
	currentConfig, err := config.NewLocalConfig(workspaceInfo.WorkingDirectoryPath, workspaceInfo.ConfigFileRelativePath)
	common.CheckError(err)
	//TODO: git pull
	if workspaceInfo.Addon.IsEnable {
		workspaceInfo = AddonEnable(workspaceInfo)
		currentConfig.AddonWebTerminal(workspaceInfo.Name, workspaceInfo.WorkingDirectoryPath)
	}

	//2. docker-compose
	//2.1. 获取compose数据
	configYamlStr, err := currentConfig.ToYaml()
	common.CheckError(err)
	linkComposeFileContent, err := currentConfig.Workspace.LinkCompose.ToYaml()
	common.CheckError(err)
	hasChanged := workspaceInfo.IsChangeConfig(configYamlStr, linkComposeFileContent) // 是否改变
	if hasChanged {                                                                   // 改变包括了初始化
		// log
		if workspaceInfo.ID != "" {
			common.SmartIDELog.Info(i18nInstance.Start.Info_workspace_changed)
		} else {
			common.SmartIDELog.Info(i18nInstance.Start.Info_workspace_create)
		}

		// 获取compose配置
		tempDockerCompose, ideBindingPort, sshBindingPort =
			currentConfig.ConvertToDockerCompose(common.SSHRemote{}, workspaceInfo.GetProjectDirctoryName(), "", true, "", nil) // 转换为docker compose格式

		// 更新端口绑定列表，只在改变的时候才需要赋值
		workspaceInfo.Extend = workspace.WorkspaceExtend{Ports: currentConfig.GetPortMappings()}

		// 保存 docker-compose 、config 文件到临时文件夹
		workspaceInfo.ConfigYaml = *currentConfig
		workspaceInfo.TempDockerCompose = tempDockerCompose
		err = workspaceInfo.SaveTempFiles() // 保存docker-compose文件
		common.CheckError(err)
	} else {
		// 先保存，确保临时文件存在 以及是最新的
		err = workspaceInfo.SaveTempFiles()
		common.CheckError(err)

		tempDockerCompose, ideBindingPort, sshBindingPort =
			currentConfig.LoadDockerComposeFromTempFile(common.SSHRemote{}, workspaceInfo.TempYamlFileAbsolutePath)
	}
	//2.2. 扩展信息
	workspaceInfo.Extend = workspaceInfo.GetWorkspaceExtend()

	//3. 容器
	//3.1. 校验 docker-compose 文件对应的环境是否已经启动
	isDockerComposeRunning := isDockerComposeRunning(ctx, cli, workspaceInfo.WorkingDirectoryPath, currentConfig.GetServiceNames())

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
				_, err = cli.NetworkCreate(ctx, network, types.NetworkCreate{})
				common.CheckError(err)
				common.SmartIDELog.InfoF(i18nInstance.Start.Info_create_network, network)
			}
		}

		// 运行docker-compose命令
		// e.g. docker-compose -f {docker-compose文件路径} --project-directory {工作目录} up -d
		pwd, _ := os.Getwd()
		commands := []string{"docker-compose", "-f", workspaceInfo.TempYamlFileAbsolutePath, "--project-directory", pwd, "up", "-d"}
		common.SmartIDELog.Debug(strings.Join(commands, " "))
		err := common.EXEC.Realtime(strings.Join(commands, " "), "")
		common.CheckError(err)
	}

	serviceNames := currentConfig.GetServiceNames()
	//4. 获取启动的容器列表
	dockerComposeContainers := GetLocalContainersWithServices(ctx, cli, workspaceInfo.WorkingDirectoryPath, serviceNames)
	devContainerName := getDevContainerName(dockerComposeContainers, currentConfig.Workspace.DevContainer.ServiceName)
	if devContainerName == "" {
		common.SmartIDELog.Error(fmt.Errorf("从services %v 中获取 开发容器名称失败", serviceNames))
	}
	if currentConfig.Workspace.DevContainer.Volumes.HasGitConfig.Value() {
		config.GitConfig(false, devContainerName, cli, &compose.Service{}, common.SSHRemote{}, k8s.ExecInPodRequest{})
	}
	docker := *common.NewDocker(cli)
	dockerContainerName := strings.ReplaceAll(devContainerName, "/", "")
	config.LocalContainerGitSet(docker, dockerContainerName)                  //git 设置
	localContainerCredentialCache(docker, dockerContainerName, workspaceInfo) // 缓存git 用户名、密码

	//5. 保存 workspace
	//5.1.
	if hasChanged {
		common.SmartIDELog.Info(i18nInstance.Start.Info_workspace_saving)

		if workspaceInfo.Name == "" {
			workspaceInfo.Name = devContainerName
		}
		workspaceInfo.TempDockerCompose = tempDockerCompose
	}
	//5.2.
	saveDataAndReloadWorkSpaceId(&workspaceInfo)

	if appinsight.Global.CmdType == "new" {
		yamlExecuteFun(*currentConfig, workspaceInfo, appinsight.Cli_Local_New, "", workspaceInfo.ID)
	} else {
		yamlExecuteFun(*currentConfig, workspaceInfo, appinsight.Cli_Local_Start, "", workspaceInfo.ID)
	}

	//5.3.
	if hasChanged {
		common.SmartIDELog.InfoF(i18nInstance.Start.Info_workspace_saved, workspaceInfo.ID)
	}

	// update .ssh/config file for vscode remote
	workspaceInfo.UpdateSSHConfig()

	// add publickey content into .ssh/authorized_keys
	config.AddPublicKeyIntoAuthorizedkeys(docker, dockerContainerName)

	// 如果是不进行端口转发，后续就不需要运行
	if isUnforward {
		return workspaceInfo, nil
	}

	//6. 执行函数内容
	endPostExecuteFun(dockerContainerName, docker)

	//7. 使用浏览器打开web ide
	if currentConfig.Workspace.DevContainer.IdeType != config.IdeTypeEnum_SDKOnly {
		common.SmartIDELog.Info(i18nInstance.Start.Info_running_openbrower)
		// vscode启动时候默认打开文件夹处理
		var url string
		switch currentConfig.Workspace.DevContainer.IdeType {
		case config.IdeTypeEnum_VsCode:
			url = fmt.Sprintf("http://localhost:%v/?folder=vscode-remote://localhost:%v%v",
				ideBindingPort, ideBindingPort, workspaceInfo.GetContainerWorkingPathWithVolumes())
		case config.IdeTypeEnum_JbProjector:
			url = fmt.Sprintf(`http://localhost:%v`, ideBindingPort)
		case config.IdeTypeEnum_Opensumi:
			url = fmt.Sprintf(`http://localhost:%v/?workspaceDir=/home/project`, ideBindingPort)
		default:
			url = fmt.Sprintf(`http://localhost:%v`, ideBindingPort)
		}

		common.SmartIDELog.Info(i18nInstance.VmStart.Info_warting_for_webide)
		isUrlReady := false
		for !isUrlReady {
			resp, err := http.Get(url)
			if (err == nil) && (resp.StatusCode == 200) {
				isUrlReady = true
				err := common.OpenBrowser(url)
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

	//9. tunnel
	sshPassword := workspaceInfo.TempDockerCompose.GetSSHPassword(currentConfig.Workspace.DevContainer.ServiceName)
	sshRemote, err := common.NewSSHRemote("localhost", sshBindingPort, model.CONST_DEV_CONTAINER_CUSTOM_USER, sshPassword, "")
	common.CheckError(err)
	options := tunnel.AutoTunnelMultipleOptions{}
	for _, portMap := range workspaceInfo.Extend.Ports {
		options.AppendPortMapping(tunnel.PortMapTypeEnum(portMap.PortMapType), portMap.OriginHostPort, portMap.CurrentHostPort,
			portMap.HostPortDesc, portMap.ContainerPort)
	}
	tunnel.AutoTunnel(sshRemote.Connection, options)

	return workspaceInfo, nil
}

func localContainerCredentialCache(docker common.Docker, dockerContainerName string, workspaceInfo workspace.WorkspaceInfo) {

	if workspaceInfo.GitRepoAuthType != workspace.GitRepoAuthType_Basic {
		return
	}

	common.SmartIDELog.Info("容器缓存git 用户名、密码")
	command := fmt.Sprintf(`git config --global user.name "%v" && git config --global user.password "%v" && git config --global credential.helper store`,
		workspaceInfo.GitUserName, workspaceInfo.GitPassword)

	out, err := docker.Exec(context.Background(), dockerContainerName, "/usr/bin -c", strings.Split(command, " "), []string{})
	common.CheckError(err)
	common.SmartIDELog.Debug(out)

	common.SmartIDELog.Debug(out)

}

// docker-compose 对应的容器是否运行
func isDockerComposeRunning(ctx context.Context, cli *client.Client, workingDir string, serviceNames []string) bool {
	isDockerComposeRunning := false

	dockerComposeContainers := GetLocalContainersWithServices(ctx, cli, workingDir, serviceNames)
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
