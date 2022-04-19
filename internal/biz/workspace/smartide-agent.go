/*
 * @Author: kenan
 * @Date: 2022-04-15 13:38:10
 * @LastEditors: kenan
 * @LastEditTime: 2022-04-19 10:57:02
 * @FilePath: /smartide-cli/internal/biz/workspace/smartide-agent.go
 * @Description:
 *
 * Copyright (c) 2022 by kenanlu@leansoftx.com, All Rights Reserved.
 */
package workspace

import (
	"fmt"

	"github.com/leansoftX/smartide-cli/pkg/common"
)

func InstallSmartideAgent(sshRemote common.SSHRemote, containerId string) {
	sshRemote.ExecSSHCommandRealTime(fmt.Sprintf("docker exec -d %s  /bin/sh -c \"curl -o /home/smartide/smartide-agent -OL 'https://smartidedl.blob.core.chinacloudapi.cn/smartide-agent/lasted/smartide-agent-linux' &&curl  -o /home/smartide/.smartide-agent.yaml  -OL 'https://smartidedl.blob.core.chinacloudapi.cn/smartide-agent/lasted/.smartide-agent.yaml' &&sudo chmod +x /home/smartide/smartide-agent  && /home/smartide/smartide-agent\"", containerId))

}
