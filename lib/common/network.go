package common

import (
	"fmt"
	"net"
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
func IsPortAvailable(port int) bool {

	address := fmt.Sprintf("%s:%d", "0.0.0.0", port)
	listener, err := net.Listen("tcp", address)
	if err != nil {

		SmartIDELog.Warning(fmt.Sprintf("port %s is taken: %s", address, err))

		return false
	}

	defer listener.Close()
	return true
}

// 检查当前端口是否被占用，并返回一个可用端口
func CheckAndGetAvailablePort(checkPort int, step int) (usablePort int) {
	if step <= 0 {
		step = 100
	}
	usablePort = checkPort

	isPortUnable := false
	for !isPortUnable {

		if !IsPortAvailable(usablePort) {
			usablePort += 100
		} else {
			isPortUnable = true
		}
	}

	return usablePort
}
