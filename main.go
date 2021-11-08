/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import (
	"encoding/json"
	"strings"

	"github.com/leansoftX/smartide-cli/cmd"
	"github.com/leansoftX/smartide-cli/lib/common"
	"github.com/leansoftX/smartide-cli/lib/i18n"

	_ "embed"
)

func main() {
	// print version
	versionInfo := formatVerion()
	common.SmartIDELog.Console(versionInfo.VersionNumber)

	// command line startup
	cmd.Execute(versionInfo)
}

// running before main
func init() {

	common.SmartIDELog.InitLogger("")

}

//go:embed stable.txt
var stable string

//go:embed stable.json
var stableJson string

// 格式化版本号，在stable.txt文件中读取
// 注：embed 不支持 “..”， 即上级目录
func formatVerion() (smartVersion cmd.SmartVersion) {

	// 转换为结构体
	json.Unmarshal([]byte(stableJson), &smartVersion)

	// 版本号赋值
	smartVersion.VersionNumber = stable
	if stable == "$(version)" {
		common.SmartIDELog.Warning(i18n.GetInstance().Main.Error.Version_not_build)
	} else if strings.ToLower(stable[0:1]) != "v" {
		smartVersion.VersionNumber = "v" + smartVersion.VersionNumber
	}

	// commit id 截取最后8位
	if len(smartVersion.TargetCommitish) > 8 {
		smartVersion.TargetCommitishShort = smartVersion.TargetCommitish[(len(smartVersion.TargetCommitish) - 8):]
	}

	return smartVersion

}
