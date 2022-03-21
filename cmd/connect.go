/*
 * @Author: jason chen (jasonchen@leansoftx.com, http://smallidea.cnblogs.com)
 * @Description:
 * @Date: 2022-02-25
 * @LastEditors:
 * @LastEditTime:
 */
package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/leansoftX/smartide-cli/cmd/start"
	"github.com/leansoftX/smartide-cli/internal/apk/appinsight"
	"github.com/leansoftX/smartide-cli/internal/biz/config"
	"github.com/leansoftX/smartide-cli/internal/biz/workspace"
	"github.com/leansoftX/smartide-cli/internal/model"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/spf13/cobra"
)

//
var connectCmd = &cobra.Command{
	Use:     "connect",
	Short:   i18nInstance.Connect.Info_help_short,
	Long:    i18nInstance.Connect.Info_help_long,
	Example: `  smartide connect`,
	Run: func(cmd *cobra.Command, args []string) {
		connectedWorkspaceIds := []string{} // 已启动工作区

		currentAuth, err := checkLogin(cmd)
		common.CheckError(err)
		if (currentAuth == model.Auth{}) {
			common.SmartIDELog.Error("用户登录信息为空！")
			return
		}
		common.SmartIDELog.Info("login for: " + currentAuth.LoginUrl)

		for {
			connect(cmd, currentAuth, args, &connectedWorkspaceIds)

			time.Sleep(time.Second * 10)
		}
	},
}

// 检查并获取当前登录用户的信息
func checkLogin(cmd *cobra.Command) (currentAuth model.Auth, err error) {
	// 确保登录
	isLogged := false
	for !isLogged {
		// 查找所有的工作区
		currentAuth, err = workspace.GetCurrentUser()
		common.CheckError(err)

		if currentAuth != (model.Auth{}) && currentAuth.Token != "" && currentAuth.Token != nil {
			// 从api 获取workspace
			_, err = workspace.GetServerWorkspaceList(currentAuth)
			if err != nil {
				common.SmartIDELog.Importance(err.Error())
				common.SmartIDELog.Importance("token 已失效，请重新登录！")

				loginCmd.Run(cmd, []string{currentAuth.LoginUrl})
			} else {
				isLogged = true
			}
		} else {
			common.SmartIDELog.Importance("运行 connect 命令前，请先登录！")

			loginUrl := ""
			fmt.Printf("请输入服务端地址（默认为%v）：", config.GlobalSmartIdeConfig.DefaultLoginUrl)
			fmt.Scanln(&loginUrl)
			if loginUrl != "" {
				loginCmd.Run(cmd, []string{loginUrl})
			} else {
				loginCmd.Run(cmd, []string{})
			}

		}
	}

	return
}

func connect(cmd *cobra.Command, currentAuth model.Auth, args []string, connectedWorkspaceIds *[]string) {
	//
	serverWorkSpaces, err := workspace.GetServerWorkspaceList(currentAuth)
	var startedServerWorkspaces []workspace.WorkspaceInfo
	for _, item := range serverWorkSpaces {
		if item.ServerWorkSpace.Status == model.WorkspaceStatusEnum_Start {
			startedServerWorkspaces = append(startedServerWorkspaces, item)
		}
	}
	common.CheckError(err)

	// print
	if len(startedServerWorkspaces) == 0 {
		common.SmartIDELog.ImportanceWithConfig(common.LogConfig{RepeatDependSecond: -1}, "请等待server工作区启动！")
	}

	// go routine 启动所有工作区
	//ai记录
	var trackEvent string
	for _, val := range args {
		trackEvent = trackEvent + " " + val
	}

	// 启动工作区
	for _, workspaceInfo := range startedServerWorkspaces {
		if workspaceInfo.ServerWorkSpace.Status == model.WorkspaceStatusEnum_Start &&
			!common.Contains(*connectedWorkspaceIds, workspaceInfo.ServerWorkSpace.NO) {

			executeStartCmdFunc := func(yamlConfig config.SmartIdeConfig) {
				var imageNames []string
				for _, service := range yamlConfig.Workspace.Servcies {
					imageNames = append(imageNames, service.Image)
				}
				appinsight.SetTrack(cmd.Use, Version.TagName, trackEvent, string(workspaceInfo.Mode), strings.Join(imageNames, ","))
			}

			//
			start.ExecuteServerVmStartCmd(workspaceInfo, executeStartCmdFunc)

			// 加入到已连接数组
			*connectedWorkspaceIds = append(*connectedWorkspaceIds, workspaceInfo.ServerWorkSpace.NO)
		}
	}
}

func init() {

}
