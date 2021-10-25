package common

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

//
type SSHRemote struct {
	SSHHost     string
	SSHPort     int
	SSHUserName string
	SSHPassword string
	Connection  *ssh.Client
}

// 实例
func (instance *SSHRemote) Instance(host string, port int, userName, password string) {

	if (instance.Connection == &ssh.Client{}) || instance.Connection == nil {
		instance.SSHHost = host
		instance.SSHPort = port
		instance.SSHUserName = userName
		instance.SSHPassword = password

		connection, err := connectionDial(host, port, userName, password)
		if err != nil {
			SmartIDELog.Error(err, "create ssh connection error:")
		}
		instance.Connection = connection
	}

}

// 判断端口是否可以（未被占用）
func (instance *SSHRemote) IsPortAvailable(port int) bool {
	command := fmt.Sprintf("sudo ss -tulwn | grep :%v", port)
	output, err := instance.ExeSSHCommand(command)
	if err != nil {
		if output != "" || err.Error() != "Process exited with status 1" {
			SmartIDELog.Error(err, output)
		}
	}

	return !strings.Contains(output, ":"+strconv.Itoa(port))
}

// 检查当前端口是否被占用，并返回一个可用端口
func (instance *SSHRemote) CheckAndGetAvailablePort(checkPort int, step int) (usablePort int) {
	if step <= 0 {
		step = 100
	}
	usablePort = checkPort

	isPortUnable := false
	for !isPortUnable {

		if !instance.IsPortAvailable(usablePort) {
			usablePort += 100
		} else {
			isPortUnable = true
		}
	}

	return usablePort
}

// 执行ssh command，在session模式下，standard output 只能在执行结束的时候获取到
func (instance *SSHRemote) ExeSSHCommand(sshCommand string) (outContent string, err error) {
	session, err := instance.Connection.NewSession()
	if err != nil {
		panic(err)
	}

	// 在ssh主机上执行命令
	out, err := session.CombinedOutput(sshCommand)
	/* 	session.StdoutPipe()
	   	session.Stdout = os.Stdout
	   	session.Stderr = os.Stderr */
	outContent = string(out)

	defer session.Close()

	if err != nil {
		if outContent == "" && err.Error() == "Process exited with status 1" {
			SmartIDELog.Warning("ssh 执行command 遇到空错误，已跳过！")
		}
	}

	return outContent, err
}

// 执行ssh command，在session模式下实时输出日志
func (instance *SSHRemote) ExeSSHCommandNotOutput(sshCommand string) (err error) {
	session, err := instance.Connection.NewSession()
	if err != nil {
		panic(err)
	}

	// 在ssh主机上执行命令
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Run(sshCommand)

	defer session.Close() //
	return err
}

// 连接到远程主机
func connectionDial(sshHost string, sshPort int, sshUserName, sshPassword string) (clientConn *ssh.Client, err error) {

	// initialize SSH connection
	var clientConfig *ssh.ClientConfig

	if len(sshUserName) > 0 {

		if len(strings.TrimSpace(sshPassword)) == 0 {
			SmartIDELog.Error("密码不能为空！")
		}

		clientConfig = &ssh.ClientConfig{
			User: sshUserName,
			Auth: []ssh.AuthMethod{
				ssh.Password(sshPassword),
			},
			Timeout: 30 * time.Second, // 30 秒超时
			// 解决 “ssh: must specify HostKeyCallback” 的问题
			HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
				return nil
			},
		}

	} else { // 如果用户不输入用户名和密码，则尝试使用ssh key pair的方式链接远程服务器
		var hostKey ssh.PublicKey
		key, err := ioutil.ReadFile("/home/user/.ssh/id_rsa")
		checkError(err, "unable to read private key:")

		// Create the Signer for this private key.
		signer, err := ssh.ParsePrivateKey(key)
		checkError(err, "unable to parse private key:")

		clientConfig = &ssh.ClientConfig{
			User: "user",
			Auth: []ssh.AuthMethod{
				// Use the PublicKeys method for remote authentication.
				ssh.PublicKeys(signer),
			},
			HostKeyCallback: ssh.FixedHostKey(hostKey),
		}

	}

	addr := fmt.Sprintf("%v:%v", sshHost, sshPort)
	return ssh.Dial("tcp", addr, clientConfig)
}

// 检查错误
func checkError(err error, info string) {
	if err != nil {
		fmt.Printf("%s. error: %s\n", info, err)
		SmartIDELog.Error(err, info)
	}
}
