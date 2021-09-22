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

	"github.com/leansoftX/smartide-cli/cmd"
	"github.com/leansoftX/smartide-cli/lib/common"
)

func main() {

	cmd.Execute()
}

// running before main
func init() {
	//TODO 日志存储在当前用户的目录下
	wd, _ := os.Getwd()
	logFilePath := filepath.Join(wd, "smartide.log")

	if !common.FileIsExit(logFilePath) {
		os.MkdirAll(logFilePath, os.ModeAppend)
	}

	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("open log file failed, err:", err)
		return
	}
	log.SetOutput(logFile)
	log.SetFlags(log.Llongfile | log.Lmicroseconds | log.Ldate)
}
