/*
 * @Author: kenan
 * @Date: 2022-03-14 09:54:06
 * @LastEditors: kenan
 * @LastEditTime: 2022-03-17 09:32:06
 * @FilePath: /smartide-cli/internal/model/ws_log.go
 * @Description:
 *
 * Copyright (c) 2022 by kenanlu@leansoftx.com, All Rights Reserved.
 */
package model

import (
	"time"
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
