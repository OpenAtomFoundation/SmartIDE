/*
 * @Author: kenan
 * @Date: 2021-12-23 15:08:31
 * @LastEditors: kenan
 * @LastEditTime: 2021-12-30 22:16:28
 * @Description: file content
 */
package start

import (
	"bufio"
	"context"
	"net/http"
	"os/exec"
	"os/signal"
	"syscall"

	"fmt"
	"os"
	"path/filepath"
	"strings"
	_ "unsafe"

	"github.com/leansoftX/smartide-cli/internal/biz/config"
	"github.com/leansoftX/smartide-cli/internal/biz/workspace"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/leansoftX/smartide-cli/pkg/kubectl"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/homedir"
	//
	// Uncomment to load all auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth"
	//
	// Or uncomment to load specific auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/openstack"
)

func ExecuteK8sStartCmd(workspaceInfo workspace.WorkspaceInfo, yamlExecuteFun func(yamlConfig config.SmartIdeConfig)) {
	{

		kubeConfig, err := kubectl.InitKubeConfig(false)
		common.CheckError(err)
		clientset, err := kubectl.NewClientSet(kubeConfig)
		common.CheckError(err)
		lastIndex := strings.LastIndex(workspaceInfo.GitCloneRepoUrl, "/")
		repoName := strings.Replace(workspaceInfo.GitCloneRepoUrl[lastIndex+1:], ".git", "", -1)
		namespace := "default"

		// git init boathouse-calculator && cd boathouse-calculator     //新建仓库并进入文件夹
		// git config core.sparsecheckout true //设置允许克隆子目录
		// echo '.ide/.ide.yaml' >> .git/info/sparse-checkout //设置要克隆的仓库的子目录路径(更改tt*为你的文件目录)   //空格别漏
		// git remote add origin git@github.com:idcf-boat-house/boathouse-calculator.git  //这里换成你要克隆的项目和库
		// git pull origin master

		repoDir := ""
		sshPath := ""
		if home := homedir.HomeDir(); home != "" {
			repoDir = filepath.Join(home, ".ide", repoName)
			ideDir := filepath.Join(home, ".ide")
			sshPath = filepath.Join(home, ".ssh/")
			workspaceInfo.WorkingDirectoryPath = repoDir
			if !common.IsExit(ideDir) {
				os.MkdirAll(ideDir, os.ModePerm)
			}
			if common.IsExit(repoDir) {
				os.RemoveAll(repoDir)
			}
			cmd := exec.Command("git", "init", repoName)
			cmd.Dir = ideDir
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err := cmd.Run()
			common.CheckError(err)
			cmd = exec.Command("git", "config", "core.sparsecheckout", "true")
			cmd.Dir = repoDir
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err = cmd.Run()
			common.CheckError(err)
			f, err := os.Create(filepath.Join(repoDir, ".git/info/sparse-checkout"))
			if err != nil {
				common.CheckError(err)
			}
			f.Close()
			common.AppendToFile(filepath.Join(repoDir, ".git/info/sparse-checkout"), ".ide/.ide.yaml")
			cmd = exec.Command("git", "remote", "add", "origin", workspaceInfo.GitCloneRepoUrl)
			cmd.Dir = repoDir
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err = cmd.Run()
			common.CheckError(err)
			cmd = exec.Command("git", "fetch")
			cmd.Dir = repoDir
			out, err := cmd.Output()
			common.SmartIDELog.Debug(string(out))
			common.CheckError(err)
			if workspaceInfo.Branch != "" {
				cmd = exec.Command("git", "checkout", workspaceInfo.Branch)
				cmd.Dir = repoDir
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				out, err := cmd.Output()
				common.SmartIDELog.Debug(string(out))
				common.CheckError(err)
			} else {

				cmd = exec.Command("git", "branch", "-a")
				cmd.Dir = repoDir
				cmd.Stderr = os.Stderr
				out, cmdErr := cmd.Output()
				common.CheckError(cmdErr)
				branches := strings.Split(string(out), "\n")
				//isContainMaster := false
				for _, branch := range branches {
					if strings.Contains(branch, "main") || strings.Contains(branch, "master") {
						workspaceInfo.Branch = strings.ReplaceAll(strings.Trim(branch, " "), "remotes/origin/", "")
						cmd = exec.Command("git", "checkout", workspaceInfo.Branch)
						cmd.Dir = repoDir
						cmd.Stdout = os.Stdout
						cmd.Stderr = os.Stderr
						err = cmd.Run()
						common.CheckError(err)
					}

				}
			}

		}

		//get .ide.yaml
		//1.3. 初始化配置文件对象

		currentConfig := config.NewConfig(workspaceInfo.WorkingDirectoryPath, workspaceInfo.ConfigFilePath, "")
		deploymentsClient := clientset.AppsV1().Deployments(namespace)
		serviceName := currentConfig.Workspace.DevContainer.ServiceName
		//workspaceInfo.ConfigYaml.Workspace.DevContainer.ServiceName
		deployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name: serviceName,
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: int32Ptr(1),
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app": serviceName,
					},
				},
				Template: apiv1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"app": serviceName,
						},
					},
					Spec: apiv1.PodSpec{
						Containers: []apiv1.Container{
							{
								Name:  serviceName,
								Image: "registry.cn-hangzhou.aliyuncs.com/smartide/smartide-base:latest",
								Ports: []apiv1.ContainerPort{
									{
										Name:          "http",
										Protocol:      apiv1.ProtocolTCP,
										ContainerPort: 3000,
									},
								},
							},
						},
					},
				},
			},
		}
		// Create Deployment
		var result *appsv1.Deployment
		var pod apiv1.Pod
		if _, err = deploymentsClient.Get(context.TODO(), serviceName, metav1.GetOptions{}); err != nil {
			if !errors.IsNotFound(err) {
				common.CheckError(err)
				return
			}
			//如果不存在则创建deployment
			fmt.Println("Creating deployment...")
			if result, err = deploymentsClient.Create(context.TODO(), deployment, metav1.CreateOptions{}); err != nil {
				common.CheckError(err)
				return
			}

			fmt.Printf("Created deployment %q.\n", result.GetObjectMeta().GetName())
			err, pod = getPod(serviceName, clientset, namespace)
			common.CheckError(err)
			err = copySshDir(err, kubeConfig, clientset, pod, namespace, sshPath, serviceName)
			common.CheckError(err)
			gitcmd := fmt.Sprint("su smartide -c ' cd /home/project && ", "git clone ", workspaceInfo.GitCloneRepoUrl, " && git checkout ", workspaceInfo.Branch, "'")
			kubectl.ExecInPod(kubeConfig, clientset, namespace, pod.Name, gitcmd, serviceName)

		} else {
			//如果存在则更新deployement
			if result, err = deploymentsClient.Update(context.TODO(), deployment, metav1.UpdateOptions{}); err != nil {
				common.CheckError(err)
				common.CheckError(err)
				return
			}
			fmt.Printf("Updated deployment %q.\n", result.GetObjectMeta().GetName())
			err, pod = getPod(serviceName, clientset, namespace)

		}

		// stopCh control the port forwarding lifecycle. When it gets closed the
		// port forward will terminate
		stopCh := make(chan struct{}, 1)
		// readyCh communicate when the port forward is ready to get traffic
		readyCh := make(chan struct{})
		// stream is used to tell the port forwarder where to place its output or
		// where to expect input if needed. For the port forwarding we just need
		// the output eventually
		stream := genericclioptions.IOStreams{
			In:     os.Stdin,
			Out:    os.Stdout,
			ErrOut: os.Stderr,
		}

		// managing termination signal from the terminal. As you can see the stopCh
		// gets closed to gracefully handle its termination.
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-sigs
			fmt.Println("Bye...")
			close(stopCh)
		}()

		go func() {
			// PortForward the pod specified from its port 9090 to the local port
			// 8080
			err := kubectl.PortForwardAPod(kubectl.PortForwardAPodRequest{
				RestConfig: kubeConfig,
				Pod: apiv1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      pod.Name,
						Namespace: namespace,
					},
				},
				LocalPort: currentConfig.Workspace.DevContainer.Ports["webide"],
				PodPort:   3000,
				Streams:   stream,
				StopCh:    stopCh,
				ReadyCh:   readyCh,
			})
			common.CheckError(err)
		}()

		select {
		case <-readyCh:
			break
		}

		println("Port forwarding is ready to get traffic. have fun!")

		var url string
		switch strings.ToLower(currentConfig.Workspace.DevContainer.IdeType) {
		case "vscode":
			url = fmt.Sprintf("http://localhost:%v/?folder=vscode-remote://localhost:%v/home/project/%v",
				6800, 6800, workspaceInfo.GetProjectDirctoryName())
		case "jb-projector":
			url = fmt.Sprintf(`http://localhost:%v`, 6800)
		default:
			url = fmt.Sprintf(`http://localhost:%v`, 6800)
		}
		// common.SmartIDELog.Info(i18nInstance.Start.Info_open_in_brower, url)
		go func(checkUrl string) {
			isUrlReady := false
			for !isUrlReady {
				resp, err := http.Get(checkUrl)
				if (err == nil) && (resp.StatusCode == 200) {
					isUrlReady = true
					common.OpenBrowser(checkUrl)
				}
			}
		}(url)
		prompt()
	}
}

func getPod(serviceName string, clientset *kubernetes.Clientset, namespace string) (error, apiv1.Pod) {
	podReq, _ := labels.NewRequirement("app", selection.Equals, []string{serviceName})
	selector := labels.NewSelector()
	selector = selector.Add(*podReq)
	pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: selector.String(),
	})
	for len(pods.Items) <= 0 {
		getPod(serviceName, clientset, namespace)
	}
	pod := pods.Items[len(pods.Items)-1]
	return err, pod
}

func copySshDir(err error, kubeConfig *rest.Config, clientset *kubernetes.Clientset, pod apiv1.Pod, namespace string, sshPath string, serviceName string) error {
	err = kubectl.CopyToPod(kubeConfig, clientset, pod.Name, namespace, "/home/smartide/.ssh", sshPath)
	common.CheckError(err)

	_, _, erro := kubectl.ExecInPod(kubeConfig, clientset, namespace, pod.Name, "sudo mv -f /home/.ssh  /home/smartide/ && sudo chown -R smartide:smartide /home/smartide/.ssh ", serviceName)
	return erro
}

func prompt() {
	fmt.Printf("-> Press Return key to continue.")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		break
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	fmt.Println()
}

func int32Ptr(i int32) *int32 { return &i }
