/*
SmartIDE - Dev Containers
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
	"regexp"
	"strconv"
)

var numPattern = regexp.MustCompile(`^\d+$|^\d+[.]\d+$`)

/*
*
判断字符串是否为 纯数字 包涵浮点型
*/
func IsNumber(s string) bool {
	_, err := strconv.Atoi(s)
	return numPattern.MatchString(s) && err == nil
}
