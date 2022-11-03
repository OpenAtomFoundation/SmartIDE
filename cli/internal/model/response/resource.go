/*
 * @Date: 2022-11-03 14:30:12
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-11-03 14:30:13
 * @FilePath: /cli/internal/model/response/resource.go
 */

package response

type ServerResourceResponse struct {
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
