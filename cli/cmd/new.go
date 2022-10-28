/*
 * @Date: 2022-04-20 17:08:53
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-10-28 11:40:55
 * @FilePath: /cli/cmd/new.go
 */
package cmd

import (
	"strings"

	newExtend "github.com/leansoftX/smartide-cli/cmd/new"
	"github.com/leansoftX/smartide-cli/internal/apk/appinsight"
	"github.com/leansoftX/smartide-cli/internal/biz/config"
	"github.com/leansoftX/smartide-cli/internal/biz/workspace"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/leansoftX/smartide-cli/pkg/k8s"

	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new",
	Short: i18nInstance.New.Info_help_short,
	Long:  i18nInstance.New.Info_help_long,
	Example: `smartide new <template_type> -T {type_name}
smartide new <template_type> -T {type_name} --host {vm_host_address|vm_host_id} --username {vm_username} --password {vm_password}
smartide new <template_type> -T {type_name} --k8s {kubernetes_context} --kubeconfig {kube_config_file_path}`,
	Run: func(cmd *cobra.Command, args []string) {
		// ai记录
		var trackEvent string
		for _, val := range args {
			trackEvent = trackEvent + " " + val
		}

		workspaceInfo, err := getWorkspaceFromCmd(cmd, args) // 加载工作区信息
		entryptionKey4Workspace(workspaceInfo)               // 申明需要加密的文本
		common.CheckError(err)
		executeStartCmdFunc := func(yamlConfig config.SmartIdeConfig) {
			if config.GlobalSmartIdeConfig.IsInsightEnabled != config.IsInsightEnabledEnum_Enabled {
				common.SmartIDELog.Debug("Application Insights disabled")
				return
			}
			var imageNames []string
			for _, service := range yamlConfig.Workspace.Servcies {
				imageNames = append(imageNames, service.Image)
			}
			appinsight.SetTrack(cmd.Use, Version.TagName, trackEvent, string(workspaceInfo.Mode), strings.Join(imageNames, ","))
		}
		if workspaceInfo.Mode == workspace.WorkingMode_Local { // 本地模式
			newExtend.LocalNew(cmd, args, workspaceInfo, executeStartCmdFunc)

		} else if workspaceInfo.Mode == workspace.WorkingMode_Remote { // 远程模式
			newExtend.VmNew(cmd, args, workspaceInfo, executeStartCmdFunc)

		} else if workspaceInfo.Mode == workspace.WorkingMode_K8s { // k8s
			k8sUtil, err := k8s.NewK8sUtil(workspaceInfo.K8sInfo.KubeConfigFilePath,
				workspaceInfo.K8sInfo.Context,
				workspaceInfo.K8sInfo.Namespace)
			common.CheckError(err)

			if workspaceInfo.CliRunningEnv == workspace.CliRunningEnvEnum_Client {
				newExtend.K8sNew_Local(cmd, args, k8sUtil, workspaceInfo, executeStartCmdFunc)
			} else {
				newExtend.K8sNew_Server(cmd, args, k8sUtil, workspaceInfo, executeStartCmdFunc)
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

	newCmd.Flags().StringP("workspacename", "w", "", "工作区名称")

	newCmd.Flags().StringP("k8s", "k", "", i18nInstance.Start.Info_help_flag_k8s)
	newCmd.Flags().StringP("kubeconfig", "", "", "自定义 kube config 文件的本地路径")
}
