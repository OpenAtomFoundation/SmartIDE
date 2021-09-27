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
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/leansoftX/smartide-cli/lib/common"
	"github.com/leansoftX/smartide-cli/lib/i18n"
	"github.com/leansoftX/smartide-cli/lib/tunnel"

	//"strconv"

	"github.com/spf13/cobra"
)

var instanceI18nStart = i18n.GetInstance().Start

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: instanceI18nStart.Info.Help_short, //"快速创建并启动SmartIDE开发环境",
	Long:  instanceI18nStart.Info.Help_long,
	Run: func(cmd *cobra.Command, args []string) {

		//0. 提示文本
		fmt.Println(i18n.GetInstance().Start.Info.Info_start)

		//1. 获取docker compose的文件内容
		var yamlFileCongfig YamlFileConfig
		yamlFileCongfig.GetConfig()
		dockerCompose, sshBindingPort := yamlFileCongfig.ConvertToDockerCompose()
		dockerCompose.ConvertToStr()
		yamlFilePath, _ := dockerCompose.SaveFile(yamlFileCongfig.Workspace.DevContainer.ServiceName)
		fmt.Printf("SSH转发端口：%v ", sshBindingPort) //TODO: 国际化	// 提示用户ssh端口绑定到了本地的某个端口

		pwd, _ := os.Getwd()
		//fmt.Printf("current dir : %s \n", pwd)

		//2. 创建容器
		//2.1. 创建网络
		ctx := context.Background()
		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			panic(err)
		}
		for network := range dockerCompose.Networks {
			networkList, _ := cli.NetworkList(ctx, types.NetworkListOptions{})
			isContain := false
			for _, item := range networkList {
				if item.Name == network {
					isContain = true
					break
				}
			}
			if !isContain {
				cli.NetworkCreate(ctx, network, types.NetworkCreate{})
				fmt.Print("创建网络 " + network) //TODO: 国际化
			}
		}

		//2.2. 运行docker-compose命令
		// docker-compose -f /Users/jasonchen/.ide/docker-compose-product-service-dev.yaml --project-directory /Users/jasonchen/Project/boat-house/boat-house-backend/src/product-service/api up -d
		// e.g. exec.Command("docker-compose", "-f "+yamlFilePath+" up -d")
		composeCmd := exec.Command("docker-compose", "-f", yamlFilePath, "--project-directory", pwd, "up", "-d") // "--project-directory", envPath,
		composeCmd.Stdout = os.Stdout
		composeCmd.Stderr = os.Stderr
		if composeCmdErr := composeCmd.Run(); composeCmdErr != nil {
			common.SmartIDELog.Fatal(composeCmdErr)
		}

		//3. 使用浏览器打开web ide
		fmt.Println(instanceI18nStart.Info.Info_running_openbrower) //TODO: 增加等待某某网址的提示
		url := fmt.Sprintf(`http://localhost:%v`, yamlFileCongfig.Workspace.DevContainer.WebidePort)
		isUrlReady := false
		for !isUrlReady {
			resp, err := http.Get(url)
			if (err == nil) && (resp.StatusCode == 200) {
				isUrlReady = true
				common.OpenBrowser(url)
			}

		}

		//99. 结束
		fmt.Println(instanceI18nStart.Info.Info_end)

		// tunnel
		for {
			//TODO: 端口冲突
			tunnel.AutoTunnelMultiple(fmt.Sprintf("localhost:%v", sshBindingPort), "root", "root123") //TODO: 登录的用户名，密码要能够从配置文件中读取出来
			time.Sleep(time.Second * 10)
		}

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
