/*
 * @Author: kenan
 * @Date: 2022-03-14 09:54:06
 * @LastEditors: kenan
 * @LastEditTime: 2022-10-11 16:45:44
 * @FilePath: /cli/internal/biz/workspace/ws_log.go
 * @Description:
 *
 * Copyright (c) 2022 by kenanlu@leansoftx.com, All Rights Reserved.
 */
package workspace

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/leansoftX/smartide-cli/internal/model"
	"github.com/leansoftX/smartide-cli/pkg/common"
)

/*
*description: ws接口日志结构体
* param Ws_id 工作区ID
* param Level 日志级别1:info,2:warning,3:debug,4:error
* param Type  日志类别1:smartide-cli,2:smartide-server
* param Status 执行状态 1:未启动,2:启动中,3:执行完毕,4:执行错误
 */
type WorkspaceLog struct {
	Title    string    `json:"title" `
	ParentId int       `json:"parentID"`
	Content  string    `json:"content" `
	Ws_id    string    `json:"ws_id" ` //工作区ID
	Level    int       `json:"level" ` //日志级别1:info,2:warning,3:debug,4:error
	Type     int       `json:"type"`   //日志类别1:smartide-cli,2:smartide-server
	StartAt  time.Time `json:"startAt"`
	EndAt    time.Time `json:"endAt" `
	Status   int       `json:"status" ` //执行状态 1:未启动,2:启动中,3:执行完毕,4:执行错误
}

type ActionEnum int

const (
	ActionEnum_Workspace_Start           ActionEnum = 1
	ActionEnum_Workspace_Stop            ActionEnum = 2
	ActionEnum_Workspace_RemoveContainer ActionEnum = 3
	ActionEnum_Workspace_Remove          ActionEnum = 4
	ActionEnum_Workspace_Connect         ActionEnum = 5
	ActionEnum_Ingress_Enable            ActionEnum = 6
	ActionEnum_Ingress_Disable           ActionEnum = 7
	ActionEnum_SSH_Enable                ActionEnum = 8
	ActionEnum_SSH_Disable               ActionEnum = 9
)

/*
* description: 根据action:1 start,2 stop ,3 deleteContainer,4 delete 获取日志parentID
* param wid 工作区id
* param action
 */
func GetParentId(wid string, action ActionEnum, token string, apiHost string) (praentId int, err error) {
	// 查询当前工作区日志parentid
	var title = ""
	var response = ""
	praentId = 0
	switch action {
	case ActionEnum_Workspace_Start:
		title = "启动工作区"
	case ActionEnum_Workspace_Stop:
		title = "停止工作区"
	case ActionEnum_Workspace_RemoveContainer:
		title = "删除工作区容器"
	case ActionEnum_Workspace_Remove:
		title = "清理工作区环境"
	case ActionEnum_Workspace_Connect:
		title = "客户端启动工作区"
	case ActionEnum_Ingress_Enable:
		title = "创建Ingress"
	case ActionEnum_Ingress_Disable:
		title = "删除Ingress"
	case ActionEnum_SSH_Enable:
		title = "创建SSH通道"
	case ActionEnum_SSH_Disable:
		title = "删除SSH通道"
	}
	url := fmt.Sprint(apiHost, "/api/smartide/wslog/find")

	httpClient := common.CreateHttpClientEnableRetry()
	response, err = httpClient.Get(url,
		map[string]string{"title": title, "ws_id": wid, "parentID": "0"},
		map[string]string{"Content-Type": "application/json", "x-token": token})
	if response != "" {
		l := &model.WorkspaceLogResponse{}
		if err = json.Unmarshal([]byte(response), l); err == nil {
			if l.Code == 0 && l.Data.ResServerWorkspaceLog.ID > 0 {
				return int(l.Data.ResServerWorkspaceLog.ID), err
			}
		}
	}
	return praentId, err

}

func GetWorkspaceNo(wid string, token string, apiHost string) (no string, err error) {
	// 查询当前工作区日志parentid
	var response = ""
	url := fmt.Sprint(apiHost, "/api/smartide/workspace/find")
	httpClient := common.CreateHttpClientEnableRetry()
	response, err = httpClient.Get(url,
		map[string]string{"id": wid},
		map[string]string{"Content-Type": "application/json", "x-token": token})
	if response != "" {
		l := &model.WorkspaceResponse{}
		//err = json.Unmarshal([]byte(response), l)
		if err = json.Unmarshal([]byte(response), l); err == nil {
			if l.Code == 0 && l.Data.ResmartideWorkspace.NO != "" {
				return l.Data.ResmartideWorkspace.NO, err
			}
		}
	}
	return no, err

}

func CreateWsLog(wid string, token string, apiHost string, title string, content string) (parentId int, err error) {
	var response = ""
	url := fmt.Sprint(apiHost, "/api/smartide/wslog/create")
	httpClient := common.CreateHttpClientEnableRetry()
	response, err = httpClient.PostJson(url,
		map[string]interface{}{
			"ws_id":   wid,
			"title":   title,
			"content": content,
			"level":   1,
			"type":    1,
			"startAt": time.Now(),
			"endAt":   time.Now(),
		}, map[string]string{"Content-Type": "application/json", "x-token": token})
	if response != "" {
		l := &model.WorkspaceLogResponse{}
		//err = json.Unmarshal([]byte(response), l)
		if err = json.Unmarshal([]byte(response), l); err == nil {
			if l.Code == 0 {
				return l.Data.ResServerWorkspaceLog.ParentId, nil
			}
		}
	}
	return -1, err
}

func UpdateWsLog(wid string, token string, apiHost string, title string, content string) (err error) {
	var response = ""
	url := fmt.Sprint(apiHost, "/api/smartide/wslog/create")
	httpClient := common.CreateHttpClientEnableRetry()
	response, err = httpClient.PostJson(url,
		map[string]interface{}{
			"ws_id":   wid,
			"title":   title,
			"content": content,
			"level":   1,
			"type":    1,
			"startAt": time.Now(),
			"endAt":   time.Now(),
		}, map[string]string{"Content-Type": "application/json", "x-token": token})
	if response != "" {
		l := &model.WorkspaceLogResponse{}
		//err = json.Unmarshal([]byte(response), l)
		if err = json.Unmarshal([]byte(response), l); err == nil {
			if l.Code == 0 {
				return nil
			}
		}
	}
	return err
}
