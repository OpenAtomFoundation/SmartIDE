/*
 * @Date: 2022-06-07 15:38:21
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-09-22 08:58:47
 * @FilePath: /cli/cmd/remove/server.go
 */

package remove

import (
	"errors"
	"time"

	"github.com/leansoftX/smartide-cli/cmd/server"
	"github.com/leansoftX/smartide-cli/internal/biz/workspace"
	"github.com/leansoftX/smartide-cli/internal/model"
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
		if serverWorkSpace.ServerWorkSpace.Status == model.WorkspaceStatusEnum_Remove ||
			serverWorkSpace.ServerWorkSpace.Status == model.WorkspaceStatusEnum_Error_Remove ||
			serverWorkSpace.ServerWorkSpace.Status == model.WorkspaceStatusEnum_ContainerRemoved ||
			serverWorkSpace.ServerWorkSpace.Status == model.WorkspaceStatusEnum_Error_ContainerRemoved {
			isRemoved = true
			desc := serverWorkSpace.ServerWorkSpace.Status.GetDesc()
			common.SmartIDELog.Info(desc)
		}

		time.Sleep(time.Second * 15)
	}

	return nil
}
