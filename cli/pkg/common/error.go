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
	"strings"
)

// 数组中是否包含
func CheckError(err error, headers ...string) {
	errFunc := func(err error) {}
	CheckErrorFunc(err, errFunc, headers...)
}

// 是否为退出错误
func IsExitError(err error) bool {
	if err == nil {
		return false
	}
	_, isExitError := err.(*exec.ExitError)
	if isExitError && err.Error() != "" {
		SmartIDELog.Debug(err.Error())
	}
	return isExitError
}

// 数组中是否包含
func CheckErrorFunc(err error, afterFunc func(err error), headers ...string) {
	// 是否附加的信息中包含错误
	hasErrorMsg := false
	for _, str := range headers {
		if strings.Contains(str, "error:") {
			hasErrorMsg = true
			break
		}
	}

	//
	afterFunc(err)

	// err对象应该包含错误信息
	if hasErrorMsg || err != nil {
		if err == nil {
			err = errors.New("error_ ")
		}
		SmartIDELog.Error(err, headers...)
	}

}
