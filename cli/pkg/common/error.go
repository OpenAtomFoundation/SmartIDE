/*
 * @Author: jason chen (jasonchen@leansoftx.com, http://smallidea.cnblogs.com)
 * @Description:
 * @Date: 2021-11
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-06-08 11:58:34
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
