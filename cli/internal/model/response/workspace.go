/*
 * @Author: kenan
 * @Date: 2022-02-15 19:32:44
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-11-03 15:00:04
 * @FilePath: /cli/internal/model/response/workspace.go
 * @Description:
 *
 * Copyright (c) 2022 by kenanlu@leansoftx.com, All Rights Reserved.
 */
package response

import (
	"time"
)

type WorkspaceStatusDictionaryResponse struct {
	Code int `json:"code"`
	Data struct {
		ResysDictionary struct {
			SysDictionaryDetails []struct {
				Value int    `json:"value"`
				Label string `json:"label"`
			} `json:"sysDictionaryDetails"`
		} `json:"resysDictionary"`
	} `json:"data"`
	Msg string `json:"msg"`
}

type WorkspaceListResponse struct {
	Code int    `json:"code"`
	Data Data   `json:"data"`
	Msg  string `json:"msg"`
}

type WorkspaceLogResponse struct {
	Code int `json:"code"`
	Data struct {
		ResServerWorkspaceLog ServerWorkspaceLogResponse `json:"rewsLog"`
	}
	Msg string `json:"msg"`
}

type GetWorkspaceSingleResponse struct {
	Code int `json:"code"`
	Data struct {
		ResmartideWorkspace ServerWorkspaceResponse `json:"resmartideWorkspaces"`
	} `json:"data"`
	Msg string `json:"msg"`
}

type Data struct {
	List []ServerWorkspaceResponse `json:"list"`
}

//	{
//	  "code": 0,
//	  "data": {
//	    "list": [
//	      {
//	        "ID": 1,
//	        "CreatedAt": "2022-03-22T06:28:36.596Z",
//	        "UpdatedAt": "2022-03-22T06:28:36.596Z",
//	        "title": "启动工作区",
//	        "parentID": 0,
//	        "content": "启动工作区",
//	        "ws_id": "SWS001",
//	        "level": 0,
//	        "type": 0,
//	        "startAt": "2022-03-22T06:28:36.196Z",
//	        "endAt": "2022-03-22T06:28:36.196Z",
//	        "status": 0
//	      }
//	    ],
//	    "total": 1,
//	    "page": 0,
//	    "pageSize": 0
//	  },
//	  "msg": "获取成功"
//	}
type LogData struct {
	List     []ServerWorkspaceLogResponse `json:"list"`
	Total    int64                        `json:"total"`
	Page     int                          `json:"page"`
	PageSize int                          `json:"pageSize"`
}

// 工作区
type ServerWorkspaceResponse struct {
	GVA_MODEL
	NO   string `json:"no" `
	Name string `json:"name" `

	GitRepoUrl               string `json:"gitRepoUrl" `
	Branch                   string `json:"branch"`
	ConfigFilePath           string `json:"configFilePath"`
	ConfigFileContent        string `json:"configFileContent"`
	TempDockerComposeContent string `json:"tempDockerComposeContent" `
	LinkDockerCompose        string `json:"linkDockerCompose" `
	Extend                   string `json:"extend" `

	Status WorkspaceStatusEnum `json:"status" `

	ResourceID int                    `json:"ResourceID" `
	Resource   ServerResourceResponse `json:"Resource"`

	OwnerGUID string `json:"ownerGuid" `
	OwnerName string `json:"ownerName" `

	KubeNamespace                 string                            `json:"kubeNamespace" `
	KubeIngressAuthenticationType KubeIngressAuthenticationTypeEnum `json:"kubeIngressAuthenticationType"`
	KubeIngressLoginUserName      string                            `json:"kubeIngressUserName" `
	KubeIngressLoginPassword      string                            `json:"kubeIngressPassword" `

	// 最大的cpu
	K8sUsedCup float32 `json:"cup" `
	// 最大的内存
	K8sUsedMemory float32 `json:"memory" `
	// 模板库的git clone url
	TemplateGitUrl string `json:"templateGitUrl" `
	// 端口配置 eg:[{label:apps-ports-3001,value:3001}]
	PortConfigs []PortConfig `json:"ports" `
}

// 端口配置
type PortConfig struct {
	// 描述
	Label string `json:"label" `
	// 端口号
	Port uint `json:"value" `
}

// "ID": 1,
// "CreatedAt": "2022-03-22T06:28:36.596Z",
// "UpdatedAt": "2022-03-22T06:28:36.596Z",
// "title": "启动工作区",
// "parentID": 0,
// "content": "启动工作区",
// "ws_id": "SWS001",
// "level": 0,
// "type": 0,
// "startAt": "2022-03-22T06:28:36.196Z",
// "endAt": "2022-03-22T06:28:36.196Z",
// "status": 0
type ServerWorkspaceLogResponse struct {
	GVA_MODEL
	Title      string     `json:"title"`
	ParentId   int        `json:"parentID" `
	Content    string     `json:"content" `
	Ws_id      string     `json:"ws_id" `
	Level      int        `json:"level" `
	Type       int        `json:"type" `
	StartAt    *time.Time `json:"startAt" `
	EndAt      *time.Time `json:"endAt" `
	Status     int        `json:"status" `
	TekEventId string     `json:"tekEventId"`
}

// 资源

// 工作区状态
type WorkspaceStatusEnum int

const (
	// 初始化
	WorkspaceStatusEnum_Init WorkspaceStatusEnum = 0
	//
	WorkspaceStatusEnum_Pending           WorkspaceStatusEnum = 101
	WorkspaceStatusEnum_Pending_NsCreated WorkspaceStatusEnum = 111

	WorkspaceStatusEnum_Start WorkspaceStatusEnum = 199

	WorkspaceStatusEnum_Stopping WorkspaceStatusEnum = 201
	WorkspaceStatusEnum_Stop     WorkspaceStatusEnum = 299

	WorkspaceStatusEnum_Removing WorkspaceStatusEnum = 301
	WorkspaceStatusEnum_Remove   WorkspaceStatusEnum = 399

	WorkspaceStatusEnum_ContainerRemoving WorkspaceStatusEnum = 311
	WorkspaceStatusEnum_ContainerRemoved  WorkspaceStatusEnum = 319

	WorkspaceStatusEnum_Error_Start            WorkspaceStatusEnum = -100
	WorkspaceStatusEnum_Error_Stop             WorkspaceStatusEnum = -201
	WorkspaceStatusEnum_Error_Remove           WorkspaceStatusEnum = -301
	WorkspaceStatusEnum_Error_ContainerRemoved WorkspaceStatusEnum = -302
)

func (workspaceStatus WorkspaceStatusEnum) GetDesc() string {

	desc := ""
	switch workspaceStatus {
	case WorkspaceStatusEnum_Init:
		desc = "Initialization"
	case WorkspaceStatusEnum_Pending:
		desc = "Pending"
	case WorkspaceStatusEnum_Remove:
		desc = "Cleaned"
	case WorkspaceStatusEnum_Removing:
		desc = "Cleaning"
	case WorkspaceStatusEnum_ContainerRemoving:
		desc = "Removing"
	case WorkspaceStatusEnum_ContainerRemoved:
		desc = "Removed"
	case WorkspaceStatusEnum_Stop:
		desc = "Stopped"
	case WorkspaceStatusEnum_Stopping:
		desc = "Stopping"
	case WorkspaceStatusEnum_Start:
		desc = "Running"
	default:
		desc = "Pending"
	}
	if int(workspaceStatus) < 0 {
		desc = "Error"
	}

	return desc
}
