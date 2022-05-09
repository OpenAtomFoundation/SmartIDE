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
func (workspace *WorkspaceInfo) GetWorkspaceExtend() WorkspaceExtend {

	var extend WorkspaceExtend
	if workspace.Mode != WorkingMode_K8s {
		if workspace.TempDockerCompose.IsNil() || workspace.ConfigYaml.IsNil() {
			return extend
		}
	}

	// 兼容链接docker-compose 和 不链接docker-compose 两种方式
	originServices := workspace.ConfigYaml.Workspace.Servcies
	if workspace.ConfigYaml.IsLinkDockerComposeFile() {
		originServices = workspace.LinkDockerCompose.Services
	}

	// 遍历 services
	for serviceName, originService := range originServices {
		isDevService := serviceName == workspace.ConfigYaml.Workspace.DevContainer.ServiceName // 是否开发容器
		originServicePorts := originService.Ports                                              // 原始端口

		// 开发容器时
		if isDevService {
			// ssh 端口
			portSSH := fmt.Sprintf("%v:%v", model.CONST_Local_Default_BindingPort_SSH, model.CONST_Container_SSHPort)
			if !common.Contains(originServicePorts, portSSH) { // 是否包含
				originServicePorts = append(originServicePorts, portSSH)
			}

			// webide 端口
			containerWebIDEPort := workspace.ConfigYaml.GetContainerWebIDEPort()
			if containerWebIDEPort != nil { // 判断是否获取到webide的端口，在sdk-only模式下没有webide
				portWebide := fmt.Sprintf("%v:%v", model.CONST_Local_Default_BindingPort_WebIDE, *containerWebIDEPort)
				if !common.Contains(originServicePorts, portWebide) {
					originServicePorts = append(originServicePorts, portWebide)
				}
			}

		}

		// 遍历原始端口
		for _, temp := range originServicePorts {
			ports := strings.Split(temp, ":")
			originLocalPort, _ := strconv.Atoi(ports[0])
			containerPort, _ := strconv.Atoi(ports[1])
			label := ""

			// 从端口描述信息中查找
			for key, port := range workspace.ConfigYaml.Workspace.DevContainer.Ports {
				if port == originLocalPort {
					label = key
					break
				}
			}

			// 如果是默认的端口，直接给描述
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

			// 如果描述信息不为空，就添加
			if label != "" {
				// 查找当前的绑定端口
				currentLocalPort := -1
				for currentServiceName, currentService := range workspace.TempDockerCompose.Services {
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
				extend.Ports = append(extend.Ports, *portMapInfo)
			}

		}
	}

	//
	for label, port := range workspace.ConfigYaml.Workspace.DevContainer.Ports {

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

	return extend
}
