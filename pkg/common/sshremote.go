/*
 * @Author: jason chen (jasonchen@leansoftx.com, http://smallidea.cnblogs.com)
 * @Description:
 * @Date: 2021-11
 * @LastEditors:
 * @LastEditTime:
 */
package common

import (
	"bytes"
	"errors"
	"path"

	//"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	scp "github.com/bramvdbogaerde/go-scp"
	"github.com/howeyc/gopass"
	"github.com/leansoftX/smartide-cli/internal/apk/i18n"
	"golang.org/x/crypto/ssh"
)

//
type SSHRemote struct {
	SSHHost        string
	SSHPort        int
	SSHUserName    string
	SSHPassword    string
	SSHKey         string
	SSHKeyPassword string
	SSHKeyPath     string
	Connection     *ssh.Client
}

var i18nInstance = i18n.GetInstance()

// 实例
func NewSSHRemote(host string, port int, userName, password string) (instance SSHRemote, err error) {

	instance = SSHRemote{}

	if (instance.Connection == &ssh.Client{}) || instance.Connection == nil {
		instance.SSHHost = host
		instance.SSHPort = port
		instance.SSHUserName = userName
		instance.SSHPassword = password

		connection, err := connectionDial(host, port, userName, password)
		if err != nil {
			return instance, err
		}

		instance.Connection = connection
	}

	return instance, nil
}

/*
// 实例
func (instance *SSHRemote) Instance(host string, port int, userName, password string) error {

	if (instance.Connection == &ssh.Client{}) || instance.Connection == nil {
		instance.SSHHost = host
		instance.SSHPort = port
		instance.SSHUserName = userName
		instance.SSHPassword = password

		connection, err := connectionDial(host, port, userName, password)
		if err != nil {
			return err
		}

		instance.Connection = connection
	}

	return nil
} */

