/*
 * @Author: jason chen (jasonchen@leansoftx.com, http://smallidea.cnblogs.com)
 * @Description: config
 * @Date: 2021-11
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-09-16 15:55:51
 */
package config

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"

	"io/ioutil"

	"github.com/leansoftX/smartide-cli/internal/apk/user"
	"github.com/leansoftX/smartide-cli/internal/model"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/leansoftX/smartide-cli/pkg/docker/compose"
	"github.com/leansoftX/smartide-cli/pkg/k8s"
	"gopkg.in/yaml.v2"
)

// 转换为yaml格式的文本
func (yamlFileConfig SmartIdeConfig) ToYaml() (result string, err error) {
	// 配置文件内容
	bytes, err := yaml.Marshal(yamlFileConfig)
	if err != nil {
		return result, err
	}
	result = string(bytes)

	result = strings.ReplaceAll(result, "\\'", "'")

	return result, err
}

// 从临时的 docker-compose 文件中加载配置
func (yamlFileConfig *SmartIdeConfig) LoadDockerComposeFromTempFile(sshRemote common.SSHRemote,
	tempDockerComposeFilePath string) (composeYaml compose.DockerComposeYml, ideBindingPort int, sshBindingPort int) {
	var yamlFileBytes []byte
	var err error

	//1. 变量赋值
	isVmCommand := false // 是否为 vm 命令模式，比如smartide vm start
	if (sshRemote != common.SSHRemote{}) {
		isVmCommand = true
	}

	// 读取生成的 docker-compose 文件
	if isVmCommand {
		remoteTempDockerComposeFilePath := common.FilePahtJoin4Linux(tempDockerComposeFilePath) //
		common.SmartIDELog.InfoF(i18nInstance.Config.Info_read_docker_compose, remoteTempDockerComposeFilePath)

		command := fmt.Sprintf(`cat %v`, remoteTempDockerComposeFilePath)
		output, err := sshRemote.ExeSSHCommand(command)
		common.CheckError(err, output)
		yamlFileBytes = []byte(output)

	} else {
		common.SmartIDELog.InfoF(i18nInstance.Config.Info_read_docker_compose, tempDockerComposeFilePath)

		// read and parse
		yamlFileBytes, err = ioutil.ReadFile(tempDockerComposeFilePath)
		common.CheckError(err)

	}

	// 解析docker-compose文件
	err = yaml.Unmarshal([]byte(yamlFileBytes), &composeYaml) // 为dockerCompose赋值
	if err != nil {
		common.SmartIDELog.Fatal(err)
	}

	// 获取 webide 、ssh 绑定端口
	if yamlFileConfig.Orchestrator.Type == OrchestratorTypeEnum_Compose ||
		yamlFileConfig.Orchestrator.Type == OrchestratorTypeEnum_Allinone { // 只有在compose模式下，在会去compose文件中查找端口是否申明
		for serviceName, service := range composeYaml.Services {
			if serviceName != yamlFileConfig.Workspace.DevContainer.ServiceName {
				continue
			}
			for _, port := range service.Ports {
				containerWebIDEPort := yamlFileConfig.GetContainerWebIDEPort()
				if containerWebIDEPort != nil && strings.Contains(port, ":"+strconv.Itoa(*containerWebIDEPort)) { // webide 端口
					index := strings.Index(port, ":")
					if index > 0 {
						ideBindingPort, _ = strconv.Atoi(port[:index])
						if ideBindingPort > 0 {
							common.SmartIDELog.DebugF(i18nInstance.Common.Info_ssh_webide_host_port, ideBindingPort)
							continue
						}
					}
				} else if strings.Contains(port, ":"+strconv.Itoa(model.CONST_Container_SSHPort)) { // ssh 端口
					index := strings.Index(port, ":")
					if index > 0 {
						sshBindingPort, _ = strconv.Atoi(port[:index])
						if sshBindingPort > 0 {
							common.SmartIDELog.DebugF(i18nInstance.Common.Info_ssh_host_port, sshBindingPort)
							continue
						}
					}
				}
			}
		}
	}
	//TODO: 在k8s 的yaml文件中查找端口是否申明

	return composeYaml, ideBindingPort, sshBindingPort
}

