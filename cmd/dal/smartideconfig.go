package dal

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/leansoftX/smartide-cli/lib/common"
	"gopkg.in/yaml.v2"
)

const configFileName = ".ide.config"

var SmartIdeHome string
var SmartIdeConfig Config

//userhome下的config
type Config struct {
	TemplateRepo   string `yaml:"template-repo"`
	ImagesRegistry string `yaml:"images-registry"`
}

//加载config配置文件
func (c *Config) LoadConfigYaml() *Config {
	ideConfigPath := filepath.Join(SmartIdeHome, configFileName)
	isIdeConfigExist := common.IsExit(ideConfigPath)
	if isIdeConfigExist {
		yamlByte, err := os.ReadFile(ideConfigPath)
		common.SmartIDELog.Error(err, i18nInstance.Config.Err_read_config, ideConfigPath)
		err = yaml.Unmarshal(yamlByte, &c)
		common.SmartIDELog.Error(err, i18nInstance.Config.Err_read_config, ideConfigPath)
	} else {
		c.TemplateRepo = "https://gitee.com/smartide/smartide-templates.git"
		// c.TemplateRepo = "https://github.com/smartide/smartide-templates.git"
		// c.ImagesRegistry = "docker.io"
		c.ImagesRegistry = "registry.cn-hangzhou.aliyuncs.com"
		c.SaveConfigYaml()
	}
	return c
}

//保存config
func (c *Config) SaveConfigYaml() {
	ideConfigPath := filepath.Join(SmartIdeHome, configFileName)
	templatesByte, err := yaml.Marshal(&c)
	common.SmartIDELog.Error(err)
	err = ioutil.WriteFile(ideConfigPath, templatesByte, 0777)
	common.SmartIDELog.Error(err)
}

func init() {
	//全局idehome
	home, err := os.UserHomeDir()
	common.CheckError(err)
	SmartIdeHome = filepath.Join(home, ".ide")

	//创建userhome/.ide
	templatesFolderIsExist := common.IsExit(SmartIdeHome)
	if !templatesFolderIsExist {
		err = os.MkdirAll(SmartIdeHome, os.ModePerm)
		common.CheckError(err)
	}

	//全局userhome/.ide/.ide.config
	SmartIdeConfig.LoadConfigYaml()
}
