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

package start

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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
func GetLocalContainersWithServices(ctx context.Context, cli *client.Client,
	workingDir string, dockerComposeServices []string) []DockerComposeContainer {

	var dockerComposeContainers []DockerComposeContainer // result define

	// home dir
	if workingDir[0:1] == "~" {
		homeDir, _ := os.UserHomeDir()
		workingDir = filepath.Join(homeDir, workingDir[1:])
	}

	//通过cli客户端对象去执行ContainerList(其实docker ps 不就是一个docker正在运行容器的一个list嘛)
	containers, err2 := cli.ContainerList(ctx, types.ContainerListOptions{})
	common.CheckError(err2)
	dockerComposeContainers = convertOriginContainer(containers, workingDir, dockerComposeServices)

	// 打印
	PrintDockerComposeContainers(dockerComposeContainers)

	return dockerComposeContainers
}

// 检测远程服务器的环境，是否安装docker、docker-compose、git
func GetRemoteContainersWithServices(sshRemote common.SSHRemote,
	workingDir string, dockerComposeServices []string) (dockerComposeContainers []DockerComposeContainer, err error) {

	// https://docs.docker.com/engine/api/v1.41/#operation/ContainerList
	command := "sudo curl -s --unix-socket /var/run/docker.sock http://dummy/containers/json "
	output, err := sshRemote.ExeSSHCommand(command)
	common.CheckError(err)

	if len(output) >= 0 { // 有返回结果的时候才需要转换
		var originContainers []types.Container
		err = json.Unmarshal([]byte(output), &originContainers)

		// home dir
		if workingDir[0:1] == "~" {
			homeDir, _ := sshRemote.GetRemoteHome()
			workingDir = common.FilePahtJoin4Linux(homeDir, workingDir[1:])
		}

		dockerComposeContainers = convertOriginContainer(originContainers, workingDir, dockerComposeServices)
	}

	return dockerComposeContainers, err
}

// 转换结构体
func convertOriginContainer(containers []types.Container,
	workingDir string, dockerComposeServices []string) (dockerComposeContainers []DockerComposeContainer) {
	for _, container := range containers {
		currentServiceName := container.Labels["com.docker.compose.service"]
		currentWorkingDir := container.Labels["com.docker.compose.project.working_dir"]
		if workingDir != "" && currentWorkingDir != workingDir { // 工作目录不匹配
			continue
		}
		if common.Contains(dockerComposeServices, currentServiceName) {
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
				ServiceName:   currentServiceName,
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
