/*
 * @Author: jason chen (jasonchen@leansoftx.com, http://smallidea.cnblogs.com)
 * @Description:
 * @Date: 2021-11
 * @LastEditors:
 * @LastEditTime:
 */
package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/leansoftX/smartide-cli/internal/apk/i18n"
	"github.com/leansoftX/smartide-cli/internal/model"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"gopkg.in/yaml.v2"
)

// 国际化
var i18nInstance = i18n.GetInstance()

func NewConfigRemote(localWorkingDir string, configFilePath string, configContent string) (result *SmartIdeConfig) {
	return newConfig(localWorkingDir, configFilePath, configContent, true)
}

// 构造函数
func NewConfig(localWorkingDir string, configFilePath string, configContent string) (result *SmartIdeConfig) {
	return newConfig(localWorkingDir, configFilePath, configContent, false)
}

func newConfig(localWorkingDir string, configFilePath string, configContent string, isRemote bool) (result *SmartIdeConfig) {
	result = &SmartIdeConfig{}

	// 配置文件的路径
	if len(configFilePath) <= 0 {
		configFilePath = model.CONST_Default_ConfigRelativeFilePath
	}

	// 工作目录的路径
	if len(localWorkingDir) <= 0 {
		dirName, err := os.Getwd()
		if err != nil {
			common.SmartIDELog.Fatal(err)
		}
		localWorkingDir = dirName
	}

	// 加载配置
	if configContent != "" {
		err := yaml.Unmarshal([]byte(configContent), &result)
		common.CheckError(err)

	} else {
		if !isRemote { // 只有本地模式下，才会从本地加载配置文件
			result = loadConfigWithYamlFile(localWorkingDir, configFilePath)

		}

	}

	// 私有成员赋值
	result.Workspace.DevContainer.configRelativeFilePath = configFilePath
	result.Workspace.DevContainer.workingDirectoryPath = localWorkingDir

	// 配置文件的相对路径
	result.Workspace.DevContainer.configRelativeFilePath =
		strings.Replace(result.Workspace.DevContainer.configRelativeFilePath, result.Workspace.DevContainer.workingDirectoryPath, "", -1)

		// 置为空
	result.Workspace.DevContainer.bindingPorts = []PortMapInfo{}

	return result
}

// 从yaml文件中获取配置
func loadConfigWithYamlFile(workingDirectoryPath, configRelativeFilePath string) (result *SmartIdeConfig) {

	result = &SmartIdeConfig{}
	configFilePath := filepath.Join(workingDirectoryPath, configRelativeFilePath)

	// check file exit
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		common.SmartIDELog.Error(err, i18nInstance.Main.Err_file_not_exit, configFilePath)
	}

	// read
	yamlFile, err := ioutil.ReadFile(configFilePath)
	common.CheckError(err)

	// parse
	err = yaml.Unmarshal(yamlFile, &result)
	common.CheckError(err)

	// 配置文件格式验证
	err = result.Valid()
	common.CheckError(err)

	return result
}
