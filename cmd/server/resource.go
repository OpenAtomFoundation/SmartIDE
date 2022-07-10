/*
 * @Date: 2022-06-30 22:17:22
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-07-04 08:44:19
 * @FilePath: /smartide-cli/cmd/server/resource.go
 */
package server

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/leansoftX/smartide-cli/internal/model"
	"github.com/leansoftX/smartide-cli/pkg/common"
)

//
type ServerResourceResponse struct {
	Code int `json:"code"`
	Data struct {
		ResmartideResource ServerResource `json:"resmartideResources"`
	} `json:"data"`
	Msg string `json:"msg"`
}

type ServerResource struct {
	ID                 int    `json:"ID" gorm:"comment:ID"`
	Name               string `json:"name" gorm:"comment:名称"`
	KubeConfig         string `json:"kube_config" gorm:"column:kube_config;type:text;comment:kube config;"`
	KubeContext        string `json:"kube_context" gorm:"column:kube_context;comment:kube context name;"`
	KubeIngressBaseDNS string `json:"kube_ingress_base_dns" gorm:"column:kube_base_dns;comment:kube base dns;"`
	OwnerGUID          string `json:"ownerGuid" form:"ownerGuid" gorm:"column:owner_guid;type:string;comment:资源的创建人即所有者 GUID;"`
	OwnerName          string `json:"ownerName" form:"ownerName" gorm:"column:owner_name;type:string;comment:资源的创建人即所有者 账号;"`
	IsDefault          bool   `json:"isDefault" form:"isDefault" gorm:"column:isDefault;type:string;comment:是否默认;"`
	Port               int    `json:"port" form:"port" gorm:"column:port;type:int;comment:端口号;"`
	GlobalPublic       bool   `json:"globalPublic" form:"globalPublic" gorm:"column:globalPublic;type:string;comment:全局共享资源;"`
	TeamID             int    `json:"teamId" form:"teamId" gorm:"column:teamId;type:int;commnent:团队ID"`
}

func GetResourceByID(auth model.Auth, resourceID string) (serverResource *ServerResource, err error) {
	url := fmt.Sprint(auth.LoginUrl, "/api/smartide/resource/find")
	queryparams := map[string]string{}
	queryparams["ID"] = resourceID
	response, err := common.Get(url,
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

	l := &ServerResourceResponse{}
	err = json.Unmarshal([]byte(response), l)
	if err != nil {
		return nil, err
	}

	if l.Code != 0 {
		return nil, fmt.Errorf("服务器访问错误，code：%v", l.Code)
	}
	return &l.Data.ResmartideResource, err
}
