/*
 * @Author: kenan
 * @Date: 2022-04-15 13:38:10
 * @LastEditors: kenan
 * @LastEditTime: 2022-05-16 21:26:57
 * @FilePath: /smartide-cli/internal/biz/workspace/smartide-agent.go
 * @Description:
 *
 * Copyright (c) 2022 by kenanlu@leansoftx.com, All Rights Reserved.
 */
package workspace

import (
	"fmt"

	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/spf13/cobra"
)

const (
	Flags_ServerHost      = "serverhost"
	Flags_ServerToken     = "servertoken"
	Flags_ServerOwnerGuid = "serverownerguid"
)

func InstallSmartideAgent(sshRemote common.SSHRemote, containerId string, cmd *cobra.Command) {
	fflags := cmd.Flags()
	host, _ := fflags.GetString(Flags_ServerHost)
	token, _ := fflags.GetString(Flags_ServerToken)
	ownerguid, _ := fflags.GetString(Flags_ServerOwnerGuid)

	agentInstallCmd := fmt.Sprintf("docker exec -d %s  /bin/sh -c \"curl -o /home/smartide/smartide-agent -OL 'https://smartidedl.blob.core.chinacloudapi.cn/smartide-agent/latest/smartide-agent-linux' && sudo chmod +x /home/smartide/smartide-agent  && cd /home/smartide;./smartide-agent --serverhost %s --servertoken %s --serverownerguid %s\"", containerId, host, token, ownerguid)

	sshRemote.ExecSSHCommandRealTime(agentInstallCmd)

}
