package common

import "strings"

// 数组中是否包含
func CheckError(err error, headers ...string) {
	// 是否附加的信息中包含错误
	hasErrorMsg := false
	for _, str := range headers {
		if strings.Contains(str, "error:") {
			hasErrorMsg = true
			break
		}
	}

	// err对象应该包含错误信息
	if hasErrorMsg || (err != nil && err.Error() != "") {
		SmartIDELog.Error(err, headers...)
	}
}
