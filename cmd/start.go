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
	"net/http"

	"github.com/leansoftX/smartide-cli/lib/common"
	"github.com/leansoftX/smartide-cli/lib/i18n"

	//"strconv"
	"time"

	"github.com/spf13/cobra"
)

var instanceI18nStart = i18n.GetInstance().Start

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: instanceI18nStart.Info.Help_short, //"快速创建并启动SmartIDE开发环境",
	Long:  instanceI18nStart.Info.Help_long,
	Run: func(cmd *cobra.Command, args []string) {

		//1. 获取docker compose的文件内容
		var yamlFileCongfig YamlFileConfig
		yamlFileCongfig.GetConfig()
		dockerCompose := yamlFileCongfig.ConvertToDockerCompose()
		fmt.Print(dockerCompose)

		// 提示文本
		fmt.Println(i18n.GetInstance().Start.Info.Info_start)

		//2. 运行docker-compose命令

		//3. 使用浏览器打开web ide
		fmt.Println(instanceI18nStart.Info.Info_running_openbrower)
		time.Sleep(10 * 1000) //TODO: 检测docker container的运行状态是否为running
		url := fmt.Sprintf(`http://localhost:%v`, yamlFileCongfig.Workspace.DevContainer.WebidePort)
		isUrlReady := false
		for !isUrlReady {
			resp, err := http.Get(url)
			if (err == nil) && (resp.StatusCode == 200) {
				isUrlReady = true
				common.OpenBrowser(url)
			}

		}

		fmt.Println(instanceI18nStart.Info.Info_end)

		//TODO tunnel

	},
}

func init() {
	//rootCmd.AddCommand(startCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// startCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// startCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
