/*
 * @Date: 2022-04-20 17:08:53
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-07-19 15:05:51
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

	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new",
	Short: i18nInstance.New.Info_help_short,
	Long:  i18nInstance.New.Info_help_long,
	Example: `  smartide new
  smartide new <templatetype> -t {typename}`,
	Run: func(cmd *cobra.Command, args []string) {

		//ai记录
		var trackEvent string
		for _, val := range args {
			trackEvent = trackEvent + " " + val
		}

		workspaceInfo, err := getWorkspaceFromCmd(cmd, args)
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

		}

	},
}

func init() {
	newCmd.Flags().StringP("type", "t", "", i18nInstance.New.Info_help_flag_type)
	newCmd.Flags().BoolVarP(&removeCmdFlag.IsContinue, "yes", "y", false, "目录不为空，是否清空文件夹！")
	newCmd.Flags().BoolVarP(&removeCmdFlag.IsUnforward, "unforward", "", false, "是否禁止端口转发")

	newCmd.Flags().StringP("host", "o", "", i18nInstance.Start.Info_help_flag_host)
	newCmd.Flags().StringP("workspacename", "w", "", "工作区名称")
	newCmd.Flags().IntP("port", "p", 22, i18nInstance.Start.Info_help_flag_port)
	newCmd.Flags().StringP("username", "u", "", i18nInstance.Start.Info_help_flag_username)
	newCmd.Flags().StringP("password", "", "", i18nInstance.Start.Info_help_flag_password)
}
