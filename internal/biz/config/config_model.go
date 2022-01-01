/*
 * @Author: jason chen (jasonchen@leansoftx.com, http://smallidea.cnblogs.com)
 * @Description:
 * @Date: 2021-11
 * @LastEditors:
 * @LastEditTime:
 */
package config

import (
	"strings"

	"github.com/leansoftX/smartide-cli/internal/model"
	"github.com/leansoftX/smartide-cli/pkg/docker/compose"
)

// smartide 的配置
// docker-compose.yaml https://docs.docker.com/compose/compose-file/
type SmartIdeConfig struct {
	//
	Version string `yaml:"version"`
	//
	Orchestrator struct {
		// 类型，比如docker-compose、smartide、k8s
		Type string `yaml:"type"`
		// docker-compose 文件用到的版本号
		Version string `yaml:"version"`
	} `yaml:"orchestrator"`
	//
	Workspace struct {
		// 开发容器申明
		DevContainer struct {
			// 服务名称
			ServiceName string `yaml:"service-name"`
			// 端口申明
			Ports map[string]int `yaml:"ports"`
			// 容器运行起来后，在webide的terminal中执行的shell命令
			Command []string `yaml:"command"`
			// ide-type web ide类型
			IdeType string `yaml:"ide-type"`
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
		} `yaml:"dev-container"`

		// 链接的docker-compose文件路径
		DockerComposeFile string `yaml:"docker-compose-file"`
		// 允许要启动的容器，docker-compose 中的services节点
		Servcies map[string]compose.Service `yaml:"services"`
		// 组网，docker-compose 中的 Networks 节点
		Networks map[string]compose.Network `yaml:",omitempty" json:"networks,omitempty"` //
		// 挂载卷配置，docker-compose 中的 Volumes 节点
		Volumes map[string]compose.Volume `yaml:"volumes,omitempty"`
		// 密钥，docker-compose 中的 Secrets 节点
		Secrets map[string]compose.YmlSecret `yaml:"secrets,omitempty"`
	} `yaml:"workspace"`
}

// 端口映射信息
type PortMapInfo struct {
	ServiceName      string
	OriginLocalPort  int
	CurrentLocalPort int
	LocalPortDesc    string
	ContainerPort    int
	PortMapType      PortMapTypeEnum
}

//
type PortMapTypeEnum string

const (
	PortMapInfo_Full        PortMapTypeEnum = "full"
	PortMapInfo_OnlyLabel   PortMapTypeEnum = "label"
	PortMapInfo_OnlyCompose PortMapTypeEnum = "compose"
)

//
func NewPortMap(
	mapType PortMapTypeEnum, orginLocalPort int, currentLocalPort int, localPortDesc string, containerPort int, serviceName string) *PortMapInfo {
	result := &PortMapInfo{
		ServiceName:      serviceName,
		OriginLocalPort:  orginLocalPort,
		CurrentLocalPort: currentLocalPort,
		LocalPortDesc:    localPortDesc,
		ContainerPort:    containerPort,
		PortMapType:      mapType,
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
	switch strings.ToLower(w.Workspace.DevContainer.IdeType) {
	case "vscode":
		return model.CONST_Container_WebIDEPort
	case "jb-projector":
		return model.CONST_Container_JetBrainsIDEPort
	default:
		return model.CONST_Container_WebIDEPort
	}

}
