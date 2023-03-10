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

package remove

import (
	"errors"
	"time"

	"github.com/leansoftX/smartide-cli/cmd/server"
	"github.com/leansoftX/smartide-cli/internal/biz/workspace"
	"github.com/leansoftX/smartide-cli/internal/model/response"
	"github.com/leansoftX/smartide-cli/pkg/common"
)

// 删除服务器上的远程主机类型工作区
func RemoveServerWorkSpaceInClient(workspaceIdStr string, workspaceInfo workspace.WorkspaceInfo, isOnlyRemoveContainer bool) error {
	currentServerAuth, _ := workspace.GetCurrentUser()

	// 触发remove
	datas := make(map[string]interface{})
	if isOnlyRemoveContainer {
		datas["isOnlyRemoveContainer"] = true
	}
	err := server.Trigger_Action("remove", workspaceIdStr, currentServerAuth, datas)
	if err != nil {
		return err
	}

	// 轮询检查工作区状态
	common.SmartIDELog.Info("等待服务器删除工作区...")
	isRemoved := false
	for !isRemoved {
		serverWorkSpace, err := workspace.GetWorkspaceFromServer(currentServerAuth, workspaceInfo.ID, workspace.CliRunningEnvEnum_Client)
		if serverWorkSpace == nil {
			return errors.New("工作区数据查询为空！")
		}
		if err != nil {
			return err
		}
		if serverWorkSpace.ServerWorkSpace.Status == response.WorkspaceStatusEnum_Remove ||
			serverWorkSpace.ServerWorkSpace.Status == response.WorkspaceStatusEnum_Error_Remove ||
			serverWorkSpace.ServerWorkSpace.Status == response.WorkspaceStatusEnum_ContainerRemoved ||
			serverWorkSpace.ServerWorkSpace.Status == response.WorkspaceStatusEnum_Error_ContainerRemoved {
			isRemoved = true
			desc := serverWorkSpace.ServerWorkSpace.Status.GetDesc()
			common.SmartIDELog.Info(desc)
		}

		time.Sleep(time.Second * 15)
	}

	return nil
}
