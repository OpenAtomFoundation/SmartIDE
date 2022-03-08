/*
 * @Author: kenan
 * @Date: 2022-02-15 19:32:44
 * @LastEditors: kenan
 * @LastEditTime: 2022-02-17 14:03:13
 * @FilePath: /smartide-cli/internal/model/workspace.go
 * @Description:
 *
 * Copyright (c) 2022 by kenanlu@leansoftx.com, All Rights Reserved.
 */
package model

import (
	"time"
)

type WorkspaceListResponse struct {
	Code int    `json:"code"`
	Data Data   `json:"data"`
	Msg  string `json:"msg"`
}

type Data struct {
	List []ServerWorkspace `json:"list"`
}

type GVA_MODEL struct {
	ID        uint      // 主键ID
	CreatedAt time.Time // 创建时间
	UpdatedAt time.Time // 更新时间
}

// 工作区
type ServerWorkspace struct {
	GVA_MODEL
	NO                       string `json:"no" `
	Name                     string `json:"name" `
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
}

// 资源
type Resource struct {
	GVA_MODEL
	Type               ReourceTypeEnum        `json:"type"`
	Name               string                 `json:"name"`
	AuthenticationType AuthenticationTypeEnum `json:"authentication_type"`
	IP                 string                 `json:"ip" `
	Status             ResourceStatusEnum     `json:"status" `
	SSHUserName        string                 `json:"username" `
	SSHPassword        string                 `json:"password" `
	SSHKey             string                 `json:"sshkey" `
	OwnerGUID          string                 `json:"ownerGuid" `
	OwnerName          string                 `json:"ownerName" `
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
	WorkspaceStatusEnum_Pending WorkspaceStatusEnum = 101
	WorkspaceStatusEnum_Start   WorkspaceStatusEnum = 109
	WorkspaceStatusEnum_Stop    WorkspaceStatusEnum = 201
)
