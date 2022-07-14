/*
 * @Date: 2022-05-31 09:36:33
 * @LastEditors: kenan
 * @LastEditTime: 2022-06-21 15:14:58
 * @FilePath: /smartide-cli/cmd/start/k8s_sws_serverEnv.go
 */

package start

import (
	smartideServer "github.com/leansoftX/smartide-cli/cmd/server"
	"github.com/leansoftX/smartide-cli/internal/biz/config"
	"github.com/leansoftX/smartide-cli/internal/biz/workspace"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/leansoftX/smartide-cli/pkg/kubectl"
	"github.com/spf13/cobra"
)

func ExecuteK8sServerStartCmd(cmd *cobra.Command, k8sUtil kubectl.KubernetesUtil,
	workspaceInfo workspace.WorkspaceInfo,
	yamlExecuteFun func(yamlConfig config.SmartIdeConfig)) error {
	// 错误反馈
	serverFeedback := func(err error) {
		if workspaceInfo.CliRunningEnv != workspace.CliRunningEvnEnum_Server {
			return
		}
		if err != nil {
			smartideServer.Feedback_Finish(smartideServer.FeedbackCommandEnum_Start, cmd, false, nil, workspace.WorkspaceInfo{}, err.Error(), "")
		}
		common.CheckError(err)
	}

	//1. 下载.kube/config文件到本地
	err := k8sUtil.CreateKubeConfig(workspaceInfo.K8sInfo.KubeConfigContent)
	serverFeedback(err)

	// 工作区
	workspaceInfo_, err := ExecuteK8sStartCmd(cmd, k8sUtil, workspaceInfo, yamlExecuteFun)
	serverFeedback(err)

	workspaceInfo = *workspaceInfo_

	//9. 反馈给smartide server
	common.SmartIDELog.Info("feedback...")
	pod, _, _ := GetDevContainerPod(k8sUtil, workspaceInfo.K8sInfo.TempK8sConfig)
	containerWebIDEPort := workspaceInfo.ConfigYaml.GetContainerWebIDEPort()
	err = smartideServer.Feedback_Finish(smartideServer.FeedbackCommandEnum_Start, cmd, true, containerWebIDEPort, workspaceInfo, "", pod.Name)
	serverFeedback(err)

	return err
}
