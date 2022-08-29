/*
 * @Author: jason chen (jasonchen@leansoftx.com, http://smallidea.cnblogs.com)
 * @Description:
 * @Date: 2021-11
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-08-18 11:06:52
 */
package config

import (
	"strings"

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
	OrchestratorTypeEnum_K8S      OrchestratorTypeEnum = "k8s"
	OrchestratorTypeEnum_Compose  OrchestratorTypeEnum = "docker-compose"
	OrchestratorTypeEnum_Allinone OrchestratorTypeEnum = "allinone"
)

type IdeTypeEnum string

const (
	IdeTypeEnum_VsCode      IdeTypeEnum = "vscode"
	IdeTypeEnum_JbProjector IdeTypeEnum = "jb-projector"
	IdeTypeEnum_Opensumi    IdeTypeEnum = "opensumi"
	IdeTypeEnum_Theia       IdeTypeEnum = "theia"
	IdeTypeEnum_SDKOnly     IdeTypeEnum = "sdk-only"
)

// 容器配置
type ContainerConfig struct {
	// 持久化配置列表
	PersistentVolumes []PersistentVolumeConfig `yaml:"persistentVolumes"`
}

type PersistentVolumeDirectoryTypeEnum string

const (
	PersistentVolumeDirectoryTypeEnum_Project PersistentVolumeDirectoryTypeEnum = "project"

	PersistentVolumeDirectoryTypeEnum_DbData PersistentVolumeDirectoryTypeEnum = "database"
	PersistentVolumeDirectoryTypeEnum_Agent  PersistentVolumeDirectoryTypeEnum = "agent"
	PersistentVolumeDirectoryTypeEnum_Other  PersistentVolumeDirectoryTypeEnum = "other"
)

// 持久化配置
type PersistentVolumeConfig struct {
	// 容器内的目录，映射到外部持久化存储介质
	MountPath string `yaml:"mountPath"`
	// volume 类型
	DirectoryType PersistentVolumeDirectoryTypeEnum `yaml:"directoryType"`
}

// 自定义的bool类型，用于兼容多种的bool设置方式
type CustomBool string

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
		HasGitConfig CustomBool `yaml:"git-config"`
		HasSshKey    CustomBool `yaml:"ssh-key"`
	} `yaml:"volumes"`

	// 绑定的端口列表
	bindingPorts []PortMapInfo

	// 默认 yaml配置文件的相对路径（相对于工作目录）
	configRelativeFilePath string //= CONST_Default_ConfigRelativeFilePath

	// 本地模式时，本地工作目录（相对于工作目录，但是在.ide.yaml中写的是相对于“.ide”目录，后续会通过程序进行转换）
	workingDirectoryPath string //= "."

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

		// 容器申明，不是必须的
		Containers map[string]ContainerConfig `yaml:"containers"`

		// 链接的docker-compose文件路径
		DockerComposeFile string `yaml:"docker-compose-file"`

		// k8s 的部署文件（通配符）
		KubeDeployFileExpression string `yaml:"kube-deploy-files,omitempty"`

		// 允许要启动的容器，docker-compose 中的services节点
		Servcies map[string]compose.Service `yaml:"services,omitempty"`
		// 组网，docker-compose 中的 Networks 节点
		Networks map[string]compose.Network `yaml:"networks,omitempty"`
		// 挂载卷配置，docker-compose 中的 Volumes 节点
		Volumes map[string]compose.Volume `yaml:"volumes,omitempty"`
		// 密钥，docker-compose 中的 Secrets 节点
		Secrets map[string]compose.YmlSecret `yaml:"secrets,omitempty"`

		// 链接的compose配置
		LinkCompose *compose.DockerComposeYml
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

		// 容器申明，不是必须的
		Containers map[string]ContainerConfig `yaml:"containers"`

		// k8s 的部署文件（通配符）
		KubeDeployFileExpression string `yaml:"kube-deploy-files,omitempty"`

		//
		//Namespace coreV1.Namespace

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
	// 宿主机的预设 SSH 端口, Origin ==> Preset/Default, not Remote
	OriginHostPort int `json:"OriginLocalPort"`
	// 宿主机的当前 SSH 端口
	CurrentHostPort int `json:"CurrentLocalPort"`
	// 端口描述
	HostPortDesc string `json:"LocalPortDesc"`
	// 容器端口
	ContainerPort int `json:"ContainerPort"`
	// 类型
	PortMapType PortMapTypeEnum `json:"PortMapType"`

	// 本地绑定的端口
	ClientPort int `json:"ClientPort"`
	// 本地绑定的旧端口
	OldClientPort int `json:"OldClientPort"`

	// 关联的目录
	RefDirecotry string `json:"RefDirecotry"`

	// 是否允许开放Ingress
	//EnableIngressUrl bool `json:"EnableIngressUrl"`

	// 关联的ingress url（公网地址）
	IngressUrl string `json:"IngressUrl"`

	// 关联的SSH External Port
	SSHPort string `json:"SSHPort"`

	//端口号是否已连接
	IsConnected bool `json:"isConnected"`
}

// GetSSHPortAtLocalHost 获取localhost上的ssh端口
func (p PortMapInfo) GetSSHPortAtLocalHost() int {
	return p.ClientPort
}

//
type PortMapTypeEnum string

const (
	PortMapInfo_Full        PortMapTypeEnum = "full"
	PortMapInfo_OnlyLabel   PortMapTypeEnum = "label"
	PortMapInfo_OnlyCompose PortMapTypeEnum = "compose"
	PortMapInfo_K8S_Service PortMapTypeEnum = "k8s_service"
)

func (customBool CustomBool) Value() bool {
	/* switch customBool.(type) {
	case int:
		tmp := customBool.(int)
		return tmp == 1
	case string:
		tmp := strings.ToLower(customBool.(string))
		return tmp == "true" || tmp == "1"
	case bool:
		return customBool.(bool)
	} */

	tmp := strings.ToLower(string(customBool))
	return tmp == "true" || tmp == "1"
}

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
func (w SmartIdeConfig) GetContainerWebIDEPort() (port *int) {
	switch w.Workspace.DevContainer.IdeType {
	case IdeTypeEnum_VsCode:
		tmp := model.CONST_Container_WebIDEPort
		port = &tmp
	case IdeTypeEnum_JbProjector:
		tmp := model.CONST_Container_JetBrainsIDEPort
		port = &tmp
	case IdeTypeEnum_Opensumi:
		tmp := model.CONST_Container_OpensumiIDEPort
		port = &tmp
		/* default:
		return -1 */
	}

	return port
}

//
func (originK8sConfig SmartIdeK8SConfig) GetProjectDirctory() string {
	containerGitCloneDir := "/home/project"
	for containerName, item := range originK8sConfig.Workspace.Containers { // git clone 的目录是否设置
		if containerName == originK8sConfig.Workspace.DevContainer.ServiceName {
			for _, volume := range item.PersistentVolumes {
				if volume.DirectoryType == PersistentVolumeDirectoryTypeEnum_Project {
					containerGitCloneDir = volume.MountPath
					break
				}
			}
			break
		}
	}
	return containerGitCloneDir
}
