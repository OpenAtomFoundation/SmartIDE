/*
 * @Author: kenan
 * @Date: 2021-12-29 14:26:42
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-05-15 23:11:56
 * @Description: file content
 */

package common

import (
	"bufio"
	"bytes"
	"os"
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
	sshConfigPath := PathJoin(sshDirectory, "config")
	if !IsExit(sshConfigPath) { // 是否存在
		f, err := os.Create(sshConfigPath) // 创建
		if err != nil {
			return err
		}
		f.Close()
	}

	// 写入配置
	b, err := fs.CheckFileContainsStr(sshConfigPath, "StrictHostKeyChecking no") // 检查是否存在指定的文本
	if err != nil {
		return err
	}

	// 是否重置
	lines := `## smartide ##
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

// fileName:文件名字(带全路径)
// content: 写入的内容
func (fs *fileOperation) AppendToFile(fileName string, content string) error {
	// 以只写的模式，打开文件
	f, err := os.OpenFile(fileName, os.O_WRONLY, 0644)
	if err != nil {
		return err
	} else {
		// 查找文件末尾的偏移量
		n, _ := f.Seek(0, os.SEEK_END)
		// 从末尾的偏移量开始写入内容
		_, err = f.WriteAt([]byte(content), n)
	}
	defer f.Close()
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
