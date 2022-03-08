/*
 * @Author: jason chen (jasonchen@leansoftx.com, http://smallidea.cnblogs.com)
 * @Description:
 * @Date: 2021-11
 * @LastEditors:
 * @LastEditTime:
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

	/* 	address := fmt.Sprintf(":%d", port)
	   	ln, err := net.Listen("tcp", address)
	   	if err != nil {
	   		return false
	   	}

	   	defer ln.Close()
	   	return true */

	/*
		address := fmt.Sprintf("%s:%d", "localhost", port)
		listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", "localhost", port))
		if err != nil {
			SmartIDELog.Debug(fmt.Sprintf("tcp port %s is taken: %s", address, err))
			return false
		}
		defer listener.Close()

		address = fmt.Sprintf("%s:%d", "0.0.0.0", port)
		conn, err := net.Listen("tcp", address)
		if err != nil {
			SmartIDELog.Debug(fmt.Sprintf("tcp port %s is taken: %s", address, err))
			return false
		}
		defer conn.Close()
	*/

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
