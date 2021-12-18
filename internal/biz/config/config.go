/*
 * @Author: jason chen (jasonchen@leansoftx.com, http://smallidea.cnblogs.com)
 * @Description:
 * @Date: 2021-11
 * @LastEditors:
 * @LastEditTime:
 */
package config

import (
	"fmt"
	"strconv"
	"strings"

	"io/ioutil"

	"github.com/leansoftX/smartide-cli/internal/apk/user"
	"github.com/leansoftX/smartide-cli/internal/model"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/leansoftX/smartide-cli/pkg/docker/compose"
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
	for serviceName, service := range composeYaml.Services {
		if serviceName != yamlFileConfig.Workspace.DevContainer.ServiceName {
			continue
		}
		for _, port := range service.Ports {
			if strings.Contains(port, ":"+strconv.Itoa(model.CONST_Container_WebIDEPort)) { // webide 端口
				index := strings.Index(port, ":")
				if index > 0 {
					ideBindingPort, _ = strconv.Atoi(port[:index])
					if ideBindingPort > 0 {
						common.SmartIDELog.InfoF(i18nInstance.Common.Info_ssh_webide_host_port, ideBindingPort)
						continue
					}
				}
			} else if strings.Contains(port, ":"+strconv.Itoa(model.CONST_Container_SSHPort)) { // ssh 端口
				index := strings.Index(port, ":")
				if index > 0 {
					sshBindingPort, _ = strconv.Atoi(port[:index])
					if sshBindingPort > 0 {
						common.SmartIDELog.InfoF(i18nInstance.Common.Info_ssh_host_port, sshBindingPort)
						continue
					}
				}
			}
		}
	}

	return composeYaml, ideBindingPort, sshBindingPort
}

