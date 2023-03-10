/*
SmartIDE - CLI
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
