/*
 * @Date: 2022-03-23 16:15:38
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-05-17 10:39:48
 * @FilePath: /smartide-cli/cmd/start/k8s.go
 */

package start

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/leansoftX/smartide-cli/internal/apk/i18n"
	"github.com/leansoftX/smartide-cli/internal/biz/config"
	"github.com/leansoftX/smartide-cli/internal/biz/workspace"
	"github.com/leansoftX/smartide-cli/internal/dal"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/leansoftX/smartide-cli/pkg/kubectl"
	"gopkg.in/yaml.v2"

	k8sScheme "k8s.io/client-go/kubernetes/scheme"

	appV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
)

// 执行k8s start
func ExecuteK8sStartCmd(workspaceInfo workspace.WorkspaceInfo, yamlExecuteFun func(yamlConfig config.SmartIdeConfig)) {
	common.SmartIDELog.Info(i18nInstance.Start.Info_k8s_init)

	kubernetes, _ := kubectl.NewKubernetes(workspaceInfo.K8sInfo.Namespace)

	//1. 检测是否安装kubectl
	common.SmartIDELog.Info("检测kubectl（v1.23.0）是否安装到 \"用户目录/.ide\"")
	err := checkAndInstallKubectl(kubernetes)
	common.CheckError(err)

	//2. 切换到指定的k8s
	common.SmartIDELog.Info("切换到指定的 k8s context: " + workspaceInfo.K8sInfo.Context)
	currentContext, err := kubernetes.ExecKubectlCommandCombined("config current-context", "")
	common.CheckError(err)
	currentContext = strings.TrimSpace(currentContext)
	if workspaceInfo.K8sInfo.Context != currentContext { //
		err = kubernetes.ExecKubectlCommandRealtime(fmt.Sprintf("config use-context %v", workspaceInfo.K8sInfo.Context), "", false)
		common.CheckError(err)
		common.SmartIDELog.Info("切换k8s上下文到 " + workspaceInfo.K8sInfo.Context)
	}

	//3. 解析 .k8s.ide.yaml 文件（是否需要注入到deploy.yaml文件中）
	common.SmartIDELog.Info("下载配置文件 及 关联k8s yaml文件")
	gitRepoRootDirPath, configFileRelativePath, k8sYamlRelativePaths, err := downloadConfigAndLinkFiles(workspaceInfo)
	common.CheckError(err)
	//3.1. 解析配置文件 + 关联的k8s yaml
	common.SmartIDELog.Info(fmt.Sprintf("解析配置文件 %v", workspaceInfo.ConfigFileRelativePath))
	configFileAbsolutePath := common.PathJoin(gitRepoRootDirPath, configFileRelativePath) // 配置文件绝对路径
	var k8sYamlAbsolutePaths []string                                                     // 关联k8s yaml文件的绝对路径
	for _, relativePath := range k8sYamlRelativePaths {
		k8sYamlAbsolutePaths = append(k8sYamlAbsolutePaths, common.PathJoin(gitRepoRootDirPath, relativePath))
	}
	originK8sConfig, err := config.NewK8SConfig(configFileAbsolutePath, k8sYamlAbsolutePaths, "", "")
	common.CheckError(err)
	if originK8sConfig == nil {
		common.SmartIDELog.Error("配置文件解析失败！")
		return // 解决下面的warning问题，没有实际作用
	}

	//4. 是否 配置文件 & k8s yaml 有改变
	hasChanged, err := hasChanged(workspaceInfo, *originK8sConfig) // 配置文件 或者 关联k8s yaml是否有改变
	common.CheckError(err)
	tempK8sConfig := workspaceInfo.K8sInfo.TempK8sConfig
	checkPod, err := getDevContainerPodReady(*kubernetes, *originK8sConfig) // pod 是否运行正常
	isReady := checkPod && err == nil
	if hasChanged || !isReady {
		//4.1. 尝试删除deployment、service
		if workspaceInfo.ID != "" { // id 为空的时候时，是 update；尝试先删除deployment、service
			common.SmartIDELog.Info("删除 service && deployment ")

			for _, deployment := range workspaceInfo.K8sInfo.OriginK8sYaml.Workspace.Deployments {
				command := fmt.Sprintf("delete deployment -l %v ", deployment.Name)
				err = kubernetes.ExecKubectlCommandRealtime(command, "", false)
				common.CheckError(err)
			}

			for _, service := range workspaceInfo.K8sInfo.OriginK8sYaml.Workspace.Services {
				command := fmt.Sprintf("delete service -l %v ", service.Name)
				err = kubernetes.ExecKubectlCommandRealtime(command, "", false)
				common.CheckError(err)
			}
		}

		//4.2. 保存配置文件（用于kubectl apply）
		common.SmartIDELog.Info("保存临时配置文件")
		repoName := common.GetRepoName(workspaceInfo.GitCloneRepoUrl)
		// ★★★★★ 把所有k8s kind转换为一个临时的k8s yaml文件
		tempK8sConfig = originK8sConfig.ConvertToTempK8SYaml(repoName, workspaceInfo.K8sInfo.Namespace, originK8sConfig.GetSystemUserName())
		tempK8sYamlFileRelativePath, err := tempK8sConfig.SaveK8STempYaml(gitRepoRootDirPath)
		// ★★★★★ 保存到目录（临时k8s yaml文件的绝对路径）
		tempK8sYamlAbsolutePath := common.PathJoin(gitRepoRootDirPath, tempK8sYamlFileRelativePath)
		common.CheckError(err)
		//4.3. 赋值属性
		if workspaceInfo.Name == "" {
			workspaceInfo.Name = repoName
		}
		workspaceInfo.WorkingDirectoryPath = gitRepoRootDirPath
		workspaceInfo.ConfigFileRelativePath = configFileRelativePath
		workspaceInfo.ConfigYaml = *originK8sConfig.ConvertToSmartIdeConfig()
		workspaceInfo.TempYamlFileAbsolutePath = tempK8sYamlAbsolutePath
		workspaceInfo.K8sInfo.OriginK8sYaml = *originK8sConfig
		workspaceInfo.K8sInfo.TempK8sConfig = tempK8sConfig

		//4.4. 执行kubectl 命令进行部署
		common.SmartIDELog.Info("执行kubectl 命令进行部署")
		err = kubernetes.ExecKubectlCommandRealtime(fmt.Sprintf("apply -f %v", tempK8sYamlAbsolutePath), "", false)
		common.CheckError(err)
		common.SmartIDELog.Info(i18nInstance.Start.Info_k8s_created)

		//4.5. git config + ssh config + git clone
		//e.g. kubectl exec -it pod-name -- /bin/bash -c "command(s)"
		err = execPod(workspaceInfo, kubernetes, originK8sConfig, tempK8sConfig)
		common.CheckError(err)

		//6. 抽取端口
		for _, service := range tempK8sConfig.Workspace.Services {
			for _, k8sContainerPortInfo := range service.Spec.Ports {
				var portMapInfo config.PortMapInfo
				portMapInfo.PortMapType = config.PortMapInfo_K8S_Service
				portMapInfo.ServiceName = service.Name
				portMapInfo.ContainerPort = k8sContainerPortInfo.TargetPort.IntValue()
				portMapInfo.OriginHostPort = int(k8sContainerPortInfo.Port)
				portMapInfo.CurrentHostPort = portMapInfo.OriginHostPort
				portMapInfo.OldClientPort = portMapInfo.OriginHostPort
				portMapInfo.ClientPort = portMapInfo.OriginHostPort
				for label, value := range originK8sConfig.Workspace.DevContainer.Ports {
					if value == int(k8sContainerPortInfo.Port) {
						portMapInfo.PortMapType = config.PortMapInfo_OnlyLabel
						portMapInfo.HostPortDesc = label
						break
					}
				}
				workspaceInfo.Extend.Ports = workspaceInfo.Extend.Ports.AppendOrUpdate(&portMapInfo)

			}
		}

	}

	//6. 端口转发，依然需要检查对应的端口是否占用
	common.SmartIDELog.Info("端口转发...")
	//6.2. 端口转发，并记录到extend
	_, _, err = getDevContainerPod(*kubernetes, tempK8sConfig)
	common.CheckError(err)

	function1 := func(k8sServiceName string, availableClientPort, hostOriginPort, index int) {
		if availableClientPort != hostOriginPort {
			common.SmartIDELog.InfoF(i18n.GetInstance().Common.Info_port_binding_result2, availableClientPort, hostOriginPort, hostOriginPort)
		} else {
			common.SmartIDELog.InfoF(i18n.GetInstance().Common.Info_port_binding_result, availableClientPort, hostOriginPort)
		}

		portMapInfo := workspaceInfo.Extend.Ports[index]
		portMapInfo.OldClientPort = portMapInfo.ClientPort
		portMapInfo.ClientPort = availableClientPort
		workspaceInfo.Extend.Ports[index] = portMapInfo

		forwardCommand := fmt.Sprintf("--namespace %v port-forward svc/%v %v:%v --address 0.0.0.0 ",
			workspaceInfo.K8sInfo.Namespace, k8sServiceName, availableClientPort, hostOriginPort)
		output, err := kubernetes.ExecKubectlCommandCombined(forwardCommand, "")
		common.SmartIDELog.Debug(output)
		for err != nil || strings.Contains(output, "error forwarding port") {
			if err != nil {
				common.SmartIDELog.Importance(err.Error())
			}
			output, err = kubernetes.ExecKubectlCommandCombined(forwardCommand, "")
			common.SmartIDELog.Debug(output)
		}

	}

	for index, portMapInfo := range workspaceInfo.Extend.Ports {
		unusedClientPort, err := common.CheckAndGetAvailableLocalPort(portMapInfo.ClientPort, 100)
		common.SmartIDELog.Error(err)

		go function1(portMapInfo.ServiceName, unusedClientPort, portMapInfo.CurrentHostPort, index)

	}

	//8. 保存到db
	workspaceId, err := dal.InsertOrUpdateWorkspace(workspaceInfo) //TODO 使用新的方法，保存config 和 关联的k8syaml，以及生成的yaml
	common.CheckError(err)
	common.SmartIDELog.InfoF(i18nInstance.Start.Info_workspace_saved, workspaceId)

	//9. 使用浏览器打开web ide
	common.SmartIDELog.Info(i18nInstance.Start.Info_running_openbrower)
	err = waitingAndOpenBrower(workspaceInfo, *originK8sConfig)
	common.CheckError(err)

	//99. 结束
	common.SmartIDELog.Info(i18nInstance.Start.Info_end)

}

