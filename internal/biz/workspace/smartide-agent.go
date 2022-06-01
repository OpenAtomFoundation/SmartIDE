/*
 * @Author: kenan
 * @Date: 2022-04-15 13:38:10
 * @LastEditors: kenan
 * @LastEditTime: 2022-06-01 16:27:13
 * @FilePath: /smartide-cli/internal/biz/workspace/smartide-agent.go
 * @Description:
 *
 * Copyright (c) 2022 by kenanlu@leansoftx.com, All Rights Reserved.
 */
package workspace

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/pkg/sftp"
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

	agentInstallCmd := fmt.Sprintf("docker exec -d %s  /bin/sh -c \"curl -o /smartide-agent -OL 'https://smartidedl.blob.core.chinacloudapi.cn/smartide-agent/latest/smartide-agent-linux' && sudo chmod +x /smartide-agent  && cd /;./smartide-agent --serverhost %s --servertoken %s --serverownerguid %s\"", containerId, host, token, ownerguid)

	localFilePath := filepath.Join("/usr/local/bin/smartide-agent")
	if common.IsExit(localFilePath) {
		//1. 将tkn 中的agent 拷贝到远程主机
		sftpClient, err := sftp.NewClient(sshRemote.Connection)
		if err != nil { //创建客户端
			fmt.Println("创建客户端sftpClient失败", err)
			return
		}
		srcFile, _ := os.Open(localFilePath)
		dstFile, _ := sftpClient.Create("~/smartide-agent")
		defer func() {
			_ = srcFile.Close()
			_ = dstFile.Close()
		}()

		buf := make([]byte, 1024)
		for {
			n, err := srcFile.Read(buf)
			if err != nil {
				if err != io.EOF {
					log.Fatalln("error occurred:", err)
				} else {
					break
				}
			}
			_, _ = dstFile.Write(buf[:n])
		}

		// 2.将远程主机文件拷贝的容器内
		agentInstallCmd = fmt.Sprintf("docker cp %s:%s /smartide-agent &docker exec -d %s  /bin/sh -c \"sudo chmod +x /smartide-agent  && cd /;./smartide-agent --serverhost %s --servertoken %s --serverownerguid %s\"", containerId, "~/smartide-agent", containerId, host, token, ownerguid)
	}

	sshRemote.ExecSSHCommandRealTime(agentInstallCmd)

}