// 把自定义的配置转换为 docker compose
func (yamlFileConfig *SmartIdeConfig) ConvertToDockerCompose(sshRemote common.SSHRemote, projectName string,
	remoteWorkingDir string, isCheckUnuesedPorts bool, userName string) (composeYaml compose.DockerComposeYml, ideBindingPort int, sshBindingPort int) {

	ideBindingPort = model.CONST_Local_Default_BindingPort_WebIDE // webide
	sshBindingPort = model.CONST_Local_Default_BindingPort_SSH    // ssh
	var dockerCompose compose.DockerComposeYml

	//1.
	//1.1. 变量赋值
	isRemoteMode := false // 是否为 vm 命令模式，比如smartide vm start
	if (sshRemote != common.SSHRemote{}) {
		isRemoteMode = true
	}

	//1.2. 文件格式检查
	if yamlFileConfig.Orchestrator.Type == OrchestratorTypeEnum_Compose ||
		yamlFileConfig.Orchestrator.Type == OrchestratorTypeEnum_Allinone {
		if yamlFileConfig.IsLinkDockerComposeFile() { // 链接了 docker-compose 文件
			if !isRemoteMode {
				// 检查docker-compose文件是否存在
				localDockerComposeFilePath, _ := yamlFileConfig.GetLocalLinkDockerComposeFile() // 本地docker compose文件的路径

				// 检查文件是否存在
				if !common.IsExist(localDockerComposeFilePath) {
					message := fmt.Sprintf(i18nInstance.Config.Err_file_not_exit, yamlFileConfig.Workspace.DockerComposeFile)
					common.SmartIDELog.Error(message)
				}
			}

		} else { // 没有链接 docker-compose 文件
			// 检查是否有services节点
			if len(yamlFileConfig.Workspace.Servcies) <= 0 {
				common.SmartIDELog.Error(i18nInstance.Config.Err_services_not_exit)
			}
		}
	}

	//2. 转换为docker-compose - 基本转换
	//2.1.
	dockerCompose, err := yamlFileConfig.getDockerCompose(sshRemote, remoteWorkingDir)
	common.CheckError(err)

	//2.2. 检查devContainer中定义的service时候存在于services中
	if _, ok := dockerCompose.Services[yamlFileConfig.Workspace.DevContainer.ServiceName]; !ok { // 是否定义了devContainer节点对应的service
		err := fmt.Sprintf(i18nInstance.Config.Err_devcontainer_not_contains, yamlFileConfig.Workspace.DevContainer.ServiceName)
		common.SmartIDELog.Error(err)
	}

	//3. 转换为docker compose - 端口绑定
	//3.1. 端口映射
	for serviceName, service := range dockerCompose.Services {

		// 如果设置了container name，就在container name前面加 project name（文件夹名称）
		if service.ContainerName != "" {
			service.ContainerName = projectName + "-" + service.ContainerName
			dockerCompose.Services[serviceName] = service
		}

		// 绑定端口被占用的问题
		if isCheckUnuesedPorts {
			hasChange := false
			for index, portMapStr := range service.Ports {
				binding := strings.Split(portMapStr, ":")
				bindingPortOld, err := strconv.Atoi(binding[0])
				common.CheckError(err)

				containerPort, err := strconv.Atoi(binding[1])
				common.CheckError(err)

				bindingPortNew, err := checkAndGetAvailableRemotePort(sshRemote, bindingPortOld, 10) // 检测端口是否被占用
				common.CheckError(err)
				if bindingPortOld != bindingPortNew {
					service.Ports[index] = fmt.Sprintf("%v:%v", bindingPortNew, containerPort)

					common.SmartIDELog.DebugF("localhost:%v (%v 被占用) -> container:%v", bindingPortNew, bindingPortOld, containerPort)
					hasChange = true

					// ide、ssh端口更新
					if serviceName == yamlFileConfig.Workspace.DevContainer.ServiceName {
						containerWebIDEPort := yamlFileConfig.GetContainerWebIDEPort()
						if containerWebIDEPort != nil && containerPort == *containerWebIDEPort {
							ideBindingPort = bindingPortNew
						} else if containerPort == model.CONST_Container_SSHPort {
							sshBindingPort = bindingPortNew
						}
					}
				} else {
					common.SmartIDELog.DebugF("localhost:%v -> container:%v", bindingPortOld, containerPort)
				}
				yamlFileConfig.setPort4Label(containerPort, bindingPortOld, bindingPortNew, serviceName)
			}
			if hasChange {
				dockerCompose.Services[serviceName] = service
			}
		}

		// 注入webide、ssh端口，确保存在两个必要的端口
		if serviceName == yamlFileConfig.Workspace.DevContainer.ServiceName {

			// webide port
			containerWebIDEPort := yamlFileConfig.GetContainerWebIDEPort()
			if containerWebIDEPort != nil &&
				!service.ContainContainerPort(*containerWebIDEPort) &&
				yamlFileConfig.Workspace.DevContainer.IdeType != IdeTypeEnum_SDKOnly {
				// 是否检查端口被占用
				if isCheckUnuesedPorts {
					// webide port
					originIdeBindingPort := ideBindingPort
					newIdeBindingPort, err0 := checkAndGetAvailableRemotePort(sshRemote, originIdeBindingPort, 100) // 检测端口是否被占用
					common.CheckError(err0)

					yamlFileConfig.setPort4Label(*containerWebIDEPort, originIdeBindingPort, newIdeBindingPort, serviceName)

					if newIdeBindingPort != originIdeBindingPort {
						ideBindingPort = newIdeBindingPort
					}
				}

				//
				service.AppendPort(strconv.Itoa(ideBindingPort) + ":" + strconv.Itoa(*containerWebIDEPort))
			}

			// ssh port
			if !service.ContainContainerPort(model.CONST_Container_SSHPort) {
				// 是否检查端口被占用
				if isCheckUnuesedPorts {
					// ssh port
					originSshBindingPort := sshBindingPort
					newSshBindingPort, err0 := checkAndGetAvailableRemotePort(sshRemote, originSshBindingPort, 100) // 检测端口是否被占用
					common.CheckError(err0)

					yamlFileConfig.setPort4Label(model.CONST_Container_SSHPort, originSshBindingPort, newSshBindingPort, serviceName)

					if newSshBindingPort != originSshBindingPort {
						sshBindingPort = newSshBindingPort
					}
				}

				//
				service.AppendPort(strconv.Itoa(sshBindingPort) + ":" + strconv.Itoa(model.CONST_Container_SSHPort))
			}

			dockerCompose.Services[serviceName] = service
		}
	}
	//3.2. 遍历端口描述，添加遗漏的端口
	for label, port := range yamlFileConfig.Workspace.DevContainer.Ports {
		hasContain := false
		for _, item := range yamlFileConfig.Workspace.DevContainer.bindingPorts {
			if item.OriginHostPort == port {
				hasContain = true
				break
			}
		}

		if !hasContain {
			common.SmartIDELog.Importance(fmt.Sprintf("在service定义中找不到端口: %v (%v) ", port, label))
			portMap := NewPortMap(PortMapInfo_OnlyLabel, port, -1, label, -1, "")
			yamlFileConfig.Workspace.DevContainer.bindingPorts = append(yamlFileConfig.Workspace.DevContainer.bindingPorts, *portMap)
		}

	}

	//4. ssh volume配置
	for serviceName, service := range dockerCompose.Services {
		if serviceName == yamlFileConfig.Workspace.DevContainer.ServiceName {

			if yamlFileConfig.Workspace.DevContainer.Volumes.HasSshKey.Value() {
				SSHVolumesConfig(isRemoteMode, &service, sshRemote, userName)
			}

			if yamlFileConfig.Workspace.DevContainer.Volumes.HasGitConfig.Value() {
				GitConfig(isRemoteMode, "", nil, &service, sshRemote, k8s.ExecInPodRequest{})
			}

			dockerCompose.Services[serviceName] = service

		}
	}

	//5. 项目目录 volume
	for serviceName, service := range dockerCompose.Services {
		if serviceName == yamlFileConfig.Workspace.DevContainer.ServiceName {

			// 查找目录映射的volume
			indexProjectVolume := -1
			volumeProject := ""
			for indexVolume, volume := range service.Volumes {
				if strings.Contains(volume, "/home/project") {
					indexProjectVolume = indexVolume
					volumeProject = volume
				}
			}

			// 当前工作目录
			twd, err := os.Getwd()
			if isRemoteMode {
				twd = sshRemote.ConvertFilePath(yamlFileConfig.GetWorkingDirectoryPath())
			}
			common.CheckError(err)

			// 设置目录映射值
			isWindows := runtime.GOOS == "windows"
			if indexProjectVolume > -1 { // 当存在配置时，需要吧把 “.” 替换为当前目录
				// 本地模式下，把“.”替换为当前目录
				if strings.Index(volumeProject, ".") == 0 {
					if isWindows && !isRemoteMode {
						service.Volumes[indexProjectVolume] = "\\'" + twd + volumeProject[1:] + "\\'"
					} else {
						service.Volumes[indexProjectVolume] = twd + volumeProject[1:]
					}

					// 重置
					dockerCompose.Services[serviceName] = service
				}

			} else { // insert default project volume
				if isWindows && !isRemoteMode {
					service.Volumes = append(service.Volumes, fmt.Sprintf("\\'%v:%v\\'", twd, "/home/project/"+projectName))
				} else {
					service.Volumes = append(service.Volumes, fmt.Sprintf("%v:%v", twd, "/home/project/"+projectName))
				}

				// 重置
				dockerCompose.Services[serviceName] = service

			}

			break
		}
	}

	//6.替换images地址，功能暂时注释，待完善
	// for serviceName, service := range dockerCompose.Services {
	// 	if service.Image.Name != "" {
	// 		if strings.Index(service.Image.Name, "registry.cn-hangzhou.aliyuncs.com") < 0 {
	// 			service.Image.Name = fmt.Sprintf("%v/%v", GlobalSmartIdeConfig.ImagesRegistry, service.Image.Name)
	// 		} else {
	// 			imageName := strings.Split(service.Image.Name, "registry.cn-hangzhou.aliyuncs.com")
	// 			service.Image.Name = fmt.Sprintf("%v%v", GlobalSmartIdeConfig.ImagesRegistry, imageName[1])
	// 		}
	// 	}
	// 	dockerCompose.Services[serviceName] = service
	// }

	//7.获取uid,gid设置到环境变量
	for serviceName, service := range dockerCompose.Services {

		if service.Environment == nil {
			service.Environment = map[string]string{}
		} else {
			for k, v := range service.Environment {
				common.SmartIDELog.DebugF("ENV---%v-----%v: %v", serviceName, k, v)
				service.Environment[k] = v
			}
		}
		// 只有IDE容器需要动态赋值uid,gid
		if serviceName == yamlFileConfig.Workspace.DevContainer.ServiceName {
			if isRemoteMode {
				uid, gid := sshRemote.GetRemoteUserInfo()
				service.Environment[model.CONST_LOCAL_USER_UID] = uid
				service.Environment[model.CONST_LOCAL_USER_GID] = gid
			} else {
				localuser := user.GetUserInfo()
				service.Environment[model.CONST_LOCAL_USER_UID] = localuser.Uid
				service.Environment[model.CONST_LOCAL_USER_GID] = localuser.Gid
			}

			if service.Environment[model.CONST_ENV_NAME_LoalUserPassword] == "" {
				service.Environment[model.CONST_ENV_NAME_LoalUserPassword] = model.CONST_DEV_CONTAINER_USER_DEFAULT_PASSWORD //smartide123.@IDE
			}

		}

		dockerCompose.Services[serviceName] = service
	}

	return dockerCompose, ideBindingPort, sshBindingPort
}