func execPod(workspaceInfo workspace.WorkspaceInfo,
	kubernetes *kubectl.Kubernetes,
	originK8sConfig *config.SmartIdeK8SConfig, tempK8sConfig config.SmartIdeK8SConfig) error {
	// 等待启动
	common.SmartIDELog.Info("等待 deployment 启动...")
	for {
		ready, err := getDevContainerPodReady(*kubernetes, *originK8sConfig)
		if err != nil {
			common.SmartIDELog.Importance(err.Error())
		}
		if ready {
			common.SmartIDELog.Debug("deployment pending")
			break
		} else {
			time.Sleep(time.Second * 2)
		}
	}
	devContainerPod, _, err := getDevContainerPod(*kubernetes, tempK8sConfig)
	if err != nil {
		return err
	}

	//5.1. git config
	if originK8sConfig.Workspace.DevContainer.Volumes.HasGitConfig.Value() {
		common.SmartIDELog.Info("git config")
		err = kubernetes.CopyLocalGitConfigToPod(*devContainerPod)
		if err != nil {
			return err
		}
	}

	//5.2. ssh config
	if originK8sConfig.Workspace.DevContainer.Volumes.HasSshKey.Value() {
		common.SmartIDELog.Info("ssh config")
		err = kubernetes.CopyLocalSSHConfigToPod(*devContainerPod)
		if err != nil {
			return err
		}
	}

	//5.3. git clone
	common.SmartIDELog.Info("git clone")
	containerGitCloneDir := "/home/project"
	for containerName, item := range originK8sConfig.Workspace.Containers { // git clone 的目录是否设置
		if containerName == originK8sConfig.Workspace.DevContainer.ServiceName {
			for _, volume := range item.PersistentVolumes {
				if volume.DirectoryType == config.PersistentVolumeDirectoryTypeEnum_Project {
					containerGitCloneDir = volume.MountPath
					break
				}
			}
			break
		}
	}
	err = kubernetes.GitClone(*devContainerPod, workspaceInfo.GitCloneRepoUrl, containerGitCloneDir, workspaceInfo.Branch)
	if err != nil {
		return err
	}

	//5.4. 复制config文件
	common.SmartIDELog.Info("copy config file")
	//repoName := common.GetRepoName(workspaceInfo.GitCloneRepoUrl)
	configFileAbsolutePath := common.PathJoin(workspaceInfo.WorkingDirectoryPath, workspaceInfo.ConfigFileRelativePath)
	err = copyConfigToPod(*kubernetes, *devContainerPod, containerGitCloneDir, configFileAbsolutePath)
	if err != nil {
		return err
	}

	return nil
}