// 把自定义的配置转换为docker compose
func (yamlFileConfig *SmartIdeConfig) ConvertToDockerCompose(sshRemote common.SSHRemote, projectName string,
	remoteConfigDir string, isCheckUnuesedPorts bool) (composeYaml compose.DockerComposeYml, ideBindingPort int, sshBindingPort int) {

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
	if strings.ToLower(yamlFileConfig.Orchestrator.Type) == "docker-compose" {
		if yamlFileConfig.IsLinkDockerComposeFile() { // 链接了 docker-compose 文件
			if !isRemoteMode {
				// 检查docker-compose文件是否存在
				localDockerComposeFilePath, _ := yamlFileConfig.GetLocalLinkDockerComposeFile() // 本地docker compose文件的路径

				// 检查文件是否存在
				if !common.IsExit(localDockerComposeFilePath) {
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
	//2.1. 链接 docker-compose 文件
	if yamlFileConfig.Workspace.DockerComposeFile != "" {
		var dockerComposeFileBytes []byte // docker-compose文件的流信息
		var err error

		if isRemoteMode {
			// 获取docker-compose文件在远程主机上的路径
			remoteDockerComposeFilePath := common.FilePahtJoin(common.OS_Linux, remoteConfigDir, yamlFileConfig.Workspace.DockerComposeFile)
			common.SmartIDELog.InfoF(i18nInstance.Config.Info_read_docker_compose, remoteDockerComposeFilePath)

			// 在远程主机上加载docker-compose文件
			command := fmt.Sprintf(`cat %v`, remoteDockerComposeFilePath)
			output, err := sshRemote.ExeSSHCommand(command)
			common.CheckError(err, output)
			dockerComposeFileBytes = []byte(output)

		} else {
			// read and parse
			common.SmartIDELog.InfoF(i18nInstance.Config.Info_read_docker_compose, yamlFileConfig.Workspace.DockerComposeFile)
			linkDockerComposeFilePath, _ := yamlFileConfig.GetLocalLinkDockerComposeFile()
			dockerComposeFileBytes, err = ioutil.ReadFile(linkDockerComposeFilePath)
			common.CheckError(err)

		}

		err = yaml.Unmarshal([]byte(dockerComposeFileBytes), &dockerCompose) // 为dockerCompose赋值
		common.CheckError(err)

	} else {
		// 确保不会有引用类型的问题
		configContentBytes, err := yaml.Marshal(yamlFileConfig)
		common.CheckError(err)
		var notReferenceConfig SmartIdeConfig
		yaml.Unmarshal(configContentBytes, &notReferenceConfig)

		// 使用新对象赋值
		dockerCompose.Version = notReferenceConfig.Orchestrator.Version
		dockerCompose.Services = notReferenceConfig.Workspace.Servcies
		dockerCompose.Networks = notReferenceConfig.Workspace.Networks
		dockerCompose.Volumes = notReferenceConfig.Workspace.Volumes
		dockerCompose.Secrets = notReferenceConfig.Workspace.Secrets

	}

	//2.2. 检查devContainer中定义的service时候存在于services中
	if _, ok := dockerCompose.Services[yamlFileConfig.Workspace.DevContainer.ServiceName]; !ok { // 是否定义了devContainer节点对应的service
		err := fmt.Sprintf(i18nInstance.Config.Err_devcontainer_not_contains, yamlFileConfig.Workspace.DevContainer.ServiceName) //TODO：国际化
		common.SmartIDELog.Error(err)
	}

	//3. 转换为docker compose - 端口绑定
	if isRemoteMode { //3.1. vm 命令模式下，即remote远程主机，只需要自动绑定ide端口，但不需要绑定22

		// 端口映射
		for serviceName, service := range dockerCompose.Services {
			if serviceName == yamlFileConfig.Workspace.DevContainer.ServiceName {
				// 是否检查端口被占用
				if isCheckUnuesedPorts {
					newIdeBindingPort := sshRemote.CheckAndGetAvailableRemotePort(ideBindingPort, 10) // 在远程主机上获取一个未被占用的端口
					if newIdeBindingPort != ideBindingPort {
						ideBindingPort = newIdeBindingPort
					}
					yamlFileConfig.setPort4Label(model.CONST_Container_WebIDEPort, ideBindingPort, newIdeBindingPort, serviceName)
				}

				service.AppendPort(strconv.Itoa(ideBindingPort) + ":" + strconv.Itoa(model.CONST_Container_WebIDEPort))
				dockerCompose.Services[serviceName] = service
			}

			// 绑定端口被占用的问题
			if isCheckUnuesedPorts {
				hasChange := false
				for index, port := range service.Ports {
					binding := strings.Split(port, ":")
					bindingPortOld, err := strconv.Atoi(binding[0])
					common.CheckError(err)

					containerPort, err := strconv.Atoi(binding[1])
					common.CheckError(err)

					bindingPortNew := sshRemote.CheckAndGetAvailableRemotePort(bindingPortOld, 10) // 在远程主机上检测端口是否被占用
					if bindingPortOld != bindingPortNew {
						service.Ports[index] = fmt.Sprintf("%v:%v", bindingPortNew, binding[1])
						common.SmartIDELog.InfoF("%v -> %v", port, service.Ports[index])
						hasChange = true

					}
					yamlFileConfig.setPort4Label(containerPort, bindingPortOld, bindingPortNew, serviceName)
				}
				if hasChange {
					dockerCompose.Services[serviceName] = service
				}
			}
		}
	} else { //3.2. 本地模式（非远程模式下），需要ide端口、22端口的绑定

		// 端口映射
		for serviceName, service := range dockerCompose.Services {
			if serviceName == yamlFileConfig.Workspace.DevContainer.ServiceName {
				if isCheckUnuesedPorts {
					newSshBindingPort := common.CheckAndGetAvailableLocalPort(sshBindingPort, 100) //
					if newSshBindingPort != sshBindingPort {
						sshBindingPort = newSshBindingPort

					}
					yamlFileConfig.setPort4Label(model.CONST_Container_SSHPort, sshBindingPort, newSshBindingPort, serviceName)

					newIdeBindingPort := common.CheckAndGetAvailableLocalPort(ideBindingPort, 10) //
					if newIdeBindingPort != ideBindingPort {
						ideBindingPort = newIdeBindingPort

					}
					yamlFileConfig.setPort4Label(model.CONST_Container_WebIDEPort, ideBindingPort, newIdeBindingPort, serviceName)
				}

				service.AppendPort(strconv.Itoa(sshBindingPort) + ":" + strconv.Itoa(model.CONST_Container_SSHPort))
				service.AppendPort(strconv.Itoa(ideBindingPort) + ":" + strconv.Itoa(model.CONST_Container_WebIDEPort))

				dockerCompose.Services[serviceName] = service
			}

			// 绑定端口被占用的问题
			if isCheckUnuesedPorts {
				hasChange := false
				for index, portMap := range service.Ports {
					binding := strings.Split(portMap, ":")
					bindingPortOld, err := strconv.Atoi(binding[0])
					common.CheckError(err)

					containerPort, err := strconv.Atoi(binding[1])
					common.CheckError(err)

					// 获取到一个可用的端口
					bindingPortNew := common.CheckAndGetAvailableLocalPort(bindingPortOld, 10)
					if bindingPortOld != bindingPortNew {
						service.Ports[index] = fmt.Sprintf("%v:%v", bindingPortNew, binding[1])
						common.SmartIDELog.InfoF("%v -> %v", portMap, service.Ports[index])
						hasChange = true

					}
					yamlFileConfig.setPort4Label(containerPort, bindingPortOld, bindingPortNew, serviceName)

				}
				if hasChange {
					dockerCompose.Services[serviceName] = service
				}
			}

		}
	}
	//3.1. 遍历端口描述，添加遗漏的端口
	for label, port := range yamlFileConfig.Workspace.DevContainer.Ports {

		hasContain := false
		for _, item := range yamlFileConfig.Workspace.DevContainer.bindingPorts {
			if item.OriginLocalPort == port {
				hasContain = true
				break
			}
		}

		if !hasContain {
			portMap := NewPortMap(PortMapInfo_OnlyLabel, port, -1, label, -1, "")
			yamlFileConfig.Workspace.DevContainer.bindingPorts = append(yamlFileConfig.Workspace.DevContainer.bindingPorts, *portMap)
		}

	}

	//4. ssh volume配置
	sshKey := yamlFileConfig.Workspace.DevContainer.Volumes.SshKey
	gitconfig := yamlFileConfig.Workspace.DevContainer.Volumes.GitConfig
	for serviceName, service := range dockerCompose.Services {
		if serviceName == yamlFileConfig.Workspace.DevContainer.ServiceName {
			service = SSHVolumesConfig(sshKey, isRemoteMode, service, sshRemote)
			service = GitConfig(gitconfig, isRemoteMode, "", nil, service, sshRemote)
			dockerCompose.Services[serviceName] = service

		}
	}

	//5. 项目目录 volume
	for serviceName, service := range dockerCompose.Services {
		if serviceName == yamlFileConfig.Workspace.DevContainer.ServiceName {
			for indexVolume, volume := range service.Volumes {
				if strings.Contains(volume, "/home/project") {
					service.Volumes[indexVolume] = strings.Replace(volume, "/home/project", "/home/project/"+projectName, -1)
				}
			}
			dockerCompose.Services[serviceName] = service
			break
		}
	}

	//6.替换images地址
	for serviceName, service := range dockerCompose.Services {
		if service.Image.Name != "" {
			if strings.Index(service.Image.Name, "registry.cn-hangzhou.aliyuncs.com") < 0 {
				service.Image.Name = fmt.Sprintf("%v/%v", GlobalSmartIdeConfig.ImagesRegistry, service.Image.Name)
			} else {
				imageName := strings.Split(service.Image.Name, "registry.cn-hangzhou.aliyuncs.com")
				service.Image.Name = fmt.Sprintf("%v%v", GlobalSmartIdeConfig.ImagesRegistry, imageName[1])
			}
		}
		dockerCompose.Services[serviceName] = service
	}

	//7.获取uid,gid设置到环境变量
	for serviceName, service := range dockerCompose.Services {
		service.Environment = map[string]string{}

		if service.Environment == nil {
			service.Environment = map[string]string{}
		}

		if service.Environment == nil {
			service.Environment = map[string]string{}
		}

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
		dockerCompose.Services[serviceName] = service
	}

	return dockerCompose, ideBindingPort, sshBindingPort
}
