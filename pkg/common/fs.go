/*
 * @Author: kenan
 * @Date: 2021-12-29 14:26:42
 * @LastEditors: kenan
 * @LastEditTime: 2022-01-13 17:46:33
 * @Description: file content
 */

package common

import (
	"bufio"
	"bytes"
	"os"
	"strconv"
)

// fileName:文件名字(带全路径)
// content: 写入的内容
func AppendToFile(fileName string, content string) error {
	// 以只写的模式，打开文件
	f, err := os.OpenFile(fileName, os.O_WRONLY, 0644)
	if err != nil {
		SmartIDELog.Error("cacheFileList.yml file create failed. err: " + err.Error())
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

func CheckFileContainsStr(path, str string) (bool, error) {
	b := false
	cmp := []byte(str)
	// log.Println(cmp)
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()
	input := bufio.NewScanner(f)
	for input.Scan() {
		//log.Println(input.Bytes())
		info := input.Bytes()
		if bytes.Contains(info, cmp) {
			// log.Println(info)
			b = true
			break
		}

		tokens := bytes.SplitN(input.Bytes(), []byte(" "), 10)
		//log.Println(tokens)
		totalSize += float64(len(input.Bytes()))
		if len(tokens) < 9 {
			continue
		}
		strconv.ParseInt(string(tokens[8]), 10, 0)
	}
	totalSize = totalSize / (1 << 20)
	return b, nil
}