// 是否改变
func hasChanged(workspaceInfo workspace.WorkspaceInfo, originK8sConfig config.SmartIdeK8SConfig) (bool, error) {
	if workspaceInfo.ConfigYaml.IsNil() {
		return true, nil
	}

	// 配置文件是否相同
	currentConfigYaml, err := originK8sConfig.ConvertToConfigYaml()
	if err != nil {
		return false, err
	}
	if oldConfigYaml, _ := workspaceInfo.ConfigYaml.ToYaml(); oldConfigYaml != currentConfigYaml {
		return true, err
	}

	// 关联的k8s yaml文件是否相同
	oldK8sYaml, _ := workspaceInfo.K8sInfo.OriginK8sYaml.ConvertToK8sYaml()
	currentK8sYaml, _ := originK8sConfig.ConvertToK8sYaml()
	if oldK8sYaml != currentK8sYaml {
		return true, nil
	}

	return false, err
}

// 复制config文件到pod
func copyConfigToPod(k kubectl.Kubernetes, pod coreV1.Pod, podDestGitRepoPath string, configFilePath string) error {
	// 目录
	configFileDir := path.Dir(configFilePath)
	tempDirPath := common.PathJoin(configFileDir, ".temp")
	if !common.IsExit(tempDirPath) {
		err := os.MkdirAll(tempDirPath, os.ModePerm)
		if err != nil {
			return err
		}
	}

	// 把文件写入到临时文件中
	input, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return err
	}
	tempConfigFilePath := common.PathJoin(tempDirPath, "config-temp.yaml")
	err = ioutil.WriteFile(tempConfigFilePath, input, 0644)
	if err != nil {
		return err
	}

	// 增加.gitigorne文件
	gitignoreFile := common.PathJoin(configFileDir, ".gitignore")
	err = ioutil.WriteFile(gitignoreFile, []byte("/.temp/"), 0644)
	if err != nil {
		return err
	}

	// copy
	destDir := common.PathJoin(podDestGitRepoPath, ".ide")
	err = k.CopyToPod(pod, gitignoreFile, destDir)
	if err != nil {
		return err
	}
	err = k.CopyToPod(pod, tempDirPath, destDir)
	if err != nil {
		return err
	}

	return nil
}

