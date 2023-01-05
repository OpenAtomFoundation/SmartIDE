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
