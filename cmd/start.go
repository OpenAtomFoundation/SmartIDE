package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/leansoftX/smartide-cli/lib/common"
	"github.com/leansoftX/smartide-cli/lib/docker/compose"
	"github.com/leansoftX/smartide-cli/lib/i18n"
	"github.com/leansoftX/smartide-cli/lib/tunnel"

	"github.com/spf13/cobra"
)

var instanceI18nStart = i18n.GetInstance().Start

var ideyamlfile string

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: instanceI18nStart.Info.Help_short,
	Long:  instanceI18nStart.Info.Help_long,
	Run: func(cmd *cobra.Command, args []string) {

		//0. 提示文本
		fmt.Println(i18n.GetInstance().Start.Info.Info_start)

		//0.1. 校验是否能正常执行docker
		dockerCmd := exec.Command("docker", "-v")
		if dockerErr := dockerCmd.Run(); dockerErr != nil {
			common.SmartIDELog.Error(instanceI18nStart.Error.Docker_Err)
		}
		dockerpsCmd := exec.Command("docker", "ps")
		if dockerpsErr := dockerpsCmd.Run(); dockerpsErr != nil {
			common.SmartIDELog.Error(instanceI18nStart.Error.DockerPs_Err)
		}

		//0.2. 校验是否能正常执行 docker-compose
		dockercomposeCmd := exec.Command("docker-compose", "version")
		if dockercomposeErr := dockercomposeCmd.Run(); dockercomposeErr != nil {
			common.SmartIDELog.Error(instanceI18nStart.Error.Docker_Compose_Err)
		}

		//1. 获取docker compose的文件内容
		var yamlFileCongfig YamlFileConfig
		if ideyamlfile != "" { //增加指定yaml文件启动
			yamlFileCongfig.SetYamlFilePath(ideyamlfile)
		}
		yamlFileCongfig.GetConfig() // 读取配置
		var dockerCompose compose.DockerComposeYml
		dockeComposeYamlFilePath := dockerCompose.GetDockerComposeFilePath(yamlFileCongfig.Workspace.DevContainer.ServiceName) // 获取docker compose文件的路径

		//1.1. 校验docker compose文件对应的环境是否已经启动
		dockerComposeContainers := getDockerComposeContainers(dockeComposeYamlFilePath, dockerCompose.Services)
		if len(dockerComposeContainers) > 0 {
			common.SmartIDELog.Error("容器已经启动！") //TODO 如果已经启动，需要监听stop
		}

		//1.2. 保存文件
		dockerCompose, ideBindingPort, sshBindingPort := yamlFileCongfig.ConvertToDockerCompose() // 转换为docker compose格式
		err := dockerCompose.SaveFile(dockeComposeYamlFilePath)                                   // 保存docker compose文件
		if err != nil {
			common.SmartIDELog.Error(err, "save docker compose file:")
		}

		//1.3. print
		fmt.Printf("docker-compose: %v \n", dockeComposeYamlFilePath)
		fmt.Printf("SSH转发端口：%v \n", sshBindingPort) //TODO: 国际化	// 提示用户ssh端口绑定到了本地的某个端口

		//2. 创建容器
		ctx := context.Background()
		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			panic(err)
		}

		//2.1. 创建网络
		for network := range dockerCompose.Networks {
			networkList, _ := cli.NetworkList(ctx, types.NetworkListOptions{})
			isContain := false
			for _, item := range networkList {
				if item.Name == network {
					isContain = true
					break
				}
			}
			if !isContain {
				cli.NetworkCreate(ctx, network, types.NetworkCreate{})
				fmt.Print("创建网络 " + network) //TODO: 国际化
			}
		}

		//2.2. 运行docker-compose命令
		// e.g. docker-compose -f /Users/jasonchen/.ide/docker-compose-product-service-dev.yaml --project-directory /Users/jasonchen/Project/boat-house/boat-house-backend/src/product-service/api up -d
		pwd, _ := os.Getwd()
		composeCmd := exec.Command("docker-compose", "-f", dockeComposeYamlFilePath, "--project-directory", pwd, "up", "-d")

		composeCmd.Stdout = os.Stdout
		composeCmd.Stderr = os.Stderr
		if composeCmdErr := composeCmd.Run(); composeCmdErr != nil {
			common.SmartIDELog.Fatal(composeCmdErr)
		}

		//2.3. 配置gitconfig
		//ConfigGit()

		//3. 使用浏览器打开web ide
		fmt.Println(instanceI18nStart.Info.Info_running_openbrower) //TODO: 增加等待某某网址的提示

		var url string
		//指定yaml文件启动时候默认打开文件夹处理
		if ideyamlfile != "" {
			url = fmt.Sprintf("http://localhost:%v/?folder=vscode-remote://localhost:%v/home/project", ideBindingPort, ideBindingPort)
		} else {
			url = fmt.Sprintf(`http://localhost:%v`, ideBindingPort)
		}
		fmt.Println("打开", url) //TODO: 国际化
		isUrlReady := false
		for !isUrlReady {
			resp, err := http.Get(url)
			if (err == nil) && (resp.StatusCode == 200) {
				isUrlReady = true
				common.OpenBrowser(url)
			}

		}

		//99. 结束
		fmt.Println(instanceI18nStart.Info.Info_end)

		// tunnel， 死循环，不结束
		for {
			tunnel.AutoTunnelMultiple(fmt.Sprintf("localhost:%v", sshBindingPort), "root", "root123") //TODO: 登录的用户名，密码要能够从配置文件中读取出来
			time.Sleep(time.Second * 10)
		}

	},
}

