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

	//
	for serviceName, originService := range originServices {
		isDevService := serviceName == workspace.ConfigYaml.Workspace.DevContainer.ServiceName

		// 原始端口
		originServicePorts := originService.Ports
		if isDevService {
			// ssh 端口
			if workspace.Mode == WorkingMode_Local || workspace.Mode == WorkingMode_K8s {
				portSSH := fmt.Sprintf("%v:%v", model.CONST_Local_Default_BindingPort_SSH, model.CONST_Container_SSHPort)
				if !common.Contains(originServicePorts, portSSH) {
					originServicePorts = append(originServicePorts, portSSH)
				}
			}

			// webide 端口
			portWebide := fmt.Sprintf("%v:%v", model.CONST_Local_Default_BindingPort_WebIDE, workspace.ConfigYaml.GetContainerWebIDEPort())
			if !common.Contains(originServicePorts, portWebide) {
				originServicePorts = append(originServicePorts, portWebide)
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
					label = "WebIDE"
				} else if originLocalPort == model.CONST_Local_Default_BindingPort_SSH {
					label = "SSH"
				}
			}
			//TODO 端口绑定存在漏洞，并没有指定是哪个sercie，多个service可能会出现重名的情况

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
			if item.OriginLocalPort == port {
				hasContain = true
				break
			}
		}

		if !hasContain {
			portMap := config.NewPortMap(config.PortMapInfo_OnlyLabel, port, -1, label, -1, "")
			extend.Ports = append(extend.Ports, *portMap)
		}

	}

	return extend
}
