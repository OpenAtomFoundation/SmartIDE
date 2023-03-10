/*
SmartIDE - CLI
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
	} else if stable != "" && !strings.Contains(strings.ToLower(stable), "v") {
		smartVersion.VersionNumber = "v" + smartVersion.VersionNumber
	}

	// commit id 截取最后8位
	if len(smartVersion.TargetCommitish) > 8 {
		smartVersion.TargetCommitishShort = smartVersion.TargetCommitish[(len(smartVersion.TargetCommitish) - 8):]
	}

	return smartVersion

}
