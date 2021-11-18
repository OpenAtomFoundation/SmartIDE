package common

import (
	"bytes"
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
	"github.com/leansoftX/smartide-cli/lib/i18n"
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
}

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
	command := fmt.Sprintf("ss -tulwn | grep :%v", port)
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

// git clone
func (instance *SSHRemote) GitClone(gitRepoUrl string, workSpaceDir string) (outContent string, err error) {

	if strings.TrimSpace(gitRepoUrl) == "" {
		SmartIDELog.Error(i18n.GetInstance().Common.Error.Err_sshremote_param_repourl_none)
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
				SmartIDELog.Console(i18n.GetInstance().Common.Info.Info_privatekey_is_overwrite)
				fmt.Scanln(&isOverwrite)
			} else {
				SmartIDELog.Debug(i18n.GetInstance().Common.Debug.Debug_same_not_overwrite)
				isOverwrite = "n"
			}

		} else { // 提示私钥文件是否覆盖（不覆盖就无法执行git clone）
			SmartIDELog.Console(i18n.GetInstance().Common.Info.Info_whether_overwrite)
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
	gitDirPath := strings.Replace(FilePahtJoin4Linux(workSpaceDir, ".git"), "~/", "", -1) // 把路径变成 “a/b/c” 的形式，不支持 “./a/b/c”、“～/a/b/c”、“./a/b/c”
	cloneCommand := fmt.Sprintf(`[[ ! -d "%v" ]] && rm -rf %v && git clone %v %v || echo "%v"`,
		gitDirPath, workSpaceDir, gitRepoUrl, workSpaceDir, i18n.GetInstance().Common.Info.Info_gitrepo_cloned) // .git 文件如果不存在，在需要git clone
	outContent, err = instance.ExeSSHCommand(cloneCommand)
	if err != nil {
		SmartIDELog.Debug(err.Error())
	}
	// instance.ExecSSHCommandRealTime2(cloneCommand)

	// 需要录入密码的情况
	newGitRepoUrl := strings.ToLower(gitRepoUrl)
	if strings.Contains(outContent, "could not read Password for") { // 常规录入密码
		SmartIDELog.Console(i18n.GetInstance().Common.Info.Info_please_enter_password)
		passwordBytes, _ := gopass.GetPasswdMasked()
		password := string(passwordBytes)

		index := strings.LastIndex(newGitRepoUrl, "@")
		if index < 0 {
			newGitRepoUrl = strings.Replace(newGitRepoUrl, "https://", "https://"+password+"@", -1)
			newGitRepoUrl = strings.Replace(newGitRepoUrl, "http://", "http://"+password+"@", -1)
		} else {
			header := newGitRepoUrl[:strings.Index(newGitRepoUrl, "//")+2]
			newGitRepoUrl = header + password + newGitRepoUrl[index:]
		}
		SmartIDELog.Debug(newGitRepoUrl)

	} else {
		SmartIDELog.Debug(outContent)
	}

	// 需要确认
	outContent, err = instance.continueConnectingAndGoOn(newGitRepoUrl, cloneCommand)
	CheckError(err, outContent)

	return outContent, err
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
				SmartIDELog.Console(i18n.GetInstance().Common.Info.Info_please_enter_password)

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
				SmartIDELog.Info("重置密码成功 .")
				break
			} else {
				SmartIDELog.Debug(str)
			}
		}
	}(in, stdoutB)

	err = session.Run("ssh-keygen -f ~/.ssh/id_rsa -p")
	CheckError(err)
}

// Are you sure you want to continue connecting (yes/no)?
func (instance *SSHRemote) continueConnectingAndGoOn(repoUrl string, cloneCommand string) (output string, err error) {

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
			} //TODO 什么时候可以结束？

			t = output.Len()

			SmartIDELog.Debug(">>" + str)

			if strings.Contains(str, "Are you sure you want to continue connecting") {
				SmartIDELog.Debug(i18n.GetInstance().Common.Debug.Debug_auto_connect_gitrepo)
				_, err = in.Write([]byte("yes\n"))
				CheckError(err)

				break
			} else if strings.Contains(str, "already exists and is not an empty directory.") {
				SmartIDELog.Debug("dir exists")
				break
			} else if strings.Contains(str, i18n.GetInstance().Common.Info.Info_gitrepo_cloned) {
				SmartIDELog.Debug("goroutine , exit")
				break
			} else if strings.Contains(str, "fatal: Authentication failed for") {
				SmartIDELog.Debug("> fatal: Authentication failed for") //TODO 双语
				break
			} else if strings.Contains(str, "Password for") {
				SmartIDELog.Console(i18n.GetInstance().Common.Info.Info_please_enter_password)
				passwordBytes, _ := gopass.GetPasswdMasked()
				_, err = in.Write([]byte(string(passwordBytes) + "\n"))
				CheckError(err)

				//break
			}
		}
	}(in, stdoutB)

	err = session.Run(cloneCommand)
	return stdoutB.String(), err
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
			SmartIDELog.Debug(i18n.GetInstance().Common.Debug.Debug_empty_error)
		}
	}

	/* 	// 错误判断
	   	if strings.Contains(outContent, "error:") || strings.Contains(outContent, "fatal:") {
	   		return "", errors.New(outContent)
	   	} */

	// 记录日志，有些情况下不想输出信息，比如cat id_rsa时
	if isConsoleAndLog {
		outContent = strings.Trim(outContent, "\n")
		SmartIDELog.Debug(fmt.Sprintf("%v >> `%v`", sshCommand, outContent))
	}

	return outContent, err
}

// 执行ssh command，在session模式下实时输出日志
func (instance *SSHRemote) ExecSSHCommandRealTime(sshCommand string) (err error) {
	session, err := instance.Connection.NewSession()
	CheckError(err)

	// 在ssh主机上执行命令
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Run(sshCommand)

	defer session.Close() //
	return err
}

//
func (instance *SSHRemote) ExecSSHCommandRealTime2(sshCommand string) (err error) {

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

			if strings.Contains(str, ":error") || strings.Contains(str, ":fatal") {
				SmartIDELog.Error(str)
			} else {
				SmartIDELog.Info(str)
			}
		}
	}(in, stdoutB)

	return session.Run(sshCommand)
	//CheckError(err)
}

func (instance *SSHRemote) RemoteUpload(filesMaps map[string]string) (err error) {
	// initialize SSH connection
	var clientConfig *ssh.ClientConfig

	if len(instance.SSHPassword) > 0 {

		if len(strings.TrimSpace(instance.SSHPassword)) == 0 {
			SmartIDELog.Error("密码不能为空！")
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
	// SmartIDELog.InfoF("连接到远程主机 %v@%v:%v ...", sshUserName, sshHost, sshPort)

	// initialize SSH connection
	var clientConfig *ssh.ClientConfig
	if sshPort <= 0 {
		sshPort = 22
	}

	if len(sshPassword) > 0 {

		if len(strings.TrimSpace(sshPassword)) == 0 {
			SmartIDELog.Error(i18n.GetInstance().Common.Error.Err_password_none)
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
