/*
 * @Author: jason chen (jasonchen@leansoftx.com, http://smallidea.cnblogs.com)
 * @Description:
 * @Date: 2021-11
 * @LastEditors:
 * @LastEditTime:
 */
package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/leansoftX/smartide-cli/cmd/start"
	"github.com/leansoftX/smartide-cli/internal/biz/workspace"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/spf13/cobra"
) // stopCmd represents the stop command

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: i18nInstance.Stop.Info_help_short,
	Long:  i18nInstance.Stop.Info_help_long,
	Run: func(cmd *cobra.Command, args []string) {

		common.SmartIDELog.Info(i18nInstance.Stop.Info_start)

		// 获取 workspace 信息
		common.SmartIDELog.Info(i18nInstance.Main.Info_workspace_loading)
		workspaceInfo := loadWorkspaceWithDb(cmd, args)

		// 判断是否有工作区数据
		if workspaceInfo.IsNil() {
			common.SmartIDELog.Error(i18nInstance.Main.Err_workspace_none)
		}

		// 执行对应的stop
		if workspaceInfo.Mode == workspace.WorkingMode_Local {
			stopLocal(workspaceInfo)

		} else {
			stopRemote(workspaceInfo)

		}

		common.SmartIDELog.Info(i18nInstance.Stop.Info_end)

	},
}

// 关闭服务器上的远程工作区，使用server mode的参数
func stopServerRemoteByParams(serverWorkSpaceId string) {
	// 直接运行stop

	// 反馈到server
}

// 关闭服务器上的远程工作区，向server传递ID，由服务端处理stop
func stopServerRemoteById(serverWorkSpaceId string) {
	// 1. 请求服务端，触发stop

	// 2. 请求服务端，获取工作区的状态

}

// 停止本地容器
func stopLocal(workspace workspace.WorkspaceInfo) {
	// 校验是否能正常执行docker
	err := start.CheckLocalEnv()
	common.CheckError(err)

	// 本地执行docker-compose
	composeCmd := exec.Command("docker-compose", "-f", workspace.TempDockerComposeFilePath,
		"--project-directory", workspace.WorkingDirectoryPath, "stop")
	composeCmd.Stdout = os.Stdout
	composeCmd.Stderr = os.Stderr
	if composeCmdErr := composeCmd.Run(); composeCmdErr != nil {
		common.SmartIDELog.Fatal(composeCmdErr)
	}
}

// 停止远程容器
func stopRemote(workspaceInfo workspace.WorkspaceInfo) {
	// ssh 连接
	common.SmartIDELog.Info(i18nInstance.Stop.Info_sshremote_connection_creating)
	sshRemote, err := common.NewSSHRemote(workspaceInfo.Remote.Addr, workspaceInfo.Remote.SSHPort, workspaceInfo.Remote.UserName, workspaceInfo.Remote.Password)
	common.CheckError(err)

	// 项目文件夹是否存在
	if !sshRemote.IsCloned(workspaceInfo.WorkingDirectoryPath) {
		msg := fmt.Sprintf(i18nInstance.Stop.Err_env_project_dir_remove, workspaceInfo.ID)
		common.SmartIDELog.Error(msg)
	}

	// 检查临时文件夹是否存在
	if !sshRemote.IsExit(workspaceInfo.TempDockerComposeFilePath) || !sshRemote.IsExit(workspaceInfo.ConfigYaml.GetConfigYamlFilePath()) {
		workspaceInfo.SaveTempFilesForRemote(sshRemote)
	}

	// 检查环境
	err = start.CheckRemoveEnv(sshRemote)
	common.CheckError(err)

	// 停止容器
	common.SmartIDELog.Info(i18nInstance.Stop.Info_docker_stopping)
	command := fmt.Sprintf("docker-compose -f %v --project-directory %v stop",
		common.FilePahtJoin4Linux(workspaceInfo.TempDockerComposeFilePath), common.FilePahtJoin4Linux(workspaceInfo.WorkingDirectoryPath))
	err = sshRemote.ExecSSHCommandRealTime(command)
	common.CheckError(err)

}

func init() {
	stopCmd.Flags().StringVarP(&configYamlFileRelativePath, "filepath", "f", "", i18nInstance.Stop.Info_help_flag_filepath)

}
