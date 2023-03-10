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

package new

import (
	"encoding/json"
	"fmt"
	"strings"

	smartideServer "github.com/leansoftX/smartide-cli/cmd/server"

	"github.com/leansoftX/smartide-cli/internal/apk/i18n"
	"github.com/leansoftX/smartide-cli/internal/biz/workspace"
	"github.com/leansoftX/smartide-cli/internal/model"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/leansoftX/smartide-cli/pkg/k8s"
	"github.com/spf13/cobra"
)

var i18nInstance = i18n.GetInstance()

func preRun(cmd *cobra.Command, workspaceInfo workspace.WorkspaceInfo, logAction workspace.ActionEnum, k8sUtil *k8s.KubernetesUtil) func(err error) {
	mode, _ := cmd.Flags().GetString("mode")
	isModeServer := strings.ToLower(mode) == "server"
	// 错误反馈
	serverFeedback := func(err error) {
		if !isModeServer {
			return
		}
		if err != nil {
			errFeedback, ok := err.(*model.FeedbackError)
			if !ok {
				tmp := model.CreateFeedbackError(err.Error(), true)
				errFeedback = &tmp
			} else { // 不需要重试的错误，直接删除namespace
				if !errFeedback.IsRetry && k8sUtil != nil && workspaceInfo.Mode == workspace.WorkingMode_K8s {
					k8sUtil.ExecKubectlCommandCombined("delete namespace --force "+workspaceInfo.K8sInfo.Namespace, "")
				}
			}
			errMsgBits, _ := json.Marshal(errFeedback)
			smartideServer.Feedback_Finish(smartideServer.FeedbackCommandEnum_New, cmd, false, nil, workspaceInfo, string(errMsgBits), "")
			common.CheckError(err) // exit
		}
	}

	// 日志
	if apiHost, _ := cmd.Flags().GetString(smartideServer.Flags_ServerHost); apiHost != "" {
		wsURL := fmt.Sprint(strings.ReplaceAll(strings.ReplaceAll(apiHost, "https", "ws"), "http", "ws"), "/ws/smartide/ws")
		common.WebsocketStart(wsURL)
		token, _ := cmd.Flags().GetString(smartideServer.Flags_ServerToken)
		if token != "" {
			if workspaceIdStr, _ := cmd.Flags().GetString(smartideServer.Flags_ServerWorkspaceid); workspaceIdStr != "" {
				if no, _ := workspace.GetWorkspaceNo(workspaceIdStr, token, apiHost); no != "" {
					if pid, err := workspace.GetParentId(no, logAction, token, apiHost); err == nil && pid > 0 {
						common.SmartIDELog.Ws_id = no
						common.SmartIDELog.ParentId = pid
					}
				}
			}

		}
	}

	return serverFeedback
}
