/*
 * @Date: 2022-06-07 09:26:55
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-08-01 10:43:59
 * @FilePath: /cli/cmd/start/k8s_sws_clientEnv.go
 */

package start

import (
	"errors"
	"fmt"
	"strings"

	smartideServer "github.com/leansoftX/smartide-cli/cmd/server"
	"github.com/leansoftX/smartide-cli/internal/biz/config"
	"github.com/leansoftX/smartide-cli/internal/biz/workspace"
	"github.com/leansoftX/smartide-cli/internal/model"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/leansoftX/smartide-cli/pkg/kubectl"
)

// 在本地start 远程服务器上的k8s工作区
func ExecuteServerK8sStartByClientEnvCmd(workspaceInfo workspace.WorkspaceInfo,
	yamlExecuteFun func(yamlConfig config.SmartIdeConfig)) error {

	//0. 验证
	common.SmartIDELog.Info(i18nInstance.VmStart.Info_starting)
	// 检查工作区的状态
	if workspaceInfo.ServerWorkSpace.Status != model.WorkspaceStatusEnum_Start {
		if workspaceInfo.ServerWorkSpace.Status == model.WorkspaceStatusEnum_Pending || workspaceInfo.ServerWorkSpace.Status == model.WorkspaceStatusEnum_Init {
			return errors.New("当前工作区正在启动中，请等待！")

		} else if workspaceInfo.ServerWorkSpace.Status == model.WorkspaceStatusEnum_Stop {
			return errors.New("当前工作区已停止！")

		} else {
			return errors.New("当前工作区运行异常！")

		}
	}

	//1. 创建k8sUtil 对象
	k8sUtil, err := kubectl.NewK8sUtil(workspaceInfo.K8sInfo.KubeConfigFilePath,
		workspaceInfo.K8sInfo.Context,
		workspaceInfo.K8sInfo.Namespace)
	if err != nil {
		return err
	}
	err = k8sUtil.Check()
	if err != nil {
		return err
	}

	//2. 端口转发，依然需要检查对应的端口是否占用
	common.SmartIDELog.Info("端口转发...")
	//2.1. 端口转发，并记录到extend
	//_, _, err = GetDevContainerPod(*k8sUtil, workspaceInfo.K8sInfo.TempK8sConfig)
	if err != nil {
		return err
	}
	//2.2. func
	function1 := func(k8sServiceName string, availableClientPort, hostOriginPort, index int) {
		if availableClientPort != hostOriginPort {
			common.SmartIDELog.InfoF("[端口转发] localhost:%v( %v 被占用) -> Service: %v  ", availableClientPort, hostOriginPort, hostOriginPort)
		} else {
			common.SmartIDELog.InfoF("[端口转发] localhost:%v -> Service: %v  ", availableClientPort, hostOriginPort)
		}

		portMapInfo := workspaceInfo.Extend.Ports[index]
		portMapInfo.OldClientPort = portMapInfo.ClientPort
		portMapInfo.ClientPort = availableClientPort
		workspaceInfo.Extend.Ports[index] = portMapInfo

		forwardCommand := fmt.Sprintf("port-forward svc/%v %v:%v --address 0.0.0.0 ",
			k8sServiceName, availableClientPort, hostOriginPort)
		output, err := k8sUtil.ExecKubectlCommandCombined(forwardCommand, "")
		common.SmartIDELog.Debug(output)
		for err != nil || strings.Contains(output, "error forwarding port") {
			if err != nil {
				common.SmartIDELog.ImportanceWithError(err)
			}
			output, err = k8sUtil.ExecKubectlCommandCombined(forwardCommand, "")
			common.SmartIDELog.Debug(output)
		}

	}
	//2.3. 遍历端口
	for index, portMapInfo := range workspaceInfo.Extend.Ports {
		unusedClientPort, err := common.CheckAndGetAvailableLocalPort(portMapInfo.ClientPort, 100)
		common.SmartIDELog.Error(err)

		go function1(portMapInfo.ServiceName, unusedClientPort, portMapInfo.CurrentHostPort, index)

	}

	//9. 更新server端的extend字段
	currentAuth, err := workspace.GetCurrentUser()
	if err != nil {
		return err
	}
	err = smartideServer.FeeadbackExtend(currentAuth, workspaceInfo)
	if err != nil {
		common.SmartIDELog.ImportanceWithError(err)
	}
	common.SmartIDELog.Info("本地端口绑定信息 更新完成！")

	return nil
}
