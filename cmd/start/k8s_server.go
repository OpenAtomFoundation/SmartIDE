/*
 * @Date: 2022-05-31 09:36:33
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-05-31 10:17:54
 * @FilePath: /smartide-cli/cmd/start/k8s_server.go
 */

package start

import (
	"io/ioutil"

	smartideServer "github.com/leansoftX/smartide-cli/cmd/server"
	"github.com/leansoftX/smartide-cli/internal/biz/config"
	"github.com/leansoftX/smartide-cli/internal/biz/workspace"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/leansoftX/smartide-cli/pkg/kubectl"
	"github.com/spf13/cobra"
)

func ExecuteK8sServerStartCmd(cmd *cobra.Command, workspaceInfo workspace.WorkspaceInfo, yamlExecuteFun func(yamlConfig config.SmartIdeConfig)) error {
	// 错误反馈
	serverFeedback := func(err error) {
		if !(workspaceInfo.CliRunningEnv != workspace.CliRunningEvnEnum_Server) {
			return
		}
		if err != nil {
			smartideServer.Feedback_Finish(smartideServer.FeedbackCommandEnum_Start, cmd, false, nil, workspace.WorkspaceInfo{}, err.Error(), "")
		}
	}

	// 下载.kube/config文件到本地
	err := ioutil.WriteFile("~/.kube/config", []byte(workspaceInfo.K8sInfo.KubeConfigContent), 0777)
	if err != nil {
		serverFeedback(err)
	}

	// 执行命令
	err = ExecuteK8sStartCmd(workspaceInfo, yamlExecuteFun)
	if err != nil {
		serverFeedback(err)
	}

	//9. 反馈给smartide server
	common.SmartIDELog.Info("feedback...")
	k8sUtil, err := kubectl.NewK8sUtil(workspaceInfo.K8sInfo.KubeConfigFilePath, workspaceInfo.K8sInfo.Context, workspaceInfo.K8sInfo.Namespace)
	if err != nil {
		return err
	}
	pod, _, _ := getDevContainerPod(*k8sUtil, workspaceInfo.K8sInfo.TempK8sConfig)
	containerWebIDEPort := workspaceInfo.ConfigYaml.GetContainerWebIDEPort()
	err = smartideServer.Feedback_Finish(smartideServer.FeedbackCommandEnum_Start, cmd, true, containerWebIDEPort, workspaceInfo, "", pod.Name)
	if err != nil {
		return err
	}

	return err
}
