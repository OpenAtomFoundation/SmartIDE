package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/leansoftX/smartide-cli/cmd/lib"
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
		common.SmartIDELog.Info(i18n.GetInstance().Start.Info.Info_start)

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
		var yamlFileCongfig lib.YamlFileConfig
		if ideyamlfile != "" { //增加指定yaml文件启动
			yamlFileCongfig.SetYamlFilePath(ideyamlfile)
		}
		yamlFileCongfig.GetConfig() // 读取配置
		var dockerCompose compose.DockerComposeYml
		dockeComposeYamlFilePath :=
			dockerCompose.GetTmpDockerComposeFilePath(yamlFileCongfig.Workspace.DevContainer.ServiceName) // 获取临时docker-compose文件的路径
		dockerCompose, ideBindingPort, sshBindingPort := yamlFileCongfig.ConvertToDockerCompose(common.SSHRemote{}, "") // 转换为docker compose格式

		//1.1. 校验docker compose文件对应的环境是否已经启动
		dockerComposeContainers := getDockerComposeContainers(dockeComposeYamlFilePath, dockerCompose.Services)
		if len(dockerComposeContainers) > 0 {
			common.SmartIDELog.Error(instanceI18nStart.Error.Docker_started) //TODO 如果已经启动，需要监听stop
		}

		//1.2. 生成docker-compose文件内容并保存
		err := dockerCompose.SaveFile(dockeComposeYamlFilePath) // 保存docker-compose文件
		common.CheckError(err)

		//1.3. print
		common.SmartIDELog.InfoF(instanceI18nStart.Info.Info_docker_compose_filepath, dockeComposeYamlFilePath)
		common.SmartIDELog.InfoF(instanceI18nStart.Info.Info_ssh_tunnel, sshBindingPort) // 提示用户ssh端口绑定到了本地的某个端口

		//2. 创建容器
		ctx := context.Background()
		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		common.CheckError(err)

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
				common.SmartIDELog.InfoF(instanceI18nStart.Info.Info_create_network, network)
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
		//3. 使用浏览器打开web ide
		common.SmartIDELog.Info(instanceI18nStart.Info.Info_running_openbrower)

		var url string
		//vscode启动时候默认打开文件夹处理
		if yamlFileCongfig.Workspace.DevContainer.IdeType == "vscode" {
			url = fmt.Sprintf("http://localhost:%v/?folder=vscode-remote://localhost:%v/home/project", ideBindingPort, ideBindingPort)
		} else {
			url = fmt.Sprintf(`http://localhost:%v`, ideBindingPort)
		}
		common.SmartIDELog.Info(instanceI18nStart.Info.Info_open_in_brower, url)
		isUrlReady := false
		for !isUrlReady {
			resp, err := http.Get(url)
			if (err == nil) && (resp.StatusCode == 200) {
				isUrlReady = true
				common.OpenBrowser(url)
			}

		}

		// 启动的容器列表
		dockerComposeContainers = getDockerComposeContainers(dockeComposeYamlFilePath, dockerCompose.Services)
		var containerName string
		for _, container := range dockerComposeContainers {
			if container.ServiceName == yamlFileCongfig.Workspace.DevContainer.ServiceName {
				containerName = container.ContainerName
				break
			}
		}
		docker := *common.NewDocker(cli)
		out := ""
		out, err = docker.Exec(context.Background(), strings.ReplaceAll(containerName, "/", ""), "/bin", []string{"chmod", "-R", "700", "/root"}, []string{})
		common.CheckError(err)
		common.SmartIDELog.Debug(out)

		//99. 结束
		common.SmartIDELog.Info(instanceI18nStart.Info.Info_end)

		// tunnel， 死循环，不结束
		for {
			tunnel.AutoTunnelMultiple(fmt.Sprintf("localhost:%v", sshBindingPort), "root", "root123", dockerCompose.GetLocalBindingPorts()) //TODO: 登录的用户名，密码要能够从配置文件中读取出来
			time.Sleep(time.Second * 10)
		}

	},
}

// 获取docker compose运行起来对应的容器
func getDockerComposeContainers(dockerComposeFilePath string, dockerComposeServices map[string]compose.Service) []DockerComposeContainer {

	var dockerComposeContainers []DockerComposeContainer // result define

	//0. valid
	if !common.IsExit(dockerComposeFilePath) {
		return dockerComposeContainers
	}

	//第一步：获取ctx
	ctx := context.Background()

	//获取cli客户端对象
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	common.CheckError(err)

	//通过cli客户端对象去执行ContainerList(其实docker ps 不就是一个docker正在运行容器的一个list嘛)
	containers, err2 := cli.ContainerList(ctx, types.ContainerListOptions{})
	common.CheckError(err2)

	//
	for serviceName, _ := range dockerComposeServices {

		for _, container := range containers {

			if container.Labels["com.docker.compose.service"] == serviceName {
				var ports []string
				for _, port := range container.Ports {
					str := fmt.Sprintf("%v:%v", port.PublicPort, port.PrivatePort)
					if !common.Contains(ports, str) { // 限制重复的端口绑定信息
						ports = append(ports, str)
					}
				}

				dockerComposeContainer := DockerComposeContainer{
					ServiceName:   serviceName,
					ContainerName: strings.Join(container.Names, ","),
					State:         container.State,
					Image:         container.Image,
					Ports:         strings.Join(ports, ";"),
				}
				dockerComposeContainers = append(dockerComposeContainers, dockerComposeContainer)
				break
			}

		}
	}

	// 打印
	printDockerComposeContainers(dockerComposeContainers)

	return dockerComposeContainers
}

// 打印 service 列表
func printDockerComposeContainers(dockerComposeContainers []DockerComposeContainer) {
	if len(dockerComposeContainers) <= 0 {
		return
	}
	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
	fmt.Fprintln(w, "service\tstate\timage\tports\t")
	for _, service := range dockerComposeContainers {
		line := fmt.Sprintf("%v\t%v\t%v\t%v\t", service.ServiceName, service.State, service.Image, service.Ports)
		fmt.Fprintln(w, line)
	}
	w.Flush()
}

// 容器信息
type DockerComposeContainer struct {
	ServiceName   string
	ContainerName string
	//Command       string
	Image string
	Ports string
	State string
}

func init() {

	startCmd.Flags().StringVarP(&ideyamlfile, "filepath", "f", "", instanceI18nStart.Info.Help_flag_filepath)

}
