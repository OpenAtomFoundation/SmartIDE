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
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	cmdCommon "github.com/leansoftX/smartide-cli/cmd/common"
	newExtend "github.com/leansoftX/smartide-cli/cmd/new"
	"github.com/leansoftX/smartide-cli/internal/apk/appinsight"
	"github.com/leansoftX/smartide-cli/internal/biz/config"
	"github.com/leansoftX/smartide-cli/internal/biz/workspace"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/leansoftX/smartide-cli/pkg/k8s"
	"github.com/spf13/cobra"
	coreV1 "k8s.io/api/core/v1"
)

var newCmd = &cobra.Command{
	Use:   "new",
	Short: i18nInstance.New.Info_help_short,
	Long:  i18nInstance.New.Info_help_long,
	Example: `smartide new <template_type> -T {type_name}
smartide new <template_type> -T {type_name} --host {vm_host_address|vm_host_id} --username {vm_username} --password {vm_password}
smartide new <template_type> -T {type_name} --k8s {kubernetes_context} --kubeconfig {kube_config_file_path}`,
	Run: func(cmd *cobra.Command, args []string) {

		appinsight.Global.CmdType = "new"
		workspaceInfo, err := cmdCommon.GetWorkspaceFromCmd(cmd, args) // 加载工作区信息
		entryptionKey4Workspace(workspaceInfo)                         // 申明需要加密的文本
		common.CheckError(err)
		executeStartCmdFunc := func(yamlConfig config.SmartIdeConfig, workspaceInfo workspace.WorkspaceInfo, cmdtype, userguid, workspaceid string) {
			var imageNames []string
			for _, service := range yamlConfig.Workspace.LinkCompose.Services {
				imageNames = append(imageNames, service.Image)
			}
			if workspaceInfo.CliRunningEnv == workspace.CliRunningEvnEnum_Server {
				serveruserguid := ""
				if workspaceInfo.ServerWorkSpace != nil {
					serveruserguid = workspaceInfo.ServerWorkSpace.OwnerGUID
				}
				appinsight.SetWorkSpaceTrack(cmdtype, args, string(workspaceInfo.Mode), serveruserguid, workspaceInfo.ID, "", "", strings.Join(imageNames, ","))

			} else {
				clientmachinename, _ := os.Hostname()
				appinsight.SetWorkSpaceTrack(cmdtype, args, string(workspaceInfo.Mode), "", "", workspaceid, clientmachinename, strings.Join(imageNames, ","))
			}
		}
		appinsight_k8sFunc := func(yamlConfig config.SmartIdeK8SConfig, workspaceInfo workspace.WorkspaceInfo, cmdtype, userguid, workspaceid string) {
			var imageNames []string
			if len(yamlConfig.Workspace.Deployments) == 0 {
				//pod
				for i := 0; i < len(yamlConfig.Workspace.Others); i++ {
					other := yamlConfig.Workspace.Others[i]

					re := reflect.ValueOf(other)
					kindName := ""
					if re.Kind() == reflect.Ptr {
						re = re.Elem()
					}
					kindName = fmt.Sprint(re.FieldByName("Kind"))
					if kindName == "Pod" {
						var tmpPod *coreV1.Pod
						switch other.(type) {
						case coreV1.Pod:
							tmp := other.(coreV1.Pod)
							tmpPod = &tmp
						default:
							tmpPod = other.(*coreV1.Pod)
						}
						for _, container := range tmpPod.Spec.Containers {
							imageNames = append(imageNames, container.Image)
						}
					}
				}
			} else {
				//Deployment
				for _, deployment := range yamlConfig.Workspace.Deployments {
					for _, container := range deployment.Spec.Template.Spec.Containers {
						imageNames = append(imageNames, container.Image)
					}
				}
			}

			if workspaceInfo.CliRunningEnv == workspace.CliRunningEvnEnum_Server {
				serveruserguid := ""
				if workspaceInfo.ServerWorkSpace != nil {
					serveruserguid = workspaceInfo.ServerWorkSpace.OwnerGUID
				}
				appinsight.SetWorkSpaceTrack(cmdtype, args, string(workspaceInfo.Mode), serveruserguid, workspaceInfo.ID, "", "", strings.Join(imageNames, ","))

			} else {
				clientmachinename, _ := os.Hostname()
				appinsight.SetWorkSpaceTrack(cmdtype, args, string(workspaceInfo.Mode), "", "", workspaceid, clientmachinename, strings.Join(imageNames, ","))
			}
		}

		if workspaceInfo.Mode == workspace.WorkingMode_Local { // 本地模式
			newExtend.LocalNew(cmd, args, workspaceInfo, executeStartCmdFunc)

		} else if workspaceInfo.Mode == workspace.WorkingMode_Remote { // 远程模式
			newExtend.VmNew(cmd, args, workspaceInfo, executeStartCmdFunc)

		} else if workspaceInfo.Mode == workspace.WorkingMode_K8s { // k8s
			k8sUtil, err := k8s.NewK8sUtilWithContent(workspaceInfo.K8sInfo.KubeConfigContent,
				workspaceInfo.K8sInfo.Context,
				workspaceInfo.K8sInfo.Namespace)
			common.CheckError(err)

			if workspaceInfo.CliRunningEnv == workspace.CliRunningEnvEnum_Client {
				newExtend.K8sNew_Local(cmd, args, k8sUtil, workspaceInfo, appinsight_k8sFunc)
			} else {
				newExtend.K8sNew_Server(cmd, args, k8sUtil, workspaceInfo, appinsight_k8sFunc)
			}

		}

		// 如果在本地运行，需要驻守
		if workspaceInfo.CliRunningEnv == workspace.CliRunningEnvEnum_Client {
			for {
				time.Sleep(time.Millisecond * 300)
			}

		}

	},
}

func init() {
	newCmd.Flags().StringP("type", "T", "", i18nInstance.New.Info_help_flag_type)
	newCmd.Flags().BoolVarP(&removeCmdFlag.IsContinue, "yes", "y", false, "目录不为空，是否清空文件夹！")
	newCmd.Flags().BoolVarP(&removeCmdFlag.IsUnforward, "unforward", "", false, "是否禁止端口转发")

	newCmd.Flags().StringP("host", "o", "", i18nInstance.Start.Info_help_flag_host)
	newCmd.Flags().IntP("port", "p", 22, i18nInstance.Start.Info_help_flag_port)
	newCmd.Flags().StringP("username", "u", "", i18nInstance.Start.Info_help_flag_username)
	newCmd.Flags().StringP("password", "", "", i18nInstance.Start.Info_help_flag_password)

	newCmd.Flags().StringP("serverownerguid", "g", "", i18nInstance.Start.Info_help_flag_ownerguid)
	newCmd.Flags().StringP("repourl", "r", "", i18nInstance.Start.Info_help_flag_repourl)
	newCmd.Flags().StringP("branch", "b", "", i18nInstance.Start.Info_help_flag_branch)
	newCmd.Flags().StringP("gitusername", "", "", "访问当前git库的用户信息")
	newCmd.Flags().StringP("gitpassword", "", "", "对当前git库拥有访问权限的令牌")

	newCmd.Flags().StringP("workspacename", "w", "", "工作区名称")

	newCmd.Flags().StringP("k8s", "k", "", i18nInstance.Start.Info_help_flag_k8s)
	newCmd.Flags().StringP("kubeconfig", "", "", "自定义 kube config 文件的本地路径")
}
