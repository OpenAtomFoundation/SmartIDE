/*
 * @Author: jason chen (jasonchen@leansoftx.com, http://smallidea.cnblogs.com)
 * @Description:
 * @Date: 2021-11
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-04-21 10:05:27
 */
package start

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/leansoftX/smartide-cli/pkg/common"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

// 容器信息
type DockerComposeContainer struct {
	ServiceName   string
	ContainerName string
	//Command       string
	Image   string
	ImageID string
	Ports   []string
	State   string
}

// 获取docker compose运行起来对应的容器
func GetLocalContainersWithServices(ctx context.Context, cli *client.Client, dockerComposeServices []string) []DockerComposeContainer {

	var dockerComposeContainers []DockerComposeContainer // result define

	//通过cli客户端对象去执行ContainerList(其实docker ps 不就是一个docker正在运行容器的一个list嘛)
	containers, err2 := cli.ContainerList(ctx, types.ContainerListOptions{})
	common.CheckError(err2)
	dockerComposeContainers = convertOriginContainer(containers, dockerComposeServices)

	// 打印
	PrintDockerComposeContainers(dockerComposeContainers)

	return dockerComposeContainers
}

// 检测远程服务器的环境，是否安装docker、docker-compose、git
func GetRemoteContainersWithServices(sshRemote common.SSHRemote, dockerComposeServices []string) (dockerComposeContainers []DockerComposeContainer, err error) {

	// https://docs.docker.com/engine/api/v1.41/#operation/ContainerList
	command := "sudo curl -s --unix-socket /var/run/docker.sock http://dummy/containers/json "
	output, err := sshRemote.ExeSSHCommand(command)
	common.CheckError(err)

	if len(output) >= 0 { // 有返回结果的时候才需要转换
		var originContainers []types.Container
		err = json.Unmarshal([]byte(output), &originContainers)
		dockerComposeContainers = convertOriginContainer(originContainers, dockerComposeServices)
	}

	return dockerComposeContainers, err
}

// 转换结构体
func convertOriginContainer(containers []types.Container, dockerComposeServices []string) (dockerComposeContainers []DockerComposeContainer) {
	//
	for _, serviceName := range dockerComposeServices {

		for _, container := range containers {

			if container.Labels["com.docker.compose.service"] == serviceName {
				var ports []string
				for _, port := range container.Ports {
					if port.PublicPort <= 0 {
						continue
					}
					str := fmt.Sprintf("%v:%v", port.PublicPort, port.PrivatePort)
					if !common.Contains(ports, str) { // 限制重复的端口绑定信息
						ports = append(ports, str)
					}
				}

				// 去掉/
				containerName := ""
				for _, name := range container.Names {
					tmp := ""
					if strings.Index(name, "/") == 0 {
						tmp = name[1:]
					} else {
						tmp = name
					}
					if len(containerName) > 0 {
						containerName += "," + tmp
					} else {
						containerName += tmp
					}
				}

				dockerComposeContainer := DockerComposeContainer{
					ServiceName:   serviceName,
					ContainerName: containerName,
					State:         container.State,
					Image:         container.Image,
					ImageID:       container.ImageID,
					Ports:         ports,
				}
				dockerComposeContainers = append(dockerComposeContainers, dockerComposeContainer)
				break
			}

		}
	}

	return dockerComposeContainers
}

// 打印 service 列表
func PrintDockerComposeContainers(dockerComposeContainers []DockerComposeContainer) {
	if len(dockerComposeContainers) <= 0 {
		return
	}
	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
	fmt.Fprintln(w, i18nInstance.Common.Info_table_header_containers)
	for _, service := range dockerComposeContainers {
		line := fmt.Sprintf("%v\t%v\t%v\t%v\t", service.ServiceName, service.State, service.Image, strings.Join(service.Ports, "; "))
		fmt.Fprintln(w, line)
	}
	w.Flush()
}
