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

package config

import (
	"fmt"
	"path/filepath"

	"github.com/leansoftX/smartide-cli/pkg/docker/compose"
)

// 获取webterminal的service内容
func GetWebTerminalCompose(contaninerName, pwd string) compose.Service {
	projectName := filepath.Base(pwd)
	var webTerminalService compose.Service
	webTerminalService.ContainerName = contaninerName
	webTerminalService.Image = fmt.Sprintf("%v/smartide/smartide-webterminal", GlobalSmartIdeConfig.ImagesRegistry)
	webTerminalService.Restart = "always"
	webTerminalService.AppendPort("6860:6860")
	webTerminalService.Volumes = append(webTerminalService.Volumes, "/var/run/docker.sock:/var/run/docker.sock")
	webTerminalService.Networks = append(webTerminalService.Networks, "smartide-network")
	webTerminalService.Environment = map[string]string{
		"LOCAL_USER_GID":        "1000",
		"LOCAL_USER_PASSWORD":   "root123",
		"LOCAL_USER_UID":        "1000",
		"ROOT_PASSWORD":         "root123",
		"TERMINAL_USER":         "smartide",
		"TERMINAL_DOCKER_LABEL": fmt.Sprintf("com.docker.compose.project=%v", projectName), // 过滤当前工作区
	}
	return webTerminalService
}

// 增加addon webterminal的配置信息
func (c *SmartIdeConfig) AddonWebTerminal(webterminalName, pwd string) {
	if len(c.Workspace.Networks) == 0 {
		c.Workspace.Networks = make(map[string]compose.Network)
		c.Workspace.Networks["smartide-network"] = compose.Network{External: true}
	} else {
		if _, ok := c.Workspace.Networks["smartide-network"]; !ok {
			c.Workspace.Networks["smartide-network"] = compose.Network{External: true}
		}
	}
	if len(c.Workspace.Servcies) == 0 {
		c.Workspace.Servcies = make(map[string]compose.Service)
	}
	webterminServiceName := fmt.Sprintf("%v_smartide-webterminal", webterminalName)
	c.Workspace.Servcies[webterminServiceName] = GetWebTerminalCompose(webterminServiceName, pwd)
}
