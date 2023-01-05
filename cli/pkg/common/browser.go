/*
SmartIDE - Dev Containers
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
