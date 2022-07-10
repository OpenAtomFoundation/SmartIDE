/*
 * @Author: jason chen (jasonchen@leansoftx.com, http://smallidea.cnblogs.com)
 * @Description:
 * @Date: 2021-11
 * @LastEditors:
 * @LastEditTime:
 */
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
