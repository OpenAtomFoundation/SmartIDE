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
	"net/http"
	"os"
	"strings"

	"github.com/leansoftX/smartide-cli/lib/common"
	"github.com/leansoftX/smartide-cli/lib/i18n"

	//"strconv"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	network2 "github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/spf13/cobra"
)

var instanceI18nStart = i18n.GetInstance().Start

// docker run
func dockerRun(cli *client.Client, smartIDEImage, smartIDEName string, hostMapping map[string]string, networks []string, tunnelIP string) { //TODO: Command
	// 提示文本
	fmt.Println(i18n.GetInstance().Start.Info.Info_start)

	// check tunnel
	//isTunnel := common.Contains(args, "-tunnel") //TODO: 不区分大小写

	//get current path
	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
		return
	}

	// e.g. docker run -i --user root --name=smartide --init -p 3030:3000 --expose 3001 -p 3001:3001 -v "$(pwd):/home/project" registry.cn-hangzhou.aliyuncs.com/smartide/smartide-node:latest --inspect=0.0.0.0:3001

	// port binding
	portBinding := nat.PortMap{}
	portSets := nat.PortSet{}
	for containerPort, bindingPort := range hostMapping {
		portBinding[nat.Port(containerPort+"/tcp")] = []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: bindingPort}}
		portSets[nat.Port(containerPort)] = struct{}{}
	}
	if len(tunnelIP) > 0 {
		portBinding[nat.Port("22/tcp")] = []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: "6022"}}
	}

	// docker host config
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

	// docker network config
	networkConfig := &network2.NetworkingConfig{
		EndpointsConfig: map[string]*network2.EndpointSettings{},
	}
	//	networkMapping[item] = &network.EndpointSettings{}
	gatewayConfig := &network2.EndpointSettings{
		//Gateway: item,
		Aliases: networks,
	}
	networkConfig.EndpointsConfig["bridge"] = gatewayConfig //?? 如何配置多个容器网络

	//
	ctx := context.Background()
	/* cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	} */

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
			User:         "root",
			Image:        imageName,
			ExposedPorts: portSets, // 容器对外暴露的端口，注意不是宿主机的端口
		}, hostCfg, networkConfig, nil, smartIDEName)
		if err != nil {
			panic(err)
		}

		if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
			panic(err)
		}
		fmt.Println(resp.ID)
	}
}

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: instanceI18nStart.Info.Help_short, //"快速创建并启动SmartIDE开发环境",
	Long:  instanceI18nStart.Info.Help_long,
	Run: func(cmd *cobra.Command, args []string) {

		var yamlFileCongfig YamlFileConfig
		yamlFileCongfig.GetConfig()

		// 提示文本
		fmt.Println(i18n.GetInstance().Start.Info.Info_start)

		//
		ctx := context.Background()
		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			panic(err)
		}

		// networks
		var containerNetworks []string
		for _, dependency := range yamlFileCongfig.Workspace.Dependencies {
			for _, service := range dependency.DockerCompose.servcies {

				for _, network := range service.Networks {
					if !common.Contains(containerNetworks, network) {
						// 依赖容器的网络，缓存下来用于去重
						containerNetworks = append(containerNetworks, network)
						// 创建网络
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
						}

					}
				}
			}
		}

		// web ide 容器
		containerAndHostPortMapping := map[string]string{
			"3000":                                 yamlFileCongfig.Workspace.IdePort,
			yamlFileCongfig.Workspace.AppDebugPort: yamlFileCongfig.Workspace.AppHostPort,
		}
		dockerRun(cli, yamlFileCongfig.Workspace.Image, yamlFileCongfig.Workspace.AppName, containerAndHostPortMapping, nil, "")

		// web ide 需要的其他容器
		for _, dependency := range yamlFileCongfig.Workspace.Dependencies {
			for serviceName, service := range dependency.DockerCompose.servcies {

				var portMapping map[string]string = make(map[string]string)
				for _, portStr := range service.Ports {
					ports := strings.Split(portStr, ":")
					portMapping[ports[0]] = ports[1]
				}

				dockerRun(cli, service.Image, serviceName, portMapping, service.Networks, "")
			}

			//dependency.containerAndHostPortMapping
		}

		// 使用浏览器打开web ide
		fmt.Println(instanceI18nStart.Info.Info_running_openbrower)
		time.Sleep(10 * 1000) //TODO: 检测docker container的运行状态是否为running
		url := fmt.Sprintf(`http://localhost:%v`, yamlFileCongfig.Workspace.IdePort)
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
