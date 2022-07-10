/*
 * @Author: jason chen (jasonchen@leansoftx.com, http://smallidea.cnblogs.com)
 * @Description:
 * @Date: 2021-11
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-04-02 11:23:16
 */
package common

import (
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// 打开浏览器（跨平台，支持windows、macos、linux）
func OpenBrowser(url string) (err error) {
	switch runtime.GOOS {
	case "linux":
		output, err := exec.Command("ls -l /usr/share/xsessions/").CombinedOutput()
		if err == nil {
			SmartIDELog.Debug(string(output))
			if !strings.Contains(string(output), "No such file or directory") { // 有图形化的界面，才打开
				err = exec.Command("xdg-open", url).Start()
			} else {
				return errors.New("当前非桌面版！")
			}
		}

	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}

	return err

}
