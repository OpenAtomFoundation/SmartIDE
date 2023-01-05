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
