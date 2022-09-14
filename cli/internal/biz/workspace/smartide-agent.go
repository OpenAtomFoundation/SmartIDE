/*
 * @Author: kenan
 * @Date: 2022-04-15 13:38:10
 * @LastEditors: kenan
 * @LastEditTime: 2022-09-14 10:07:02
 * @FilePath: /cli/internal/biz/workspace/smartide-agent.go
 * @Description:
 *
 * Copyright (c) 2022 by kenanlu@leansoftx.com, All Rights Reserved.
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

func InstallSmartideAgent(sshRemote common.SSHRemote, containerId string, cmd *cobra.Command, wid uint) {
	fflags := cmd.Flags()
	host, _ := fflags.GetString(Flags_ServerHost)
	token, _ := fflags.GetString(Flags_ServerToken)
	ownerguid, _ := fflags.GetString(Flags_ServerOwnerGuid)

	agentInstallCmd := fmt.Sprintf("docker exec -d %s  /bin/sh -c \"curl -o /smartide-agent -OL 'https://smartidedl.blob.core.chinacloudapi.cn/smartide-agent/latest/smartide-agent-linux' && sudo chmod +x /smartide-agent  && cd /;./smartide-agent --serverhost %s --servertoken %s --serverownerguid %s --workspaceId %v \"", containerId, host, token, ownerguid, wid)

	localFilePath := filepath.Join("/usr/local/bin/smartide-agent")
	if common.IsExist(localFilePath) {
		//1. 将tkn 中的agent 拷贝到远程主机

		homeDir, _ := sshRemote.GetRemoteHome()
		dstPath := filepath.Join(homeDir, "smartide-agent")
		if err := sshRemote.CopyFile(localFilePath, dstPath); err != nil {
			common.SmartIDELog.Error(fmt.Sprintf("smartide-agent install fail:%s", err.Error()))

		}

		// 2.将远程主机文件拷贝的容器内
		agentInstallCmd = fmt.Sprintf(" docker exec -d %s  /bin/sh -c \"sudo cp -rf /home/smartide/smartide-agent /smartide-agent && sudo chmod +x /smartide-agent  && cd /;./smartide-agent --serverhost %s --servertoken %s --serverownerguid %s --workspaceId %v\"", containerId, host, token, ownerguid, wid)
	}

	sshRemote.ExecSSHCommandRealTime(agentInstallCmd)

}
