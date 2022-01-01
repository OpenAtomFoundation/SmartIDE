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
	"path/filepath"
	"strings"

	"github.com/leansoftX/smartide-cli/internal/biz/workspace"
	"github.com/leansoftX/smartide-cli/internal/dal"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/spf13/cobra"
)

var resetCmdFalgs struct {

	// 是否确定删除
	IsContinue bool

	// 是否删除远程主机上的文件夹
	IsRemoveDirectory bool

	// 删除compose对应的所有镜像
	IsRemoveAllComposeImages bool

	// 是否删除全部
	IsAll bool
}

// initCmd represents the init command
var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: i18nInstance.Reset.Info_help_short, //,
	Long:  i18nInstance.Reset.Info_help_long,  //i18nInstance.Version.Info_help_long,
	Run: func(cmd *cobra.Command, args []string) {
		// 打印日志
		common.SmartIDELog.Info(i18nInstance.Reset.Info_start)

		// 提示 是否确认
		if !resetCmdFalgs.IsContinue { // 如果设置了参数yes，那么默认就是确认删除
			isEnableRemove := ""
			common.SmartIDELog.Console(i18nInstance.Reset.Warn_confirm_all_remove)
			fmt.Scanln(&isEnableRemove)
			if strings.ToLower(isEnableRemove) != "y" {
				return
			}
		}

		// 打印全部工作区信息
		printWorkspaces()

		// 逐个删除工作区
		common.SmartIDELog.Info(i18nInstance.Reset.Info_workspace_remove_all)
		workspaces, err := dal.GetWorkspaceList()
		common.CheckError(err)
		for _, workspaceInfo := range workspaces {

			// 打印日志
			common.SmartIDELog.InfoF(i18nInstance.Reset.Info_workspace_removing, workspaceInfo.ID)

			// ssh remote 链接检查
			if workspaceInfo.Mode == workspace.WorkingMode_Remote {
				ssmRemote := common.SSHRemote{}
				common.SmartIDELog.InfoF(i18nInstance.Main.Info_ssh_connect_check, workspaceInfo.Remote.Addr, workspaceInfo.Remote.SSHPort)
				err = ssmRemote.CheckDail(workspaceInfo.Remote.Addr, workspaceInfo.Remote.SSHPort, workspaceInfo.Remote.UserName, workspaceInfo.Remote.Password)
				if err != nil {
					if resetCmdFalgs.IsAll { // 删除所有的时候，不顾及太多
						common.SmartIDELog.Importance(err.Error())
					} else {
						common.SmartIDELog.Error(err.Error())
					}
					continue
				}

			}

			// 删除工作区
			common.Block{
				Try: func() {
					//1.1. 删除远程主机的工作区
					if workspaceInfo.Mode == workspace.WorkingMode_Local { // 本地模式
						// 删除对应的容器\镜像
						removeLocalMode(workspaceInfo, resetCmdFalgs.IsRemoveAllComposeImages || resetCmdFalgs.IsAll, resetCmdFalgs.IsAll)

					} else if workspaceInfo.Mode == workspace.WorkingMode_Remote { // 远程模式
						// 删除对应的容器\镜像\工作目录
						removeRemoteMode(workspaceInfo,
							resetCmdFalgs.IsRemoveAllComposeImages || resetCmdFalgs.IsAll, resetCmdFalgs.IsRemoveDirectory || resetCmdFalgs.IsAll, resetCmdFalgs.IsAll)

					}

					//1.2. 删除数据
					dal.RemoveWorkspace(workspaceInfo.ID)

				},
				Catch: func(e common.Exception) {
					msg := fmt.Sprintf("%v", e)
					if resetCmdFalgs.IsAll { // 删除所有的时候，不顾及太多
						common.SmartIDELog.Importance(msg)
					} else {
						common.SmartIDELog.Error(msg)
					}

				},
				Finally: func() {

				},
			}.Do()

		}

		// 删除所有
		if resetCmdFalgs.IsAll {
			dirname, err := os.UserHomeDir()
			common.CheckError(err)

			// 删除.db文件
			sqliteFilePath := filepath.Join(dirname, dal.SqliteFilePath)
			err = os.Remove(sqliteFilePath)
			common.CheckError(err)
			common.SmartIDELog.InfoF(i18nInstance.Reset.Info_db_remove, sqliteFilePath)

			// 删除模板文件
			configFilePath := filepath.Join(dirname, ".ide/.ide.config")
			err = os.Remove(configFilePath)
			common.CheckError(err)
			common.SmartIDELog.InfoF(i18nInstance.Reset.Info_config_remove, configFilePath)

			// 删除配置文件
			templatesDirctoryPath := filepath.Join(dirname, ".ide/templates")
			os.RemoveAll(templatesDirctoryPath)
			common.SmartIDELog.InfoF(i18nInstance.Reset.Info_template_remove, templatesDirctoryPath)
		}

		// end
		common.SmartIDELog.Info(i18nInstance.Reset.Info_end)
	},
}

func init() {
	resetCmd.Flags().BoolVarP(&resetCmdFalgs.IsContinue, "yes", "y", false, i18nInstance.Reset.Info_help_flag_yes)

	resetCmd.Flags().BoolVarP(&resetCmdFalgs.IsRemoveAllComposeImages, "image", "i", false, i18nInstance.Reset.Info_help_flag_image)
	resetCmd.Flags().BoolVarP(&resetCmdFalgs.IsRemoveDirectory, "floder", "f", false, i18nInstance.Reset.Info_help_flag_floder)
	resetCmd.Flags().BoolVarP(&resetCmdFalgs.IsAll, "all", "a", false, i18nInstance.Reset.Info_help_flag_all)
}
