/*
 * @Author: kenan
 * @Date: 2022-02-15 19:32:44
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-09-09 09:57:36
 * @FilePath: /cli/internal/model/workspace.go
 * @Description:
 *
 * Copyright (c) 2022 by kenanlu@leansoftx.com, All Rights Reserved.
 */
package model

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
		ResServerWorkspaceLog ServerWorkspaceLog `json:"rewsLog"`
	}
	Msg string `json:"msg"`
}

type WorkspaceResponse struct {
	Code int `json:"code"`
	Data struct {
		ResmartideWorkspace ServerWorkspace `json:"resmartideWorkspaces"`
	} `json:"data"`
	Msg string `json:"msg"`
}

type Data struct {
	List []ServerWorkspace `json:"list"`
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
	List     []ServerWorkspaceLog `json:"list"`
	Total    int64                `json:"total"`
	Page     int                  `json:"page"`
	PageSize int                  `json:"pageSize"`
}

type GVA_MODEL struct {
	ID        uint      // 主键ID
	CreatedAt time.Time // 创建时间
	UpdatedAt time.Time // 更新时间
}

// 工作区
type ServerWorkspace struct {
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

	ResourceID int      `json:"ResourceID" `
	Resource   Resource `json:"Resource"`

	OwnerGUID string `json:"ownerGuid" `
	OwnerName string `json:"ownerName" `

	KubeNamespace                 string                            `json:"kubeNamespace" `
	KubeIngressAuthenticationType KubeIngressAuthenticationTypeEnum `json:"kubeIngressAuthenticationType"`
	KubeIngressLoginUserName      string                            `json:"kubeIngressUserName" `
	KubeIngressLoginPassword      string                            `json:"kubeIngressPassword" `
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

type ServerWorkspaceLog struct {
	GVA_MODEL
	Title    string     `json:"title"`
	ParentId int        `json:"parentID" `
	Content  string     `json:"content" `
	Ws_id    string     `json:"ws_id" `
	Level    int        `json:"level" `
	Type     int        `json:"type" `
	StartAt  *time.Time `json:"startAt" `
	EndAt    *time.Time `json:"endAt" `
	Status   int        `json:"status" `
}

// 资源
type Resource struct {
	GVA_MODEL
	Type               ReourceTypeEnum        `json:"type"`
	Name               string                 `json:"name"`
	AuthenticationType AuthenticationTypeEnum `json:"authentication_type"`
	IP                 string                 `json:"ip" `
	Port               int                    `json:"port" `
	Status             ResourceStatusEnum     `json:"status" `
	// ssh 模式下的用户名 && k8s ingress 用户名
	UserName string `json:"username" `
	// ssh 模式 用户名密码验证方式的密码 && k8s ingress 密码
	Password string `json:"password" `
	SSHKey   string `json:"sshkey" `

	KubeConfigContent string `json:"kube_config"`
	KubeBaseDNS       string `json:"kube_ingress_base_dns" `
	KubeContext       string `json:"kube_context" `

	//KubeUserName           string                     `json:"kube_ingress_user_name"`
	//KubePassword           string                     `json:"kube_ingress_password"`

	OwnerGUID string `json:"ownerGuid" `
	OwnerName string `json:"ownerName" `
}

// 资源类型
type ReourceTypeEnum int

const (
	ReourceTypeEnum_Local  ReourceTypeEnum = 1
	ReourceTypeEnum_Remote ReourceTypeEnum = 2
	ReourceTypeEnum_K8S    ReourceTypeEnum = 3
)

// 认证方式
type AuthenticationTypeEnum int

const (
	AuthenticationTypeEnum_SSH        AuthenticationTypeEnum = 1
	AuthenticationTypeEnum_PAT        AuthenticationTypeEnum = 2
	AuthenticationTypeEnum_Password   AuthenticationTypeEnum = 3
	AuthenticationTypeEnum_KubeConfig AuthenticationTypeEnum = 4
)

// kube认证方式
type KubeIngressAuthenticationTypeEnum int

const (
	KubeAuthenticationTypeEnum_None  KubeIngressAuthenticationTypeEnum = 1
	KubeAuthenticationTypeEnum_Basic KubeIngressAuthenticationTypeEnum = 2
	KubeAuthenticationTypeEnum_Oauth KubeIngressAuthenticationTypeEnum = 3
)

// 资源状态
type ResourceStatusEnum int

const (
	// 初始化
	ResourceStatusEnum_Init ResourceStatusEnum = 0
	//
	ResourceStatusEnum_Pending           ResourceStatusEnum = 101
	ResourceStatusEnum_Start             ResourceStatusEnum = 109
	ResourceStatusEnum_Stop              ResourceStatusEnum = 201
	ResourceStatusEnum_Error_Unreachable ResourceStatusEnum = -100
)

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
