package docker

import (
	"context"
	"io"
	"log"
	"os"

	"github.com/leansoftX/smartide-cli/internal/apk/i18n"
	"github.com/leansoftX/smartide-cli/pkg/common"

	//"strconv"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	network2 "github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

// docker run
func dockerCreateAndRun(cli *client.Client, smartIDEImage, smartIDEName string, hostMapping map[string]string, networks []string, tunnelIP string) { //TODO: Command
	// 提示文本
	common.SmartIDELog.Info(i18n.GetInstance().Start.Info_start)

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
	// networkMapping[item] = &network.EndpointSettings{}
	gatewayConfig := &network2.EndpointSettings{
		//Gateway: networks[0],
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
		common.SmartIDELog.Info(i18n.GetInstance().Start.Info_running_container)
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
		common.SmartIDELog.Info(resp.ID)
	}
}

/* 示例
//
		ctx := context.Background()
		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			panic(err)
		}

		// networks
		var containerNetworks []string
		for _, dependency := range yamlFileCongfig.Workspace.Dependencies {
			for _, service := range dependency.DockerCompose.Servcies {

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
		dockerRun(cli, yamlFileCongfig.Workspace.Image, yamlFileCongfig.Workspace.AppName, containerAndHostPortMapping, containerNetworks, "")

		// web ide 需要的其他容器
		for _, dependency := range yamlFileCongfig.Workspace.Dependencies {
			for serviceName, service := range dependency.DockerCompose.Servcies {

				var portMapping map[string]string = make(map[string]string)
				for _, portStr := range service.Ports {
					ports := strings.Split(portStr, ":")
					portMapping[ports[0]] = ports[1]
				}

				dockerRun(cli, service.Image, serviceName, portMapping, service.Networks, "")
			}

			//dependency.containerAndHostPortMapping
		}
*/
