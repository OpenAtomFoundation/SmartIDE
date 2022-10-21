/*
 * @Date: 2022-03-23 16:15:38
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-10-21 10:08:31
 * @FilePath: /cli/cmd/start/k8s.go
 */

package start

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/leansoftX/smartide-cli/internal/biz/config"
	"github.com/leansoftX/smartide-cli/internal/biz/workspace"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/leansoftX/smartide-cli/pkg/k8s"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	k8sScheme "k8s.io/client-go/kubernetes/scheme"

	appV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
)

// 执行k8s start
func ExecuteK8sStartCmd(cmd *cobra.Command, k8sUtil k8s.KubernetesUtil, workspaceInfo workspace.WorkspaceInfo, yamlExecuteFun func(yamlConfig config.SmartIdeConfig)) (*workspace.WorkspaceInfo, error) {
	common.SmartIDELog.Info(i18nInstance.Start.Info_k8s_init)

	if workspaceInfo.K8sInfo.Namespace == "" {
		workspaceInfo.K8sInfo.Namespace = k8sUtil.Namespace
	}
	runAsUserName := "smartide"

	if common.SmartIDELog.Ws_id != "" {
		execSSHPolicy(workspaceInfo, common.ServerHost, common.ServerToken, common.ServerUserGuid)

	}

	//3. 解析 .k8s.ide.yaml 文件（是否需要注入到deploy.yaml文件中）
	common.SmartIDELog.Info("下载配置文件 及 关联k8s yaml文件")
	gitRepoRootDirPath, configFileRelativePath, _, err := downloadConfigAndLinkFiles(workspaceInfo)
	if err != nil {
		return nil, err
	}
	//3.1. 解析配置文件 + 关联的k8s yaml
	common.SmartIDELog.Info(fmt.Sprintf("解析配置文件 %v", workspaceInfo.ConfigFileRelativePath))
	originK8sConfig, err := config.NewK8sConfig(gitRepoRootDirPath, configFileRelativePath)
	if err != nil {
		return nil, err
	}
	if originK8sConfig == nil {
		return nil, errors.New("配置文件解析失败！") // 解决下面的warning问题，没有实际作用
	}

	//4. 是否 配置文件 & k8s yaml 有改变
	hasChanged, err := hasChanged(workspaceInfo, *originK8sConfig) // 配置文件 或者 关联k8s yaml是否有改变
	if err != nil {
		return nil, err
	}
	tempK8sConfig := workspaceInfo.K8sInfo.TempK8sConfig
	checkPodReady, err := getDevContainerPodReady(k8sUtil, *originK8sConfig) // pod 是否运行正常
	isReady := checkPodReady && err == nil
	if hasChanged || !isReady {
		//4.1. 尝试删除deployment、service
		if workspaceInfo.ID != "" { // id 为空的时候时，是 update；尝试先删除deployment、service
			common.SmartIDELog.Info("删除 service && deployment ")

			for _, deployment := range workspaceInfo.K8sInfo.OriginK8sYaml.Workspace.Deployments {
				command := fmt.Sprintf("delete deployment -l %v ", deployment.Name)
				err = k8sUtil.ExecKubectlCommandRealtime(command, "", false)
				if err != nil {
					return nil, err
				}
			}

			for _, service := range workspaceInfo.K8sInfo.OriginK8sYaml.Workspace.Services {
				command := fmt.Sprintf("delete service -l %v ", service.Name)
				err = k8sUtil.ExecKubectlCommandRealtime(command, "", false)
				if err != nil {
					return nil, err
				}
			}
		}

		//4.2. 保存配置文件（用于kubectl apply）
		common.SmartIDELog.Info("保存临时配置文件")
		repoName := common.GetRepoName(workspaceInfo.GitCloneRepoUrl)
		// ★★★★★ 把所有k8s kind转换为一个临时的k8s yaml文件
		labels := getK8sLabels(cmd, workspaceInfo) // 获取k8s模式下的label
		tempK8sConfig = originK8sConfig.ConvertToTempK8SYaml(repoName, workspaceInfo.K8sInfo.Namespace, originK8sConfig.GetSystemUserName(), labels)
		tempK8sYamlFileRelativePath, err := tempK8sConfig.SaveK8STempYaml(gitRepoRootDirPath)
		// ★★★★★ 保存到目录（临时k8s yaml文件的绝对路径）
		tempK8sYamlAbsolutePath := filepath.Join(gitRepoRootDirPath, tempK8sYamlFileRelativePath)
		if err != nil {
			return nil, err
		}

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
		err = k8sUtil.ExecKubectlCommandRealtime(fmt.Sprintf("apply -f %v", tempK8sYamlAbsolutePath), "", false)
		if err != nil {
			return nil, err
		}
		common.SmartIDELog.Info(i18nInstance.Start.Info_k8s_created)

		//4.5. 执行相关操作， git config + ssh config + git clone + agent
		//e.g. kubectl exec -it pod-name -- /bin/bash -c "command(s)"
		err = execPod(cmd, workspaceInfo, &k8sUtil, originK8sConfig, tempK8sConfig, runAsUserName) // ★★★★★
		if err != nil {
			return nil, err
		}

		//4.6. 抽取端口
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
				if strings.Contains(portMapInfo.HostPortDesc, "tools-webide") { // 如果是webide，就设置项目文件夹路径
					portMapInfo.RefDirecotry = originK8sConfig.GetProjectDirctory()
				}
				workspaceInfo.Extend.Ports = workspaceInfo.Extend.Ports.AppendOrUpdate(&portMapInfo)

			}
		}

	}

	//6. 端口转发，依然需要检查对应的端口是否占用
	common.SmartIDELog.Info("端口转发...")
	//6.2. 端口转发，并记录到extend
	_, _, err = GetDevContainerPod(k8sUtil, tempK8sConfig)
	if err != nil {
		return nil, err
	}

	function1 := func(k8sServiceName string, availableClientPort, hostOriginPort, index int) {
		forwardCommand := fmt.Sprintf("port-forward svc/%v %v:%v --address 0.0.0.0 ",
			k8sServiceName, availableClientPort, hostOriginPort)
		output, err := k8sUtil.ExecKubectlCommandCombined(forwardCommand, "")
		common.SmartIDELog.Debug(output)
		for err != nil || strings.Contains(output, "error forwarding port") {
			if err != nil {
				common.SmartIDELog.ImportanceWithError(err)
			}
			output, err = k8sUtil.ExecKubectlCommandCombined(forwardCommand, "")
			common.SmartIDELog.Debug(output)
		}

	}

	for index, portMapInfo := range workspaceInfo.Extend.Ports {
		unusedClientPort, err := common.CheckAndGetAvailableLocalPort(portMapInfo.ClientPort, 100)
		common.SmartIDELog.Error(err)
		updatePortInfo(unusedClientPort, portMapInfo.CurrentHostPort, &workspaceInfo, index)
		go function1(portMapInfo.ServiceName, unusedClientPort, portMapInfo.CurrentHostPort, index)

	}

	if workspaceInfo.CliRunningEnv == workspace.CliRunningEnvEnum_Client {
		//8. 保存到db
		saveDataAndReloadWorkSpaceId(&workspaceInfo)
		common.SmartIDELog.InfoF(i18nInstance.Start.Info_workspace_saved, workspaceInfo.ID)

		//9. 使用浏览器打开web ide
		if originK8sConfig.Workspace.DevContainer.IdeType != config.IdeTypeEnum_SDKOnly {
			common.SmartIDELog.Info(i18nInstance.Start.Info_running_openbrower)
			err = waitingAndOpenBrower(workspaceInfo, *originK8sConfig)
			if err != nil {
				return nil, err
			}
		}
	}

	// ssh config update
	workspaceInfo.UpdateSSHConfig()

	return &workspaceInfo, nil
}

