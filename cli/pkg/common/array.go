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

import "strings"

// 数组中是否包含
func Contains(slice []string, item string) bool {
	set := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		set[s] = struct{}{}
	}

	_, ok := set[item]
	return ok
}

// 数组中是否包含(模糊匹配)
func Contains4StringArry(slice []string, item string) bool {
	for _, s := range slice {
		if strings.Contains(s, item) {
			return true
		}

	}

	return false
}

// 数组中是否包含某个元素
func Contains4Int(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// 剔除空元素
func RemoveEmptyItem(slice []string) []string {
	if len(slice) == 0 {
		return slice
	}
	for i, v := range slice {
		if strings.TrimSpace(v) == "" {
			if i+1 > len(slice) {
				slice = slice[:i]
			} else {
				slice = append(slice[:i], slice[i+1:]...)
			}
			return RemoveEmptyItem(slice)
		}
	}
	return slice
}

// 剔除空元素
func RemoveItem(slice []string, item string) []string {
	if len(slice) == 0 {
		return slice
	}
	for i, v := range slice {
		if v == item {
			if i+1 > len(slice) {
				slice = slice[:i]
			} else {
				slice = append(slice[:i], slice[i+1:]...)
			}
			return slice
		}
	}
	return slice
}

/**
 * 数组去重 去空
 */
func RemoveDuplicatesAndEmpty(a []string) (ret []string) {
	a_len := len(a)
	for i := 0; i < a_len; i++ {
		if (i > 0 && a[i-1] == a[i]) || len(a[i]) == 0 {
			continue
		}
		ret = append(ret, a[i])
	}
	return
}
