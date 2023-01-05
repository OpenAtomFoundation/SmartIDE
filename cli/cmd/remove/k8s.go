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

package remove

import (
	"fmt"
	"os"
	"strings"

	"github.com/leansoftX/smartide-cli/internal/biz/workspace"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/leansoftX/smartide-cli/pkg/k8s"
	"github.com/spf13/cobra"
)

// 删除k8s资源
func RemoveK8s(k8sUtil k8s.KubernetesUtil, workspaceInfo workspace.WorkspaceInfo) error {

	// 移除k8s资源
	common.SmartIDELog.Info("移除k8s资源...")

	output, err := k8sUtil.ExecKubectlCommandCombined("delete namespace --force "+workspaceInfo.K8sInfo.Namespace, "")
	common.SmartIDELog.Debug(output)
	notFoundMsg := "Error from server (NotFound)"
	if strings.Contains(output, notFoundMsg) {
		msg := fmt.Sprintf("%v: namespaces \"%v\" not found", notFoundMsg, workspaceInfo.K8sInfo.Namespace)
		common.SmartIDELog.Importance(msg)
		return nil
	}
	if err != nil {
		return err
	}

	// 删除本地.ide目录下的文件
	common.SmartIDELog.Info("移除本地缓存文件...")
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	repoName := common.GetRepoName(workspaceInfo.GitCloneRepoUrl)
	if repoName != "" {
		filePath := common.PathJoin(home, ".ide", repoName)
		os.RemoveAll(filePath)
	}

	//remove config note from .ssh/config file
	workspaceInfo.RemoveSSHConfig()

	return nil
}

// 删除远程工作区对应的k8s
func RemoveServerK8s(k8sUtil k8s.KubernetesUtil,
	cmd *cobra.Command, workspaceInfo workspace.WorkspaceInfo,
	isRemoveAllComposeImages bool, isForce bool,
	podName string) error {

	//2. 移除资源
	err := RemoveK8s(k8sUtil, workspaceInfo)
	if err != nil {
		return err
	}

	/* 	//9. 反馈给smartide server
	   	common.SmartIDELog.Info("feedback...")
	   	containerWebIDEPort := workspaceInfo.ConfigYaml.GetContainerWebIDEPort()
	   	err = smartideServer.Feedback_Finish(smartideServer.FeedbackCommandEnum_Remove, cmd, true, containerWebIDEPort, workspaceInfo, "", podName)
	   	if err != nil {
	   		return err
	   	} */

	return nil
}
