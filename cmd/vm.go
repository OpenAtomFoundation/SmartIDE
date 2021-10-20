package cmd

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	/*"strings" */

	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
	"gopkg.in/yaml.v2"

	"github.com/leansoftX/smartide-cli/lib/common"
	"github.com/leansoftX/smartide-cli/lib/tunnel"
)

var vmCmd = &cobra.Command{
	Use:   "vm",
	Short: "vm start、stop、remove",
	Long:  instanceI18nStop.Info.Help_long,
	Run: func(cmd *cobra.Command, args []string) {
		//var smartIDEName = "smartide"

	},
}

var host string
var port int
var username string
var password string
var repourl string

const repoRoot string = "project"

/* smartide vm start/stop/remove
--host, -H XXX.XXX.XXX.XXX
--username, -U XXX （可选）
--password, -P XXX（可选）
--repourl, -R https://github.com/idcf-boat-house/boathouse-calculator

e.g.  vm start --host {host}:22 --username {username} --password {password} --repourl https://github.com/idcf-boat-house/boathouse-calculator.git
*/
var vmStartCmd = &cobra.Command{
	Use:   "start",
	Short: "vm start、stop、remove",
	Long:  instanceI18nStop.Info.Help_long,
	Run: func(cmd *cobra.Command, args []string) {

		// 1. 连接到远程主机
		fmt.Println("连接到远程主机 ...")
		clientConn, err := connectionDial(host, port, username, password)
		checkError(err, "")

		// 2. 在远程主机上执行相应的命令
		repoName := getRepoName(repourl)
		repoWorkspace := "~/" + repoRoot + "/" + repoName

		//2.1. 执行git clone
		command := fmt.Sprintf(`git clone %v %v`, repourl, repoWorkspace)
		fmt.Printf("%v ...\n", command)
		output, _ := exeSSHCommand(clientConn, command)
		common.SmartIDELog.Info(output)

		//2.2. git pull
		gitPullCommand := fmt.Sprintf("cd %v && git pull cd ~", repoWorkspace)
		exeSSHCommand(clientConn, gitPullCommand)

		//2.3. 读取配置.ide.yaml 并 转换为docker-compose
		fmt.Println("读取代码库下的配置文件 ...")
		command = fmt.Sprintf(`
		cd %v
		cat ./.ide/.ide.yaml
		`, repoWorkspace)
		output, err = exeSSHCommand(clientConn, command)
		checkError(err, output)
		fmt.Println(output) // 打印配置文件的内容
		yamlContent := output
		var yamlFileCongfig YamlFileConfig
		yamlFileCongfig.GetConfigWithStr(yamlContent)
		dockerCompose, ideBindingPort, _ := yamlFileCongfig.ConvertToDockerCompose()

		//2.4. 创建网络
		fmt.Println("创建网络 ...")
		networkCreateCommand := ""
		for network := range dockerCompose.Networks {
			networkCreateCommand += "docker network create " + network + "\n "
		}
		output, _ = exeSSHCommand(clientConn, networkCreateCommand)
		common.SmartIDELog.Info(output)

		//2.5. 在远程vm上生成docker-compose文件，运行docker-compose up
		fmt.Println("docker-compose up ...")
		bytesDockerComposeContent, err := yaml.Marshal(&dockerCompose)
		strDockerComposeContent := strings.ReplaceAll(string(bytesDockerComposeContent), "\"", "\\\"") // 文本中包含双引号
		checkError(err, string(bytesDockerComposeContent))
		commandCreateDockerComposeFile := fmt.Sprintf(`
		mkdir -p ~/.ide
		echo "%v" >> ~/.ide/docker-compose-%v.yaml
		docker-compose -f ~/.ide/docker-compose-%v.yaml --project-directory %v up -d
		`, strDockerComposeContent, repoName, repoName, repoWorkspace)
		output, err = exeSSHCommand(clientConn, commandCreateDockerComposeFile)
		checkError(err, commandCreateDockerComposeFile+"\n"+output)

		//3. 当前主机绑定到远程端口
		var addrMapping map[string]string = map[string]string{}
		remotePortBindings := GetPortBindings(dockerCompose)
		// 查找所有远程主机的端口
		for bindingPort := range remotePortBindings {
			portInt, _ := strconv.Atoi(bindingPort)
			unusedClientPort := strconv.Itoa(common.CheckAndGetAvailablePort(portInt, 100))
			addrMapping["localhost:"+unusedClientPort] = "localhost:" + bindingPort
			fmt.Printf("localhost:%v 绑定到 %v:%v", unusedClientPort, host, bindingPort)
			fmt.Println()
		}
		// 执行绑定
		tunnel.TunnelMultiple(clientConn, addrMapping)

		//4. 打开浏览器
		url := fmt.Sprintf(`http://localhost:%v`, ideBindingPort)
		fmt.Printf("等待WebIDE启动 %s ... \n", url) //TODO: 国际化
		go func(checkUrl string) {
			isUrlReady := false
			for !isUrlReady {
				resp, err := http.Get(checkUrl)
				if (err == nil) && (resp.StatusCode == 200) {
					isUrlReady = true
					common.OpenBrowser(checkUrl)
					fmt.Printf("打开 %v \n", checkUrl)
				}
			}
		}(url)

		//5. 死循环进行驻守
		for {
			time.Sleep(500)
		}
	},
}

