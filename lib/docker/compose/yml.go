/*
docker-compose 文件解携，https://docs.docker.com/compose/compose-file/
*/

package compose

import (
	"strconv"
	"strings"

	"github.com/leansoftX/smartide-cli/lib/common"
	yaml "gopkg.in/yaml.v2"
)

// 完整的配置文件
type DockerComposeYml struct {
	Version  string               `yaml:"version"`            // 版本号
	Services map[string]Service   `yaml:"services"`           // 服务配置
	Volumes  map[string]Volume    `yaml:"volumes,omitempty"`  // 挂载卷配置
	Networks map[string]Network   `yaml:"networks,omitempty"` // 网络配置
	Secrets  map[string]YmlSecret `yaml:"secrets,omitempty"`  // 密钥
	// TODO configs
	// TODO Variable substitution
	// TODO Extension fields

	//SmartIDE SmartIDE `yaml:"smartide,omitempty"` // 一些自定义的信息
}

//
func (c *DockerComposeYml) IsNil() bool {
	return c.Version == "" && len(c.Services) == 0 && len(c.Volumes) == 0 && len(c.Networks) == 0 && len(c.Secrets) == 0
}

//
func (c *DockerComposeYml) IsNotNil() bool {
	return !c.IsNil()
}

// 获取本地端口 和 容器端口的绑定关系
func (c *DockerComposeYml) GetPortBindings() map[string]string {
	var result map[string]string = map[string]string{}
	for _, service := range c.Services {
		for _, portBinding := range service.Ports {
			ports := strings.Split(portBinding, ":")
			if len(ports) == 2 {
				result[ports[0]] = ports[1]
			} else {
				result[ports[0]] = ports[0]
			}
		}
	}
	return result
}

// 获取被绑定的本地端口
func (c *DockerComposeYml) GetLocalBindingPorts() []int {
	portBindings := c.GetPortBindings()
	result := []int{}

	for key, _ := range portBindings {
		port, err := strconv.Atoi(key)
		common.CheckError(err)
		result = append(result, port)
	}

	return result
}

// 把结构化对象转换为string
func (c *DockerComposeYml) ToString() string {
	if c.IsNil() {
		return ""
	}

	d, err := yaml.Marshal(&c)
	common.CheckError(err)

	return string(d)
}
