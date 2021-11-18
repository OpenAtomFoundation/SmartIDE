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
	"github.com/leansoftX/smartide-cli/cmd/lib"
	"github.com/leansoftX/smartide-cli/cmd/start"
	"github.com/leansoftX/smartide-cli/lib/common"
	"github.com/leansoftX/smartide-cli/lib/i18n"
	"github.com/spf13/cobra"
) // stopCmd represents the stop command

var instanceI18nStop = i18n.GetInstance().Stop

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: instanceI18nStop.Info.Help_short,
	Long:  instanceI18nStop.Info.Help_long,
	Run: func(cmd *cobra.Command, args []string) {

		common.SmartIDELog.Info(instanceI18nStop.Info.Info_start)
		//0.1. 校验是否能正常执行docker
		err := start.CheckLocalEnv()
		common.CheckError(err)

		// 获取 workspace 信息
		common.SmartIDELog.Info("读取工作区信息...")
		workspaceInfo := loadWorkspaceWithDb(cmd, args)

		// 执行
		if (workspaceInfo != dal.WorkspaceInfo{}) {
			if workspaceInfo.Mode == dal.WorkingMode_Local {
				stopLocal(workspaceInfo.WorkingDirectoryPath, workspaceInfo.ConfigFilePath)
			} else {
				//TODO
				stopRemove()
			}
		}

		common.SmartIDELog.Info(instanceI18nStop.Info.Info_end)

	},
}

//
func stopLocal(workingDir string, configFilePath string) {
	//1. 获取docker compose的文件内容
	var yamlFileCongfig lib.YamlFileConfig
	/* 	if configFilePath != "" { //增加指定yaml文件启动
		yamlFileCongfig.SetYamlFilePath(configFilePath)
	} */
	yamlFileCongfig.SetWorkspace(workingDir, configFilePath)
	yamlFileCongfig.GetConfig()
	yamlFileCongfig.ConvertToDockerCompose(common.SSHRemote{}, "", true)
	//servicename := yamlFileCongfig.Workspace.DevContainer.ServiceName

	repoUrl := getLocalGitRepoUrl()
	projectName := getRepoName(repoUrl)
	common.SmartIDELog.Info(fmt.Sprintf("项目名称: %v", projectName))

	yamlFilePath := yamlFileCongfig.GetTempDockerComposeFilePath(yamlFileCongfig.GetLocalWorkingDirectry(), projectName)

	pwd, _ := os.Getwd()
	composeCmd := exec.Command("docker-compose", "-f", yamlFilePath, "--project-directory", pwd, "stop")
	composeCmd.Stdout = os.Stdout
	composeCmd.Stderr = os.Stderr
	if composeCmdErr := composeCmd.Run(); composeCmdErr != nil {
		common.SmartIDELog.Fatal(composeCmdErr)
	}
}

//
func stopRemove() {

}

func init() {
	stopCmd.Flags().StringVarP(&configYamlFileRelativePath, "filepath", "f", "", instanceI18nStop.Info.Help_flag_filepath)

}
