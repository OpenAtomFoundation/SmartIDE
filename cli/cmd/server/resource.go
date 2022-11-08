/*
 * @Date: 2022-06-30 22:17:22
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-11-03 14:51:03
 * @FilePath: /cli/cmd/server/resource.go
 */
package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/leansoftX/smartide-cli/internal/model"
	apiResponse "github.com/leansoftX/smartide-cli/internal/model/response"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"gorm.io/gorm"
)

type ServerResourceResponse struct {
	Code int `json:"code"`
	Data struct {
		ResmartideResource ServerResource `json:"resmartideResources"`
	} `json:"data"`
	Msg string `json:"msg"`
}

type ServerResource struct {
	ID                 int `json:"ID" gorm:"comment:ID"`
	CreatedAt          time.Time
	UpdatedAt          time.Time
	DeletedAt          gorm.DeletedAt
	Type               int    `json:"type" gorm:"not null;type:smallint;default:2;comment:资源类型"`
	Name               string `json:"name" gorm:"comment:名称"`
	AuthenticationType int    `json:"authentication_type" gorm:"not null;type:smallint;default:3;comment:认证类型"`
	IP                 string `json:"ip" gorm:"not null;type:string;"`
	Status             int    `json:"status" gorm:"not null;type:smallint;default:0;comment:状态"`
	SSHUserName        string `json:"username" gorm:"column:user_name;type:string;comment:（k8s）用户名;"`
	SSHPassword        string `json:"password" gorm:"column:password;type:string;comment:(k8s)密码;"`
	SSHKey             string `json:"sshkey" gorm:"column:ssh_key;type:text;comment:public access tokent;"`
	KubeIngressBaseDNS string `json:"kube_ingress_base_dns" gorm:"column:kube_base_dns;comment:kube base dns;"`
	KubeConfig         string `json:"kube_config" gorm:"column:kube_config;type:text;comment:kube config;"`
	KubeContext        string `json:"kube_context" gorm:"column:kube_context;comment:kube context name;"`
	OwnerGUID          string `json:"ownerGuid" form:"ownerGuid" gorm:"column:owner_guid;type:string;comment:资源的创建人即所有者 GUID;"`
	OwnerName          string `json:"ownerName" form:"ownerName" gorm:"column:owner_name;type:string;comment:资源的创建人即所有者 账号;"`
	IsDefault          bool   `json:"isDefault" form:"isDefault" gorm:"column:isDefault;type:string;comment:是否默认;"`
	Port               int    `json:"port" form:"port" gorm:"column:port;type:int;comment:端口号;"`
	GlobalPublic       bool   `json:"globalPublic" form:"globalPublic" gorm:"column:globalPublic;type:string;comment:全局共享资源;"`
	TeamID             int    `json:"teamId" form:"teamId" gorm:"column:teamId;type:int;commnent:团队ID"`
	CertType           int    `json:"certtype" form:"certtype" gorm:"column:certtype;type:int;commnent:认证策略类型,1:http,2:静态证书https,3:动态证书https"`
	CertCrt            string `json:"certcrt" form:"certcrt" gorm:"column:certcrt;type:string;comment:Https证书"`
	CertKey            string `json:"certkey" form:"certkey" gorm:"column:certkey;type:string;comment:Https证书秘钥"`
}

func GetResourceByID(auth model.Auth, resourceID string) (serverResource *ServerResource, err error) {
	url := fmt.Sprint(auth.LoginUrl, "/api/smartide/resource/find")
	queryparams := map[string]string{}
	queryparams["ID"] = resourceID

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

func UpdateResourceByID(auth model.Auth, serverResource *ServerResource) (err error) {
	url := fmt.Sprint(auth.LoginUrl, "/api/smartide/resource/update")
	httpClient := common.CreateHttpClientEnableRetry()
	var serverResourceMap map[string]interface{}
	data, _ := json.Marshal(serverResource)
	json.Unmarshal(data, &serverResourceMap)
	response, err := httpClient.Put(url,
		serverResourceMap, map[string]string{"Content-Type": "application/json", "x-token": auth.Token.(string)})
	if response != "" {
		l := &apiResponse.GetWorkspaceSingleResponse{}
		if err = json.Unmarshal([]byte(response), l); err == nil {
			if l.Code == 0 {
				return nil
			}
		}
	}
	return err
}
