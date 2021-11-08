package cmd

import (
	"fmt"
	"net/http"
	"path/filepath"

	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"github.com/leansoftX/smartide-cli/cmd/lib"
	"github.com/leansoftX/smartide-cli/lib/common"
	"github.com/leansoftX/smartide-cli/lib/docker/compose"
	"github.com/leansoftX/smartide-cli/lib/i18n"
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
var branch string = "main"

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
	Short: i18n.GetInstance().VmStart.Info.Help_short,
	Long:  i18n.GetInstance().VmStart.Info.Help_long,
	Run: func(cmd *cobra.Command, args []string) {

		common.SmartIDELog.Info(i18n.GetInstance().VmStart.Info.Info_starting)

		//1. 连接到远程主机
		common.SmartIDELog.Info(i18n.GetInstance().VmStart.Info.Info_connect_remote)
		var sshRemote common.SSHRemote
		sshRemote.Instance(host, port, username, password)

		//2. 在远程主机上执行相应的命令
		repoName := getRepoName(repourl)
		repoWorkspace := "~/" + repoRoot + "/" + repoName

		//2.1. 执行git clone
		common.SmartIDELog.Info(i18n.GetInstance().VmStart.Info.Info_git_clone)
		output, err := sshRemote.GitClone(repourl, repoWorkspace)
		common.CheckError(err, output)

		//2.2. git checkout
		checkoutCommand := "git fetch && "
		if branch != "" {
			checkoutCommand += "git checkout " + branch
		} else { // 有可能当前目录所处的分支非主分支
			// 获取分支列表，确认主分支是master 还是 main
			branches, _ := sshRemote.ExeSSHCommand(fmt.Sprintf("cd %v && git branch", repoWorkspace))
			branches = strings.ReplaceAll(branches, " ", "")
			isContainMaster := strings.Contains(branches, "master\n") || branches == "master"
			if isContainMaster {
				checkoutCommand += "git checkout master"
			} else {
				checkoutCommand += "git checkout main"
			}

		}

		//2.3. git checkout & pull
		common.SmartIDELog.Info(i18n.GetInstance().VmStart.Info.Info_git_checkout_and_pull)
		gitPullCommand := fmt.Sprintf("cd %v && %v && git pull && cd ~", repoWorkspace, checkoutCommand)
		output, err = sshRemote.ExeSSHCommand(gitPullCommand)
		if err != nil {
			common.SmartIDELog.Warning(err.Error(), output)
		}

		//2.4. 读取配置.ide.yaml 并 转换为docker-compose
		relativeYamlFilePath := "/.ide/.ide.yaml"
		if ideyamlfile != "" {
			relativeYamlFilePath = ideyamlfile // 指定配置文件的路径
		}
		ideYamlFilePath := common.FilePahtJoin(common.OS_Linux, repoWorkspace, relativeYamlFilePath) //fmt.Sprintf(`%v/.ide/.ide.yaml`, repoWorkspace)
		common.SmartIDELog.Info(fmt.Sprintf(i18n.GetInstance().VmStart.Info.Info_read_config, ideYamlFilePath))
		catCommand := fmt.Sprintf(`cat %v`, ideYamlFilePath)
		output, err = sshRemote.ExeSSHCommand(catCommand)
		common.CheckError(err, output)
		yamlContent := output
		var yamlFileCongfig lib.YamlFileConfig
		yamlFileCongfig.GetConfigWithStr(yamlContent)
		dockerCompose, ideBindingPort, _ := yamlFileCongfig.ConvertToDockerCompose(sshRemote, filepath.Dir(ideYamlFilePath))

		//2.4. 创建网络
		common.SmartIDELog.Info(i18n.GetInstance().VmStart.Info.Info_create_network)
		networkCreateCommand := ""
		for network := range dockerCompose.Networks {
			networkCreateCommand += "docker network create " + network + "\n "
		}
		sshRemote.ExeSSHCommand(networkCreateCommand)

		//2.5. 在远程vm上生成docker-compose文件，运行docker-compose up
		common.SmartIDELog.Info(i18n.GetInstance().VmStart.Info.Info_compose_up)
		bytesDockerComposeContent, err := yaml.Marshal(&dockerCompose)
		printServices(dockerCompose.Services)                                                          // 打印services
		strDockerComposeContent := strings.ReplaceAll(string(bytesDockerComposeContent), "\"", "\\\"") // 文本中包含双引号
		common.CheckError(err, string(bytesDockerComposeContent))
		commandCreateDockerComposeFile := fmt.Sprintf(`
		mkdir -p ~/.ide
		rm -rf ~/.ide/docker-compose-%v.yaml
		echo "%v" >> ~/.ide/docker-compose-%v.yaml
		docker-compose -f ~/.ide/docker-compose-%v.yaml --project-directory %v up -d
		`, repoName, strDockerComposeContent, repoName, repoName, repoWorkspace)
		err = sshRemote.ExeSSHCommandNotOutput(commandCreateDockerComposeFile)
		common.CheckError(err, commandCreateDockerComposeFile)

		//3. 当前主机绑定到远程端口
		var addrMapping map[string]string = map[string]string{}
		remotePortBindings := dockerCompose.GetPortBindings()
		unusedLocalPort4IdeBindingPort := ideBindingPort // 未使用的本地端口，与ide端口对应
		// 查找所有远程主机的端口
		for bindingPort, containerPort := range remotePortBindings {
			portInt, _ := strconv.Atoi(bindingPort)
			unusedLocalPort := common.CheckAndGetAvailablePort(portInt, 100) // 得到一个未被占用的本地端口
			if portInt == ideBindingPort && unusedLocalPort != ideBindingPort {
				unusedLocalPort4IdeBindingPort = unusedLocalPort
			}
			addrMapping["localhost:"+strconv.Itoa(unusedLocalPort)] = "localhost:" + bindingPort
			msg := fmt.Sprintf("localhost:%v -> %v:%v -> container:%v", unusedLocalPort, host, bindingPort, containerPort)
			common.SmartIDELog.Info(msg)
		}
		// 执行绑定
		tunnel.TunnelMultiple(sshRemote.Connection, addrMapping)

		//4. 打开浏览器
		var url string
		//vscode启动时候默认打开文件夹处理
		if yamlFileCongfig.Workspace.DevContainer.IdeType == "vscode" {
			url = fmt.Sprintf("http://localhost:%v/?folder=vscode-remote://localhost:%v/home/project", unusedLocalPort4IdeBindingPort, unusedLocalPort4IdeBindingPort)
		} else {
			url = fmt.Sprintf(`http://localhost:%v`, unusedLocalPort4IdeBindingPort)
		}
		common.SmartIDELog.Info(i18n.GetInstance().VmStart.Info.Info_warting_for_webide) //TODO: 国际化
		go func(checkUrl string) {
			isUrlReady := false
			for !isUrlReady {
				resp, err := http.Get(checkUrl)
				if (err == nil) && (resp.StatusCode == 200) {
					isUrlReady = true
					common.OpenBrowser(checkUrl)
					common.SmartIDELog.InfoF(i18n.GetInstance().VmStart.Info.Info_open_brower, checkUrl)
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

// get repo name
func getRepoName(repoUrl string) string {
	/* _, err := url.ParseRequestURI(repoUrl)
	if err != nil {
		panic(err)
	} */
	index := strings.LastIndex(repoUrl, "/")
	return strings.Replace(repoUrl[index+1:], ".git", "", -1)
}

//
func init() {

	vmStartCmd.Flags().StringVarP(&host, "host", "o", "", "远程IP")
	vmStartCmd.MarkFlagRequired("host")
	vmStartCmd.Flags().IntVarP(&port, "port", "p", 22, "SSH 端口，默认为22")
	vmStartCmd.Flags().StringVarP(&username, "username", "u", "", "SSH 登录用户")
	vmStartCmd.MarkFlagRequired("username")
	vmStartCmd.Flags().StringVarP(&password, "password", "t", "", "SSH 用户密码")
	vmStartCmd.Flags().StringVarP(&repourl, "repourl", "r", "", "远程代码仓库的克隆地址")
	vmStartCmd.MarkFlagRequired("repourl")

	vmStartCmd.Flags().StringVarP(&ideyamlfile, "filepath", "f", "", "指定yaml文件路径")
	vmStartCmd.Flags().StringVarP(&branch, "branch", "b", "", "指定git分支")

	vmCmd.AddCommand(vmStartCmd)

}
