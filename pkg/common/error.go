/*
 * @Author: jason chen (jasonchen@leansoftx.com, http://smallidea.cnblogs.com)
 * @Description:
 * @Date: 2021-11
 * @LastEditors:
 * @LastEditTime:
 */
package common

import "strings"

// 数组中是否包含
func CheckError(err error, headers ...string) {
	errFunc := func(err error) {}
	CheckErrorFunc(err, errFunc, headers...)
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
	if hasErrorMsg || (err != nil && err.Error() != "") {
		SmartIDELog.Error(err, headers...)
	}

}
