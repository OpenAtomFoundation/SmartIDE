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
	"fmt"
	"strings"
	"time"

	"github.com/leansoftX/smartide-cli/cmd"
	"github.com/leansoftX/smartide-cli/internal/apk/i18n"
	"github.com/leansoftX/smartide-cli/pkg/common"

	_ "embed"
)

func main() {
	//
	defer func() {
		if err := recover(); err != nil {
			common.SmartIDELog.Fatal(err)
		}
	}()

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

var BuildTime string

// 格式化版本号，在stable.txt文件中读取
// 注：embed 不支持 “..”， 即上级目录
func formatVerion() (smartVersion cmd.SmartVersion) {

	// 转换为结构体
	json.Unmarshal([]byte(stableJson), &smartVersion)

	// 编译时间
	smartVersion.BuildTime, _ = time.ParseInLocation("2006-01-02 15:04:05", BuildTime, time.Local)

	// 版本号赋值
	smartVersion.VersionNumber = stable
	if stable == "$(version)" {
		smartVersion.VersionNumber = fmt.Sprintf(i18n.GetInstance().Main.Info_version_local, smartVersion.BuildTime.Format("2006-01-02 15:04:05"))
		common.SmartIDELog.Importance(i18n.GetInstance().Main.Err_version_not_build)
	} else if strings.ToLower(stable[0:1]) != "v" {
		smartVersion.VersionNumber = "v" + smartVersion.VersionNumber
	}

	// commit id 截取最后8位
	if len(smartVersion.TargetCommitish) > 8 {
		smartVersion.TargetCommitishShort = smartVersion.TargetCommitish[(len(smartVersion.TargetCommitish) - 8):]
	}

	return smartVersion

}
