/*
 * @Author: kenan
 * @Date: 2022-02-15 17:18:27
 * @LastEditors: kenan
 * @LastEditTime: 2022-02-20 19:41:11
 * @FilePath: /smartide-cli/internal/biz/workspace/workspace.go
 * @Description:
 *
 * Copyright (c) 2022 by kenanlu@leansoftx.com, All Rights Reserved.
 */
package workspace

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/leansoftX/smartide-cli/internal/biz/config"
	"github.com/leansoftX/smartide-cli/internal/model"
	"github.com/leansoftX/smartide-cli/pkg/common"
)

func GetServerWorkspaceList(auth model.Auth) (ws []WorkspaceInfo, err error) {
	//ws = workspaces
	url := fmt.Sprint(auth.LoginUrl, "/api/smartide/workspace/getList")
	headers := map[string]string{
		"Content-Type": "application/json",
		//"x-token":      auth.Token.(string),
	}
	if auth.Token != nil {
		headers["x-token"] = auth.Token.(string)
	}
	response, err := common.Get(url, map[string]string{}, headers)

	if err != nil {
		return nil, err
	}
	if response == "" {
		return nil, errors.New("服务器返回空！")
	}

	l := &model.WorkspaceListResponse{}
	err = json.Unmarshal([]byte(response), l)
	common.CheckError(err)
	if l.Code == 0 && len(l.Data.List) > 0 {
		{
			for _, serverWorkSpace := range l.Data.List {
				workspaceInfo, err := CreateWorkspaceInfoFromServer(serverWorkSpace)
				common.CheckError(err)
				ws = append(ws, workspaceInfo)
			}
		}
	}
	return ws, err
}

// 获取工作区详情
func GetWorkspaceFromServer(auth model.Auth, no string) (workspaceInfo WorkspaceInfo, err error) {
	url := fmt.Sprint(auth.LoginUrl, "/api/smartide/workspace/getList")
	response, err := common.Get(url, map[string]string{},
		map[string]string{
			"Content-Type": "application/json",
			"x-token":      auth.Token.(string),
		})

	if err != nil {
		return WorkspaceInfo{}, err
	}
	if response == "" {
		return WorkspaceInfo{}, errors.New("服务器访问空数据！")
	}

	l := &model.WorkspaceListResponse{}
	err = json.Unmarshal([]byte(response), l)
	if err != nil {
		return WorkspaceInfo{}, err
	}

	if l.Code == 0 && len(l.Data.List) > 0 {
		{
			for _, w := range l.Data.List {
				if strings.EqualFold(w.NO, no) {
					workspaceInfo, err = CreateWorkspaceInfoFromServer(w)
					break
				}
			}
		}
	}
	return workspaceInfo, err
}

// 获取当前用户
func GetCurrentUser() (auth model.Auth, err error) {
	c := &config.GlobalSmartIdeConfig
	for i, a := range c.Auths {
		if a.CurrentUse {
			auth = c.Auths[i]
		}
	}
	return auth, err
}
