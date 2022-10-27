/*
 * @Date: 2022-04-20 11:03:53
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-10-27 10:55:23
 * @FilePath: /cli/cmd/new/global.go
 */

package new

import (
	"fmt"
	"strings"

	smartideServer "github.com/leansoftX/smartide-cli/cmd/server"
	"github.com/leansoftX/smartide-cli/internal/apk/i18n"
	"github.com/leansoftX/smartide-cli/internal/biz/workspace"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/spf13/cobra"
)

var i18nInstance = i18n.GetInstance()

func preRun(cmd *cobra.Command, workspaceInfo workspace.WorkspaceInfo) func(err error) {
	mode, _ := cmd.Flags().GetString("mode")
	isModeServer := strings.ToLower(mode) == "server"
	// 错误反馈
	serverFeedback := func(err error) {
		if !isModeServer {
			return
		}
		if err != nil {
			smartideServer.Feedback_Finish(smartideServer.FeedbackCommandEnum_New, cmd, false, nil, workspaceInfo, err.Error(), "")
			common.CheckError(err)
		}
	}

	if apiHost, _ := cmd.Flags().GetString(smartideServer.Flags_ServerHost); apiHost != "" {
		wsURL := fmt.Sprint(strings.ReplaceAll(strings.ReplaceAll(apiHost, "https", "ws"), "http", "ws"), "/ws/smartide/ws")
		common.WebsocketStart(wsURL)
		token, _ := cmd.Flags().GetString(smartideServer.Flags_ServerToken)
		if token != "" {
			if workspaceIdStr, _ := cmd.Flags().GetString(smartideServer.Flags_ServerWorkspaceid); workspaceIdStr != "" {
				if no, _ := workspace.GetWorkspaceNo(workspaceIdStr, token, apiHost); no != "" {
					if pid, err := workspace.GetParentId(no, workspace.ActionEnum_Workspace_Start, token, apiHost); err == nil && pid > 0 {
						common.SmartIDELog.Ws_id = no
						common.SmartIDELog.ParentId = pid
					}
				}
			}

		}
	}

	return serverFeedback
}
