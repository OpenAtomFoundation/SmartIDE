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
	smartideServer "github.com/leansoftX/smartide-cli/cmd/server"
	"github.com/leansoftX/smartide-cli/internal/apk/i18n"
	"github.com/leansoftX/smartide-cli/internal/biz/workspace"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/spf13/cobra"
)

var i18nInstance = i18n.GetInstance()

// 错误反馈
func serverFeedback(workspaceInfo workspace.WorkspaceInfo, cmd *cobra.Command, err error) {
	if workspaceInfo.CliRunningEnv != workspace.CliRunningEvnEnum_Server {
		return
	}
	if err != nil {
		smartideServer.Feedback_Finish(smartideServer.FeedbackCommandEnum_Start, cmd, false, nil, workspaceInfo, err.Error(), "")
	}

	common.SmartIDELog.Error(err)
}
