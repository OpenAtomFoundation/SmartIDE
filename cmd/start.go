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
	"io"
	"log"
	"os"
	"os/exec"
	"runtime"

	"github.com/leansoftX/smartide-cli/lib/i18n"

	//"strconv"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/spf13/cobra"
)

func openbrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
	}

}

var instanceI18nStart = i18n.GetInstance().Start

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: instanceI18nStart.Info.Help_short, //"快速创建并启动SmartIDE开发环境",
	Long:  instanceI18nStart.Info.Help_long,
	Run: func(cmd *cobra.Command, args []string) {

		var yamlFileCongfig YamlFileConfig
		yamlFileCongfig.GetConfig()

		var smartIDEPort = yamlFileCongfig.Workspace.IdePort
		var smartIDEImage = yamlFileCongfig.Workspace.Image
		var smartIDEName = yamlFileCongfig.Workspace.AppName
		smartIDEImageDefaultPort := "3000"

		fmt.Println(i18n.GetInstance().Start.Info.Info_start)

		//get current path
		currentDir, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
			return
		}

		// e.g. docker run -i --user root --name=smartide --init -p 3030:3000 --expose 3001 -p 3001:3001 -v "$(pwd):/home/project" registry.cn-hangzhou.aliyuncs.com/smartide/smartide-node:latest --inspect=0.0.0.0:3001

		// port binding
		portBinding := nat.PortMap{
			nat.Port(yamlFileCongfig.Workspace.AppDebugPort): []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: yamlFileCongfig.Workspace.AppHostPort}},
			nat.Port(smartIDEImageDefaultPort + "/tcp"):      []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: smartIDEPort}},
		}

		// docker run parameters
		hostCfg := &container.HostConfig{
			Mounts: []mount.Mount{
				{
					Type:   mount.TypeBind,
					Source: currentDir,
					Target: "/home/project",
				}},
			PortBindings: portBinding,
			RestartPolicy: container.RestartPolicy{
				Name: "always",
			},
			// AutoRemove: true,
		}

		ctx := context.Background()
		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			panic(err)
		}

		//
		startErr := cli.ContainerStart(ctx, smartIDEName, types.ContainerStartOptions{})
		if startErr != nil {
			fmt.Println(instanceI18nStart.Info.Info_running_container)
			imageName := smartIDEImage
			out, err := cli.ImagePull(ctx, imageName, types.ImagePullOptions{})
			if err != nil {
				panic(err)
			}
			io.Copy(os.Stdout, out)
			resp, err := cli.ContainerCreate(ctx, &container.Config{
				User:  "root",
				Image: imageName,
				ExposedPorts: nat.PortSet{
					nat.Port(yamlFileCongfig.Workspace.AppDebugPort): {},
					nat.Port(smartIDEImageDefaultPort):               {},
				}, // 容器对外暴露的端口，注意不是宿主机的端口
			}, hostCfg, nil, nil, smartIDEName)
			if err != nil {
				panic(err)
			}

			if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
				panic(err)
			}
			fmt.Println(resp.ID)
		}

		// 使用浏览器打开web ide
		fmt.Println(instanceI18nStart.Info.Info_running_openbrower)
		/* outStatus, err := cli.ContainerStats(ctx, smartIDEName, false)
		fmt.Println(outStatus) */
		time.Sleep(10 * 1000) //TODO: 检测docker container的运行状态是否为running
		url := fmt.Sprintf(`http://localhost:%v`, smartIDEPort)
		openbrowser(url)

		fmt.Println(instanceI18nStart.Info.Info_end)

	},
}

func init() {
	rootCmd.AddCommand(startCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// startCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// startCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
