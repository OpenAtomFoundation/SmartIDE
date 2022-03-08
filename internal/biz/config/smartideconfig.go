/*
 * @Author: jason chen (jasonchen@leansoftx.com, http://smallidea.cnblogs.com)
 * @Description:
 * @Date: 2021-11
 * @LastEditors: kenan
 * @LastEditTime: 2022-02-16 16:31:52
 */
package config

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/leansoftX/smartide-cli/internal/model"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"gopkg.in/yaml.v2"
)

const configFileName = ".ide.config"

var SmartIdeHome string
var GlobalSmartIdeConfig GlobalConfig

//userhome下的config
type GlobalConfig struct {
	TemplateRepo    string       `yaml:"template-repo"`
	ImagesRegistry  string       `yaml:"images-registry"`
	DefaultLoginUrl string       `yaml:"default-login-url"`
	Auths           []model.Auth `yaml:"auths"`
}

func GetCurrentAuth(auths []model.Auth) model.Auth {
	for _, item := range auths {
		if item.CurrentUse {
			return item
		}
	}
	return model.Auth{}
}

//加载config配置文件
func (c *GlobalConfig) LoadConfigYaml() *GlobalConfig {
	ideConfigPath := filepath.Join(SmartIdeHome, configFileName)
	isIdeConfigExist := common.IsExit(ideConfigPath)
	if isIdeConfigExist {
		yamlByte, err := os.ReadFile(ideConfigPath)
		common.SmartIDELog.Error(err, i18nInstance.Config.Err_read_config, ideConfigPath)
		err = yaml.Unmarshal(yamlByte, &c)
		common.SmartIDELog.Error(err, i18nInstance.Config.Err_read_config, ideConfigPath)
	} else { // 如果配置文件不存在
		c.TemplateRepo = "https://gitee.com/smartide/smartide-templates.git"
		c.DefaultLoginUrl = model.CONST_LOGIN_URL
		c.ImagesRegistry = "registry.cn-hangzhou.aliyuncs.com"

		c.SaveConfigYaml()
	}

	// 检测默认登录地址是否存在
	if c.DefaultLoginUrl == "" {
		c.DefaultLoginUrl = model.CONST_LOGIN_URL
		c.SaveConfigYaml()
	}

	return c
}

//保存config
func (c *GlobalConfig) SaveConfigYaml() {
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
	GlobalSmartIdeConfig.LoadConfigYaml()
}
