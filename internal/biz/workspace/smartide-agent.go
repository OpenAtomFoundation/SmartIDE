/*
 * @Author: kenan
 * @Date: 2022-04-15 13:38:10
 * @LastEditors: kenan
 * @LastEditTime: 2022-04-20 20:29:10
 * @FilePath: /smartide-cli/internal/biz/workspace/smartide-agent.go
 * @Description:
 *
 * Copyright (c) 2022 by kenanlu@leansoftx.com, All Rights Reserved.
 */
package workspace

import (
	"fmt"
	"strings"

	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/spf13/cobra"
)

const (
	Flags_ServerHost = "serverhost"
)

func InstallSmartideAgent(sshRemote common.SSHRemote, containerId string, cmd *cobra.Command) {
	fflags := cmd.Flags()
	host, _ := fflags.GetString(Flags_ServerHost)
	agentInstallCmd := ""
	// 测试环境
	if i := strings.Index(host, "test"); i > -1 {
		agentInstallCmd = fmt.Sprintf("docker exec -d %s  /bin/sh -c \"curl -o /home/smartide/smartide-agent -OL 'https://smartidedl.blob.core.chinacloudapi.cn/smartide-agent/lasted/test/smartide-agent-linux' && curl  -o /home/smartide/.smartide-agent.yaml  -OL 'https://smartidedl.blob.core.chinacloudapi.cn/smartide-agent/lasted/test/.smartide-agent.yaml' && sudo chmod +x /home/smartide/smartide-agent  && cd /home/smartide;./smartide-agent\"", containerId)
	} else {
		agentInstallCmd = fmt.Sprintf("docker exec -d %s  /bin/sh -c \"curl -o /home/smartide/smartide-agent -OL 'https://smartidedl.blob.core.chinacloudapi.cn/smartide-agent/lasted/prod/smartide-agent-linux' && curl  -o /home/smartide/.smartide-agent.yaml  -OL 'https://smartidedl.blob.core.chinacloudapi.cn/smartide-agent/lasted/prod/.smartide-agent.yaml' && sudo chmod +x /home/smartide/smartide-agent  && cd /home/smartide;./smartide-agent\"", containerId)
	}
	sshRemote.ExecSSHCommandRealTime(agentInstallCmd)

}
