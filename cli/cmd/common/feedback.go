/*
 * @Date: 2022-11-04 14:53:22
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-11-04 14:56:21
 * @FilePath: /cli/cmd/common/feedback.go
 */

package common

import (
	"encoding/json"
	"strings"

	"github.com/leansoftX/smartide-cli/internal/biz/workspace"
	"github.com/leansoftX/smartide-cli/internal/model"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/spf13/cobra"

	smartideServer "github.com/leansoftX/smartide-cli/cmd/server"
)

func GetFeedbackFunc(cmd *cobra.Command, workspaceInfo workspace.WorkspaceInfo) func(error) {
	mode, _ := cmd.Flags().GetString("mode")
	isModeServer := strings.ToLower(mode) == "server"

	return func(err error) {
		if !isModeServer {
			return
		}
		if err != nil {
			errFeedback, ok := err.(*model.FeedbackError)
			if !ok {
				tmp := model.CreateFeedbackError(err.Error(), true)
				errFeedback = &tmp
			}
			errMsgBits, _ := json.Marshal(errFeedback)
			smartideServer.Feedback_Finish(smartideServer.FeedbackCommandEnum_New, cmd, false, nil, workspaceInfo, string(errMsgBits), "")
			common.CheckError(err) // exit
		}
	}
}
