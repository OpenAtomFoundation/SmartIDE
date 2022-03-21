/*
 * @Author: jason chen (jasonchen@leansoftx.com, http://smallidea.cnblogs.com)
 * @Description:
 * @Date: 2021-11
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-03-16 14:27:23
 */
package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/leansoftX/smartide-cli/cmd/server"
	"github.com/leansoftX/smartide-cli/cmd/start"
	"github.com/leansoftX/smartide-cli/internal/biz/workspace"
	"github.com/leansoftX/smartide-cli/internal/model"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/spf13/cobra"
) // stopCmd represents the stop command

var stopCmd = &cobra.Command{
	Use:     "stop",
	Short:   i18nInstance.Stop.Info_help_short,
	Long:    i18nInstance.Stop.Info_help_long,
	Example: `  smartide stop {workspaceid} `,
	Run: func(cmd *cobra.Command, args []string) {

		common.SmartIDELog.Info(i18nInstance.Stop.Info_start)
		mode, _ := cmd.Flags().GetString("mode")

		// 获取 workspace 信息
		common.SmartIDELog.Info(i18nInstance.Main.Info_workspace_loading)
		var workspaceInfo workspace.WorkspaceInfo
		// 当前登录信息
		currentAuth, err := workspace.GetCurrentUser()
		common.CheckError(err)
		workspaceIdStr := getWorkspaceIdFromFlagsAndArgs(cmd, args)
		if strings.ToLower(mode) == "server" || strings.Contains(workspaceIdStr, "SWS") { // 当mode=server时，从server端反调数据

			workspaceInfo, err = workspace.GetWorkspaceFromServer(currentAuth, workspaceIdStr)

			if err == nil {
				if workspaceInfo.ID == "" || workspaceInfo.ServerWorkSpace.NO == "" {
					err = fmt.Errorf("没有查询到 %v 对应的工作区数据！", workspaceIdStr)
				}
			}
			if err != nil {
				msg := err.Error()
				server.Feedback_Finish(server.FeedbackCommandEnum_Stop, cmd, false, 0, workspaceInfo, msg)
			}
			common.CheckError(err)

		} else {
			workspaceInfo = loadWorkspaceWithDb(cmd, args)
		}

		if strings.ToLower(mode) == "server" {
			msg := ""
			// 远程主机上停止
			err := stopRemote(workspaceInfo) //todo return error
			if err != nil {
				msg = err.Error()
			}

			// feeadback
			common.SmartIDELog.Info("反馈运行结果...")
			err = server.Feedback_Finish(server.FeedbackCommandEnum_Stop, cmd, err == nil, 0, workspaceInfo, msg)
			common.CheckError(err)

		} else if workspaceInfo.Mode == workspace.WorkingMode_Server { // 录入的是服务端工作区id
			// 触发stop
			err = server.Trigger_Action("stop", workspaceIdStr, currentAuth.LoginUrl, currentAuth, make(map[string]interface{}))
			common.CheckError(err)

			// 轮询检查工作区状态
			isStop := false
			for !isStop {
				serverWorkSpace, err := workspace.GetWorkspaceFromServer(currentAuth, workspaceInfo.ID)
				if err != nil {
					common.SmartIDELog.Importance(err.Error())
				}
				if serverWorkSpace.ServerWorkSpace.Status == model.WorkspaceStatusEnum_Stop ||
					serverWorkSpace.ServerWorkSpace.Status == model.WorkspaceStatusEnum_Error_Stop {
					isStop = true
				}

				time.Sleep(time.Second * 15)
			}

		} else { // 普通模式下
			// 判断是否有工作区数据
			if workspaceInfo.IsNil() {
				common.SmartIDELog.Error(i18nInstance.Main.Err_workspace_none)
			}

			// 执行对应的stop
			if workspaceInfo.Mode == workspace.WorkingMode_Local {
				stopLocal(workspaceInfo)

			} else {
				err := stopRemote(workspaceInfo)
				common.CheckError(err)

			}

		}

		common.SmartIDELog.Info(i18nInstance.Stop.Info_end)

	},
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
func stopRemote(workspaceInfo workspace.WorkspaceInfo) error {
	// ssh 连接
	common.SmartIDELog.Info(i18nInstance.Stop.Info_sshremote_connection_creating)
	sshRemote, err := common.NewSSHRemote(workspaceInfo.Remote.Addr, workspaceInfo.Remote.SSHPort, workspaceInfo.Remote.UserName, workspaceInfo.Remote.Password)
	if err != nil {
		return err
	}

	// 项目文件夹是否存在
	if !sshRemote.IsCloned(workspaceInfo.WorkingDirectoryPath) {
		msg := fmt.Sprintf(i18nInstance.Stop.Err_env_project_dir_remove, workspaceInfo.ID)
		return errors.New(msg)
	}

	// 检查临时文件夹是否存在
	if !sshRemote.IsExit(workspaceInfo.TempDockerComposeFilePath) || !sshRemote.IsExit(workspaceInfo.ConfigYaml.GetConfigYamlFilePath()) {
		workspaceInfo.SaveTempFilesForRemote(sshRemote)
	}

	// 检查环境
	err = start.CheckRemoveEnv(sshRemote)
	if err != nil {
		return err
	}

	// 停止容器
	common.SmartIDELog.Info(i18nInstance.Stop.Info_docker_stopping)
	command := fmt.Sprintf("docker-compose -f %v --project-directory %v stop",
		common.FilePahtJoin4Linux(workspaceInfo.TempDockerComposeFilePath), common.FilePahtJoin4Linux(workspaceInfo.WorkingDirectoryPath))
	err = sshRemote.ExecSSHCommandRealTime(command)
	if err != nil {
		return err
	}

	return nil
}

func init() {
	//stopCmd.Flags().StringVarP(&configYamlFileRelativePath, "filepath", "f", "", i18nInstance.Stop.Info_help_flag_filepath)

}
