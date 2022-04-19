/*
 * @Author: jason chen (jasonchen@leansoftx.com, http://smallidea.cnblogs.com)
 * @Description:
 * @Date: 2021-11
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-04-11 10:36:28
 */
package config

import (
	"github.com/leansoftX/smartide-cli/internal/model"
	"github.com/leansoftX/smartide-cli/pkg/docker/compose"

	appV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"

	networkingV1 "k8s.io/api/networking/v1"
)

//
/* type SmartIdeConfigBase struct {
	//
	Version string `yaml:"version"`
	//
	Orchestrator struct {
		// 类型，比如docker-compose、smartide、k8s
		Type string `yaml:"type"`
		// docker-compose 文件用到的版本号
		Version string `yaml:"version"`
	} `yaml:"orchestrator"`
} */

//
type OrchestratorTypeEnum string

const (
	OrchestratorTypeEnum_K8S     OrchestratorTypeEnum = "k8s"
	OrchestratorTypeEnum_Compose OrchestratorTypeEnum = "docker-compose"
)

type IdeTypeEnum string

const (
	IdeTypeEnum_VsCode      IdeTypeEnum = "vscode"
	IdeTypeEnum_JbProjector IdeTypeEnum = "jb-projector"
	IdeTypeEnum_Opensumi    IdeTypeEnum = "opensumi"
	IdeTypeEnum_Theia       IdeTypeEnum = "theia"
	IdeTypeEnum_SDKOnly     IdeTypeEnum = "sdk-only"
)

// 开发容器配置
type DevContainerConfig struct {
	// 服务名称
	ServiceName string `yaml:"service-name"`
	// 端口申明
	Ports map[string]int `yaml:"ports"`
	// 容器运行起来后，在webide的terminal中执行的shell命令
	Command []string `yaml:"command"`
	// ide-type web ide类型
	IdeType IdeTypeEnum `yaml:"ide-type"`
	// 磁盘映射
	Volumes struct {
		GitConfig string `yaml:"git-config"`
		SshKey    string `yaml:"ssh-key"`
	} `yaml:"volumes"`

	// 绑定的端口列表
	bindingPorts []PortMapInfo

	// 默认 yaml配置文件的相对路径（相对于工作目录）
	configRelativeFilePath string //= CONST_Default_ConfigRelativeFilePath

	// 本地模式时，本地工作目录（相对于工作目录，但是在.ide.yaml中写的是相对于“.ide”目录，后续会通过程序进行转换）
	workingDirectoryPath string //= "." //TODO 应该是一个私有成员

	// web ide 对外访问的端口
	// WebidePort string `yaml:"webide-port"`
} //`yaml:"dev-container"`

// smartide 的配置
// docker-compose.yaml https://docs.docker.com/compose/compose-file/
type SmartIdeConfig struct {
	//
	Version string `yaml:"version"`
	//
	Orchestrator struct {
		// 类型，比如docker-compose、smartide、k8s
		Type OrchestratorTypeEnum `yaml:"type"`
		// docker-compose 文件用到的版本号
		Version string `yaml:"version"`
	} `yaml:"orchestrator"`

	//
	Workspace struct {
		// 开发容器申明
		DevContainer DevContainerConfig `yaml:"dev-container"`

		// 链接的docker-compose文件路径
		DockerComposeFile string `yaml:"docker-compose-file"`

		// k8s 的部署文件（通配符）
		KubeDeployFiles string `yaml:"kube-deploy-files,omitempty"`

		// 允许要启动的容器，docker-compose 中的services节点
		Servcies map[string]compose.Service `yaml:"services,omitempty"`
		// 组网，docker-compose 中的 Networks 节点
		Networks map[string]compose.Network `yaml:"networks,omitempty"`
		// 挂载卷配置，docker-compose 中的 Volumes 节点
		Volumes map[string]compose.Volume `yaml:"volumes,omitempty"`
		// 密钥，docker-compose 中的 Secrets 节点
		Secrets map[string]compose.YmlSecret `yaml:"secrets,omitempty"`
	} `yaml:"workspace"`
}

type SmartIdeK8SConfig struct {
	//
	Version string `yaml:"version"`
	//
	Orchestrator struct {
		// 类型，比如docker-compose、smartide、k8s
		Type OrchestratorTypeEnum `yaml:"type"`
		// docker-compose 文件用到的版本号
		Version string `yaml:"version"`
	} `yaml:"orchestrator"`

	//
	Workspace struct {
		// 开发容器申明
		DevContainer DevContainerConfig `yaml:"dev-container"`

		// k8s 的部署文件（通配符）
		KubeDeployFiles string `yaml:"kube-deploy-files,omitempty"`

		//
		Services []coreV1.Service

		//
		Deployments []appV1.Deployment

		//
		PVCS []coreV1.PersistentVolumeClaim

		//
		Networks []networkingV1.NetworkPolicy

		Others []interface{}
	} `yaml:"workspace"`
}

// 端口映射信息
type PortMapInfo struct {
	// 对应的service名称，对应 docker-compose的service 或 k8s的service
	ServiceName string `json:"ServiceName"`
	// 宿主机的预设端口
	OriginHostPort int `json:"OriginLocalPort"`
	// 宿主机的当前端口
	CurrentHostPort int `json:"CurrentLocalPort"`
	// 端口描述
	HostPortDesc string `json:"LocalPortDesc"`
	// 容器端口
	ContainerPort int `json:"ContainerPort"`
	// 类型
	PortMapType PortMapTypeEnum `json:"PortMapType"`

	ClientPort    int `json:"ClientPort"`
	OldClientPort int `json:"OldClientPort"`
}

//
type PortMapTypeEnum string

const (
	PortMapInfo_Full        PortMapTypeEnum = "full"
	PortMapInfo_OnlyLabel   PortMapTypeEnum = "label"
	PortMapInfo_OnlyCompose PortMapTypeEnum = "compose"
	PortMapInfo_K8S_Service PortMapTypeEnum = "k8s_service"
)

//
func NewPortMap(
	mapType PortMapTypeEnum, orginLocalPort int, currentLocalPort int, localPortDesc string, containerPort int, serviceName string) *PortMapInfo {
	result := &PortMapInfo{
		ServiceName:     serviceName,
		OriginHostPort:  orginLocalPort,
		CurrentHostPort: currentLocalPort,
		HostPortDesc:    localPortDesc,
		ContainerPort:   containerPort,
		PortMapType:     mapType,
	}

	return result
}

// 配置文件路径
func (w SmartIdeConfig) GetConfigRelativeFilePath() string {
	return w.Workspace.DevContainer.configRelativeFilePath

}

// 工作目录
func (w SmartIdeConfig) GetWorkingDirectoryPath() string {
	return w.Workspace.DevContainer.workingDirectoryPath

}

//返回容器内IDE端口，web ide的默认端口：3000，JetBrains IDE的默认端口：8887
func (w SmartIdeConfig) GetContainerWebIDEPort() int {
	switch w.Workspace.DevContainer.IdeType {
	case IdeTypeEnum_VsCode:
		return model.CONST_Container_WebIDEPort
	case IdeTypeEnum_JbProjector:
		return model.CONST_Container_JetBrainsIDEPort
	case IdeTypeEnum_Opensumi:
		return model.CONST_Container_OpensumiIDEPort
	default:
		return model.CONST_Container_WebIDEPort
	}

}
