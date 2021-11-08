package common

// 数组中是否包含
func CheckError(err error, headers ...string) {
	if err != nil {
		SmartIDELog.Error(err, headers...)
	}
}
