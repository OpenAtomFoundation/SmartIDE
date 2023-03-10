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

package workspace

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/leansoftX/smartide-cli/internal/biz/config"
	"github.com/leansoftX/smartide-cli/internal/model"
	apiResponse "github.com/leansoftX/smartide-cli/internal/model/response"
	"github.com/leansoftX/smartide-cli/pkg/common"
)

func GetServerWorkspaceList(auth model.Auth, cliRunningEnv CliRunningEvnEnum) (ws []WorkspaceInfo, err error) {
	url := fmt.Sprint(auth.LoginUrl, "/api/smartide/workspace/getList?page=1&pageSize=100")
	headers := map[string]string{
		"Content-Type": "application/json",
		//"x-token":      auth.Token.(string),
	}
	if auth.Token != nil {
		headers["x-token"] = auth.Token.(string)
	}
	httpClient := common.CreateHttpClientEnableRetry()
	response, err := httpClient.Get(url, map[string]string{}, headers)
	if err != nil {
		return nil, err
	}
	if response == "" {
		return nil, errors.New("服务器返回空！")
	}

	l := &apiResponse.WorkspaceListResponse{}
	err = json.Unmarshal([]byte(response), l)
	if err != nil {
		return ws, err
	}
	if l.Code != 0 {
		err = errors.New(l.Msg)
		return
	}
	if len(l.Data.List) > 0 {
		{
			for _, serverWorkSpace := range l.Data.List {
				workspaceInfo, err := CreateWorkspaceInfoFromServer(serverWorkSpace)
				workspaceInfo.CliRunningEnv = cliRunningEnv
				if err != nil {
					return ws, err
				}
				ws = append(ws, workspaceInfo)
			}
		}
	}
	return ws, err
}

// 获取工作区详情
func GetWorkspaceFromServer(auth model.Auth, no string, cliRunningEnv CliRunningEvnEnum) (workspaceInfo *WorkspaceInfo, err error) {
	if (auth == model.Auth{}) {
		err = errors.New("用户未登录！")
		return
	}

	no = strings.TrimSpace(no)
	if no == "" {
		return nil, errors.New("workspace id is nil")
	}
	url := fmt.Sprint(auth.LoginUrl, "/api/smartide/workspace/find")
	queryparams := map[string]string{}
	if common.IsNumber(no) {
		queryparams["id"] = no
	} else {
		queryparams["no"] = no
	}
	httpClient := common.CreateHttpClientEnableRetry()
	response, err := httpClient.Get(url,
		queryparams,
		map[string]string{
			"Content-Type": "application/json",
			"x-token":      auth.Token.(string),
		})

	if err != nil {
		return nil, err
	}
	if response == "" {
		return nil, errors.New("服务器访问空数据！")
	}

	l := &apiResponse.GetWorkspaceSingleResponse{}
	err = json.Unmarshal([]byte(response), l)
	if err != nil {
		return nil, err
	}

	if l.Code == 0 {
		var workspaceInfo_ WorkspaceInfo
		workspaceInfo_, err = CreateWorkspaceInfoFromServer(l.Data.ResmartideWorkspace)
		workspaceInfo = &workspaceInfo_
		workspaceInfo.CliRunningEnv = cliRunningEnv
	}
	return workspaceInfo, err
}

// 获取当前用户
func GetCurrentUser() (auth model.Auth, err error) {
	c := &config.GlobalSmartIdeConfig
	for i, a := range c.Auths {
		if a.CurrentUse {
			auth = c.Auths[i]
			break
		}
	}

	if auth.Token != "" {
		common.SmartIDELog.AddEntryptionKeyWithReservePart(fmt.Sprint(auth.Token))
	}

	return auth, err
}
