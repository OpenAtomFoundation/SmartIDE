/*
 * @Author: jason chen (jasonchen@leansoftx.com, http://smallidea.cnblogs.com)
 * @Description:
 * @Date: 2021-11
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-09-15 18:02:08
 */
package common

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// 判断所给路径文件/文件夹是否存在
func IsExist(path string) bool {
	if strings.TrimSpace(path) == "" {
		return false
	}

	// 替换当前用户目录
	if path[0] == '~' {
		home, _ := os.UserHomeDir()
		if home != "" {
			path = filepath.Join(home, path[1:])
		}
	}

	// 检测是否存在
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}

// 判断所给路径是否为文件夹
func IsDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

// 判断所给路径是否为文件
func IsFile(path string) bool {
	return !IsDir(path)
}

type OSType int32

const (
	OS_Windows OSType = 1
	OS_Linux   OSType = 2
)

// 使用当前系统分隔符，进行路径的拼接
func PathJoin(paths ...string) string {
	return strings.Join(paths, string(filepath.Separator))
}

// 路径组合，参数 os 可以是windows
func filePathJoin(osType OSType, paths ...string) string {
	result := filepath.Join(paths...)
	switch osType {
	case OS_Windows:
		result = strings.ReplaceAll(result, "/", "\\")
	case OS_Linux:
		result = strings.ReplaceAll(result, "\\", "/")
	}
	return result
}

// 路径组合 for linux
func FilePahtJoin4Linux(paths ...string) string {
	return filePathJoin(OS_Linux, paths...)
}
