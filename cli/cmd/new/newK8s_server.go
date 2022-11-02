/*
 * @Date: 2022-10-27 11:21:54
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-11-02 08:51:49
 * @FilePath: /cli/cmd/new/newK8s_server.go
 */
/*
 * @Date: 2022-10-27 09:35:51
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-10-27 11:21:23
 * @FilePath: /cli/cmd/new/newK8s.go
 */
/*
 * @Date: 2022-04-20 10:46:40
 * @LastEditors: kenan
 * @LastEditTime: 2022-10-20 10:10:08
 * @FilePath: /cli/cmd/new/newVm.go
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
	yamlExecuteFun func(yamlConfig config.SmartIdeConfig)) {

	//0. 错误反馈
	serverFeedback := preRun(cmd, workspaceInfo)

	_, err := start.ExecuteK8s_ServerWS_ServerEnv(cmd, *k8sUtil, workspaceInfo, yamlExecuteFun)
	serverFeedback(err)

}
