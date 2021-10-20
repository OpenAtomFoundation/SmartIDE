package common

import (
	"os"
	"path/filepath"
)

// 文件是否存在
func FileIsExit(filePath string) bool {
	_, err := os.Stat(filepath.FromSlash(filePath))
	return !os.IsNotExist(err)
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