// 验证
func (instance *SSHRemote) CheckDail(host string, port int, userName, password string) error {

	if (instance.Connection == &ssh.Client{}) || instance.Connection == nil {

		connection, err := connectionDial(host, port, userName, password)

		if err != nil {
			return err
		}

		defer connection.Close()
	}

	return nil
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
func (instance *SSHRemote) CheckAndGetAvailableRemotePort(checkPort int, step int) (usablePort int) {
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

// 获取远程主机上的当前目录
func (sshRemote *SSHRemote) GetRemotePwd() (currentDir string, err error) {
	currentDir, err = sshRemote.ExeSSHCommand("pwd")
	return currentDir, err
}

//获取远程uid,gid
func (sshRemote *SSHRemote) GetRemoteUserInfo() (Uid string, Gid string) {
	remuid, err := sshRemote.ExeSSHCommand("id -u $USER")
	remgid, remgiderr := sshRemote.ExeSSHCommand("id -g $USER")
	SmartIDELog.Debug("Remote---Uid:" + remuid)
	SmartIDELog.Debug("Remote---Gid:" + remgid)

	if remuid != "" && err == nil {
		Uid = remuid
	} else {
		Uid = "1000"
	}
	if remgid != "" && remgiderr == nil {
		Gid = remgid
	} else {
		Gid = "1000"
	}
	return Uid, Gid
}

// 当前目录是否已经clone
func (instance *SSHRemote) IsCloned(workSpaceDir string) bool {
	gitDirPath := strings.Replace(FilePahtJoin4Linux(workSpaceDir, ".git"), "~/", "", -1) // 把路径变成 “a/b/c” 的形式，不支持 “./a/b/c”、“～/a/b/c”、“./a/b/c”
	cloneCommand := fmt.Sprintf(`[[ -d "%v" ]] && echo "1" || echo "0"`,
		gitDirPath)
	outContent, err := instance.ExeSSHCommand(cloneCommand)
	CheckError(err)

	// .git 文件夹不存在，清空文件夹
	if outContent == "0" {
		instance.ExeSSHCommand("sudo rm -rf " + workSpaceDir)
	}

	return outContent == "1"
}

// 文件是否存在
func (instance *SSHRemote) IsExit(filepath string) bool {

	filepath = instance.convertFilePath(filepath)

	command := fmt.Sprintf(`[[ -f "%v" ]] && echo "1" || echo "0"`, filepath)
	outContent, err := instance.ExeSSHCommand(command)
	CheckError(err)

	return outContent == "1"
}

// 获取文件内容
func (instance *SSHRemote) GetContent(filepath string) string {

	filepath = instance.convertFilePath(filepath)

	command := fmt.Sprintf(`cat "%v"`, filepath)
	outContent, err := instance.ExeSSHCommand(command)
	CheckError(err)

	return outContent
}

// 创建文件，如果存在就附加内容
func (sshRemote *SSHRemote) CreateFileByEcho(filepath string, content string) error {

	filepath = sshRemote.convertFilePath(filepath)

	/* 	// 检查并创建文件夹
	   	dir := path.Dir(filepath)
	   	err := sshRemote.CheckAndCreateDir(dir)
	   	if err != nil {
	   		return err
	   	} */

	// 创建文件
	command := fmt.Sprintf(`sudo echo "%v" >> %v`, content, filepath)
	_, err := sshRemote.ExeSSHCommand(command)

	return err
}

// 检查并创建文件夹
func (sshRemote *SSHRemote) CheckAndCreateDir(dir string) error {
	dir = sshRemote.convertFilePath(dir)

	command := fmt.Sprintf("[[ -f \"%v\" ]] && echo \"1\" || mkdir -p \"%v\"", dir, dir)
	_, err := sshRemote.ExeSSHCommand(command)
	return err

}

// 转换文件路径为远程主机支持的
func (instance *SSHRemote) convertFilePath(filepath string) (newFilepath string) {
	newFilepath = filepath

	newFilepath = strings.ReplaceAll(filepath, "\\", "/")

	index := strings.Index(newFilepath, "~/")
	if index == 0 {
		pwd, err := instance.GetRemotePwd()
		CheckError(err)
		newFilepath = path.Join(pwd, strings.Replace(newFilepath, "~/", "", -1))
	}

	return newFilepath
}

// git clone
func (instance *SSHRemote) GitClone(gitRepoUrl string, workSpaceDir string) error {

	if instance.IsCloned(workSpaceDir) {
		SmartIDELog.Info(i18n.GetInstance().Common.Info_gitrepo_cloned)
		return nil
	}

	if strings.TrimSpace(gitRepoUrl) == "" {
		SmartIDELog.Error(i18n.GetInstance().Common.Err_sshremote_param_repourl_none)
	}
	if workSpaceDir == "" {
		workSpaceDir = getRepoName(gitRepoUrl)
	}

	// 检测是否为ssh模式
	if strings.Index(gitRepoUrl, "git@") == 0 {
		isOverwrite := "y" // 是否覆盖服务器上的私钥文件
		isAllowCopyPrivateKey := ""

		commandRsa := `[[ -f ".ssh/id_rsa" ]] && cat ~/.ssh/id_rsa || echo ""`
		remoteRsaPri, err := instance.ExeSSHCommandConsole(commandRsa, false)
		CheckError(err)
		SmartIDELog.DebugF("%v >> `%v`", commandRsa, "****")

		commandRsaPub := `[[ -f ".ssh/id_rsa.pub" ]] && cat ~/.ssh/id_rsa.pub || echo ""`
		remoteRsaPub, err := instance.ExeSSHCommandConsole(commandRsaPub, false)
		CheckError(err)
		SmartIDELog.DebugF("%v >> `%v`", commandRsaPub, "****")

		if remoteRsaPri != "" && remoteRsaPub != "" { // 文件存在时提示是否覆盖

			// 读取本地的ssh配置文件
			homeDir, err := os.UserHomeDir()
			CheckError(err)
			localRsaPub, err := ioutil.ReadFile(filepath.Join(homeDir, "/.ssh/id_rsa.pub")) // 读取本地的 id_rsa 文件
			CheckError(err)                                                                 // , string(localRsaPub)

			// 公钥 文件不同时才会提示覆盖
			if strings.TrimSpace(remoteRsaPub) != strings.TrimSpace(string(localRsaPub)) {
				SmartIDELog.Console(i18n.GetInstance().Common.Info_privatekey_is_overwrite)
				fmt.Scanln(&isOverwrite)
			} else {
				SmartIDELog.Debug(i18n.GetInstance().Common.Debug_same_not_overwrite)
				isOverwrite = "n"
			}

		} else { // 提示私钥文件是否覆盖（不覆盖就无法执行git clone）
			SmartIDELog.Console(i18n.GetInstance().Common.Info_whether_overwrite)
			fmt.Scanln(&isAllowCopyPrivateKey)
		}

		if isAllowCopyPrivateKey == "y" || isOverwrite == "y" {

			if isOverwrite == "y" {
				// 读取本地的ssh配置文件
				homeDir, err := os.UserHomeDir()
				CheckError(err)
				idRsa, err := ioutil.ReadFile(filepath.Join(homeDir, "/.ssh/id_rsa")) // 读取本地的 id_rsa 文件
				CheckError(err, string(idRsa))
				idRsaPub, err := ioutil.ReadFile(filepath.Join(homeDir, "/.ssh/id_rsa.pub")) // 读取本地的 id_rsa.pub 文件
				CheckError(err, string(idRsaPub))

				// 执行私钥文件复制
				command := fmt.Sprintf(`mkdir -p .ssh
			rm -rf ~/.ssh/id_rsa
			echo "%v" >> ~/.ssh/id_rsa
			chmod 600 ~/.ssh/id_rsa

			rm -rf ~/.ssh/id_rsa.pub
			echo "%v" >> ~/.ssh/id_rsa.pub
			chmod 600 ~/.ssh/id_rsa.pub

			`, string(idRsa), string(idRsaPub))
				output, err := instance.ExeSSHCommandConsole(command, false)
				CheckError(err, output)

				// log
				consoleCommand := strings.ReplaceAll(command, string(idRsa), "***")
				consoleCommand = strings.ReplaceAll(consoleCommand, string(idRsaPub), "***")
				SmartIDELog.DebugF("%v >> `%v`", consoleCommand, output)

				// 执行私钥密码的取消 —— 把私钥密码设置为空
				// https://docs.github.com/cn/authentication/connecting-to-github-with-ssh/working-with-ssh-key-passphrases
				instance.sshSaveEmptyPassphrase()
			}
		}
	}

	// 执行clone
	//gitDirPath := strings.Replace(FilePahtJoin4Linux(workSpaceDir, ".git"), "~/", "", -1) // 把路径变成 “a/b/c” 的形式，不支持 “./a/b/c”、“～/a/b/c”、“./a/b/c”
	cloneCommand := fmt.Sprintf(`git clone %v %v`,
		gitRepoUrl, workSpaceDir) // .git 文件如果不存在，在需要git clone
	err := instance.ExecSSHCommandRealTimeFunc(cloneCommand, func(output string) error {
		if strings.Contains(output, "error") || strings.Contains(output, "fatal") {

			newGitRepoUrl := strings.ToLower(gitRepoUrl)

			// 需要录入密码的情况
			if strings.Contains(output, "could not read Password for") { // 常规录入密码
				SmartIDELog.Console(i18n.GetInstance().Common.Info_please_enter_password)
				passwordBytes, _ := gopass.GetPasswdMasked()
				password := string(passwordBytes)

				// 添加密码到 https/http 链接中
				index := strings.LastIndex(newGitRepoUrl, "@")
				if index < 0 {
					newGitRepoUrl = strings.Replace(newGitRepoUrl, "https://", "https://"+password+"@", -1)
					newGitRepoUrl = strings.Replace(newGitRepoUrl, "http://", "http://"+password+"@", -1)
				} else {
					header := newGitRepoUrl[:strings.Index(newGitRepoUrl, "//")+2]
					newGitRepoUrl = header + password + newGitRepoUrl[index:]
				}
				SmartIDELog.Debug(newGitRepoUrl)

				// 再次运行 git clone
				instance.ExecSSHCommandRealTimeFunc(cloneCommand, nil)

			} else {
				return errors.New(output)
			}

		} else {
			SmartIDELog.ConsoleInLine(output)
			if strings.Contains(output, "done.") {
				fmt.Println()
			}
		}

		return nil
	})

	// log
	if err == nil {
		SmartIDELog.Info(i18n.GetInstance().Common.Info_gitrepo_clone_done)
	}

	return err
}

// 保存一个空密码，保证后续的git clone不需要输入私钥的密码
func (instance *SSHRemote) sshSaveEmptyPassphrase() {
	// 如果本身就是空密码，就不需要执行了
	output, _ := instance.ExeSSHCommand("ssh-keygen -f ~/.ssh/id_rsa -p")
	if !strings.Contains(output, "Enter old passphrase") {
		return
	}

	session, err := instance.Connection.NewSession()
	CheckError(err)
	defer session.Close()

	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	err = session.RequestPty("xterm", 80, 40, modes)
	CheckError(err)

	stdoutB := new(bytes.Buffer)
	session.Stdout = stdoutB
	in, _ := session.StdinPipe()

	go func(in io.Writer, output *bytes.Buffer) {

		var t int = 0

		for {
			str := string(output.Bytes()[t:])
			if str == "" {
				continue
			}

			t = output.Len()

			if strings.Contains(str, "Enter old passphrase") {
				SmartIDELog.Console(i18n.GetInstance().Common.Info_please_enter_password)

				password, err := gopass.GetPasswdMasked()
				CheckError(err)

				_, err = in.Write([]byte(string(password) + "\n"))
				CheckError(err)
			} else if strings.Contains(str, "Enter new passphrase (empty for no passphrase)") {
				_, err = in.Write([]byte("\n"))
				CheckError(err)
			} else if strings.Contains(str, "Enter same passphrase again") {
				_, err = in.Write([]byte("\n"))
				CheckError(err)
				SmartIDELog.Info(i18nInstance.Common.Info_ssh_rsa_cancel_pwd_successed)
				break
			} else {
				SmartIDELog.Debug(str)
			}
		}
	}(in, stdoutB)

	err = session.Run("ssh-keygen -f ~/.ssh/id_rsa -p")
	CheckError(err)
}

// get repo name
func getRepoName(repoUrl string) string {
	index := strings.LastIndex(repoUrl, "/")
	return strings.Replace(repoUrl[index+1:], ".git", "", -1)
}

// 执行ssh command，在session模式下，standard output 只能在执行结束的时候获取到
func (instance *SSHRemote) ExeSSHCommand(sshCommand string) (outContent string, err error) {

	return instance.ExeSSHCommandConsole(sshCommand, true)
}

// 执行ssh command，在session模式下，standard output 只能在执行结束的时候获取到
func (instance *SSHRemote) ExeSSHCommandConsole(sshCommand string, isConsoleAndLog bool) (outContent string, err error) {
	if len(sshCommand) <= 0 {
		return "", nil
	}

	session, err := instance.Connection.NewSession()
	CheckError(err)

	// 在ssh主机上执行命令
	SmartIDELog.Debug(sshCommand + " >> ...")
	out, err := session.CombinedOutput(sshCommand)
	outContent = string(out)
	defer session.Close()

	// 空错误判断
	if err != nil {
		if outContent == "" && err.Error() == "Process exited with status 1" {
			SmartIDELog.Debug(i18n.GetInstance().Common.Debug_empty_error)
		}
	}

	// 记录日志，有些情况下不想输出信息，比如cat id_rsa时
	if isConsoleAndLog {
		outContent = strings.Trim(outContent, "\n")
		SmartIDELog.Debug(fmt.Sprintf("%v >> `%v`", sshCommand, outContent))
	}

	return outContent, err
}

// 实时执行
func (instance *SSHRemote) ExecSSHCommandRealTime(sshCommand string) (err error) {

	return instance.ExecSSHCommandRealTimeFunc(sshCommand, nil)
}

// 实时执行，带函数
func (instance *SSHRemote) ExecSSHCommandRealTimeFunc(sshCommand string, yamlExecuteFun func(output string) error) (err error) {

	SmartIDELog.Debug("-> " + sshCommand)
	if (*instance == SSHRemote{}) {
		return errors.New(i18nInstance.Common.Err_ssh_dial_none)
	}

	session, err := instance.Connection.NewSession()
	CheckError(err)
	defer session.Close()

	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	err = session.RequestPty("xterm", 80, 40, modes)
	CheckError(err)

	stdoutB := new(bytes.Buffer)
	session.Stdout = stdoutB
	in, _ := session.StdinPipe()

	// 函数
	if yamlExecuteFun == nil {
		yamlExecuteFun = func(out string) error {
			if strings.Contains(out, "error") || strings.Contains(out, "fatal") {
				SmartIDELog.Error(out)
			} else {
				SmartIDELog.Console(out)
			}
			return nil
		}
	}

	var exit chan bool = make(chan bool)

	go func(in io.Writer, out *bytes.Buffer, exit chan bool) {

		var t int = 0
		for {

			isExit := false
			select {
			case <-exit:
				isExit = true
			default:
			}
			if isExit || out == nil { // 退出
				break
			}

			originMsg := out.String()
			if originMsg == "" || t > len(originMsg) {
				continue
			}
			msg := originMsg[t:] // 获取 当前的字符串
			t = len(originMsg)   // 为 下一次获取字符串做准备

			//msg = strings.ReplaceAll(msg, "\x00", "") // ??
			if msg == "" {
				continue
			}

			//yamlExecuteFun(msg)

			array := strings.Split(msg, "\r\n")
			for _, sub := range array {
				if len(sub) == 0 || sub == "\r\n" { //|| sub == "\r"
					continue
				}

				err = yamlExecuteFun(sub)
				CheckError(err)
			}

		}
	}(in, stdoutB, exit)

	err = session.Run(sshCommand)
	exit <- true
	return err
}

func (instance *SSHRemote) RemoteUpload(filesMaps map[string]string) (err error) {
	// initialize SSH connection
	var clientConfig *ssh.ClientConfig

	if len(instance.SSHPassword) > 0 {

		if len(strings.TrimSpace(instance.SSHPassword)) == 0 {
			SmartIDELog.Error(i18nInstance.Common.Err_ssh_password_required)
		}

		clientConfig = &ssh.ClientConfig{
			User: instance.SSHUserName,
			Auth: []ssh.AuthMethod{
				ssh.Password(instance.SSHPassword),
			},
			Timeout: 30 * time.Second, // 30 秒超时
			// 解决 “ssh: must specify HostKeyCallback” 的问题
			HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
				return nil
			},
		}

	} else { // 如果用户不输入用户名和密码，则尝试使用ssh key pair的方式链接远程服务器
		//var hostKey ssh.PublicKey
		homePath, err := os.UserHomeDir()
		if err != nil {
			CheckError(err)
		}
		filePath := filepath.Join(homePath, "/.ssh/id_rsa")
		key, err := ioutil.ReadFile(filePath)
		CheckError(err, "unable to read private key:")

		// Create the Signer for this private key.
		signer, err := ssh.ParsePrivateKey(key)
		CheckError(err, "unable to parse private key:")

		clientConfig = &ssh.ClientConfig{
			User: instance.SSHUserName,
			Auth: []ssh.AuthMethod{
				// Use the PublicKeys method for remote authentication.
				ssh.PublicKeys(signer),
			},
			HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
				// use OpenSSH's known_hosts file if you care about host validation
				return nil
			},
		}

	}

	addr := fmt.Sprintf("%v:%v", instance.SSHHost, instance.SSHPort)

	if err == nil {
		for k, v := range filesMaps {

			client := scp.NewClient(addr, clientConfig)
			err = client.Connect()
			if err != nil {
				fmt.Println("Couldn't establish a connection to the remote server ", err)
				return
			}
			// Open a file
			f, _ := os.Open(k)

			defer client.Close()
			// Finaly, copy the file over
			// Usage: CopyFile(fileReader, remotePath, permission)
			defer f.Close()

			err = client.CopyFile(f, v, "0777")
			if err != nil {
				fmt.Println("Error while copying file ", err)
			}

		}

		// Close client connection after the file has been copied

	}
	return
}

