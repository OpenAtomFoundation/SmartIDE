/*
 * @Author: jason chen (jasonchen@leansoftx.com, http://smallidea.cnblogs.com)
 * @Description:
 * @Date: 2021-11
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-09-05 15:17:59
 */
package start

import (
	"github.com/leansoftX/smartide-cli/internal/apk/cli"
	"github.com/leansoftX/smartide-cli/internal/apk/i18n"
	"github.com/leansoftX/smartide-cli/internal/apk/server"
	"github.com/leansoftX/smartide-cli/internal/biz/workspace"
	"github.com/leansoftX/smartide-cli/internal/model"
	"github.com/spf13/cobra"
)

var i18nInstance = i18n.GetInstance()

func getK8sLabels(cmd *cobra.Command, workspaceInfo workspace.WorkspaceInfo) map[string]string {
	labels := make(map[string]string)

	labels["smartide.workspaceName"] = workspaceInfo.Name
	labels["smartide.cliVersion"] = cli.GetCliVersionByShell() //
	labels["smartide.trigger"] = string(workspaceInfo.CliRunningEnv)
	labels["smartide.workspaceId"] = workspaceInfo.ID

	if workspaceInfo.CliRunningEnv != workspace.CliRunningEnvEnum_Client {
		userName, _ := cmd.Flags().GetString("serverusername")
		serverHost, _ := cmd.Flags().GetString("serverhost")
		serverToken, _ := cmd.Flags().GetString("servertoken")

		labels["smartide.workspaceId"] = workspaceInfo.ServerWorkSpace.NO
		labels["smartide.serverUser"] = userName
		labels["smartide.serverHost"] = serverHost

		auth := model.Auth{}
		auth.Token = serverToken
		auth.LoginUrl = serverHost
		auth.UserName = userName
		labels["smartide.serverVersion"], _ = server.GetVersionByApi(auth)
	}

	return labels
}
