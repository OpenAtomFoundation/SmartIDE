package cmd

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	/*"strings" */

	"github.com/howeyc/gopass"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"github.com/leansoftX/smartide-cli/cmd/lib"
	"github.com/leansoftX/smartide-cli/lib/common"
	"github.com/leansoftX/smartide-cli/lib/docker/compose"
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
var port int = 22
var username string
var password string
var repourl string

const repoRoot string = "project"

/* smartide vm start/stop/remove
--host, -o XXX.XXX.XXX.XXX
--port, -p 默认为22
--username, -u XXX （可选）
--password, -t XXX（可选）
--repourl, -R https://github.com/idcf-boat-house/boathouse-calculator

e.g.  vm start --host {host} --port 22 --username {username} --password {password} --repourl https://github.com/idcf-boat-house/boathouse-calculator.git
*/
var vmStartCmd = &cobra.Command{
	Use:   "start",
	Short: "vm start、stop、remove",
	Long:  instanceI18nStop.Info.Help_long,
	Run: func(cmd *cobra.Command, args []string) {

		//0. 参数
		if len(strings.TrimSpace(username)) > 0 && len(strings.TrimSpace(password)) == 0 {
			fmt.Printf("密码不能为空，请输入: ")
			passwordBytes, _ := gopass.GetPasswdMasked()
			password = string(passwordBytes)
		}

		// 1. 连接到远程主机
		fmt.Println("连接到远程主机 ...")
		var sshRemote common.SSHRemote
		sshRemote.Instance(host, port, username, password)

		// 2. 在远程主机上执行相应的命令
		repoName := getRepoName(repourl)
		repoWorkspace := "~/" + repoRoot + "/" + repoName

		//2.1. 执行git clone
		command := fmt.Sprintf(`git clone %v %v`, repourl, repoWorkspace)
		fmt.Printf("%v ...\n", command)
		output, _ := sshRemote.ExeSSHCommand(command)
		common.SmartIDELog.Info(output)

		//2.2. git pull
		fmt.Println("git pull ")
		gitPullCommand := fmt.Sprintf("cd %v && git pull && cd ~", repoWorkspace)
		output, err := sshRemote.ExeSSHCommand(gitPullCommand)
		if err != nil {
			common.SmartIDELog.Warning(err.Error(), output)
		}

		//2.3. 读取配置.ide.yaml 并 转换为docker-compose
		ideYamlFilePath := fmt.Sprintf(`%v/.ide/.ide.yaml`, repoWorkspace)
		fmt.Println("读取代码库下的配置文件(", ideYamlFilePath, ") ...")
		catCommand := fmt.Sprintf(`cat %v`, ideYamlFilePath)
		output, err = sshRemote.ExeSSHCommand(catCommand)
		checkError(err, output)
		// fmt.Println(output) // 打印配置文件的内容
		yamlContent := output
		var yamlFileCongfig lib.YamlFileConfig
		yamlFileCongfig.GetConfigWithStr(yamlContent)
		dockerCompose, ideBindingPort, _ := yamlFileCongfig.ConvertToDockerCompose(sshRemote)

		//2.4. 创建网络
		fmt.Println("创建网络 ...")
		networkCreateCommand := ""
		for network := range dockerCompose.Networks {
			networkCreateCommand += "docker network create " + network + "\n "
		}
		output, _ = sshRemote.ExeSSHCommand(networkCreateCommand)
		common.SmartIDELog.Info(output)

		//2.5. 在远程vm上生成docker-compose文件，运行docker-compose up
		fmt.Println("docker-compose up ...")
		bytesDockerComposeContent, err := yaml.Marshal(&dockerCompose)
		printServices(dockerCompose.Services) // 打印services
		fmt.Println()
		strDockerComposeContent := strings.ReplaceAll(string(bytesDockerComposeContent), "\"", "\\\"") // 文本中包含双引号
		checkError(err, string(bytesDockerComposeContent))
		commandCreateDockerComposeFile := fmt.Sprintf(`
		mkdir -p ~/.ide
		rm -rf ~/.ide/docker-compose-%v.yaml
		echo "%v" >> ~/.ide/docker-compose-%v.yaml
		docker-compose -f ~/.ide/docker-compose-%v.yaml --project-directory %v up -d
		`, repoName, strDockerComposeContent, repoName, repoName, repoWorkspace)
		err = sshRemote.ExeSSHCommandNotOutput(commandCreateDockerComposeFile)
		checkError(err, commandCreateDockerComposeFile)

		//3. 当前主机绑定到远程端口
		var addrMapping map[string]string = map[string]string{}
		remotePortBindings := lib.GetPortBindings(dockerCompose)
		unusedLocalPort4IdeBindingPort := ideBindingPort // 未使用的本地端口，与ide端口对应
		// 查找所有远程主机的端口
		for bindingPort, containerPort := range remotePortBindings {
			portInt, _ := strconv.Atoi(bindingPort)
			unusedLocalPort := common.CheckAndGetAvailablePort(portInt, 100) // 得到一个未被占用的本地端口
			if portInt == ideBindingPort && unusedLocalPort != ideBindingPort {
				unusedLocalPort4IdeBindingPort = unusedLocalPort
			}
			addrMapping["localhost:"+strconv.Itoa(unusedLocalPort)] = "localhost:" + bindingPort
			fmt.Printf("localhost:%v -> %v:%v -> container:%v", unusedLocalPort, host, bindingPort, containerPort)
			fmt.Println()
		}
		// 执行绑定
		tunnel.TunnelMultiple(sshRemote.Connection, addrMapping)

		//4. 打开浏览器
		url := fmt.Sprintf(`http://localhost:%v`, unusedLocalPort4IdeBindingPort)
		fmt.Println("等待WebIDE启动 ... ") //TODO: 国际化
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

// 打印 service 列表
func printServices(services map[string]compose.Service) {
	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
	fmt.Fprintln(w, "service\timage\tports\t")
	for serviceName, service := range services {
		line := fmt.Sprintf("%v\t%v\t%v\t", serviceName, service.Image.Name+":"+service.Image.Tag, strings.Join(service.Ports, ";"))
		fmt.Fprintln(w, line)
	}
	w.Flush()
}

// 检查错误
func checkError(err error, info string) {
	if err != nil {
		fmt.Printf("%s. error: %s\n", info, err)
		common.SmartIDELog.Error(err, info)
	}
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

	vmStartCmd.Flags().StringVarP(&host, "host", "o", "", "远程IP")
	vmStartCmd.MarkFlagRequired("host")
	vmStartCmd.Flags().IntVarP(&port, "port", "p", 22, "SSH 端口，默认为22")
	vmStartCmd.Flags().StringVarP(&username, "username", "u", "", "SSH 登录用户")
	vmStartCmd.Flags().StringVarP(&password, "password", "t", "", "SSH 用户密码")
	vmStartCmd.Flags().StringVarP(&repourl, "repourl", "r", "", "远程代码仓库的克隆地址")
	vmStartCmd.MarkFlagRequired("repourl")

	vmCmd.AddCommand(vmStartCmd)

}