// 获取docker compose运行起来对应的容器
func getDockerComposeContainers(dockerComposeFilePath string, dockerComposeServices map[string]compose.Service) []DockerComposeContainer {

	var dockerComposeContainers []DockerComposeContainer // result define

	//0. valid
	if !common.FileIsExit(dockerComposeFilePath) {
		return dockerComposeContainers
	}

	//1. docker-compose ps
	//1.1. copy
	runRemoveCommand4DockerComposeFile()                            // 删除文件
	outRunCopyCommand, err := runCopyCommand(dockerComposeFilePath) // 创建docker-compose.yaml文件
	if err != nil {
		runRemoveCommand4DockerComposeFile() // 删除文件
		common.SmartIDELog.Fatal(err, "copy file:", string(outRunCopyCommand))
	}

	//1.2. ps
	psExecCmd := exec.Command("docker-compose", "ps", "-a")
	outPSCommand, err := psExecCmd.CombinedOutput()
	outputPSCommand := string(outPSCommand)
	if err != nil {
		runRemoveCommand4DockerComposeFile() // 删除文件
		common.SmartIDELog.Fatal(err, "docker-compose ps -a:", outputPSCommand)
	}
	common.SmartIDELog.Info(outputPSCommand)
	runRemoveCommand4DockerComposeFile() // 删除文件

	//3. parse output
	/*
						e.g.
		                   Name                                  Command               State                                       Ports
		---------------------------------------------------------------------------------------------------------------------------------------------------------------------
		boathouse-calculator_boathouse-calculator_1   sh /usr/local/bin/entry_po ...   Up      0.0.0.0:7122->22/tcp,:::7122->22/tcp, 0.0.0.0:7100->3000/tcp,:::7100->3000/tcp
	*/
	lines := strings.Split(outputPSCommand, "\n")
	if len(lines) <= 2 { // 读取的结果有问题
		fmt.Println(outPSCommand)
		return dockerComposeContainers
	}
	resultLines := lines[2:]
	for _, line := range resultLines {
		// 不能为空
		if strings.TrimSpace(line) == "" {
			continue
		}

		// 拆分字符串
		params := strings.Split(strings.ReplaceAll(line, "...", " "), "  ")
		params = common.RemoveEmptyItem(params)

		// 获取基本信息
		dockerComposeContainer := DockerComposeContainer{
			ServiceName:   "",
			ContainerName: params[0],
			//Ports:         strings.TrimSpace(params[3]),
			State: strings.TrimSpace(params[2]),
		}

		// 从container name 中获取 service name
		for serviceName, item := range dockerComposeServices {
			if dockerComposeContainer.ContainerName == item.ContainerName {
				dockerComposeContainer.ServiceName = item.Name
			} else if strings.Contains(dockerComposeContainer.ContainerName, item.ContainerName) {
				dockerComposeContainer.ServiceName = item.Name
			} else if strings.Contains(dockerComposeContainer.ContainerName, serviceName) {
				dockerComposeContainer.ServiceName = item.Name
			} else {
				//TODO 查看端口映射是否相同
			}
		}

		dockerComposeContainers = append(dockerComposeContainers, dockerComposeContainer)
	}

	return dockerComposeContainers
}

// 运行拷贝文件命令
func runCopyCommand(dockerComposeFilePath string) (output string, err error) {
	command := ""
	var cpCommand *exec.Cmd

	switch runtime.GOOS {

	case "windows":
		command = fmt.Sprintf("copy %v docker-compose.yaml", dockerComposeFilePath)
		cpCommand = exec.Command("cmd", "/c", command)

	default:
		command = fmt.Sprintf("cp -i %v docker-compose.yaml", dockerComposeFilePath)
		cpCommand = exec.Command("bash", "-c", command)
	}

	bytes, err := cpCommand.CombinedOutput()
	output = string(bytes)

	return output, err
}

// 运行删除命令
func runRemoveCommand4DockerComposeFile() (output string, err error) {
	command := ""
	var removeCommand *exec.Cmd

	filePath := "docker-compose.yaml"
	switch runtime.GOOS {

	case "windows":
		command = fmt.Sprintf("del %v", filePath)
		removeCommand = exec.Command("cmd", "/c", command)

	default:
		command = fmt.Sprintf("rm -rf %v", filePath)
		removeCommand = exec.Command("bash", "-c", command)
	}

	bytes, err := removeCommand.CombinedOutput()
	output = string(bytes)

	return output, err
}

// 容器信息
type DockerComposeContainer struct {
	ServiceName   string
	ContainerName string
	//Command       string
	Ports string
	State string
}

func init() {

	startCmd.Flags().StringVarP(&ideyamlfile, "filepath", "f", "", "指定yaml文件路径")
	//rootCmd.AddCommand(startCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// startCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// startCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
