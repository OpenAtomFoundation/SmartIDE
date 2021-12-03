/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/leansoftX/smartide-cli/cmd/dal"
	"github.com/leansoftX/smartide-cli/cmd/start"
	"github.com/leansoftX/smartide-cli/lib/common"
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

		// 执行
		// 判断是否有工作区数据
		if workspaceInfo.IsNil() {
			common.SmartIDELog.Error(i18nInstance.Main.Err_workspace_none)
		}

		// 执行对应的stop
		if workspaceInfo.Mode == dal.WorkingMode_Local {
			stopLocal(workspaceInfo)

		} else {
			stopRemove(workspaceInfo)

		}

		common.SmartIDELog.Info(i18nInstance.Stop.Info_end)

	},
}

// 停止本地容器
func stopLocal(workspace dal.WorkspaceInfo) {
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
func stopRemove(workspace dal.WorkspaceInfo) {
	// ssh 连接
	common.SmartIDELog.Info(i18nInstance.Stop.Info_sshremote_connection_creating)
	var sshRemote common.SSHRemote
	err := sshRemote.Instance(workspace.Remote.Addr, workspace.Remote.SSHPort, workspace.Remote.UserName, workspace.Remote.Password)
	common.CheckError(err)

	// 检查环境
	err = start.CheckRemoveEnv(sshRemote)
	common.CheckError(err)

	// 停止容器
	common.SmartIDELog.Info(i18nInstance.Stop.Info_docker_stopping)
	command := fmt.Sprintf("docker-compose -f %v --project-directory %v stop",
		common.FilePahtJoin4Linux(workspace.TempDockerComposeFilePath), common.FilePahtJoin4Linux(workspace.WorkingDirectoryPath))
	err = sshRemote.ExecSSHCommandRealTime(command)
	common.CheckError(err)

}

func init() {
	stopCmd.Flags().StringVarP(&configYamlFileRelativePath, "filepath", "f", "", i18nInstance.Stop.Info_help_flag_filepath)

}
