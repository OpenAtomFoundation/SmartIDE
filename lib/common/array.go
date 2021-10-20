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
			slice = append(slice[:i], slice[i+1:]...)
			return RemoveEmptyItem(slice)
		}
	}
	return slice
}