// 下载配置文件 和 关联的k8s yaml文件
func downloadConfigAndLinkFiles(workspaceInfo workspace.WorkspaceInfo) (
	gitRepoRootDirPath string,
	configFileRelativePath string, linkK8sYamlRelativePaths []string, err error) {

	//3.1. 下载配置文件
	gitRepoRootDirPath, fileRelativePaths, err := downloadFilesByGit(workspaceInfo.GitCloneRepoUrl, workspaceInfo.Branch, workspaceInfo.ConfigFileRelativePath)
	if err != nil {
		return
	}
	if len(fileRelativePaths) != 1 || !strings.Contains(fileRelativePaths[0], workspaceInfo.ConfigFileRelativePath) {
		err = fmt.Errorf("配置文件 %v 不存在！", workspaceInfo.ConfigFileRelativePath)
		return
	}

	//3.2. 下载配置文件关联的yaml文件
	common.SmartIDELog.Info("下载配置文件 关联的 k8s yaml 文件")
	configFileRelativePath = fileRelativePaths[0]
	var configYaml map[string]interface{}
	configFileBytes, err := ioutil.ReadFile(common.PathJoin(gitRepoRootDirPath, configFileRelativePath))
	if err != nil {
		return
	}
	err = yaml.Unmarshal(configFileBytes, &configYaml)
	if err != nil {
		return
	}
	smartIdeConfig := config.NewConfig("", "", string(configFileBytes))
	if smartIdeConfig == nil {
		err = fmt.Errorf("配置文件 %v 内容为空！", workspaceInfo.ConfigFileRelativePath)
		return
	}
	if smartIdeConfig.Workspace.KubeDeployFiles == "" {
		return "", "", []string{}, fmt.Errorf("配置文件 %v Workspace.KubeDeployFiles 节点未配置！", workspaceInfo.ConfigFileRelativePath)
	}
	filePathExpression := path.Join(".ide", smartIdeConfig.Workspace.KubeDeployFiles) // 这里的文件路径都是相对	.ide 的，注意这里只能使用 反斜杠 的分隔符
	_, linkK8sYamlRelativePaths, err = downloadFilesByGit(workspaceInfo.GitCloneRepoUrl, workspaceInfo.Branch, filePathExpression)
	if err != nil {
		return "", "", []string{}, err
	}
	if len(linkK8sYamlRelativePaths) == 0 {
		return "", "", []string{}, fmt.Errorf("没有找到关联的 %v yaml文件！", smartIdeConfig.Workspace.KubeDeployFiles)
	}

	//configFileRelativePath := strings.Replace(configFileAbsolutePath, gitRepoRootDirPath, "", -1)
	return gitRepoRootDirPath, configFileRelativePath, linkK8sYamlRelativePaths, nil
}

