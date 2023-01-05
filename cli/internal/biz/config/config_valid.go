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

package config

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/leansoftX/smartide-cli/pkg/common"
)

// 验证配置文件格式是否正确
func (c SmartIdeK8SConfig) Valid() error {
	// Workspace.KubeDeployFiles 节点
	if c.Workspace.KubeDeployFileExpression == "" {
		return errors.New("Workspace.KubeDeployFiles 未在配置文件中定义！")
	}

	// service name 必须在 k8s 部署文件中申明
	isContainServiceName := false
	for _, deployment := range c.Workspace.Deployments {
		for _, container := range deployment.Spec.Template.Spec.Containers {
			if container.Name == c.Workspace.DevContainer.ServiceName {
				isContainServiceName = true
				break
			}
		}
	}
	if !isContainServiceName {
		for _, other := range c.Workspace.Others {
			re := reflect.ValueOf(other)
			if re.Kind() == reflect.Ptr {
				re = re.Elem()
			}
			kindName := fmt.Sprint(re.FieldByName("Kind"))
			if strings.ToLower(kindName) == "pod" && re.FieldByName("ObjectMeta").FieldByName("Name").String() == c.Workspace.DevContainer.ServiceName {
				isContainServiceName = true
				break
			}
		}
	}
	if !isContainServiceName {
		return fmt.Errorf("service (%v) 未在关联 k8s yaml 中定义！", c.Workspace.DevContainer.ServiceName)
	}

	// 申明的端口是否在service中存在
	for portLabel, port := range c.Workspace.DevContainer.Ports {
		isContain := false
		for _, service := range c.Workspace.Services {
			for _, specPort := range service.Spec.Ports {
				if specPort.Port == int32(port) {
					isContain = true
					break
				}
			}
		}
		if !isContain {
			return fmt.Errorf("端口 (%v:%v) 没有在k8s yaml文件中申明！", portLabel, port)
		}
	}

	return nil
}

// 验证配置文件格式是否正确
func (c SmartIdeConfig) Valid() error {
	// 格式不能为空
	if c.Orchestrator.Type == "" {
		return errors.New(i18nInstance.Config.Err_config_orchestrator_type_none)

	} else {
		if c.Orchestrator.Type != OrchestratorTypeEnum_Compose &&
			c.Orchestrator.Type != OrchestratorTypeEnum_K8S &&
			c.Orchestrator.Type != OrchestratorTypeEnum_Allinone {
			return errors.New(i18nInstance.Config.Err_config_orchestrator_type_valid)
		}
	}

	// 格式对应的版本
	if c.Orchestrator.Version == "" {
		msg := fmt.Sprintf(i18nInstance.Config.Err_config_orchestrator_version_none, c.Orchestrator.Type)
		return errors.New(msg)
	}

	// service name 不能为空
	if c.Workspace.DevContainer.ServiceName == "" {
		return errors.New(i18nInstance.Config.Err_config_devcontainer_servicename_none)

	} else {

		if len(c.Workspace.Servcies) > 0 {
			hasService := false
			for serviceName := range c.Workspace.Servcies {
				if serviceName == c.Workspace.DevContainer.ServiceName {
					hasService = true
					break
				}
			}
			if !hasService {
				msg := fmt.Sprintf(i18nInstance.Config.Err_config_devcontainer_services_not_exit, c.Workspace.DevContainer.ServiceName)
				return errors.New(msg)
			}
		}

		//TODO 如果是关联了docker-compose 文件
	}

	// web ide的类型不能为空
	if c.Workspace.DevContainer.IdeType == "" {
		return errors.New(i18nInstance.Config.Err_config_devcontainer_idetype_none)

	} else {
		switch c.Workspace.DevContainer.IdeType {
		case IdeTypeEnum_JbProjector, IdeTypeEnum_Opensumi, IdeTypeEnum_Theia, IdeTypeEnum_VsCode, IdeTypeEnum_SDKOnly:
			break
		default:
			return errors.New(i18nInstance.Config.Err_config_devcontainer_idetype_valid)
		}
	}

	// ports 中的端口 & 描述不能重复
	if len(c.Workspace.DevContainer.Ports) > 0 {
		var ports []int
		var labels []string
		for label, port := range c.Workspace.DevContainer.Ports {

			if common.Contains4Int(ports, port) {
				msg := fmt.Sprintf(i18nInstance.Config.Err_config_devcontainer_ports_port_reqeat, port)
				return errors.New(msg)
			} else {
				ports = append(ports, port)
			}

			if common.Contains(labels, label) {
				msg := fmt.Sprintf(i18nInstance.Config.Err_config_devcontainer_ports_label_reqeat, label)
				return errors.New(msg)
			} else {
				labels = append(labels, label)
			}

		}
	}

	// 定义了ports时，必services中有且仅有一个
	if c.Orchestrator.Type == OrchestratorTypeEnum_Compose {
		for label, port := range c.Workspace.DevContainer.Ports {
			count := 0
			for _, service := range c.Workspace.Servcies {
				for _, portStr := range service.Ports {
					array := strings.Split(portStr, ":")
					if array[0] == strconv.Itoa(port) {
						count++
					}
				}
			}

			if count == 0 {
				return fmt.Errorf("没有找到 %v:%v 的端口绑定信息", label, port)
			} else if count > 1 {
				return fmt.Errorf("%v:%v 被多个service重复绑定", label, port)
			}
		}
	}

	return nil
}

func (c SmartIdeConfig) IsNil() bool {
	return c.Workspace.DevContainer.ServiceName == "" ||
		c.Workspace.DevContainer.IdeType == "" ||
		c.Orchestrator.Type == ""
}

func (c *SmartIdeConfig) IsNotNil() bool {
	return !c.IsNil()
}