// 检查错误
func checkError(err error, info string) {
	if err != nil {
		fmt.Printf("%s. error: %s\n", info, err)
		common.SmartIDELog.Error(err)
	}
}

// 连接到远程主机
func connectionDial(sshHost string, sshPort int, sshUserName, sshPassword string) (clientConn *ssh.Client, err error) {

	// initialize SSH connection
	var clientConfig *ssh.ClientConfig

	if len(sshUserName) > 0 {

		// 输入账号就要有密码
		if len(sshPassword) <= 0 {
			fmt.Print("密码不能为空，请输入: ")
			bytes, _ := readPass(0)
			sshPassword = string(bytes)
			fmt.Println()
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

// 读取密码
func readPass(fd int) ([]byte, error) {
	// make sure an interrupt won't break the terminal
	sigint := make(chan os.Signal)
	state, err := terminal.GetState(fd)
	if err != nil {
		return nil, err
	}
	go func() {
		for _ = range sigint {
			terminal.Restore(fd, state)
			fmt.Println("^C")
			os.Exit(1)
		}
	}()
	signal.Notify(sigint, os.Interrupt)
	defer func() {
		signal.Stop(sigint)
		close(sigint)
	}()
	return terminal.ReadPassword(fd)
}

// 执行ssh command，在session模式下，standard output 只能在执行结束的时候获取到
func exeSSHCommand(clientConn *ssh.Client, sshCommand string) (outContent string, err error) {
	session, err := clientConn.NewSession()
	if err != nil {
		panic(err)
	}
	defer session.Close() //

	// 在ssh主机上执行命令
	out, err := session.CombinedOutput(sshCommand)
	session.StdoutPipe()
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	outContent = string(out)

	return outContent, err
}

// get repo name
func getRepoName(repoUrl string) string {
	_, err := url.ParseRequestURI(repoUrl)
	if err != nil {
		panic(err)
	}
	index := strings.LastIndex(repoUrl, "/")
	return strings.Replace(repoUrl[index+1:], ".git", "", -1)
}

//
func init() {

	vmStartCmd.Flags().StringVarP(&host, "host", "H", "", "远程IP")
	vmStartCmd.MarkFlagRequired("host")
	vmStartCmd.Flags().IntVarP(&port, "port", "P", 22, "SSH 端口，默认为22")
	vmStartCmd.Flags().StringVarP(&username, "username", "U", "", "SSH 登录用户")
	vmStartCmd.Flags().StringVarP(&password, "password", "T", "", "SSH 用户密码")
	vmStartCmd.Flags().StringVarP(&repourl, "repourl", "R", "", "远程IP")
	vmStartCmd.MarkFlagRequired("repourl")

	vmCmd.AddCommand(vmStartCmd)

}
