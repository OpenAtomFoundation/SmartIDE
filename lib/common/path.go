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