// 等待webide可以访问，并打开
func waitingAndOpenBrower(workspaceInfo workspace.WorkspaceInfo, originK8sConfig config.SmartIdeK8SConfig) error {
	var ideBindingPort int

	// 获取项目文件夹路径
	projectDir := "/home/project"
	for containerName, item := range originK8sConfig.Workspace.Containers {
		if containerName == originK8sConfig.Workspace.DevContainer.ServiceName {
			for _, pv := range item.PersistentVolumes {
				if pv.DirectoryType == config.PersistentVolumeDirectoryTypeEnum_Project {
					projectDir = pv.MountPath
					break
				}
			}
			break
		}
	}

	// 获取webide的url
	var url string
	switch originK8sConfig.Workspace.DevContainer.IdeType {
	case config.IdeTypeEnum_VsCode:
		portMap, err := workspaceInfo.Extend.Ports.Find("tools-webide-vscode")
		if err != nil {
			return err
		}
		ideBindingPort = portMap.ClientPort
		url = fmt.Sprintf("http://localhost:%v/?folder=vscode-remote://localhost:%v%v",
			ideBindingPort, ideBindingPort, projectDir)
	case config.IdeTypeEnum_JbProjector:
		portMap, err := workspaceInfo.Extend.Ports.Find("tools-webide-jb")
		if err != nil {
			return err
		}
		ideBindingPort = portMap.ClientPort
		url = fmt.Sprintf(`http://localhost:%v`, ideBindingPort)
	case config.IdeTypeEnum_Opensumi:
		portMap, err := workspaceInfo.Extend.Ports.Find("tools-webide-opensumi")
		if err != nil {
			return err
		}
		ideBindingPort = portMap.ClientPort
		url = fmt.Sprintf(`http://localhost:%v/?workspaceDir=%v`, ideBindingPort, projectDir)
	}
	// 等待webide启动
	common.SmartIDELog.Info(i18nInstance.VmStart.Info_warting_for_webide + " " + url)
	isUrlReady := false
	for !isUrlReady {
		resp, err := http.Get(url)
		if (err == nil) && (resp.StatusCode == 200) {
			isUrlReady = true
			common.OpenBrowser(url)
			common.SmartIDELog.InfoF(i18nInstance.VmStart.Info_open_brower, url)
		} /* else {

		} */
	}

	return nil
}

