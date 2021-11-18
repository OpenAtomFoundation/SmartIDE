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
	"strings"

	"github.com/leansoftX/smartide-cli/lib/common"
	"github.com/leansoftX/smartide-cli/lib/i18n"
	"github.com/spf13/cobra"
	"gopkg.in/src-d/go-git.v4"

	"github.com/leansoftX/smartide-cli/cmd/dal"
	"github.com/leansoftX/smartide-cli/cmd/lib"
	"github.com/leansoftX/smartide-cli/cmd/start"
)

var instanceI18nRemove = i18n.GetInstance().Remove

// removeCmd represents the remove command
var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: instanceI18nRemove.Info.Help_short,
	Long:  instanceI18nRemove.Info.Help_long,
	Example: `
	smartide remove <workspaceid>
	smartide remove <workspaceid> -y`,
	Run: func(cmd *cobra.Command, args []string) {

		common.SmartIDELog.Info(instanceI18nRemove.Info.Info_start)
		//0.1. 校验是否能正常执行docker
		err := start.CheckLocalEnv()
		common.CheckError(err)

		// 获取 workspace 信息
		common.SmartIDELog.Info("读取工作区信息...")
		workspaceInfo := loadWorkspaceWithDb(cmd, args)

		// 是否确认删除
		isYesFlag, err := cmd.Flags().GetBool("yes")
		common.CheckError(err)
		if !isYesFlag { // 如果设置了参数yes，那么默认就是确认删除
			isEnableRemove := ""
			common.SmartIDELog.Console("是否确认删除（y｜n）？")
			fmt.Scanln(&isEnableRemove)
			if strings.ToLower(isEnableRemove) != "y" {
				return
			}
		}

		// 执行
		if (workspaceInfo != dal.WorkspaceInfo{}) {
			if workspaceInfo.Mode == dal.WorkingMode_Local {
				removeLocalMode(workspaceInfo.WorkingDirectoryPath, workspaceInfo.ConfigFilePath)
			} else {
				//TODO
				removeRemoteMode()
			}
		}

		// remote workspace in db
		if workspaceInfo.ID > 0 {
			err := dal.RemoveWorkspace(workspaceInfo.ID)
			common.CheckError(err)
		}

		// log
		common.SmartIDELog.Info(instanceI18nRemove.Info.Info_end)
	},
}

// 从flag、args中获取参数信息，然后再去数据库中读取相关数据
func loadWorkspaceWithDb(cmd *cobra.Command, args []string) dal.WorkspaceInfo {
	workspaceInfo := dal.WorkspaceInfo{}
	workspaceId := getWorkspaceIdFromFlagsAndArgs(cmd, args)
	if workspaceId > 0 { // 从db中获取workspace的信息
		var err2 error
		workspaceInfo, err2 = getWorkspaceWithDbAndValid(workspaceId)
		common.CheckError(err2)

	} else { // 当没有workspace id 的时候，只能是本地模式 + 当前目录对应workspace
		// current directory
		pwd, err := os.Getwd()
		common.CheckError(err)

		// git remote url
		gitRepo, err := git.PlainOpen(pwd)
		common.CheckError(err)
		gitRemote, err := gitRepo.Remote("origin")
		common.CheckError(err)
		gitRemmoteUrl := gitRemote.Config().URLs[0]

		/* 		// host
		   		var remoteId int
		   		var remoteHost string
		   		hostFlag := cmd.Flag("host")
		   		if hostFlag != nil {
		   			host := hostFlag.Value.String()
		   			if common.IsNumber(host) {
		   				remoteId, err = strconv.Atoi(host)
		   				common.CheckError(err)
		   			} else {
		   				remoteHost = host
		   			}
		   		} */

		workspaceInfo, err = dal.GetSingleWorkspaceByParams(dal.WorkingMode_Local, pwd, gitRemmoteUrl, -1, "")
		common.CheckError(err)
		if (workspaceInfo == dal.WorkspaceInfo{}) {
			common.SmartIDELog.Error("当前工作区未创建，无须执行remote操作")
		}
	}

	return workspaceInfo
}

// 本地删除工作去对应的环境
func removeLocalMode(workingDir string, configFilePath string) error {
	//1. 获取docker compose的文件内容
	var yamlFileCongfig lib.YamlFileConfig
	/* if configFilePath != "" { //增加指定yaml文件启动
		yamlFileCongfig.SetYamlFilePath(configFilePath)
	} */
	yamlFileCongfig.SetWorkspace(workingDir, configFilePath)
	yamlFileCongfig.GetConfig()
	yamlFileCongfig.ConvertToDockerCompose(common.SSHRemote{}, "", true)

	repoUrl := getLocalGitRepoUrl()
	projectName := getRepoName(repoUrl)

	//servicename := yamlFileCongfig.Workspace.DevContainer.ServiceName
	tmpDockerComposeFilePath := yamlFileCongfig.GetTempDockerComposeFilePath(yamlFileCongfig.GetLocalWorkingDirectry(), projectName)

	pwd, _ := os.Getwd()
	composeCmd := exec.Command("docker-compose", "-f", tmpDockerComposeFilePath, "--project-directory", pwd, "down", "-v")
	composeCmd.Stdout = os.Stdout
	composeCmd.Stderr = os.Stderr
	if composeCmdErr := composeCmd.Run(); composeCmdErr != nil {
		common.SmartIDELog.Fatal(composeCmdErr)
	}

	return nil
}

//
func removeRemoteMode() {

}

func init() {
	removeCmd.Flags().StringVarP(&configYamlFileRelativePath, "filepath", "f", "", instanceI18nRemove.Info.Help_flag_filepath)
	removeCmd.Flags().Int32P("workspaceid", "w", 0, "设置后，可以删除对应的工作空间")
	removeCmd.Flags().BoolP("yes", "y", false, "设置后，将不在提示是否删除")

	//TODO 	强制删除
}
