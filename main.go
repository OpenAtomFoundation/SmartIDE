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
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/leansoftX/smartide-cli/cmd"
	"github.com/leansoftX/smartide-cli/lib/common"
	"github.com/leansoftX/smartide-cli/lib/i18n"

	_ "embed"
)

func main() {

	currentVersion := formatVerion()
	fmt.Print(currentVersion)

	cmd.Execute(currentVersion)
}

// running before main
func init() {
	dirname, err := os.UserHomeDir() // home dir
	if err != nil {
		log.Fatal(err)
	}
	logFilePath := filepath.Join(dirname, ".ide/smartide.log") // current user dir + ...

	if !common.FileIsExit(logFilePath) {
		os.MkdirAll(filepath.Join(dirname, ".ide"), os.ModePerm) // create dir
		os.Create(logFilePath)                                   // create file
	}

	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("open log file failed, err:", err)
		return
	}
	log.SetOutput(logFile)
	log.SetFlags(log.Llongfile | log.Lmicroseconds | log.Ldate)

}

//go:embed stable.txt
var version string

// 格式化版本号，在stable.txt文件中读取
// 注：embed 不支持 “..”， 即上级目录
func formatVerion() string {

	if version == "$(version)" {

		common.SmartIDELog.Warning(i18n.GetInstance().Main.Error.Version_not_build)

	} else if strings.ToLower(version[0:1]) != "v" {

		return "" + version + "\n"

	}

	return version + "\n"

}
