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

package cmd

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	cmdCommon "github.com/leansoftX/smartide-cli/cmd/common"
	"github.com/leansoftX/smartide-cli/internal/apk/appinsight"
	"github.com/leansoftX/smartide-cli/internal/biz/workspace"
	"github.com/leansoftX/smartide-cli/internal/dal"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/leansoftX/smartide-cli/pkg/k8s"
	"github.com/spf13/cobra"

	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/leansoftX/smartide-cli/cmd/remove"
	"github.com/leansoftX/smartide-cli/cmd/server"
)

var removeCmdFlag struct {
	// 是否仅删除本地的工作区
	IsOnlyRemoveWorkspaceDataRecord bool

	// 是否仅删除远程的容器
	IsOnlyRemoveContainer bool

	// 是否确定删除
	IsContinue bool

	// 是否禁止端口转发
	IsUnforward bool

	// 是否删除远程主机上的文件夹
	IsRemoveRemoteDirectory bool

	// 强制删除
	IsForce bool

	// 删除compose对应的所有镜像
	IsRemoveAllComposeImages bool
}

// 删除的模式
type RemoveMode string

const (
	RemoteMode_None                          RemoveMode = "none"
	RemoteMode_OnlyRemoveContainer           RemoveMode = "only_container"
	RemoteMode_OnlyRemoveWorkspaceDataRecord RemoveMode = "only_data_record"
)

