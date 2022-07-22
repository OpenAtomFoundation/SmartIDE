/*
 * @Author: jason chen (jasonchen@leansoftx.com, http://smallidea.cnblogs.com)
 * @Description:
 * @Date: 2021-11
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-07-22 16:30:07
 */
package config

import (
	"github.com/leansoftX/smartide-cli/internal/model"
	"github.com/leansoftX/smartide-cli/pkg/common"
)

// 获取端口 及其 描述的映射关系
func (yamlFileConfig *SmartIdeConfig) GetPortLabelMap() map[int]string {
	result := map[int]string{}

	// 去初始配置中查找
	for key, value := range yamlFileConfig.Workspace.DevContainer.Ports {
		result[value] = key
	}

	/* 	// 先到变更中查找
	   	for key, value := range _changedPortLabelMap {
	   		result[value] = key
	   	} */

	return result
}

// 获取描述 及其 端口的映射关系
func (yamlFileConfig *SmartIdeConfig) GetLabelPortMap() map[string]int {
	result := map[string]int{}

	// 去初始配置中查找
	for key, value := range yamlFileConfig.Workspace.DevContainer.Ports {
		result[key] = value
	}

	/* 	// 先到变更中查找
	   	for key, value := range _changedPortLabelMap {
	   		result[key] = value
	   	} */

	return result
}

//
func (yamlFileConfig *SmartIdeConfig) GetPortMappings() []PortMapInfo {
	return yamlFileConfig.Workspace.DevContainer.bindingPorts
}

//
func (yamlFileConfig *SmartIdeConfig) setPort4Label(containerPort int, oldPort int, newPort int, serviceName string) {
	if containerPort <= 0 {
		common.SmartIDELog.Error("containerPort <= 0")
	}
	if containerPort <= 0 {
		common.SmartIDELog.Error("oldPort <= 0")
	}
	if containerPort <= 0 {
		common.SmartIDELog.Error("newPort <= 0")
	}

	if oldPort > 0 && newPort > 0 && containerPort > 0 {
		label := yamlFileConfig.GetLabelWithPort(oldPort, newPort, containerPort)

		//
		isExit := false
		for index, item := range yamlFileConfig.Workspace.DevContainer.bindingPorts {
			if item.HostPortDesc == label && item.OriginHostPort == oldPort {
				if item.CurrentHostPort != newPort {
					item.CurrentHostPort = newPort
					yamlFileConfig.Workspace.DevContainer.bindingPorts[index] = item
				}

				isExit = true
			}
		}
		if !isExit {
			var portMapType PortMapTypeEnum
			if label == "" {
				portMapType = PortMapInfo_OnlyCompose
			} else {
				portMapType = PortMapInfo_Full
			}
			portMap := NewPortMap(portMapType, oldPort, newPort, label, containerPort, serviceName)
			yamlFileConfig.Workspace.DevContainer.bindingPorts = append(yamlFileConfig.Workspace.DevContainer.bindingPorts, *portMap)
		}

	}
}

// 获取 本地端口 对应的描述
func (yamlFileConfig *SmartIdeConfig) GetLabelWithPort(originHostPort, currentHostPort, containerPort int) string {
	label := ""

	// 先到变更中查找
	for _, value := range yamlFileConfig.Workspace.DevContainer.bindingPorts {
		if (originHostPort > 0 && value.OriginHostPort == originHostPort) ||
			(currentHostPort > 0 && value.CurrentHostPort == currentHostPort) {
			label = value.HostPortDesc
			break
		}
	}

	// 去初始配置中查找
	if len(label) <= 0 && originHostPort > 0 {
		for key, value := range yamlFileConfig.Workspace.DevContainer.Ports {
			if value == originHostPort {
				label = key
				break
			}
		}
	}

	// 是否为默认的端口，比如ide、ssh
	if label == "" {
		if originHostPort == model.CONST_Local_Default_BindingPort_WebIDE {
			if containerPort == model.CONST_Container_JetBrainsIDEPort {
				label = "tools-webide-jb"
			} else if containerPort == model.CONST_Container_OpensumiIDEPort {
				label = "tools-webide-opensumi"
			} else {
				label = "tools-webide-vscode"
			}
		} else if originHostPort == model.CONST_Local_Default_BindingPort_SSH {
			label = "tools-ssh"
		}
	}

	return label
}
