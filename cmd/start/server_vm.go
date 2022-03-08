/*
 * @Author: kenan
 * @Date: 2022-02-16 17:44:45
 * @LastEditors: kenan
 * @LastEditTime: 2022-02-21 15:37:20
 * @FilePath: /smartide-cli/cmd/start/server_vm.go
 * @Description:
 *
 * Copyright (c) 2022 by kenanlu@leansoftx.com, All Rights Reserved.
 */
package start

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	smartideServer "github.com/leansoftX/smartide-cli/cmd/server"
	"github.com/leansoftX/smartide-cli/internal/biz/config"
	"github.com/leansoftX/smartide-cli/internal/biz/workspace"
	"github.com/leansoftX/smartide-cli/internal/model"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/leansoftX/smartide-cli/pkg/tunnel"
)

// 远程服务器执行 start 命令
func ExecuteServerVmStartCmd(workspaceInfo workspace.WorkspaceInfo, yamlExecuteFun func(yamlConfig config.SmartIdeConfig)) {
	common.SmartIDELog.Info(i18nInstance.VmStart.Info_starting)

	// 检查工作区的状态
	if workspaceInfo.ServerWorkSpace.Status != model.WorkspaceStatusEnum_Start {
		if workspaceInfo.ServerWorkSpace.Status == model.WorkspaceStatusEnum_Pending || workspaceInfo.ServerWorkSpace.Status == model.WorkspaceStatusEnum_Init {
			common.SmartIDELog.Error("当前工作区正在启动中，请等待！")
		} else if workspaceInfo.ServerWorkSpace.Status == model.WorkspaceStatusEnum_Stop {
			common.SmartIDELog.Error("当前工作区已停止！")
		} else {
			common.SmartIDELog.Error("当前工作区运行异常！")
		}
	}

	//0. 连接到远程主机
	msg := fmt.Sprintf(" %v@%v:%v ...", workspaceInfo.Remote.UserName, workspaceInfo.Remote.Addr, workspaceInfo.Remote.SSHPort)
	common.SmartIDELog.Info(i18nInstance.VmStart.Info_connect_remote + msg)
	sshRemote, err := common.NewSSHRemote(workspaceInfo.Remote.Addr, workspaceInfo.Remote.SSHPort, workspaceInfo.Remote.UserName, workspaceInfo.Remote.Password)
	common.CheckError(err)

	//6. 当前主机绑定到远程端口
	common.SmartIDELog.Info(i18nInstance.VmStart.Info_tunnel_waiting) // log
	var addrMapping map[string]string = map[string]string{}
	var unusedLocalPort4IdeBindingPort int
	for i, pmi := range workspaceInfo.Extend.Ports {
		port := &workspaceInfo.Extend.Ports[i] // 指针，用于重新设定值
		if pmi.CurrentHostPort <= 0 {
			common.SmartIDELog.Importance(fmt.Sprintf("%v 绑定端口不正确 ", pmi.CurrentHostPort))
			continue
		}

		// 获取本地未使用的端口
		var unusedClientPort int
		unusedClientPort, err = common.CheckAndGetAvailableLocalPort(pmi.CurrentHostPort, 100)
		if err != nil {
			common.SmartIDELog.Importance(err.Error())
		}

		// 更新extend.ports的信息
		if port.OldClientPort == 0 { // 值为空的时候，直接赋值
			port.OldClientPort = unusedClientPort
		} else { // 值不为空的时候，使用之前的clientport字段赋值
			port.OldClientPort = port.ClientPort
		}
		port.ClientPort = unusedClientPort // 当前端口赋值
		unusedClientPortStr := strconv.Itoa(port.ClientPort)
		addrMapping["localhost:"+unusedClientPortStr] = "localhost:" + strconv.Itoa(pmi.CurrentHostPort)

		// 获取webide的本地端口
		if pmi.HostPortDesc != "" {
			unusedClientPortStr += fmt.Sprintf("(%v)", pmi.HostPortDesc)
			if strings.Contains(strings.ToLower(pmi.HostPortDesc), "webide") {
				unusedLocalPort4IdeBindingPort = workspaceInfo.Extend.Ports[i].ClientPort
			}
		}

		// 打印信息
		msg := fmt.Sprintf("localhost:%v -> %v:%v -> container:%v",
			unusedClientPortStr, workspaceInfo.Remote.Addr, pmi.OriginHostPort, pmi.ContainerPort)
		common.SmartIDELog.Info(msg)
	}

	//6.2. 执行绑定
	tunnel.TunnelMultiple(sshRemote.Connection, addrMapping)

	//8. 打开浏览器
	var checkUrl string
	//vscode启动时候默认打开文件夹处理
	common.SmartIDELog.Info(i18nInstance.VmStart.Info_warting_for_webide + fmt.Sprintf(`: %v`, unusedLocalPort4IdeBindingPort))
	switch strings.ToLower(workspaceInfo.ConfigYaml.Workspace.DevContainer.IdeType) {
	case "vscode":
		checkUrl = fmt.Sprintf("http://localhost:%v/?folder=vscode-remote://localhost:%v%v",
			unusedLocalPort4IdeBindingPort, unusedLocalPort4IdeBindingPort, workspaceInfo.GetContainerWorkingPathWithVolumes())
	case "jb-projector":
		checkUrl = fmt.Sprintf(`http://localhost:%v`, unusedLocalPort4IdeBindingPort)
	default:
		checkUrl = fmt.Sprintf(`http://localhost:%v`, unusedLocalPort4IdeBindingPort)
	}
	isUrlReady := false
	for !isUrlReady {
		resp, err := http.Get(checkUrl)
		if (err == nil) && (resp.StatusCode == 200) {
			isUrlReady = true
			//common.OpenBrowser(checkUrl) // 这里不用打开，从server中点击即可
			common.SmartIDELog.InfoF(i18nInstance.VmStart.Info_open_brower, checkUrl)
		} else {
			msg := fmt.Sprintf("%v 检测失败", checkUrl)
			common.SmartIDELog.Debug(msg)
		}
	}

	//9. 更新server端的extend字段
	currentAuth, err := workspace.GetCurrentUser()
	common.CheckError(err)
	err = smartideServer.FeeadbackExtend(currentAuth, workspaceInfo)
	if err != nil {
		common.SmartIDELog.Importance(err.Error())
	}
	common.SmartIDELog.Info("本地端口绑定信息 更新完成！")

}
