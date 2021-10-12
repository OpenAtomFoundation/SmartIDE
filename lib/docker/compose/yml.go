/*
docker-compose 文件解携，https://docs.docker.com/compose/compose-file/
*/

package compose

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

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

// 把结构化对象转换为string
func (c *DockerComposeYml) ConvertToStr() string {
	d, err := yaml.Marshal(&c)
	if err != nil {
		common.SmartIDELog.Fatal(err)
	}
	fmt.Printf("--- m dump:\n%s\n\n", string(d))

	//fmt.Print(string(dockerCompose))
	return string(d)
}

// 保存yaml文件到home目录下
func (c *DockerComposeYml) SaveFile(projectName string) (yamlFilePath string, err error) {

	dirname, err := os.UserHomeDir()
	if err != nil {
		common.SmartIDELog.Fatal(err)
	}
	yamlFileName := fmt.Sprintf("docker-compose-%s.yaml", projectName)
	yamlFileDirPath := filepath.Join(dirname, ".ide/") // current user dir + ...
	yamlFilePath = filepath.Join(yamlFileDirPath, yamlFileName)

	if !common.FileIsExit(yamlFilePath) {
		os.MkdirAll(yamlFileDirPath, os.ModePerm) // create dir
		os.Create(yamlFilePath)                   // create file
	}

	d, err := yaml.Marshal(&c)
	if err != nil {
		common.SmartIDELog.Fatal(err)
	}

	err = ioutil.WriteFile(yamlFilePath, d, 0)
	if err != nil {
		common.SmartIDELog.Fatal(err)
	}

	return yamlFilePath, err
}

//获取docker-compose.yaml文件位置
func (c *DockerComposeYml) GetDockerComposeFile(projectName string) (yamlFilePath string, err error) {

	dirname, err := os.UserHomeDir()
	if err != nil {
		common.SmartIDELog.Fatal(err)
	}
	yamlFileName := fmt.Sprintf("docker-compose-%s.yaml", projectName)
	yamlFileDirPath := filepath.Join(dirname, ".ide/") // current user dir + ...
	yamlFilePath = filepath.Join(yamlFileDirPath, yamlFileName)

	if common.FileIsExit(yamlFilePath) {
		return yamlFilePath, err
	} else {
		//common.SmartIDELog.Fatal("err")
		fmt.Sprintf("找不到docker-compose-%s.yaml", projectName)
	}

	return "", err
}
