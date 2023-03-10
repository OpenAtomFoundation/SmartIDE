/*
SmartIDE - CLI
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

package workspace

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/leansoftX/smartide-cli/internal/biz/config"
	"github.com/leansoftX/smartide-cli/internal/model"
	"github.com/leansoftX/smartide-cli/pkg/common"
)

// 从生成docker-compose 和 配置文件中获取扩展信息（端口绑定）
func (workspaceInfo *WorkspaceInfo) GetWorkspaceExtend() WorkspaceExtend {

	var extend WorkspaceExtend
	if workspaceInfo.Mode != WorkingMode_K8s {
		if workspaceInfo.TempDockerCompose.IsNil() || workspaceInfo.ConfigYaml.IsNil() {
			return extend
		}
	}

	// 兼容链接docker-compose 和 不链接docker-compose 两种方式
	composeServices := workspaceInfo.ConfigYaml.Workspace.Servcies
	if workspaceInfo.ConfigYaml.IsLinkDockerComposeFile() {
		composeServices = workspaceInfo.ConfigYaml.Workspace.LinkCompose.Services
	}

	//1. 遍历 compose 文件中的 services
	for serviceName, originService := range composeServices {
		isDevService := serviceName == workspaceInfo.ConfigYaml.Workspace.DevContainer.ServiceName // 是否开发容器
		isWebTerminal := serviceName == fmt.Sprintf("%v_smartide-webterminal", workspaceInfo.Name)
		originServicePorts := originService.Ports // 原始端口

		//1.1. 开发容器时
		if isDevService {
			// ssh 端口
			portSSH := fmt.Sprintf("%v:%v", model.CONST_Local_Default_BindingPort_SSH, model.CONST_Container_SSHPort)
			if !common.Contains(originServicePorts, portSSH) { // 是否包含
				originServicePorts = append(originServicePorts, portSSH)
			}

			// webide 端口
			containerWebIDEPort := workspaceInfo.ConfigYaml.GetContainerWebIDEPort()
			if containerWebIDEPort != nil { // 判断是否获取到webide的端口，在sdk-only模式下没有webide
				portWebide := fmt.Sprintf("%v:%v", model.CONST_Local_Default_BindingPort_WebIDE, *containerWebIDEPort)
				if !common.Contains(originServicePorts, portWebide) {
					originServicePorts = append(originServicePorts, portWebide)
				}
			}

		}

		//1.2. 遍历原始端口
		for _, temp := range originServicePorts {
			ports := strings.Split(temp, ":")
			originLocalPort, _ := strconv.Atoi(ports[0])
			containerPort, _ := strconv.Atoi(ports[1])
			label := ""

			//1.2.1 从端口描述信息中查找
			for key, port := range workspaceInfo.ConfigYaml.Workspace.DevContainer.Ports {
				if port == originLocalPort {
					label = key
					break
				}
			}

			// webterminal的端口lablel赋值
			if isWebTerminal {
				label = "tools-webterminal"
			}

			//1.2.2. 如果是默认的端口，直接给描述
			if isDevService && label == "" {
				if originLocalPort == model.CONST_Local_Default_BindingPort_WebIDE {
					if containerPort == model.CONST_Container_JetBrainsIDEPort {
						label = model.CONST_DevContainer_PortDesc_JB
					} else if containerPort == model.CONST_Container_OpensumiIDEPort {
						label = model.CONST_DevContainer_PortDesc_Opensumi
					} else {
						label = model.CONST_DevContainer_PortDesc_Vscode
					}
				} else if originLocalPort == model.CONST_Local_Default_BindingPort_SSH {
					label = model.CONST_DevContainer_PortDesc_SSH
				}
			}

			//1.2.3. 如果描述信息不为空，就添加
			if label != "" {
				// 查找当前的绑定端口
				currentLocalPort := -1
				for currentServiceName, currentService := range workspaceInfo.TempDockerCompose.Services {
					if currentServiceName == serviceName {
						for _, str := range currentService.Ports {
							currentPorts := strings.Split(str, ":")
							tmpCurrentLocalPort, _ := strconv.Atoi(currentPorts[0])
							tmpCurrentContainerPort, _ := strconv.Atoi(currentPorts[1])
							if tmpCurrentContainerPort == containerPort {
								currentLocalPort = tmpCurrentLocalPort
								break
							}

						}
					}
				}

				// 端口绑定信息
				portMapInfo := config.NewPortMap(config.PortMapInfo_Full, originLocalPort, currentLocalPort, label, containerPort, serviceName)
				if strings.Contains(portMapInfo.HostPortDesc, "tools-webide") {
					portMapInfo.RefDirecotry = workspaceInfo.GetContainerWorkingPathWithVolumes()
				}
				// set client port
				portMapInfo.OldClientPort = originLocalPort
				portMapInfo.ClientPort = currentLocalPort
				extend.Ports = append(extend.Ports, *portMapInfo)
			}

		}
	}

	//2. 配置文件中的端口
	for label, port := range workspaceInfo.ConfigYaml.Workspace.DevContainer.Ports {

		hasContain := false
		for _, item := range extend.Ports {
			if item.OriginHostPort == port {
				hasContain = true
				break
			}
		}

		if !hasContain { // 不包含在接口列表中
			portMap := config.NewPortMap(config.PortMapInfo_OnlyLabel, port, -1, label, -1, "")
			extend.Ports = append(extend.Ports, *portMap)
		}

	}

	//3. server port configs
	if workspaceInfo.ServerWorkSpace != nil {
		for _, portConfig := range workspaceInfo.ServerWorkSpace.PortConfigs {
			hasContain := false
			for _, item := range extend.Ports {
				if item.OriginHostPort == int(portConfig.Port) {
					hasContain = true
					break
				}
			}

			if !hasContain { // 不包含在接口列表中
				portMap := config.NewPortMap(config.PortMapInfo_OnlyLabel, int(portConfig.Port), -1, portConfig.Label, -1, "")
				extend.Ports = append(extend.Ports, *portMap)
			}
		}
	}

	return extend
}
