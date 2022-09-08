/*
 * @Date: 2022-09-05 11:27:09
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-09-08 10:19:57
 * @FilePath: /cli/cmd/start/k8s_client.go
 */

package start

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/leansoftX/smartide-cli/internal/biz/config"
	"github.com/leansoftX/smartide-cli/internal/biz/workspace"
	"github.com/leansoftX/smartide-cli/internal/dal"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/leansoftX/smartide-cli/pkg/k8s"
	"github.com/spf13/cobra"

	coreV1 "k8s.io/api/core/v1"
)

// 在本地启动k8s工作区
func ExecuteK8sClientStartCmd(cmd *cobra.Command, k8sUtil k8s.KubernetesUtil,
	workspaceInfo workspace.WorkspaceInfo,
	yamlExecuteFun func(yamlConfig config.SmartIdeConfig)) error {

	needStore := false
	if workspaceInfo.ID == "" {
		needStore = true
	}

	// create namespace
	// namespace 是否存在
	_, err := k8sUtil.ExecKubectlCommandCombined(" get namespace "+k8sUtil.Namespace, "")
	if _, isExitError := err.(*exec.ExitError); isExitError {
		needStore = true
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
		if err != nil {
			return err
		}
		workingRootDir := filepath.Join(home, ".ide", ".k8s") // 工作目录，repo 会clone到当前目录下
		gitRepoRootDirPath := filepath.Join(workingRootDir, common.GetRepoName(workspaceInfo.GitCloneRepoUrl))
		err = os.MkdirAll(gitRepoRootDirPath, os.ModePerm)
		if err != nil {
			return err
		}
		tempK8sNamespaceYamlAbsolutePath := filepath.Join(gitRepoRootDirPath, fmt.Sprintf("k8s_deployment_%v_temp_namespace.yaml", filepath.Base(gitRepoRootDirPath)))
		k8sYamlContent, err := config.ConvertK8sKindToString(namespaceKind)
		if err != nil {
			return err
		}
		err = ioutil.WriteFile(tempK8sNamespaceYamlAbsolutePath, []byte(k8sYamlContent), 0777)
		if err != nil {
			return err
		}

		// apply
		err = k8sUtil.ExecKubectlCommandRealtime(fmt.Sprintf("apply -f %v", tempK8sNamespaceYamlAbsolutePath), "", false)
		if err != nil {
			return err
		}

		// set value
		workspaceInfo.K8sInfo.Namespace = k8sUtil.Namespace
	}

	// store
	if needStore {
		common.SmartIDELog.Info("workspace store")
		workspaceId, err := dal.InsertOrUpdateWorkspace(workspaceInfo)
		if err != nil {
			return err
		}
		if workspaceInfo.ID == "" {
			common.SmartIDELog.Info(fmt.Sprintf("workspace id: %v", workspaceId))
			workspaceInfo.ID = fmt.Sprint(workspaceId)
		}
	}

	// 工作区
	_, err = ExecuteK8sStartCmd(cmd, k8sUtil, workspaceInfo, yamlExecuteFun)
	if err != nil {
		return err
	}

	return nil
}
