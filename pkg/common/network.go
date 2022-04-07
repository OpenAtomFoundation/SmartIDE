/*
 * @Author: jason chen (jasonchen@leansoftx.com, http://smallidea.cnblogs.com)
 * @Description:
 * @Date: 2021-11
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-03-28 23:15:56
 */
package common

import (
	"errors"
	"fmt"
	"net"
	"os/exec"
	"runtime"
)

// 获取可用端口
func GetAvailablePort() (int, error) {

	address, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:0", "0.0.0.0"))
	if err != nil {

		return 0, err
	}

	listener, err := net.ListenTCP("tcp", address)
	if err != nil {

		return 0, err
	}

	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port, nil

}

// 判断端口是否可以（未被占用）
func IsPortAvailable(port int) (result bool, err error) {

	var command *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		command = exec.Command("sh", "-c", fmt.Sprintf("lsof -i tcp:%d", port))
	case "windows":
		command = exec.Command("powershell", "/c", fmt.Sprintf("netstat -aon|findstr \":%d\"", port))
	case "darwin":
		command = exec.Command("sh", "-c", fmt.Sprintf("lsof -i tcp:%d", port))
	default:
		err = errors.New("unsupported platform")
		return
	}

	output, err := command.CombinedOutput()
	// 排除exitError
	if _, ok := err.(*exec.ExitError); ok {
		err = nil
	}
	if len(output) <= 0 {
		result = true // 端口未被占用
	} else {
		SmartIDELog.Debug(fmt.Sprintf("%v used，"+string(output), port))
	}

	return
}

// 检查当前端口是否被占用，并返回一个可用端口
func CheckAndGetAvailableLocalPort(checkPort int, step int) (usablePort int, err error) {
	if step <= 0 {
		step = 100
	}
	usablePort = checkPort

	isPortUnable := false
	for !isPortUnable {
		isPortAvailable, err0 := IsPortAvailable(usablePort)
		err = err0
		if !isPortAvailable {
			usablePort += 100
		} else {
			isPortUnable = true
		}
	}

	return
}
