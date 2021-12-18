/*
 * @Author: jason chen (jasonchen@leansoftx.com, http://smallidea.cnblogs.com)
 * @Description:
 * @Date: 2021-11
 * @LastEditors:
 * @LastEditTime:
 */
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

	return true
}

// 检查当前端口是否被占用，并返回一个可用端口
func CheckAndGetAvailableLocalPort(checkPort int, step int) (usablePort int) {
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
