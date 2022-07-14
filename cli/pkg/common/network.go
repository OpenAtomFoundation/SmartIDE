/*
 * @Author: jason chen (jasonchen@leansoftx.com, http://smallidea.cnblogs.com)
 * @Description:
 * @Date: 2021-11
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-05-30 16:21:19
 */
package common

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"
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
	//1. 使用golang进行验证
	//1.1. localhost判断
	l, err := net.Listen("tcp", "localhost:"+strconv.Itoa(port))
	if err != nil {
		SmartIDELog.Debug(fmt.Sprintf("localhost:%v used, "+err.Error(), port))
		return false, err
	}
	defer l.Close()
	//1.2. 通用判断
	if runtime.GOOS != "linux" {
		l2, err := net.Listen("tcp", ":"+strconv.Itoa(port)) // 没有ip的形式，在linux中运行异常
		if err != nil {
			SmartIDELog.Debug(fmt.Sprintf(":%v used, "+err.Error(), port))
			return false, err
		}
		defer l2.Close()
		l2.Close()
	}
	l.Close()

	//2. 使用命令行工具进行验证
	//2.1. command
	command := ""
	switch runtime.GOOS {
	case "linux":
		command = fmt.Sprintf("sudo lsof -nP -iTCP:%v -t -sTCP:LISTEN", port)
	case "windows":
		command = fmt.Sprintf("netstat -aon|findstr \":%d\"", port)
	case "darwin":
		command = fmt.Sprintf("lsof -i tcp:%d -t", port) // 输出的是进程id
	default:
		err = errors.New("unsupported platform")
		return
	}
	output, err := EXEC.CombinedOutput(command, "")
	if _, ok := err.(*exec.ExitError); ok { // 排除exitError
		err = nil
	}
	//2.2. 根据输出判断端口是否占用
	if runtime.GOOS != "windows" {
		result = strings.TrimSpace(output) == "" // 如果没有返回pid，代表可用（没有被占用）
	} else {
		if !strings.Contains(string(output), string(rune(port))) {
			result = true // 端口未被占用
		} else {
			SmartIDELog.Debug(fmt.Sprintf("%v used，"+string(output), port))
		}
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

func isErrorAddressAlreadyInUse(err error) bool {
	var eOsSyscall *os.SyscallError
	if !errors.As(err, &eOsSyscall) {
		return false
	}
	var errErrno syscall.Errno // doesn't need a "*" (ptr) because it's already a ptr (uintptr)
	if !errors.As(eOsSyscall, &errErrno) {
		return false
	}
	if errErrno == syscall.EADDRINUSE {
		return true
	}
	const WSAEADDRINUSE = 10048
	if runtime.GOOS == "windows" && errErrno == WSAEADDRINUSE {
		return true
	}
	return false
}
