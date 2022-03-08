/*
 * @Author: jason chen (jasonchen@leansoftx.com, http://smallidea.cnblogs.com)
 * @Description:
 * @Date: 2021-11
 * @LastEditors:
 * @LastEditTime:
 */
package config

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/leansoftX/smartide-cli/pkg/common"
)

// 验证配置文件格式是否正确
func (c SmartIdeConfig) Valid() error {
	// 格式不能为空
	if c.Orchestrator.Type == "" {
		return errors.New(i18nInstance.Config.Err_config_orchestrator_type_none)

	} else {
		if c.Orchestrator.Type != "docker-compose" {
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
		if c.Workspace.DevContainer.IdeType != "vscode" && c.Workspace.DevContainer.IdeType != "theia" && c.Workspace.DevContainer.IdeType != "jb-projector" {
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

	return nil
}

//
func (c SmartIdeConfig) IsNil() bool {
	return c.Workspace.DevContainer.ServiceName == "" ||
		c.Workspace.DevContainer.IdeType == "" ||
		c.Orchestrator.Type == ""
}

//
func (c *SmartIdeConfig) IsNotNil() bool {
	return !c.IsNil()
}