// 检测pod是否已经ready
func getDevContainerPodReady(kubernetes kubectl.Kubernetes, smartideK8sConfig config.SmartIdeK8SConfig) (bool, error) {
	devContainerName := smartideK8sConfig.Workspace.DevContainer.ServiceName
	for _, deployment := range smartideK8sConfig.Workspace.Deployments {
		for _, container := range deployment.Spec.Template.Spec.Containers {
			if container.Name == devContainerName {
				// deployment ready
				command := fmt.Sprintf("get deployment %v -o=yaml --namespace %v", deployment.Name, kubernetes.Namespace)
				yaml, err := kubernetes.ExecKubectlCommandCombined(command, "")
				if err != nil {
					return false, err
				}
				decode := k8sScheme.Codecs.UniversalDeserializer().Decode
				obj, _, _ := decode([]byte(yaml), nil, nil)
				deployment := obj.(*appV1.Deployment)
				deploymentReady := deployment.Status.Replicas > 0 && deployment.Status.Replicas == deployment.Status.ReadyReplicas

				// pod ready
				if deploymentReady {
					common.SmartIDELog.Info(fmt.Sprintf("deployment %v started， check pod status！", deployment.Name))
					pod, _, err := getDevContainerPod(kubernetes, smartideK8sConfig)
					if err != nil {
						return false, err
					}
					if pod.Status.Phase == coreV1.PodRunning {
						return true, nil
					}
				}

				// default
				return false, nil
			}
		}
	}

	return false, nil
}

//
func getDevContainerPod(kubernetes kubectl.Kubernetes, smartideK8sConfig config.SmartIdeK8SConfig) (
	pod *coreV1.Pod, serviceName string, err error) {

	devContainerName := smartideK8sConfig.Workspace.DevContainer.ServiceName
	for _, deployment := range smartideK8sConfig.Workspace.Deployments {
		for _, container := range deployment.Spec.Template.Spec.Containers {
			if container.Name == devContainerName {

				selector := ""
				index := 0
				for key, value := range deployment.Spec.Selector.MatchLabels {
					if index == 0 {
						selector += fmt.Sprintf("%v=%v", key, value)
					}
					index++
				}

				pod, err := kubernetes.GetPod(selector, kubernetes.Namespace)
				if err != nil {
					return pod, "", err
				}

				for _, service := range smartideK8sConfig.Workspace.Services {
					for key, value := range deployment.Spec.Selector.MatchLabels {
						if _, ok := service.Spec.Selector[key]; ok && service.Spec.Selector[key] == value {
							serviceName = service.Name
						}
					}
				}

				return pod, serviceName, nil
			}
		}
	}

	return nil, "", nil
}

