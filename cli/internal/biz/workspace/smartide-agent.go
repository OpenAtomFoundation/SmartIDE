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
	"path/filepath"

	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/spf13/cobra"
)

const (
	Flags_ServerHost      = "serverhost"
	Flags_ServerToken     = "servertoken"
	Flags_ServerOwnerGuid = "serverownerguid"
)

func InstallSmartideAgent(sshRemote common.SSHRemote) {

	localFilePath := filepath.Join("/usr/local/bin/smartide-agent")
	if common.IsExist(localFilePath) {
		//1. 将tkn 中的agent 拷贝到远程主机

		homeDir, _ := sshRemote.GetRemoteHome()
		dstPath := filepath.Join(homeDir, "smartide-agent")
		sshRemote.ExecSSHCommandRealTime(fmt.Sprintf("[[ -d  %s ]] && sudo rm -rf %s", dstPath, dstPath))
		if err := sshRemote.CopyFile(localFilePath, dstPath); err != nil {
			common.SmartIDELog.Error(fmt.Sprintf("smartide-agent install fail:%s", err.Error()))

		}
	}

}

func StartSmartideAgent(sshRemote common.SSHRemote, containerId string, cmd *cobra.Command, wid uint) {
	fflags := cmd.Flags()
	host, _ := fflags.GetString(Flags_ServerHost)
	token, _ := fflags.GetString(Flags_ServerToken)
	ownerguid, _ := fflags.GetString(Flags_ServerOwnerGuid)
	if host != "" && token != "" {
		agentInstallCmd := fmt.Sprintf("docker exec -d %s  /bin/sh -c \"curl -o /smartide-agent -OL 'https://smartidedl.blob.core.chinacloudapi.cn/smartide-agent/latest/smartide-agent-linux' && sudo chmod +x /smartide-agent  && cd /;./smartide-agent --serverhost %s --servertoken %s --serverownerguid %s --workspaceId %v \"", containerId, host, token, ownerguid, wid)

		localFilePath := filepath.Join("/usr/local/bin/smartide-agent")
		if common.IsExist(localFilePath) {
			// 2.将远程主机文件拷贝的容器内
			agentInstallCmd = fmt.Sprintf(" docker exec -d %s  /bin/sh -c \"sudo cp -rf /home/smartide/smartide-agent /smartide-agent && sudo chmod +x /smartide-agent  && cd /;./smartide-agent --serverhost %s --servertoken %s --serverownerguid %s --workspaceId %v\"", containerId, host, token, ownerguid, wid)
		}

		sshRemote.ExecSSHCommandRealTime(agentInstallCmd)
	}

}
