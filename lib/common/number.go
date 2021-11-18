package common

import (
	"regexp"
	"strconv"
)

var numPattern = regexp.MustCompile(`^\d+$|^\d+[.]\d+$`)

/**
判断字符串是否为 纯数字 包涵浮点型
*/
func IsNumber(s string) bool {
	_, err := strconv.Atoi(s)
	return numPattern.MatchString(s) && err == nil
}