// func setSSHPWD(tempK8sConfig config.SmartIdeK8SConfig, host string, token string, ownerGuid string) (err error) {

// 	for _, d := range tempK8sConfig.Workspace.Deployments {
// 		for _, c := range d.Spec.Template.Spec.Containers {
// 			i := func() (i int) {
// 				for i, ev := range c.Env {
// 					if ev.Name == model.CONST_ENV_NAME_LoalUserPassword {
// 						return i
// 					}
// 				}
// 				return -1
// 			}()
// 			if i >= 0 {
// 				p := ""
// 				if p, err = common.GetBasicPassword(host, token, ownerGuid); p != "" {
// 					c.Env[i].Value = p
// 				}
// 			} else {
// 				p := ""
// 				if p, err = common.GetBasicPassword(host, token, ownerGuid); p != "" {
// 					c.Env = append(c.Env, coreV1.EnvVar{Name: model.CONST_ENV_NAME_LoalUserPassword, Value: p})
// 				}
// 			}

// 		}
// 	}
// 	return nil
// }

// k8s模式将ssh-key 写入本地
func execSSHPolicy(workspaceInfo workspace.WorkspaceInfo, host string, token string, ownerGuid string) {
	if ws, err := common.GetWSPolicies("2", host, token, ownerGuid); err == nil {
		if len(ws) > 0 {
			detail := common.Detail{}
			if ws[len(ws)-1].Detail != "" {
				if err := json.Unmarshal([]byte(ws[len(ws)-1].Detail), &detail); err == nil {
					idRsa := detail.IdRsa
					idRsaPub := detail.IdRsaPub
					if homeDir, err := os.UserHomeDir(); err == nil {
						path := filepath.Join(homeDir, ".ssh")
						if _, err := common.PathExists(path, 0700); err == nil {

							commad := `echo -e 'Host *\n	StrictHostKeyChecking no' >>  ~/.ssh/config && sudo chmod 700 ~/.ssh/config`
							common.RunCmd(commad, true)
							commad = fmt.Sprintf(`echo -e '%v' >>  ~/.ssh/id_rsa && sudo chmod 600 ~/.ssh/id_rsa`, idRsa)
							common.RunCmd(commad, true)
							commad = fmt.Sprintf(`echo -e '%v' >>  ~/.ssh/id_rsa.pub && sudo chmod 644 ~/.ssh/id_rsa.pub`, idRsaPub)
							common.RunCmd(commad, true)

						}
					}
				}

			}

		}
	}
}