func checkAndGetAvailableRemotePort(sshRemote common.SSHRemote, bindingPortOld int, step int) (bindingPortNew int, err error) {
	isRemoteMode := sshRemote != common.SSHRemote{}
	if isRemoteMode {
		bindingPortNew = sshRemote.CheckAndGetAvailableRemotePort(bindingPortOld, step) // 在远程主机上检测端口是否被占用
	} else {
		bindingPortNew, err = common.CheckAndGetAvailableLocalPort(bindingPortOld, step) // 在本地主机上检测端口是否被占用
	}
	return
}

// 获取docker compose
func (yamlFileConfig SmartIdeConfig) getDockerCompose(sshRemote common.SSHRemote, remoteWorkingDir string) (
	dockerCompose compose.DockerComposeYml, err error) {
	isRemoteMode := false // 是否为 vm 命令模式，比如smartide vm start
	if (sshRemote != common.SSHRemote{}) {
		isRemoteMode = true
	}

	//2.1. 链接 docker-compose 文件
	if yamlFileConfig.Workspace.DockerComposeFile != "" {
		var dockerComposeFileBytes []byte // docker-compose文件的流信息
		//	var err error

		if isRemoteMode {
			// 获取docker-compose文件在远程主机上的路径
			remoteDockerComposeFilePath := common.FilePahtJoin4Linux(remoteWorkingDir, ".ide", yamlFileConfig.Workspace.DockerComposeFile)
			common.SmartIDELog.InfoF(i18nInstance.Config.Info_read_docker_compose, remoteDockerComposeFilePath)

			// 在远程主机上加载docker-compose文件
			command := fmt.Sprintf(`cat %v`, remoteDockerComposeFilePath)
			output, err := sshRemote.ExeSSHCommand(command)
			if err != nil {
				return dockerCompose, err
			}
			dockerComposeFileBytes = []byte(output)

		} else {
			// read and parse
			common.SmartIDELog.InfoF(i18nInstance.Config.Info_read_docker_compose, yamlFileConfig.Workspace.DockerComposeFile)
			linkDockerComposeFilePath, _ := yamlFileConfig.GetLocalLinkDockerComposeFile()
			dockerComposeFileBytes, err = ioutil.ReadFile(linkDockerComposeFilePath)
			if err != nil {
				return dockerCompose, err
			}

		}

		err = yaml.Unmarshal([]byte(dockerComposeFileBytes), &dockerCompose) // 为dockerCompose赋值
		if err != nil {
			return dockerCompose, err
		}

	} else {
		// 确保不会有引用类型的问题
		configContent, err := yamlFileConfig.ToYaml()
		if err != nil {
			return dockerCompose, err
		}
		var notReferenceConfig SmartIdeConfig
		yaml.Unmarshal([]byte(configContent), &notReferenceConfig)

		// 使用新对象赋值
		dockerCompose.Version = notReferenceConfig.Orchestrator.Version
		dockerCompose.Services = make(map[string]compose.Service)
		for serviceName, service := range notReferenceConfig.Workspace.Servcies {
			s := service
			dockerCompose.Services[serviceName] = s
		}
		dockerCompose.Networks = notReferenceConfig.Workspace.Networks
		dockerCompose.Volumes = notReferenceConfig.Workspace.Volumes
		dockerCompose.Secrets = notReferenceConfig.Workspace.Secrets
	}

	return dockerCompose, err
}
