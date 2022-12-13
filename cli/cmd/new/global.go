/*
 * @Date: 2022-04-20 11:03:53
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-12-12 11:02:28
 * @FilePath: /cli/cmd/new/global.go
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