// 更新Port信息
func updatePortInfo(availableClientPort int, hostOriginPort int, workspaceInfo *workspace.WorkspaceInfo, index int) {
	if availableClientPort != hostOriginPort {
		common.SmartIDELog.InfoF("[端口转发] localhost:%v( %v 被占用) -> Service: %v  ", availableClientPort, hostOriginPort, hostOriginPort)
	} else {
		common.SmartIDELog.InfoF("[端口转发] localhost:%v -> Service: %v  ", availableClientPort, hostOriginPort)
	}

	portMapInfo := workspaceInfo.Extend.Ports[index]
	portMapInfo.OldClientPort = portMapInfo.ClientPort
	portMapInfo.ClientPort = availableClientPort
	workspaceInfo.Extend.Ports[index] = portMapInfo
}

func execPod(cmd *cobra.Command, workspaceInfo workspace.WorkspaceInfo,
	kubernetes *k8s.KubernetesUtil,
	originK8sConfig *config.SmartIdeK8SConfig, tempK8sConfig config.SmartIdeK8SConfig,
	runAsUserName string) error {
	// 等待启动
	common.SmartIDELog.Info("等待 deployment 启动...")
	for {
		ready, err := getDevContainerPodReady(*kubernetes, *originK8sConfig)
		if err != nil {
			common.SmartIDELog.ImportanceWithError(err)
		}
		if ready {
			common.SmartIDELog.Debug("deployment pending")
			break
		} else {
			time.Sleep(time.Second * 2)
		}
	}
	devContainerPod, _, err := GetDevContainerPod(*kubernetes, tempK8sConfig)
	if err != nil {
		return err
	}

	//5.5. agent
	common.SmartIDELog.Info("install agent")
	if workspaceInfo.CacheEnv == workspace.CacheEnvEnum_Server {
		err := kubernetes.CopyToPod(*devContainerPod, tempK8sConfig.Workspace.DevContainer.ServiceName, common.PathJoin("/usr/local/bin", "smartide-agent"), common.PathJoin("/", "smartide-agent"), runAsUserName)
		if err != nil {
			return err
		}
		err = kubernetes.CopyLocalSSHConfigToPod(*devContainerPod, tempK8sConfig.Workspace.DevContainer.ServiceName, runAsUserName)
		if err != nil {
			return err
		}
		err = FeeadbackContainerId(cmd, workspaceInfo, devContainerPod.Name)
		if err != nil {
			return err
		}
		kubernetes.StartAgent(cmd, *devContainerPod, tempK8sConfig.Workspace.DevContainer.ServiceName, runAsUserName, workspaceInfo.ServerWorkSpace)

	}

	time.Sleep(time.Second * 10)
	//5.1. git config
	// 会通过agent生成
	if workspaceInfo.CliRunningEnv == workspace.CliRunningEnvEnum_Client {
		if originK8sConfig.Workspace.DevContainer.Volumes.HasGitConfig.Value() {
			common.SmartIDELog.Info("git config")
			err = kubernetes.CopyLocalGitConfigToPod(*devContainerPod, tempK8sConfig.Workspace.DevContainer.ServiceName, runAsUserName)
			if err != nil {
				return err
			}
		}
	}

	//5.2. ssh config
	// 本地模式下才需要拷贝ssh 公私钥文件，如果是server模式下，会通过agent下载公私钥文件
	if workspaceInfo.CliRunningEnv == workspace.CliRunningEnvEnum_Client {
		if originK8sConfig.Workspace.DevContainer.Volumes.HasSshKey.Value() {
			common.SmartIDELog.Info("ssh config")
			err = kubernetes.CopyLocalSSHConfigToPod(*devContainerPod, tempK8sConfig.Workspace.DevContainer.ServiceName, runAsUserName)
			if err != nil {
				return err
			}

			//本地k8s 模式将公钥写入容器内knowhost
			command := `sudo echo -e | cat ~/.ssh/id_rsa.pub   >>  /home/smartide/.ssh/authorized_keys && sudo chmod 644 /home/smartide/.ssh/authorized_keys`
			err = kubernetes.ExecuteCommandRealtimeInPod(*devContainerPod, tempK8sConfig.Workspace.DevContainer.ServiceName, command, runAsUserName)
			if err != nil {
				return err
			}
		}
	}

	//5.3. 缓存git 用户名、密码
	if workspaceInfo.GitRepoAuthType == workspace.GitRepoAuthType_Basic {
		common.SmartIDELog.Info("container cache git username and password...")
		command := fmt.Sprintf(` git config --global user.name '%v' && git config --global user.password '%v' && git config --global credential.helper store`,
			workspaceInfo.GitUserName, workspaceInfo.GitPassword)
		kubernetes.ExecKubectlCommandRealtime(command, "", false)
	}

	//5.4. git clone
	common.SmartIDELog.Info("git clone")
	containerGitCloneDir := originK8sConfig.GetProjectDirctory()
	actualGitRepoUrl := workspaceInfo.GitCloneRepoUrl
	if workspaceInfo.GitRepoAuthType == workspace.GitRepoAuthType_Basic {
		actualGitRepoUrl, err =
			common.AddUsernamePassword4ActualGitRpoUrl(actualGitRepoUrl, workspaceInfo.GitUserName, workspaceInfo.GitPassword)
		if err != nil {
			common.SmartIDELog.Warning(err.Error())
		}
	}
	err = kubernetes.GitClone(*devContainerPod, tempK8sConfig.Workspace.DevContainer.ServiceName, runAsUserName, actualGitRepoUrl, containerGitCloneDir, workspaceInfo.Branch)
	if err != nil {
		return err
	}

	//5.5. 复制config文件
	common.SmartIDELog.Info("copy config file")
	configFileAbsolutePath := common.PathJoin(workspaceInfo.WorkingDirectoryPath, workspaceInfo.ConfigFileRelativePath)
	err = copyConfigToPod(*kubernetes, *devContainerPod, tempK8sConfig.Workspace.DevContainer.ServiceName, containerGitCloneDir, configFileAbsolutePath, runAsUserName)
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
func copyConfigToPod(k k8s.KubernetesUtil, pod coreV1.Pod, containerName string, podDestGitRepoPath string, configFilePath string, runAsUserName string) error {
	// 目录
	configFileDir := path.Dir(configFilePath)
	tempDirPath := common.PathJoin(configFileDir, ".temp")
	if !common.IsExist(tempDirPath) {
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
	err = k.CopyToPod(pod, containerName, gitignoreFile, destDir, runAsUserName)
	if err != nil {
		return err
	}
	err = k.CopyToPod(pod, containerName, tempDirPath, destDir, runAsUserName)
	if err != nil {
		return err
	}

	return nil
}

// 复制config文件到pod
func copyAgentToPod(k k8s.KubernetesUtil, pod coreV1.Pod, containerName string, podDestGitRepoPath string, agentFilePath string, runAsUserName string) error {
	// 目录
	configFileDir := path.Dir(agentFilePath)
	tempDirPath := common.PathJoin(configFileDir, ".temp")
	if !common.IsExist(tempDirPath) {
		err := os.MkdirAll(tempDirPath, os.ModePerm)
		if err != nil {
			return err
		}
	}

	// 把文件写入到临时文件中
	input, err := ioutil.ReadFile(agentFilePath)
	if err != nil {
		return err
	}
	tempAgentFilePath := common.PathJoin(tempDirPath, "smartide-agent")
	err = ioutil.WriteFile(tempAgentFilePath, input, 0775)
	if err != nil {
		return err
	}

	// copy
	destDir := common.PathJoin("/", "smartide-agent")

	err = k.CopyToPod(pod, containerName, tempDirPath, destDir, runAsUserName)
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
	if len(fileRelativePaths) != 1 || !strings.Contains(fileRelativePaths[0], filepath.Join(workspaceInfo.ConfigFileRelativePath)) {
		err = fmt.Errorf("配置文件 %v 不存在！", workspaceInfo.ConfigFileRelativePath)
		return
	}

	//3.2. 下载配置文件关联的yaml文件
	common.SmartIDELog.Info("下载配置文件 关联的 k8s yaml 文件")
	configFileRelativePath = fileRelativePaths[0]
	var configYaml map[string]interface{}
	configFileBytes, err := ioutil.ReadFile(filepath.Join(gitRepoRootDirPath, configFileRelativePath))
	if err != nil {
		return
	}
	err = yaml.Unmarshal(configFileBytes, &configYaml)
	if err != nil {
		return
	}
	filePathExpression := configYaml["workspace"].(map[interface{}]interface{})["kube-deploy-files"].(string)
	if filePathExpression == "" {
		return "", "", []string{}, fmt.Errorf("配置文件 %v Workspace.kube-deploy-files 节点未配置！", workspaceInfo.ConfigFileRelativePath)
	}
	filePathExpression = filepath.Join(".ide", filePathExpression)

	//
	_, linkK8sYamlRelativePaths, err = downloadFilesByGit(workspaceInfo.GitCloneRepoUrl, workspaceInfo.Branch, filePathExpression)
	if err != nil {
		return "", "", []string{}, err
	}
	if len(linkK8sYamlRelativePaths) == 0 {
		return "", "", []string{}, fmt.Errorf("没有找到 %v 匹配的yaml文件！", filePathExpression)
	}

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
func getDevContainerPodReady(kubernetes k8s.KubernetesUtil, smartideK8sConfig config.SmartIdeK8SConfig) (bool, error) {
	devContainerName := smartideK8sConfig.Workspace.DevContainer.ServiceName

	if len(smartideK8sConfig.Workspace.Deployments) == 0 {
		pod, _, err := getDevContainerPod_PodDefinition(kubernetes, smartideK8sConfig)

		if err != nil {
			return false, err
		}

		isReady := pod.Status.Phase == coreV1.PodRunning
		return isReady, nil
	}

	for _, deployment := range smartideK8sConfig.Workspace.Deployments {
		for _, container := range deployment.Spec.Template.Spec.Containers {
			if container.Name == devContainerName {
				// deployment ready
				command := fmt.Sprintf("get deployment %v -o=yaml ", deployment.Name)
				yaml, err := kubernetes.ExecKubectlCommandCombined(command, "")
				if err != nil {
					return false, err
				}
				decode := k8sScheme.Codecs.UniversalDeserializer().Decode
				obj, _, _ := decode([]byte(yaml), nil, nil)
				if obj == nil {
					return false, nil
				}
				deployment := obj.(*appV1.Deployment)
				deploymentReady := deployment.Status.Replicas > 0 && deployment.Status.Replicas == deployment.Status.ReadyReplicas

				// pod ready
				if deploymentReady {
					common.SmartIDELog.Info(fmt.Sprintf("deployment %v started， check pod status！", deployment.Name))
					pod, _, err := GetDevContainerPod(kubernetes, smartideK8sConfig)
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

func getDevContainerPod_PodDefinition(kubernetes k8s.KubernetesUtil, smartideK8sConfig config.SmartIdeK8SConfig) (
	podInstance *coreV1.Pod, serviceName string, err error) {
	devContainerName := smartideK8sConfig.Workspace.DevContainer.ServiceName
	isContain := false
	for i := 0; i < len(smartideK8sConfig.Workspace.Others); i++ {
		other := smartideK8sConfig.Workspace.Others[i]

		re := reflect.ValueOf(other)
		kindName := ""
		if re.Kind() == reflect.Ptr {
			re = re.Elem()
		}
		kindName = fmt.Sprint(re.FieldByName("Kind"))
		if kindName == "Pod" {
			var tmpPod *coreV1.Pod
			switch other.(type) {
			case coreV1.Pod:
				tmp := other.(coreV1.Pod)
				tmpPod = &tmp
			default:
				tmpPod = other.(*coreV1.Pod)
			}
			for _, container := range tmpPod.Spec.Containers {
				if container.Name == devContainerName {
					podInstance, err = kubernetes.GetPodInstanceByName(tmpPod.ObjectMeta.Name)
					if err != nil {
						return nil, "", err
					}
					isContain = true
					break
				}

			}
		}

		if isContain {
			break
		}

	}

	for _, service := range smartideK8sConfig.Workspace.Services {
		for key, value := range podInstance.ObjectMeta.Labels {
			if _, ok := service.Spec.Selector[key]; ok && service.Spec.Selector[key] == value {
				serviceName = service.Name
				break
			}
		}
	}
	return
}

func getDevContainerPod_DeploymentDefinition(kubernetes k8s.KubernetesUtil, smartideK8sConfig config.SmartIdeK8SConfig) (
	podInstance *coreV1.Pod, serviceName string, err error) {
	devContainerName := smartideK8sConfig.Workspace.DevContainer.ServiceName
	isContain := false
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

				podInstance, err = kubernetes.GetPodInstanceBySelector(selector)
				if err != nil {
					return podInstance, "", err
				}

				for _, service := range smartideK8sConfig.Workspace.Services {
					for key, value := range deployment.Spec.Selector.MatchLabels {
						if _, ok := service.Spec.Selector[key]; ok && service.Spec.Selector[key] == value {
							serviceName = service.Name
							break
						}
					}
				}
				isContain = true
				break

			}
		}

		if isContain {
			break
		}
	}
	return
}

func GetDevContainerPod(kubernetes k8s.KubernetesUtil, smartideK8sConfig config.SmartIdeK8SConfig) (
	podInstance *coreV1.Pod, serviceName string, err error) {
	if len(smartideK8sConfig.Workspace.Deployments) == 0 {
		return getDevContainerPod_PodDefinition(kubernetes, smartideK8sConfig)
	} else {
		return getDevContainerPod_DeploymentDefinition(kubernetes, smartideK8sConfig)
	}

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
	workingRootDir := filepath.Join(home, ".ide", ".k8s") // 工作目录，repo 会clone到当前目录下
	return common.GIT.DownloadFilesByGit(workingRootDir, gitCloneUrl, branch, filePathExpression)
}

const (
	Flags_ServerHost      = "serverhost"
	Flags_ServerToken     = "servertoken"
	Flags_ServerOwnerGuid = "serverownerguid"
)

func FeeadbackContainerId(cmd *cobra.Command, workspaceInfo workspace.WorkspaceInfo, containerId string) error {
	fflags := cmd.Flags()
	host, _ := fflags.GetString(Flags_ServerHost)
	token, _ := fflags.GetString(Flags_ServerToken)
	url := fmt.Sprint(host, "/api/smartide/workspace/update")
	params := make(map[string]interface{})
	params["ID"] = workspaceInfo.ServerWorkSpace.ID
	params["ContainerId"] = containerId
	headers := map[string]string{
		"x-token": token,
	}

	httpClient := common.CreateHttpClientEnableRetry()
	_, err := httpClient.Put(url, params, headers)

	return err
}