// 连接到远程主机
func connectionDial(sshHost string, sshPort int, sshUserName, sshPassword string) (clientConn *ssh.Client, err error) {
	// initialize SSH connection
	var clientConfig *ssh.ClientConfig
	if sshPort <= 0 {
		sshPort = 22
	}

	if len(sshPassword) > 0 {

		if len(strings.TrimSpace(sshPassword)) == 0 {
			SmartIDELog.Error(i18n.GetInstance().Common.Err_password_none)
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
		//var hostKey ssh.PublicKey
		homePath, err := os.UserHomeDir()
		CheckError(err)
		filePath := filepath.Join(homePath, "/.ssh/id_rsa")
		key, err := ioutil.ReadFile(filePath)
		CheckError(err, "unable to read private key:")

		// Create the Signer for this private key.
		signer, err := ssh.ParsePrivateKey(key)
		CheckError(err, "unable to parse private key:")

		clientConfig = &ssh.ClientConfig{
			User: sshUserName,
			Auth: []ssh.AuthMethod{
				// Use the PublicKeys method for remote authentication.
				ssh.PublicKeys(signer),
			},
			HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
				// use OpenSSH's known_hosts file if you care about host validation
				return nil
			},
		}

	}

	addr := fmt.Sprintf("%v:%v", sshHost, sshPort)
	return ssh.Dial("tcp", addr, clientConfig)
}
