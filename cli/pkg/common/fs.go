/*
 * @Author: kenan
 * @Date: 2021-12-29 14:26:42
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-11-04 10:49:55
 * @Description: file content
 */

package common

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type fileOperation struct{}

// 文件操作相关
var FS fileOperation

func init() {
	FS = fileOperation{}

}

// 跳过host检查
func (fs *fileOperation) SkipStrictHostKeyChecking(sshDirectory string, isReset bool) error {
	sshConfigPath := filepath.Join(sshDirectory, "config")

	//  是否存在
	if !IsExist(sshConfigPath) {
		SmartIDELog.Debug(fmt.Sprintf("创建 %v 文件", sshConfigPath))

		// 文件夹是否存在
		dirPath := filepath.Dir(sshConfigPath)
		if dirPath != "" && !IsExist(dirPath) {
			err := os.MkdirAll(dirPath, os.ModePerm)
			if err != nil {
				return err
			}
		}

		// 创建文件
		f, err := os.Create(sshConfigPath)
		if err != nil {
			return err
		}
		defer f.Close()
	}

	// 写入配置
	b, err := fs.CheckFileContainsStr(sshConfigPath, "## smartide StrictHostKeyChecking ##") // 检查是否存在指定的文本
	if err != nil {
		return err
	}

	// 是否重置
	lines := `## smartide StrictHostKeyChecking ##
HOST *
StrictHostKeyChecking no`
	if isReset {
		fileContentBytes, err := os.ReadFile(sshConfigPath)
		if err != nil {
			return err
		}
		fileContent := string(fileContentBytes)
		if strings.Contains(fileContent, lines) {
			fileContent = strings.Replace(fileContent, lines, "", -1)
			return os.WriteFile(sshConfigPath, []byte(fileContent), 0644)
		}

	} else {
		if !b {
			return fs.AppendToFile(sshConfigPath, lines) // 写入
		}

	}

	return nil
}

func (fs *fileOperation) CreateOrOverWrite(filePath string, content string) error {
	return writeToFile(filePath, content, true)
}

// fileName:文件名字(带全路径)
// content: 写入的内容
func (fs *fileOperation) AppendToFile(filePath string, content string) error {
	return writeToFile(filePath, content, false)
}

// 指定路径是否存在
func (fs *fileOperation) IsExist(filePath string) bool {
	_, err := os.Stat(filePath)
	if err != nil {
		return false
	}
	return true
}

func writeToFile(filePath string, content string, isOverWrite bool) (err error) {

	if filePath == "" {
		return errors.New("文件路径不能为空！")
	}

	// 替换当前用户目录
	if filePath[0] == '~' {
		home, _ := os.UserHomeDir()
		if home != "" {
			filePath = filepath.Join(home, filePath[1:])
		}
	}

	// 文件是否存在
	if !IsExist(filePath) {

		// 文件夹是否存在
		dirPath := filepath.Dir(filePath)
		if dirPath != "" && !IsExist(dirPath) {
			err := os.MkdirAll(dirPath, os.ModePerm)
			if err != nil {
				return err
			}
		}

		// 创建文件
		f, err := os.Create(filePath)
		if err != nil {
			return err
		}
		defer f.Close()
	}

	// 文件内容
	if isOverWrite {
		err = os.WriteFile(filePath, []byte(content), 0644)
	} else { // 附加到文件中
		// 以只写的模式，打开文件
		f, err := os.OpenFile(filePath, os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		// 查找文件末尾的偏移量
		n, _ := f.Seek(0, os.SEEK_END)
		// 从末尾的偏移量开始写入内容
		_, err = f.WriteAt([]byte(content), n)
		if err != nil {
			return err
		}
		defer f.Close()
	}

	return err
}

var totalSize float64

// 检查指定文件中是否包含文本
func (fs *fileOperation) CheckFileContainsStr(path, str string) (bool, error) {
	b := false
	cmp := []byte(str)
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()
	input := bufio.NewScanner(f)

	for input.Scan() {
		info := input.Bytes()
		if bytes.Contains(info, cmp) {
			b = true
			break
		}

		tokens := bytes.SplitN(input.Bytes(), []byte(" "), 10)
		totalSize += float64(len(input.Bytes()))
		if len(tokens) < 9 {
			continue
		}
		strconv.ParseInt(string(tokens[8]), 10, 0)
	}
	totalSize = totalSize / (1 << 20)
	return b, nil
}