func checkAndInstallKubectl(kubernetes *kubectl.Kubernetes) error {

	//1. 在.ide目录下面检查kubectl文件是否存在
	//e.g. Client Version: version.Info{Major:"1", Minor:"23", GitVersion:"v1.23.5", GitCommit:"c285e781331a3785a7f436042c65c5641ce8a9e9", GitTreeState:"clean", BuildDate:"2022-03-16T15:58:47Z", GoVersion:"go1.16.8", Compiler:"gc", Platform:"linux/amd64"}
	// 1.1. 判断是否安装
	isInstallKubectl := true
	output, err := kubernetes.ExecKubectlCommandCombined("version --client", "")
	common.SmartIDELog.Debug(output)
	if !strings.Contains(output, "GitVersion:\"v1.23.0\"") {
		isInstallKubectl = false
	} else if err != nil {
		common.SmartIDELog.Importance(err.Error())
		isInstallKubectl = false
	}
	if isInstallKubectl { // 如果已经安装，将返回
		return nil
	}

	//2. 如果不存在从smartide dl中下载对应的版本
	common.SmartIDELog.Info("安装kubectl（v1.23.0）工具到 \"用户目录/.ide\" 目录...")
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	var execCommand2 *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		kubectlFilePath := strings.Join([]string{home, ".ide", "kubectl.exe"}, string(filepath.Separator)) //common.PathJoin(home, ".ide")
		command := "Invoke-WebRequest -Uri \"https://smartidedl.blob.core.chinacloudapi.cn/kubectl/v1.23.0/bin/windows/amd64/kubectl.exe\" -OutFile " + kubectlFilePath
		common.SmartIDELog.Debug(command)
		execCommand2 = exec.Command("powershell", "/c", command)
	case "darwin":
		command := `curl -OL  "https://smartidedl.blob.core.chinacloudapi.cn/kubectl/v1.23.0/bin/darwin/amd64/kubectl" \
		&& mv -f kubectl ~/.ide/kubectl \
		&& chmod +x ~/.ide/kubectl`
		common.SmartIDELog.Debug(command)
		execCommand2 = exec.Command("bash", "-c", command)
	case "linux":
		command := `curl -OL  "https://smartidedl.blob.core.chinacloudapi.cn/kubectl/v1.23.0/bin/linux/amd64/kubectl" \
		&& mv -f kubectl ~/.ide/kubectl \
		&& chmod +x ~/.ide/kubectl`
		common.SmartIDELog.Debug(command)
		execCommand2 = exec.Command("bash", "-c", command)
	}
	execCommand2.Stdout = os.Stdout
	execCommand2.Stderr = os.Stderr
	err = execCommand2.Run()
	if err != nil {
		return err
	}

	return nil
}

// 从git下载相关文件
func downloadFilesByGit(gitCloneUrl string, branch string, filePathExpression string) (
	gitRepoRootDirPath string, fileRelativePaths []string, err error) {
	// home 目录的路径
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}

	// 文件路径
	workingRootDir := common.PathJoin(home, ".ide", ".k8s") // 工作目录，repo 会clone到当前目录下
	sshPath := common.PathJoin(home, ".ssh")
	if !common.IsExit(workingRootDir) { // 目录如果不存在，就要创建
		os.MkdirAll(workingRootDir, os.ModePerm)
	}

	// 设置不进行host的检查，即clone新的git库时，不会给出提示
	err = common.FS.SkipStrictHostKeyChecking(sshPath, false)
	if err != nil {
		return
	}

	// 下载指定的文件
	//filePathExpression = common.PathJoin(".ide", filePathExpression)
	fileRelativePaths, err = common.GIT.SparseCheckout(workingRootDir, gitCloneUrl, filePathExpression, branch)
	if err != nil {
		return
	}

	// 还原.ssh config 的设置
	err = common.FS.SkipStrictHostKeyChecking(sshPath, true)
	if err != nil {
		return
	}

	gitRepoRootDirPath = common.PathJoin(workingRootDir, common.GetRepoName(gitCloneUrl)) // git repo 的根目录
	for index, _ := range fileRelativePaths {
		fileRelativePaths[index] = strings.Replace(fileRelativePaths[index], gitRepoRootDirPath, "", -1) // 把绝对路径改为相对路径
	}
	return gitRepoRootDirPath, fileRelativePaths, nil
}
