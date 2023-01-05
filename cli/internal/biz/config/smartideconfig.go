/*
SmartIDE - Dev Containers
Copyright (C) 2023 leansoftX.com

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package config

import (
	"fmt"
	"os"

	"github.com/leansoftX/smartide-cli/internal/model"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"gopkg.in/yaml.v2"
)

const configFileName = ".ide.config"

var SmartIdeHome string
var GlobalSmartIdeConfig GlobalConfig

type IsInsightEnabledEnum string

const (
	IsInsightEnabledEnum_None     IsInsightEnabledEnum = ""
	IsInsightEnabledEnum_Enabled  IsInsightEnabledEnum = "true"
	IsInsightEnabledEnum_Disabled IsInsightEnabledEnum = "false"
)

// userhome下的config
type GlobalConfig struct {
	TemplateActualRepoUrl string               `yaml:"template-repo" json:"template-repo"`
	ImagesRegistry        string               `yaml:"images-registry" json:"images-registry"`
	DefaultLoginUrl       string               `yaml:"default-login-url" json:"default-login-url"`
	Auths                 []model.Auth         `yaml:"auths" json:"auths"`
	IsInsightEnabled      IsInsightEnabledEnum `yaml:"isInsight" json:"isInsight"`
}

func GetCurrentAuth(auths []model.Auth) model.Auth {
	for _, item := range auths {
		if item.CurrentUse {
			return item
		}
	}
	return model.Auth{}
}

// 加载config配置文件
func (c *GlobalConfig) LoadConfigYaml() *GlobalConfig {
	ideConfigPath := common.PathJoin(SmartIdeHome, configFileName)
	isIdeConfigExist := common.IsExist(ideConfigPath)
	if isIdeConfigExist {
		yamlByte, err := os.ReadFile(ideConfigPath)
		common.SmartIDELog.Error(err, i18nInstance.Config.Err_read_config, ideConfigPath)
		err = yaml.Unmarshal(yamlByte, &c)
		common.SmartIDELog.Error(err, i18nInstance.Config.Err_read_config, ideConfigPath)
	} else { // 如果配置文件不存在
		c.TemplateActualRepoUrl = "https://gitee.com/smartide/smartide-templates.git"
		c.DefaultLoginUrl = model.CONST_LOGIN_URL
		c.ImagesRegistry = "registry.cn-hangzhou.aliyuncs.com"

		c.SaveConfigYaml()
	}

	// 检测默认登录地址是否存在
	if c.DefaultLoginUrl == "" {
		c.DefaultLoginUrl = model.CONST_LOGIN_URL
		c.SaveConfigYaml()
	}

	// 默认开启
	if c.IsInsightEnabled == "" {
		c.IsInsightEnabled = IsInsightEnabledEnum_Enabled
	}

	// 加密
	for _, auth := range c.Auths {

		if auth.Token != "" {
			common.SmartIDELog.AddEntryptionKeyWithReservePart(fmt.Sprint(auth.Token))
		}
	}

	return c
}

// 保存config
func (c *GlobalConfig) SaveConfigYaml() {
	ideConfigPath := common.PathJoin(SmartIdeHome, configFileName)
	templatesByte, err := yaml.Marshal(&c)
	common.SmartIDELog.Error(err)
	err = os.WriteFile(ideConfigPath, templatesByte, 0777)
	common.SmartIDELog.Error(err)
}

func init() {
	//全局idehome
	home, err := os.UserHomeDir()
	common.CheckError(err)
	SmartIdeHome = common.PathJoin(home, ".ide")

	//创建userhome/.ide
	templatesFolderIsExist := common.IsExist(SmartIdeHome)
	if !templatesFolderIsExist {
		err = os.MkdirAll(SmartIdeHome, os.ModePerm)
		common.CheckError(err)
	}

	//全局userhome/.ide/.ide.config
	GlobalSmartIdeConfig.LoadConfigYaml()
}
