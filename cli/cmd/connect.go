/*
SmartIDE - Dev Containers
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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/leansoftX/smartide-cli/cmd/start"
	"github.com/leansoftX/smartide-cli/internal/apk/appinsight"
	"github.com/leansoftX/smartide-cli/internal/biz/config"
	"github.com/leansoftX/smartide-cli/internal/biz/workspace"
	"github.com/leansoftX/smartide-cli/internal/model"
	"github.com/leansoftX/smartide-cli/internal/model/response"
	apiResponse "github.com/leansoftX/smartide-cli/internal/model/response"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/spf13/cobra"
)

var connectCmd = &cobra.Command{
	Use:     "connect",
	Short:   i18nInstance.Connect.Info_help_short,
	Long:    i18nInstance.Connect.Info_help_long,
	Example: `  smartide connect`,
	Run: func(cmd *cobra.Command, args []string) {
		// 用户登录信息
		currentAuth, err := checkLogin(cmd)
		common.CheckError(err)
		if (currentAuth == model.Auth{}) {
			common.SmartIDELog.Error("用户登录信息为空！")
			return
		}
		common.SmartIDELog.Info("login for: " + currentAuth.LoginUrl)
		//appinsight.SetCliLoginTrack(appinsight.Cli_Server_Connect,currentAuth.LoginUrl,currentAuth.UserName,args)
		// cli 运行环境
		cliRunningEnv := workspace.CliRunningEnvEnum_Client
		if value, _ := cmd.Flags().GetString("mode"); strings.ToLower(value) == "server" {
			cliRunningEnv = workspace.CliRunningEvnEnum_Server
		}

		// 是否有工作区数据
		tmpStartedServerWorkspaces, err := getServerWorkspaces(currentAuth, cliRunningEnv, []apiResponse.WorkspaceStatusEnum{response.WorkspaceStatusEnum_Start})
		common.CheckError(err)
		if len(tmpStartedServerWorkspaces) == 0 {
			common.SmartIDELog.Importance("请等待server工作区启动！")
		}

		// 轮询开始端口转发
		isUnforward, _ := cmd.Flags().GetBool("unforward")
		for {
			startedServerWorkspaces, err := getServerWorkspaces(currentAuth, cliRunningEnv, []apiResponse.WorkspaceStatusEnum{})
			if err == io.EOF { // 排除EOF的错误
				common.SmartIDELog.Importance("getServerWorkspaces EOF")
			} else {
				common.SmartIDELog.ImportanceWithError(err)
				connect(startedServerWorkspaces, cmd, args, currentAuth.LoginUrl)

				if isUnforward {
					return
				}
			}

			time.Sleep(time.Second * 10)
		}

	},
}

// 检查并获取当前登录用户的信息
func checkLogin(cmd *cobra.Command) (currentAuth model.Auth, err error) {
	//cliRunningEnv := workspace.CliRunningEnvEnum_Client
	/* if value, _ := cmd.Flags().GetString("mode"); strings.ToLower(value) == "server" {
		cliRunningEnv = workspace.CliRunningEvnEnum_Server
	} */

	// 确保登录
	isLogged := false
	for !isLogged {
		// 查找所有的工作区
		currentAuth, err = workspace.GetCurrentUser()
		common.CheckError(err)

		if currentAuth != (model.Auth{}) && currentAuth.Token != "" && currentAuth.Token != nil {
			// 从api 获取workspace
			err = getServerMenu(currentAuth)
			if err != nil {
				if !strings.Contains(err.Error(), "Client.Timeout exceeded while awaiting headers") {
					common.SmartIDELog.ImportanceWithError(err)
					common.SmartIDELog.Importance("token 已失效，请重新登录！")

					loginCmd.Run(cmd, []string{currentAuth.LoginUrl})
				} else {
					return currentAuth, err
				}

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

func getServerMenu(auth model.Auth) error {
	url := fmt.Sprint(auth.LoginUrl, "/api/menu/getMenu")
	headers := map[string]string{
		"Content-Type": "application/json",
	}
	if auth.Token != nil {
		headers["x-token"] = auth.Token.(string)
	}
	headers["x-user-id"] = "1"
	httpClient := common.CreateHttpClientEnableRetry()
	response, err := httpClient.PostJson(url, make(map[string]interface{}), headers)
	if err != nil {
		return err
	}
	if response == "" {
		return errors.New("服务器返回空！")
	}

	l := &apiResponse.DefaultResponse{}
	err = json.Unmarshal([]byte(response), l)
	if err != nil {
		return err
	}
	if l.Code != 0 {
		return errors.New(l.Msg)
	}
	return nil
}

// 已经连接的工作区id 列表
var connectedWorkspaceIds []string = []string{}

// 获取远程的工作区列表
func getServerWorkspaces(currentAuth model.Auth,
	cliRunningEnv workspace.CliRunningEvnEnum, allowStatuses []apiResponse.WorkspaceStatusEnum) (
	[]workspace.WorkspaceInfo, error) {
	var startedServerWorkspaces []workspace.WorkspaceInfo
	serverWorkSpaces, err := workspace.GetServerWorkspaceList(currentAuth, cliRunningEnv)
	for _, item := range serverWorkSpaces {
		// k8s 工作区不进行connect
		if item.Mode == workspace.WorkingMode_K8s {
			continue
		}

		// 是否包含在过滤状态中
		if len(allowStatuses) > 0 {
			isContain := false
			for _, filterItem := range allowStatuses {
				if filterItem == item.ServerWorkSpace.Status {
					isContain = true
					break
				}
			}
			if !isContain {
				continue
			}
		}

		// 添加到已启动工作区列表
		startedServerWorkspaces = append(startedServerWorkspaces, item)
	}

	return startedServerWorkspaces, err
}

// go routine 启动所有工作区
func connect(startedServerWorkspaces []workspace.WorkspaceInfo, cmd *cobra.Command, args []string, LoginUrl string) {
	// appinsight
	executeStartCmdFunc := func(yamlConfig config.SmartIdeConfig, workspaceInfo workspace.WorkspaceInfo, cmdtype, userguid, workspaceid string) {
		// if config.GlobalSmartIdeConfig.IsInsightEnabled != config.IsInsightEnabledEnum_Enabled {
		// 	common.SmartIDELog.Debug("Application Insights disabled")
		// 	return
		// }
		// var imageNames []string
		// for _, service := range yamlConfig.Workspace.Servcies {
		// 	imageNames = append(imageNames, service.Image)
		// }
		// appinsight.SetTrack(cmd.Use, Version.TagName, trackEvent, string(workspace.WorkingMode_Remote), strings.Join(imageNames, ","))
	}

	var mutex sync.Mutex
	//0. start
	forwardFunc := func(fixWorkspaceInfo workspace.WorkspaceInfo) {
		mutex.Lock()
		common.SmartIDELog.Info(fmt.Sprintf("-- workspace (%v) -------------------------------", fixWorkspaceInfo.ServerWorkSpace.NO))
		var err error
		if fixWorkspaceInfo.Mode == workspace.WorkingMode_Remote {

			var imageNames []string
			for _, service := range fixWorkspaceInfo.ConfigYaml.Workspace.LinkCompose.Services {
				imageNames = append(imageNames, service.Image)
			}
			appinsight.SetAllTrack(appinsight.Cli_Server_Connect, args, LoginUrl, fixWorkspaceInfo.ServerWorkSpace.OwnerGUID, fixWorkspaceInfo.ID, "", "", strings.Join(imageNames, ","))
			_, err = start.ExecuteServerVmStartByClientEnvCmd(fixWorkspaceInfo, executeStartCmdFunc)
		} /* else if fixWorkspaceInfo.Mode == workspace.WorkingMode_K8s {
			err = start.ExecuteServerK8sStartByClientEnvCmd(fixWorkspaceInfo, executeStartCmdFunc)
		} */

		if err != nil {
			common.SmartIDELog.ImportanceWithError(err)
			connectedWorkspaceIds = common.RemoveItem(connectedWorkspaceIds, fixWorkspaceInfo.ServerWorkSpace.NO)
		}
		time.Sleep(time.Second * 26)
		defer mutex.Unlock()

		for {
			if !common.Contains(connectedWorkspaceIds, fixWorkspaceInfo.ServerWorkSpace.NO) {
				common.SmartIDELog.Info(fmt.Sprintf("-- workspace (%v) -------------------------------", fixWorkspaceInfo.ServerWorkSpace.NO))
				common.SmartIDELog.Importance(fmt.Sprintf("当前工作区（%v）非启动状态，端口转发停止！", fixWorkspaceInfo.ServerWorkSpace.NO))
				return
			}
			time.Sleep(time.Second * 10)
		}

	}

	//2. 启动工作区
	for _, workspaceInfo := range startedServerWorkspaces {
		if workspaceInfo.ServerWorkSpace.Status == apiResponse.WorkspaceStatusEnum_Start {
			if !common.Contains(connectedWorkspaceIds, workspaceInfo.ServerWorkSpace.NO) {
				// 加入到已连接数组
				connectedWorkspaceIds = append(connectedWorkspaceIds, workspaceInfo.ServerWorkSpace.NO)

				go forwardFunc(workspaceInfo)
			}

		} else {
			if common.Contains(connectedWorkspaceIds, workspaceInfo.ServerWorkSpace.NO) {
				connectedWorkspaceIds = common.RemoveItem(connectedWorkspaceIds, workspaceInfo.ServerWorkSpace.NO)
			}

		}
	}
}

func init() {
	connectCmd.Flags().BoolP("unforward", "", false, "是否禁止端口转发")
}
