/*
 * @Date: 2022-07-11 15:38:06
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-08-16 15:20:13
 * @FilePath: /cli/pkg/docker/compose/yml.go
 */
/*
docker-compose 文件解携，https://docs.docker.com/compose/compose-file/
*/

package compose

import (
	"strings"

	"github.com/leansoftX/smartide-cli/internal/model"
	"github.com/leansoftX/smartide-cli/pkg/common"
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

func (c *DockerComposeYml) IsNil() bool {
	return c.Version == "" && len(c.Services) == 0 && len(c.Volumes) == 0 && len(c.Networks) == 0 && len(c.Secrets) == 0
}

func (c *DockerComposeYml) IsNotNil() bool {
	return !c.IsNil()
}

/* // 从compose中 获取宿主端口（可能会变） 和 容器端口的绑定关系
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
} */

/* // 获取被绑定的本地端口
func (c *DockerComposeYml) GetHostBindingPorts() []int {
	portBindings := c.GetPortBindings()
	resultLocalPorts := []int{}

	for localPort := range portBindings {
		localPortInt, err := strconv.Atoi(localPort)
		common.CheckError(err)
		if !common.Contains4Int(resultLocalPorts, localPortInt) {
			resultLocalPorts = append(resultLocalPorts, localPortInt)

		}

	}

	return resultLocalPorts
} */

func (c *DockerComposeYml) GetSSHPassword(devService string) string {
	for serviceName, service := range c.Services {
		if serviceName == devService {
			return service.Environment[model.CONST_ENV_NAME_LoalUserPassword]
		}
	}

	return ""
}

// 把结构化对象转换为string
func (c *DockerComposeYml) ToYaml() (result string, err error) {
	if c == nil {
		return
	}
	if c.IsNil() {
		return
	}

	//fmt.Println(c)

	d, err := yaml.Marshal(&c)
	common.CheckError(err)
	result = string(d)

	result = strings.ReplaceAll(result, "\\'", "'")

	//result = strings.ReplaceAll(result, "\"", "\\\"") // 文本中包含双引号
	return result, nil
}
