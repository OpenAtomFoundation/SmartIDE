/*
 * @Date: 2022-06-07 14:02:14
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-06-11 09:28:41
 * @FilePath: /smartide-cli/cmd/remove/vm.go
 */

package remove

import (
	"fmt"

	"github.com/leansoftX/smartide-cli/cmd/start"
	"github.com/leansoftX/smartide-cli/internal/biz/workspace"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/spf13/cobra"
)

// 在远程主机上运行删除命令
func RemoveRemote(workspaceInfo workspace.WorkspaceInfo,
	isRemoveAllComposeImages bool, isRemoveRemoteDirectory bool, isForce bool,
	cmd *cobra.Command) error {
	// ssh 连接
	common.SmartIDELog.Info(i18nInstance.Remove.Info_sshremote_connection_creating)
	sshRemote, err := common.NewSSHRemote(workspaceInfo.Remote.Addr, workspaceInfo.Remote.SSHPort, workspaceInfo.Remote.UserName, workspaceInfo.Remote.Password)
	if err != nil {
		return err
	}

	// 检查环境
	err = sshRemote.CheckRemoveEnv()
	if err != nil {
		return err
	}

	// 项目文件夹是否存在
	if workspaceInfo.GitCloneRepoUrl != "" && !sshRemote.IsCloned(workspaceInfo.WorkingDirectoryPath) {
		sshRemote.GitClone(workspaceInfo.GitCloneRepoUrl, workspaceInfo.WorkingDirectoryPath, common.SmartIDELog.Ws_id, cmd)
		isRemoveRemoteDirectory = true // 创建后就删掉

	}

	// 检查临时文件夹是否存在
	configYamlFilePath := workspaceInfo.ConfigYaml.GetConfigFileAbsolutePath()
	if !sshRemote.IsFileExist(workspaceInfo.TempYamlFileAbsolutePath) || !sshRemote.IsFileExist(configYamlFilePath) {
		workspaceInfo.SaveTempFilesForRemote(sshRemote)
		//isRemoveRemoteDirectory = true // 创建后就删掉

	}

	// 容器列表
	containers, err := start.GetRemoteContainersWithServices(sshRemote, workspaceInfo.ConfigYaml.GetServiceNames()) // 只能获取到运行中的容器
	if err != nil {
		return err
	}
	if len(containers) <= 0 {
		common.SmartIDELog.Importance(i18nInstance.Start.Warn_docker_container_getnone)
	}

	// 远程主机上执行 docker-compose 删除容器
	//	if len(containers) > 0 {
	common.SmartIDELog.Info(i18nInstance.Remove.Info_docker_removing)
	command := fmt.Sprintf(`docker-compose -f %v --project-directory %v down -v`,
		common.FilePahtJoin4Linux(workspaceInfo.TempYamlFileAbsolutePath), common.FilePahtJoin4Linux(workspaceInfo.WorkingDirectoryPath))
	err = sshRemote.ExecSSHCommandRealTime(command)
	if err != nil {
		return err
	}

	// 删除对应的镜像
	if isRemoveAllComposeImages {
		common.SmartIDELog.Info(i18nInstance.Remove.Info_docker_rmi_removing)

		force := ""
		if isForce {
			force = "-f"
		}

		for _, service := range workspaceInfo.TempDockerCompose.Services {
			if service.Image != "" { // 镜像信息不为空
				//imageNameAndTag := fmt.Sprintf("%v:%v", service.Image.Name, service.Image.Tag)
				command := fmt.Sprintf("docker rmi %v %v", force, service.Image)
				_, err = sshRemote.ExeSSHCommand(command)
				if err != nil {
					common.SmartIDELog.ImportanceWithError(err)
				} else {
					common.SmartIDELog.InfoF(i18nInstance.Remove.Info_docker_rmi_image_removed, service.Image)
				}
			}
		}
	}

	// 删除文件夹
	if isRemoveRemoteDirectory { // 在仅删除workspace的模式下，不删除容器
		common.SmartIDELog.Info(i18nInstance.Remove.Info_project_dir_removing)
		workingDirectoryPath := common.FilePahtJoin4Linux(workspaceInfo.WorkingDirectoryPath)
		command := fmt.Sprintf("sudo rm -rf %v", workingDirectoryPath)
		err = sshRemote.ExecSSHCommandRealTime(command)
		if err != nil {
			return err
		}

		// 成功后的提示
		common.SmartIDELog.InfoF(i18nInstance.Remove.Info_project_dir_removed, workingDirectoryPath)
	}

	//remove ssh config node from .ssh/config
	workspaceInfo.RemoveSSHConfig()
	return nil
}