// removeCmd represents the remove command
var removeCmd = &cobra.Command{
	Use:     "remove",
	Short:   i18nInstance.Remove.Info_help_short,
	Long:    i18nInstance.Remove.Info_help_long,
	Aliases: []string{"rm"},
	Example: `
	 smartide remove [--workspaceid] {workspaceid} [-y] [-w] [-i] [-f] 
	 smartide remove [--workspaceid] {workspaceid} [-y] [-s] [-c] [-i] [-f]`,
	Run: func(cmd *cobra.Command, args []string) {

		mode, _ := cmd.Flags().GetString("mode")
		workspaceIdStr := cmdCommon.GetWorkspaceIdFromFlagsOrArgs(cmd, args)
		if strings.ToLower(mode) == "server" || strings.Contains(workspaceIdStr, "SWS") {
			serverModeInfo, _ := server.GetServerModeInfo(cmd)
			if serverModeInfo.ServerHost != "" {
				wsURL := fmt.Sprint(strings.ReplaceAll(strings.ReplaceAll(serverModeInfo.ServerHost, "https", "ws"), "http", "ws"), "/ws/smartide/ws")
				action := 0
				if removeCmdFlag.IsOnlyRemoveContainer {
					action = 3
				}
				if removeCmdFlag.IsRemoveAllComposeImages && removeCmdFlag.IsRemoveRemoteDirectory {
					action = 4
				}
				common.WebsocketStart(wsURL)
				if action != 0 {
					if pid, err := workspace.GetParentId(workspaceIdStr, workspace.ActionEnum_Workspace_Remove, serverModeInfo.ServerToken, serverModeInfo.ServerHost); err == nil && pid > 0 {
						common.SmartIDELog.Ws_id = workspaceIdStr
						common.SmartIDELog.ParentId = pid
					}
				}

			}
		}

		common.SmartIDELog.Info(i18nInstance.Remove.Info_start)

		//1. 获取 workspace 信息
		common.SmartIDELog.Info(i18nInstance.Main.Info_workspace_loading) // log
		workspaceInfo, err := cmdCommon.GetWorkspaceFromCmd(cmd, args)
		entryptionKey4Workspace(workspaceInfo) // 申明需要加密的文本
		common.CheckError(err)
		if workspaceInfo.IsNil() {
			common.SmartIDELog.Error(i18nInstance.Main.Err_workspace_none)
		}

		// 检查错误并feedback
		var checkErrorFeedback = func(err error) {
			if workspaceInfo.CliRunningEnv == workspace.CliRunningEvnEnum_Server && err != nil {
				server.Feedback_Finish(server.FeedbackCommandEnum_Remove, cmd, false, nil, workspaceInfo, err.Error(), "")
				common.CheckError(err)
			}

		}

		//2. 操作类型
		//2.1. 验证互斥的操作
		if removeCmdFlag.IsOnlyRemoveContainer && removeCmdFlag.IsOnlyRemoveWorkspaceDataRecord { // 仅删除容器 和 仅删除工作区，不能同时存在
			checkErrorFeedback(errors.New(i18nInstance.Remove.Err_flag_workspace_container))
		}
		if workspaceInfo.Mode == workspace.WorkingMode_Local && removeCmdFlag.IsOnlyRemoveContainer { // 本地模式下，
			checkErrorFeedback(errors.New(i18nInstance.Remove.Err_flag_container_valid))
		}

		//2.2. 操作类型
		var removeMode RemoveMode = RemoteMode_None
		if removeCmdFlag.IsOnlyRemoveContainer {
			removeMode = RemoteMode_OnlyRemoveContainer
		} else if removeCmdFlag.IsOnlyRemoveWorkspaceDataRecord {
			removeMode = RemoteMode_OnlyRemoveWorkspaceDataRecord
		}

		//3. 提示 是否确认删除
		if !removeCmdFlag.IsContinue { // 如果设置了参数yes，那么默认就是确认删除
			isEnableRemove := ""
			common.SmartIDELog.Console(i18nInstance.Remove.Info_is_confirm_remove)
			fmt.Scanln(&isEnableRemove)
			if strings.ToLower(isEnableRemove) != "y" {
				return
			}
		}

		//4. 执行删除动作
		if workspaceInfo.CliRunningEnv == workspace.CliRunningEnvEnum_Client { //4.1. 本地执行删除

			if workspaceInfo.CacheEnv == workspace.CacheEnvEnum_Server { //4.1.1. 在本地 删除服务器中的工作区
				remove.RemoveServerWorkSpaceInClient(workspaceIdStr, workspaceInfo, removeCmdFlag.IsRemoveRemoteDirectory)

			} else { //4.1.2. 删除本地的工作区
				//

				if removeMode == RemoteMode_None || removeMode == RemoteMode_OnlyRemoveContainer {
					if workspaceInfo.Mode == workspace.WorkingMode_Local {
						appinsight.SetCliLocalTrack(appinsight.Cli_Local_Remove, args, workspaceInfo.ID, "")
						err := remove.RemoveLocal(workspaceInfo, removeCmdFlag.IsRemoveAllComposeImages, removeCmdFlag.IsForce)
						common.CheckError(err)

					} else if workspaceInfo.Mode == workspace.WorkingMode_Remote {
						appinsight.SetCliLocalTrack(appinsight.Cli_Host_Remove, args, workspaceInfo.ID, "")
						err := remove.RemoveRemote(workspaceInfo, removeCmdFlag.IsRemoveAllComposeImages, removeCmdFlag.IsRemoveRemoteDirectory, removeCmdFlag.IsForce, cmd)
						common.CheckError(err)

					} else if workspaceInfo.Mode == workspace.WorkingMode_K8s {
						appinsight.SetCliLocalTrack(appinsight.Cli_K8s_Remove, args, workspaceInfo.ID, "")
						k8sUtil, err := k8s.NewK8sUtil(workspaceInfo.K8sInfo.KubeConfigFilePath,
							workspaceInfo.K8sInfo.Context,
							workspaceInfo.K8sInfo.Namespace)
						common.CheckError(err)

						err = remove.RemoveK8s(*k8sUtil, workspaceInfo)
						common.CheckError(err)

					}
				}

				// remote workspace in db
				if removeMode == RemoteMode_None || removeMode == RemoteMode_OnlyRemoveWorkspaceDataRecord { // 在仅删除容器的模式下，不删除工作区
					common.SmartIDELog.Info(i18nInstance.Remove.Info_workspace_removing)
					id, err := strconv.Atoi(workspaceInfo.ID)
					common.CheckError(err)
					err = dal.RemoveWorkspace(id)
					common.CheckError(err)
				}
			}

		} else { //4.2. 在远程主机（tekton）上执行删除
			msg := ""
			if workspaceInfo.Mode == workspace.WorkingMode_Remote {
				err := remove.RemoveRemote(workspaceInfo, removeCmdFlag.IsRemoveAllComposeImages, removeCmdFlag.IsRemoveRemoteDirectory, removeCmdFlag.IsForce, cmd)
				checkErrorFeedback(err)
			} else if workspaceInfo.Mode == workspace.WorkingMode_K8s {
				k8sUtil, err := k8s.NewK8sUtilWithContent(workspaceInfo.K8sInfo.KubeConfigContent,
					workspaceInfo.K8sInfo.Context,
					workspaceInfo.K8sInfo.Namespace)
				checkErrorFeedback(err)

				err = remove.RemoveK8s(*k8sUtil, workspaceInfo)
				checkErrorFeedback(err)
			} else {
				common.SmartIDELog.Error(fmt.Errorf("当前 %v 模式不支持在server上运行", workspaceInfo.Mode))
			}

			// feeadback
			common.SmartIDELog.Info("反馈运行结果...")
			command := server.FeedbackCommandEnum_Remove
			if removeCmdFlag.IsOnlyRemoveContainer {
				command = server.FeedbackCommandEnum_RemoveContainer
			}
			err = server.Feedback_Finish(command, cmd, err == nil, nil, workspaceInfo, msg, "")
			common.CheckError(err)

		}

		// log
		common.SmartIDELog.Info(i18nInstance.Remove.Info_end)
		common.WG.Wait()
	},
}

// 初始化
func init() {
	removeCmd.Flags().BoolVarP(&removeCmdFlag.IsContinue, "yes", "y", false, i18nInstance.Remove.Info_flag_yes)

	removeCmd.Flags().BoolVarP(&removeCmdFlag.IsOnlyRemoveWorkspaceDataRecord, "workspace", "w", false, i18nInstance.Remove.Info_flag_workspace)
	removeCmd.Flags().BoolVarP(&removeCmdFlag.IsOnlyRemoveContainer, "container", "c", false, i18nInstance.Remove.Info_flag_container)

	removeCmd.Flags().BoolVarP(&removeCmdFlag.IsRemoveRemoteDirectory, "project", "p", false, i18nInstance.Remove.Info_flag_project)
	removeCmd.Flags().BoolVarP(&removeCmdFlag.IsRemoveAllComposeImages, "image", "i", false, i18nInstance.Remove.Info_flag_image)

	removeCmd.Flags().BoolVarP(&removeCmdFlag.IsForce, "force", "f", false, i18nInstance.Remove.Info_flag_force)
}
