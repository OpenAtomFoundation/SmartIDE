/*
 * @Date: 2022-05-31 09:36:33
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-10-27 11:28:25
 * @FilePath: /cli/cmd/start/k8s_sws_serverEnv.go
 */

package start

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	smartideServer "github.com/leansoftX/smartide-cli/cmd/server"
	"github.com/leansoftX/smartide-cli/internal/biz/config"
	"github.com/leansoftX/smartide-cli/internal/biz/workspace"
	"github.com/leansoftX/smartide-cli/internal/model/response"

	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/leansoftX/smartide-cli/pkg/k8s"
	"github.com/spf13/cobra"
	coreV1 "k8s.io/api/core/v1"
)

// 在服务器上运行 k8s start
func ExecuteK8s_ServerWS_ServerEnv(cmd *cobra.Command, k8sUtil k8s.KubernetesUtil,
	workspaceInfo workspace.WorkspaceInfo,
	yamlExecuteFun func(yamlConfig config.SmartIdeConfig)) (workspace.WorkspaceInfo, error) {
	// 错误反馈
	serverFeedback := func(err error) {
		if workspaceInfo.CliRunningEnv != workspace.CliRunningEvnEnum_Server {
			return
		}
		if err != nil {
			smartideServer.Feedback_Finish(smartideServer.FeedbackCommandEnum_Start, cmd, false, nil, workspaceInfo, err.Error(), "")
			common.CheckError(err)
		}

	}

	// create namespace
	_, err := k8sUtil.ExecKubectlCommandCombined(" get namespace "+k8sUtil.Namespace, "")
	if _, isExitError := err.(*exec.ExitError); isExitError {
		common.SmartIDELog.Info("create namespace：" + k8sUtil.Namespace)

		labels := getK8sLabels(cmd, workspaceInfo)
		// namespace
		namespaceKind := coreV1.Namespace{}
		namespaceKind.Kind = "Namespace" // 必须要赋值，否则为空
		namespaceKind.APIVersion = "v1"  // 必须要赋值，否则为空
		namespaceKind.ObjectMeta.Name = k8sUtil.Namespace
		namespaceKind = k8s.AddLabels(namespaceKind, labels).(coreV1.Namespace)

		// 创建文件
		// home 目录的路径
		home, err := os.UserHomeDir()
		serverFeedback(err)
		workingRootDir := filepath.Join(home, ".ide", ".k8s") // 工作目录，repo 会clone到当前目录下
		gitRepoRootDirPath := filepath.Join(workingRootDir, common.GetRepoName(workspaceInfo.GitCloneRepoUrl))
		err = os.MkdirAll(gitRepoRootDirPath, os.ModePerm)
		serverFeedback(err)
		tempK8sNamespaceYamlAbsolutePath := filepath.Join(gitRepoRootDirPath, fmt.Sprintf("k8s_deployment_%v_temp_namespace.yaml", filepath.Base(gitRepoRootDirPath)))
		k8sYamlContent, err := config.ConvertK8sKindToString(namespaceKind)
		serverFeedback(err)
		err = os.WriteFile(tempK8sNamespaceYamlAbsolutePath, []byte(k8sYamlContent), 0777)
		serverFeedback(err)

		// apply
		err = k8sUtil.ExecKubectlCommandRealtime(fmt.Sprintf("apply -f %v", tempK8sNamespaceYamlAbsolutePath), "", false)
		serverFeedback(err)

		// set value
		workspaceInfo.K8sInfo.Namespace = k8sUtil.Namespace
	}
	// 设置为pending状态
	smartideServer.Feedback_Pending(smartideServer.FeedbackCommandEnum_Start, response.WorkspaceStatusEnum_Pending_NsCreated, cmd, workspaceInfo, "")

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

	return workspaceInfo, err
}
