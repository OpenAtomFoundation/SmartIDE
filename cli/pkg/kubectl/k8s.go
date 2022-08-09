/*
 * @Date: 2022-03-23 16:13:54
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-08-09 14:45:50
 * @FilePath: /cli/pkg/kubectl/k8s.go
 */

package kubectl

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/spf13/cobra"

	k8sScheme "k8s.io/client-go/kubernetes/scheme"

	coreV1 "k8s.io/api/core/v1"
)

type KubernetesUtil struct {
	KubectlFilePath string
	Context         string
	Namespace       string
	Commands        string
}

func NewK8sUtilWithFile(kubeConfigFilePath string, targetContext string, ns string) (*KubernetesUtil, error) {
	return newK8sUtil(kubeConfigFilePath, "", targetContext, ns)
}

func NewK8sUtilWithContent(kubeConfigContent string, targetContext string, ns string) (*KubernetesUtil, error) {
	return newK8sUtil("", kubeConfigContent, targetContext, ns)
}

func NewK8sUtil(kubeConfigFilePath string, targetContext string, ns string) (*KubernetesUtil, error) {
	return newK8sUtil(kubeConfigFilePath, "", targetContext, ns)
}

//
func newK8sUtil(kubeConfigFilePath string, kubeConfigContent string, targetContext string, ns string) (*KubernetesUtil, error) {
	if targetContext == "" {
		return nil, errors.New("target k8s context is nil")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	//1. kubectl
	//1.1. kubectl 工具的安装路径
	kubectlFilePath := "~/.ide/kubectl"
	switch runtime.GOOS {
	case "windows":
		kubectlFilePath = common.PathJoin(home, ".ide", "kubectl")
	}
	customFlags := ""

	//1.2. 检查并安装kubectl命令行工具
	common.SmartIDELog.Info("检测kubectl（v1.23.0）是否安装到 \"用户目录/.ide\"")
	err = checkAndInstallKubectl(kubectlFilePath)
	if err != nil {
		return nil, err
	}

	//2. kubeconfig
	absoluteKubeConfigFilePath := ""
	//2.0. valid
	if kubeConfigFilePath != "" && kubeConfigContent != "" {
		return nil, errors.New("配置文件路径 和 文件内容不能同时指定")
	}
	//2.1. 指定kubeconfig
	homeDir, _ := os.UserHomeDir()
	if kubeConfigFilePath != "" {
		if strings.Index(kubeConfigFilePath, "~") == 0 {
			absoluteKubeConfigFilePath = strings.Replace(kubeConfigFilePath, "~", homeDir, -1)
		} else {
			if !filepath.IsAbs(kubeConfigFilePath) { // 非绝对路径的时候，就认为是相对用户目录
				absoluteKubeConfigFilePath = filepath.Join(homeDir, kubeConfigFilePath)
			} else {
				absoluteKubeConfigFilePath = kubeConfigFilePath
			}
		}
		if !common.IsExist(absoluteKubeConfigFilePath) {
			return nil, fmt.Errorf("%v 不存在", absoluteKubeConfigFilePath)
		}
		customFlags += fmt.Sprintf("--kubeconfig %v ", absoluteKubeConfigFilePath)
	}

	//2.2. 更新配置文件的内容
	if kubeConfigContent != "" {
		absoluteKubeConfigFilePath = filepath.Join(homeDir, ".kube/config_smartide")

		err = common.FS.CreateOrOverWrite(absoluteKubeConfigFilePath, kubeConfigContent)
		if err != nil {
			return nil, err
		}

		customFlags += fmt.Sprintf("--kubeconfig %v ", absoluteKubeConfigFilePath)
	}

	//3. 切换到指定的context
	common.SmartIDELog.Info("check default k8s context: " + targetContext)
	currentContext, err := execKubectlCommandCombined(kubectlFilePath, customFlags+" config current-context", "")
	if err != nil {
		common.SmartIDELog.Importance(err.Error())
	}
	if targetContext != strings.TrimSpace(currentContext) { //
		customFlags += fmt.Sprintf("--context %v ", targetContext)
	}

	//3.1. check
	common.SmartIDELog.Info("k8s connection check...")
	output, err := execKubectlCommandCombined(kubectlFilePath, customFlags+" get nodes -o json", "")
	if strings.Contains(output, "Unable to connect to the server") {
		return nil, errors.New(output)
	}
	if err != nil {
		return nil, err
	}

	//4. namespace 为空时，使用一个随机生成的6位字符作为namespace
	if ns == "" {
		for {
			namespace := common.RandLowStr(6)
			output, err := execKubectlCommandCombined(kubectlFilePath, customFlags+" get namespace "+namespace, "")
			if _, isExitError := err.(*exec.ExitError); !isExitError {
				common.SmartIDELog.ImportanceWithError(err)
				continue
			}
			if strings.Contains(output, "not found") {
				ns = namespace
				break
			}
		}
	}
	customFlags += fmt.Sprintf("--namespace %v ", ns)

	return &KubernetesUtil{
		KubectlFilePath: kubectlFilePath,
		Commands:        customFlags,
		Context:         targetContext,
		Namespace:       ns,
	}, nil
}

type ExecInPodRequest struct {
	ContainerName string

	Command string

	Namespace string

	//Pod apiv1.Pod
}

/* // 拷贝本地ssh config到pod
func (k *KubernetesUtil) CreateKubeConfig(kubeConfigContent string) error {
	if kubeConfigContent == "" {
		return errors.New("kube config content is empty")
	}

	kubeConfigPath := "~/.kube/config"
	err := common.FS.CreateOrOverWrite(kubeConfigPath, kubeConfigContent)

	return err
} */

/* // 检查集群是否可以连接
func check() error {
	common.SmartIDELog.Info("k8s connection check...")
	command := "get nodes -o json"
	output, err := k.ExecKubectlCommandCombined(command, "")
	if err != nil {
		return err
	}
	if strings.Contains(output, "Unable to connect to the server") {
		return errors.New(output)
	}
	return err
}
*/
// 拷贝本地ssh config到pod
func (k *KubernetesUtil) CopyLocalSSHConfigToPod(pod coreV1.Pod, runAsUser string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	// current user dir
	podCurrentUserHomeDir, err := k.GetPodCurrentUserHomeDirection(pod, runAsUser)
	if err != nil {
		return err
	}

	// copy
	sshPath := filepath.Join(home, ".ssh/")
	if runtime.GOOS == "windows" {
		sshPath = filepath.Join(home, ".ssh\\")
	}
	err = k.CopyToPod(pod, sshPath, podCurrentUserHomeDir, runAsUser) //.ssh
	if err != nil {
		return err
	}

	// chmod
	commad := `sudo echo -e 'Host *\n	StrictHostKeyChecking no' >>  ~/.ssh/config`
	k.ExecuteCommandRealtimeInPod(pod, commad, runAsUser)

	return nil
}

// 拷贝本地git config到pod
func (k *KubernetesUtil) CopyLocalGitConfigToPod(pod coreV1.Pod, runAsUser string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	cmd := exec.Command("git", "config", "--list")
	cmd.Dir = home // 运行目录设置到home
	bytes, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	gitconfigs := string(bytes)
	gitconfigs = strings.ReplaceAll(gitconfigs, "file:", "")
	for _, str := range strings.Split(gitconfigs, "\n") {
		str = strings.TrimSpace(str)
		if str == "" {
			continue
		}
		var index = strings.Index(str, "=")
		if index < 0 {
			continue
		}
		var key = str[0:index]
		var value = str[index+1:]
		if strings.Contains(key, "user.name") || strings.Contains(key, "user.email") {
			gitConfigCmd := fmt.Sprintf(`git config --global --replace-all %v '%v'`, key, value)
			err = k.ExecuteCommandRealtimeInPod(pod, gitConfigCmd, runAsUser)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// git clone
func (k *KubernetesUtil) GitClone(pod coreV1.Pod,
	runAsUser string,
	gitCloneUrl string, containerCloneDir string, branch string) error {
	// 设置目录为空时，使用默认的
	if containerCloneDir == "" {
		return errors.New("容器内克隆目录为空！")
	}

	// 直接 git clone
	cloneCommand := fmt.Sprintf(`	    
		 [[ -d '%v' ]] && echo 'git repo existed！' || ( ([[ -d '%v' ]] && rm -rf %v) && git clone %v %v)
		 `, //sudo chown -R smartide:smartide %v
		filepath.Join(containerCloneDir, ".git"),
		containerCloneDir, containerCloneDir+"/*",
		gitCloneUrl, containerCloneDir)
	err := k.ExecuteCommandRealtimeInPod(pod, cloneCommand, runAsUser)
	if err != nil {
		return err
	}

	// 切换到指定的分支
	if branch != "" {
		command := fmt.Sprintf("cd %v && git checkout %v", containerCloneDir, branch)
		err := k.ExecuteCommandRealtimeInPod(pod, command, runAsUser)
		return err
	}

	return nil
}

// 拷贝文件到pod
func (k *KubernetesUtil) CopyToPod(pod coreV1.Pod, srcPath string, destPath string, runAsUser string) error {
	//e.g. kubectl cp /tmp/foo <some-namespace>/<some-pod>:/tmp/bar
	workingDir := ""
	commnad := fmt.Sprintf("cp %v %v/%v:%v", srcPath, k.Namespace, pod.Name, destPath)
	if runtime.GOOS == "windows" {
		baseDir := filepath.Base(srcPath)
		workingDir = strings.Replace(srcPath, baseDir, "", -1)
		commnad = fmt.Sprintf("cp %v %v/%v:%v", baseDir, k.Namespace, pod.Name, destPath)
	}
	err := k.ExecKubectlCommandRealtime(commnad, workingDir, false)
	if err != nil {
		return err
	}

	if runAsUser != "" {
		podCommand := fmt.Sprintf(`sudo chown -R %v:%v ~/.ssh
sudo chmod -R 700 ~/.ssh`, runAsUser, runAsUser)
		k.ExecuteCommandCombinedInPod(pod, podCommand, "")
	}

	return nil
}

// 拷贝文件到pod
func (k *KubernetesUtil) GetPodCurrentUserHomeDirection(pod coreV1.Pod, runAsUser string) (string, error) {
	tmp, err := k.ExecuteCommandCombinedInPod(pod, "cd ~ && pwd", runAsUser)
	if err != nil && !common.IsExitError(err) {
		return "", err
	}
	array := strings.Split(tmp, "\n")
	result := ""
	for _, msg := range array {
		if !strings.Contains(msg, "Unable to use a TTY - input is not a terminal or the right kind of file") &&
			msg != "" &&
			strings.Contains(msg, "/") {
			result += msg
		}
	}

	result = strings.TrimSpace(result)

	return result, nil
	/* if filepath.IsAbs(result) {
		 return result, nil
	 } else {
		 return "", fmt.Errorf("%v 非正确的文件路径！", result)
	 } */
}

const (
	Flags_ServerHost      = "serverhost"
	Flags_ServerToken     = "servertoken"
	Flags_ServerOwnerGuid = "serverownerguid"
)

func (k *KubernetesUtil) StartAgent(cmd *cobra.Command, pod coreV1.Pod, runAsUser string) error {
	fflags := cmd.Flags()
	host, _ := fflags.GetString(Flags_ServerHost)
	token, _ := fflags.GetString(Flags_ServerToken)
	ownerguid, _ := fflags.GetString(Flags_ServerOwnerGuid)

	commad := fmt.Sprintf("sudo chmod +x /smartide-agent && cd /;./smartide-agent --serverhost %s --servertoken %s --serverownerguid %s", host, token, ownerguid)

	err := k.ExecuteCommandRealtimeInPod(pod, commad, runAsUser)
	if err != nil {
		return err
	}

	return nil
}

type ProxyWriter struct {
	file *os.File
}

func NewProxyWriter(file *os.File) *ProxyWriter {
	return &ProxyWriter{
		file: file,
	}
}

func (w *ProxyWriter) Write(p []byte) (int, error) {
	// ... do something with bytes first
	fmt.Fprintf(w.file, "%s", string(p))
	return len(p), nil
}

//
func (k *KubernetesUtil) ExecKubectlCommandRealtime(command string, dirctory string, isLoop bool) error {
	var execCommand *exec.Cmd

	kubeCommand := fmt.Sprintf("%v %v %v", k.KubectlFilePath, k.Commands, command)
	if isLoop {
		kubeCommand = fmt.Sprintf("while true; do %v; done", kubeCommand)
	}
	common.SmartIDELog.Debug(kubeCommand)
	switch runtime.GOOS {
	case "windows":
		execCommand = exec.Command("powershell", "/c", kubeCommand)
	default:
		execCommand = exec.Command("bash", "-c", kubeCommand)
	}

	if dirctory != "" {
		execCommand.Dir = dirctory
	}

	//execCommand.Stdout = os.Stdin
	execCommand.Stdout = NewProxyWriter(os.Stdout)
	execCommand.Stderr = NewProxyWriter(os.Stderr)
	return execCommand.Run()
}

func (k *KubernetesUtil) ExecKubectlCommand(command string, dirctory string, isLoop bool) error {
	var execCommand *exec.Cmd

	kubeCommand := fmt.Sprintf("%v %v %v", k.KubectlFilePath, k.Commands, command)
	if isLoop {
		kubeCommand = fmt.Sprintf("while true; do %v; done", kubeCommand)
	}
	common.SmartIDELog.Debug(kubeCommand)
	switch runtime.GOOS {
	case "windows":
		execCommand = exec.Command("powershell", "/c", kubeCommand)
	default:
		execCommand = exec.Command("bash", "-c", kubeCommand)
	}

	if dirctory != "" {
		execCommand.Dir = dirctory
	}

	return execCommand.Run()
}

// 一次性执行kubectl命令
func (k *KubernetesUtil) ExecKubectlCommandCombined(command string, dirctory string) (string, error) {
	return execKubectlCommandCombined(k.KubectlFilePath, k.Commands+" "+command, dirctory)
}

func execKubectlCommandCombined(kubectlFilePath string, command string, workingDirctory string) (string, error) {
	var execCommand *exec.Cmd

	kubeCommand := fmt.Sprintf("%v %v", kubectlFilePath, command)

	switch runtime.GOOS {
	case "windows":
		execCommand = exec.Command("powershell", "/c", kubeCommand)
	default:
		execCommand = exec.Command("bash", "-c", kubeCommand)
	}

	if workingDirctory != "" {
		execCommand.Dir = workingDirctory
	}

	bytes, err := execCommand.CombinedOutput()
	output := string(bytes)
	common.SmartIDELog.Debug(fmt.Sprintf("%v %v >> %v", workingDirctory, kubeCommand, output))
	if strings.Contains(output, "error:") || strings.Contains(output, "fatal:") {
		return "", errors.New(output)
	}

	return string(bytes), err
}

// 根据selector获取pod
func (k *KubernetesUtil) GetPod(selector string, namespace string) (*coreV1.Pod, error) {
	command := fmt.Sprintf("get pod --selector=%v -o=yaml ", selector)
	yaml, err := k.ExecKubectlCommandCombined(command, "")
	if err != nil {
		return nil, err
	}

	decode := k8sScheme.Codecs.UniversalDeserializer().Decode
	obj, _, _ := decode([]byte(yaml), nil, nil)
	list := obj.(*coreV1.List)

	if list == nil || len(list.Items) == 0 {
		return nil, errors.New("查找不到对应的pod，请检查k8s运行环境是否正常！")
	}

	item := list.Items[0]
	bytes, _ := item.MarshalJSON()
	objPod, _, _ := decode(bytes, nil, nil)
	pod := objPod.(*coreV1.Pod)

	return pod, nil
}

// 在pod中实时执行shell命令
// example: kubectl -it exec podname -- bash/sh -c
func (k *KubernetesUtil) ExecuteCommandRealtimeInPod(pod coreV1.Pod, command string, runAsUser string) error {
	//command = "su smartide -c " + command
	if runAsUser != "" && runAsUser != "root" {
		if runtime.GOOS == "windows" {
			command = strings.ReplaceAll(command, "'", "`")
		} else {
			command = strings.ReplaceAll(command, "'", "\"\"")
		}
		command = fmt.Sprintf(`su %v -c '%v'`, runAsUser, command)
	}
	kubeCommand := fmt.Sprintf(` -it exec %v -- /bin/bash -c "%v"`, pod.Name, command)

	err := k.ExecKubectlCommandRealtime(kubeCommand, "", false)
	if err != nil {
		return err
	}

	return nil
}

// 在pod中一次性执行shell命令
func (k *KubernetesUtil) ExecuteCommandCombinedInPod(pod coreV1.Pod, command string, runAsUser string) (string, error) {
	//command = "su smartide -c " + command
	if runAsUser != "" && runAsUser != "root" {
		if runtime.GOOS == "windows" {
			command = strings.ReplaceAll(command, "'", "`")
		} else {
			command = strings.ReplaceAll(command, "'", "\"\"")
		}
		command = fmt.Sprintf(`su %v -c '%v'`, runAsUser, command)
	}
	kubeCommand := fmt.Sprintf(` -it exec %v -- /bin/bash -c "%v"`, pod.Name, command)
	output, err := k.ExecKubectlCommandCombined(kubeCommand, "")
	return output, err
}

// 在pod中一次性执行shell命令
func (k *KubernetesUtil) ExecuteCommandCombinedBackgroundInPod(pod coreV1.Pod, command string, runAsUser string) {
	//command = fmt.Sprintf("su smartide -c '%v'", command)
	if runAsUser != "" && runAsUser != "root" {
		if runtime.GOOS == "windows" {
			command = strings.ReplaceAll(command, "'", "`")
		} else {
			command = strings.ReplaceAll(command, "'", "\"\"")
		}
		command = fmt.Sprintf(`su %v -c '%v'`, runAsUser, command)
	}
	kubeCommand := fmt.Sprintf(` exec  %v -- /bin/bash -c "%v"`, pod.Name, command)
	k.ExecKubectlCommandCombined(kubeCommand, "")
}

// 检查并安装kubectl工具
func checkAndInstallKubectl(kubectlFilePath string) error {

	//1. 在.ide目录下面检查kubectl文件是否存在
	//e.g. Client Version: version.Info{Major:"1", Minor:"23", GitVersion:"v1.23.5", GitCommit:"c285e781331a3785a7f436042c65c5641ce8a9e9", GitTreeState:"clean", BuildDate:"2022-03-16T15:58:47Z", GoVersion:"go1.16.8", Compiler:"gc", Platform:"linux/amd64"}
	// 1.1. 判断是否安装
	isInstallKubectl := true
	output, err := execKubectlCommandCombined(kubectlFilePath, "version --client", "")
	common.SmartIDELog.Debug(output)
	if !strings.Contains(output, "GitVersion:\"v1.23.0\"") {
		isInstallKubectl = false
	} else if err != nil {
		common.SmartIDELog.ImportanceWithError(err)
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

type WorkspaceIngress struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Metadata   struct {
		Name        string `yaml:"name"`
		Namespace   string `yaml:"namespace"`
		Annotations struct {
			NginxIngressKubernetesIoAuthType   string `yaml:"nginx.ingress.kubernetes.io/auth-type"`
			NginxIngressKubernetesIoAuthSecret string `yaml:"nginx.ingress.kubernetes.io/auth-secret"`
			NginxIngressKubernetesIoUseRegex   string `yaml:"nginx.ingress.kubernetes.io/use-regex"`
			CertManagerIoClusterIssuer         string `yaml:"cert-manager.io/cluster-issuer"`
		} `yaml:"annotations"`
	} `yaml:"metadata"`
	Spec struct {
		IngressClassName string `yaml:"ingressClassName"`
		TLS              []struct {
			Hosts      []string `yaml:"hosts"`
			SecretName string   `yaml:"secretName"`
		} `yaml:"tls"`
		Rules []struct {
			Host string `yaml:"host"`
			HTTP struct {
				Paths []struct {
					Path     string `yaml:"path"`
					PathType string `yaml:"pathType"`
					Backend  struct {
						Service struct {
							Name string `yaml:"name"`
							Port struct {
								Number int `yaml:"number"`
							} `yaml:"port"`
						} `yaml:"service"`
					} `yaml:"backend"`
				} `yaml:"paths"`
			} `yaml:"http"`
		} `yaml:"rules"`
	} `yaml:"spec"`
}

type ConfigMap struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Metadata   struct {
		Name      string `yaml:"name"`
		Namespace string `yaml:"namespace"`
	} `yaml:"metadata"`
	Data map[string]string `yaml:"data"`
}
