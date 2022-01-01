/*
 * @Author: kenan
 * @Date: 2021-12-29 14:26:42
 * @LastEditors: kenan
 * @LastEditTime: 2021-12-29 15:20:58
 * @Description: file content
 */

package common

import (
	"os"
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
