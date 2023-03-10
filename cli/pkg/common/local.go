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

package common

import (
	"errors"
	"os/exec"
	"runtime"
	"strings"

	"github.com/leansoftX/smartide-cli/internal/apk/i18n"
)

// 检查本地环境，是否安装docker、docker-compose
func CheckLocalEnv() error {
	var errMsgArray []string

	//0.1. 校验是否能正常执行docker
	dockerErr := exec.Command("docker", "-v").Run()
	dockerpsErr := exec.Command("docker", "ps").Run()
	if dockerErr != nil || dockerpsErr != nil {
		if dockerErr != nil {
			SmartIDELog.Debug(dockerErr.Error())
		}
		if dockerpsErr != nil {
			SmartIDELog.Debug(dockerpsErr.Error())
		}

		errMsgArray = append(errMsgArray, i18n.GetInstance().Main.Err_env_DockerPs)
	}

	//0.2. 校验是否能正常执行 docker-compose
	dockercomposeErr := exec.Command("docker-compose", "version").Run()
	if dockercomposeErr != nil {
		SmartIDELog.Debug(dockercomposeErr.Error())
		errMsgArray = append(errMsgArray, i18n.GetInstance().Main.Err_env_Docker_Compose)
	}

	// 错误判断
	if len(errMsgArray) > 0 {
		tmps := RemoveEmptyItem(errMsgArray)
		return errors.New(strings.Join(tmps, "; "))
	}

	return nil
}

// 检查本地环境，是否安装git
func CheckLocalGitEnv() error {
	var errMsgArray []string

	// 校验是否能正常执行 git
	gitErr := exec.Command("git", "version").Run()
	if gitErr != nil {
		errMsgArray = append(errMsgArray, i18n.GetInstance().Main.Err_env_git_check)
	}

	// 错误判断
	if len(errMsgArray) > 0 {
		return errors.New(strings.Join(errMsgArray, "; "))
	}

	return nil
}

func GetNewline() string {
	if runtime.GOOS == "windows" {
		return "\r\n"
	}
	return "\n"
}
