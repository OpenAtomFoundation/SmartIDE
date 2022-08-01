/*
 * @Date: 2022-06-07 14:21:08
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-08-01 09:47:15
 * @FilePath: /cli/cmd/remove/core.go
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
