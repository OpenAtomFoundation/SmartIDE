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

package new

import (
	"github.com/leansoftX/smartide-cli/cmd/start"
	"github.com/leansoftX/smartide-cli/internal/biz/config"
	"github.com/leansoftX/smartide-cli/pkg/k8s"

	// templateModel "github.com/leansoftX/smartide-cli/internal/biz/template/model"
	"github.com/leansoftX/smartide-cli/internal/biz/workspace"
	"github.com/spf13/cobra"
)

func K8sNew_Server(cmd *cobra.Command, args []string,
	k8sUtil *k8s.KubernetesUtil,
	workspaceInfo workspace.WorkspaceInfo,
	//selectedTemplate templateModel.SelectedTemplateTypeBo,
	yamlExecuteFun func(yamlConfig config.SmartIdeK8SConfig, workspaceInfo workspace.WorkspaceInfo, cmdtype, userguid, workspaceid string)) {

	//0. 错误反馈
	serverFeedback := preRun(cmd, workspaceInfo, workspace.ActionEnum_Workspace_Start, k8sUtil)

	_, err := start.ExecuteK8s_ServerWS_ServerEnv(cmd, *k8sUtil, workspaceInfo, yamlExecuteFun)
	serverFeedback(err)

}
