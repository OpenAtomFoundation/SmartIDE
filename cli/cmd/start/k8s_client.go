/*
SmartIDE - Dev Containers
Copyright (C) 2023 leansoftX.com

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package start

import (
	"fmt"
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
func ExecuteK8s_LocalWS_LocalEnv(cmd *cobra.Command, k8sUtil k8s.KubernetesUtil,
	workspaceInfo workspace.WorkspaceInfo,
	yamlExecuteFun func(yamlConfig config.SmartIdeK8SConfig, workspaceInfo workspace.WorkspaceInfo, cmdtype, userguid, workspaceid string)) (workspace.WorkspaceInfo, error) {

	needStore := false
	if workspaceInfo.ID == "" {
		needStore = true
	}

	//1. create namespace
	_, err := k8sUtil.ExecKubectlCommandCombined(" get namespace "+k8sUtil.Namespace, "")
	if _, isExitError := err.(*exec.ExitError); isExitError { // 如果不存在，才需要创建
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
			return workspaceInfo, err
		}
		workingRootDir := filepath.Join(home, ".ide", ".k8s") // 工作目录，repo 会clone到当前目录下
		workspaceInfo.WorkingDirectoryPath = workingRootDir   // 赋值避免出错
		gitRepoRootDirPath := filepath.Join(workingRootDir, common.GetRepoName(workspaceInfo.GitCloneRepoUrl))
		err = os.MkdirAll(gitRepoRootDirPath, os.ModePerm)
		if err != nil {
			return workspaceInfo, err
		}
		tempK8sNamespaceYamlAbsolutePath := filepath.Join(gitRepoRootDirPath, fmt.Sprintf("k8s_deployment_%v_temp_namespace.yaml", filepath.Base(gitRepoRootDirPath)))
		k8sYamlContent, err := config.ConvertK8sKindToString(namespaceKind)
		if err != nil {
			return workspaceInfo, err
		}
		err = os.WriteFile(tempK8sNamespaceYamlAbsolutePath, []byte(k8sYamlContent), 0777)
		if err != nil {
			return workspaceInfo, err
		}

		// apply
		err = k8sUtil.ExecKubectlCommandRealtime(fmt.Sprintf("apply -f %v", tempK8sNamespaceYamlAbsolutePath), "", false)
		if err != nil {
			return workspaceInfo, err
		}

		// set value
		workspaceInfo.K8sInfo.Namespace = k8sUtil.Namespace
	}

	//2. store
	if needStore {
		common.SmartIDELog.Info("workspace store")
		workspaceId, err := dal.InsertOrUpdateWorkspace(workspaceInfo)
		if err != nil {
			return workspaceInfo, err
		}
		if workspaceInfo.ID == "" {
			common.SmartIDELog.Info(fmt.Sprintf("workspace id: %v", workspaceId))
			workspaceInfo.ID = fmt.Sprint(workspaceId)
		}
	}

	// 工作区
	_, err = ExecuteK8sStartCmd(cmd, k8sUtil, workspaceInfo, yamlExecuteFun)
	if err != nil {
		return workspaceInfo, err
	}

	return workspaceInfo, nil
}
