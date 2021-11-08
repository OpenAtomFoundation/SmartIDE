/*
docker-compose 文件解携，https://docs.docker.com/compose/compose-file/
*/

package compose

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
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
func (c *DockerComposeYml) ConvertToStr() string {
	d, err := yaml.Marshal(&c)
	if err != nil {
		common.SmartIDELog.Fatal(err)
	}
	common.SmartIDELog.Info("--- m dump:\n%s\n\n", string(d))

	//fmt.Print(string(dockerCompose))
	return string(d)
}

// 获取docker compose文件名称
func (c *DockerComposeYml) GetTmpDockerComposeFilePath(projectName string) string {
	dirname, err := os.Getwd()
	if err != nil {
		common.SmartIDELog.Fatal(err)
	}

	dockerComposeFileName := fmt.Sprintf("docker-compose-%s.yaml", projectName)
	yamlFileDirPath := filepath.Join(dirname, ".ide/tmp/")

	yamlFilePath := filepath.Join(yamlFileDirPath, dockerComposeFileName)
	return yamlFilePath
}

// 保存docker compose yaml文件到home目录下
func (c *DockerComposeYml) SaveFile(dockerComposeFilePath string) (err error) {
	// 检查临时文件 是否写入到	.gitignore
	checkGitignoreContainTmpDir()

	// create dir
	dockerComposeFileDir := filepath.Dir(dockerComposeFilePath) // docker compose 文件所在的目录
	if !common.IsExit(dockerComposeFileDir) {
		os.MkdirAll(dockerComposeFileDir, os.ModePerm)
		common.SmartIDELog.Info("创建目录：" + dockerComposeFileDir)
	}

	// create file
	if !common.IsExit(dockerComposeFilePath) {
		os.Create(dockerComposeFilePath)
	}

	d, err := yaml.Marshal(&c)
	if err != nil {
		common.SmartIDELog.Fatal(err)
	}

	err = ioutil.WriteFile(dockerComposeFilePath, d, 0)
	if err != nil {
		common.SmartIDELog.Fatal(err)
	}

	return err
}

// 检测是否存在包含 tmp 的 .gitignore文件
func checkGitignoreContainTmpDir() {
	dirname, err := os.Getwd()
	if err != nil {
		common.SmartIDELog.Fatal(err)
	}

	gitignorePath := filepath.Join(dirname, ".ide/.gitignore")
	if !common.IsExit(gitignorePath) {
		dirPath := filepath.Dir(gitignorePath)

		if !common.IsExit(dirPath) {
			os.Create(dirPath)
		}

		err := ioutil.WriteFile(gitignorePath, []byte("/tmp/"), 0666)
		common.CheckError(err)
	} else {
		bytes, _ := ioutil.ReadFile(gitignorePath)
		content := string(bytes)
		if !strings.Contains(content, "/tmp/") { // 不包含时，要附加
			err = ioutil.WriteFile(gitignorePath, []byte(content+"\n/tmp/"), 0666)
			common.CheckError(err)
		}
	}
}
