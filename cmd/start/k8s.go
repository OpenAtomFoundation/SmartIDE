/*
 * @Author: kenan
 * @Date: 2021-12-23 15:08:31
 * @LastEditors: kenan
 * @LastEditTime: 2022-01-18 10:50:35
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
	"github.com/leansoftX/smartide-cli/internal/dal"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/leansoftX/smartide-cli/pkg/docker/compose"
	"github.com/leansoftX/smartide-cli/pkg/kubectl"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/homedir"
	"k8s.io/client-go/util/retry"
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

		common.SmartIDELog.Info(i18nInstance.Start.Info_k8s_init)
		kubeConfig, err := kubectl.InitKubeConfig(false, workspaceInfo.K8sInfo.Context)
		common.CheckError(err)
		clientset, err := kubectl.NewClientSet(kubeConfig)
		common.CheckError(err)

		lastIndex := strings.LastIndex(workspaceInfo.GitCloneRepoUrl, "/")
		repoName := strings.Replace(workspaceInfo.GitCloneRepoUrl[lastIndex+1:], ".git", "", -1)

		repoDir := ""
		sshPath := ""
		sshPath = getIDEYML(repoDir, repoName, sshPath, &workspaceInfo)
		currentConfig := config.NewConfig(workspaceInfo.WorkingDirectoryPath, workspaceInfo.ConfigFilePath, "")
		workspaceInfo.ConfigYaml = *currentConfig
		workspaceInfo.Extend = workspace.WorkspaceExtend{Ports: currentConfig.GetPortMappings()}
		workspaceInfo.Extend = workspaceInfo.GetWorkspaceExtend()
		configYamlStr, err := currentConfig.ToYaml()
		hasChanged := workspaceInfo.ChangeConfig(configYamlStr, "") // 是否改变
		// currentConfig.Workspace.DevContainer.Ports
		deploymentsClient := clientset.AppsV1().Deployments(workspaceInfo.K8sInfo.Namespace)
		serviceName := currentConfig.Workspace.DevContainer.ServiceName
		//workspaceInfo.ConfigYaml.Workspace.DevContainer.ServiceName
		common.SmartIDELog.Info(i18nInstance.Start.Info_k8s_inited)

		claimName := fmt.Sprint(serviceName, "-claim")
		workspaceInfo.K8sInfo.PVCName = claimName

		pvcClient := clientset.CoreV1().PersistentVolumeClaims(workspaceInfo.K8sInfo.Namespace)
		pvc := &apiv1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      claimName,
				Namespace: workspaceInfo.K8sInfo.Namespace,
			},
			Spec: apiv1.PersistentVolumeClaimSpec{
				AccessModes: []apiv1.PersistentVolumeAccessMode{apiv1.ReadWriteMany},
				Resources: apiv1.ResourceRequirements{
					Requests: apiv1.ResourceList{
						apiv1.ResourceName(apiv1.ResourceStorage): resource.MustParse("1Gi"),
					},
				},
			},
		}
		var resultPvc *apiv1.PersistentVolumeClaim
		if _, err = pvcClient.Get(context.TODO(), claimName, metav1.GetOptions{}); err != nil {
			if !errors.IsNotFound(err) {
				common.CheckError(err)
				return
			}
			//如果不存在则创建deployment
			// common.SmartIDELog.Info(i18nInstance.Start.Info_k8s_creating)
			if resultPvc, err = pvcClient.Create(context.TODO(), pvc, metav1.CreateOptions{}); err != nil {
				common.CheckError(err)
				return
			}
			common.SmartIDELog.Debug(fmt.Sprintf("Created PVC %q.\n", resultPvc.GetObjectMeta().GetName()))
		}
		// else {
		// 	// retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// 	// 	if resultPvc, err = pvcClient.Update(context.TODO(), pvc, metav1.UpdateOptions{}); err != nil {
		// 	// 		common.CheckError(err)
		// 	// 	}
		// 	// 	common.SmartIDELog.Debug(fmt.Sprintf("Updated PVC %q.\n", resultPvc.GetObjectMeta().GetName()))

		// 	// 	return err
		// 	// })
		// 	// common.CheckError(retryErr)
		// }

		deploymentName := fmt.Sprint(serviceName, "-deployment")
		workspaceInfo.K8sInfo.DeploymentName = deploymentName
		deployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name: deploymentName,
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
								VolumeMounts: []apiv1.VolumeMount{
									{
										Name:      claimName,
										MountPath: "/home/project",
									},
								},
							},
						},
						Volumes: []apiv1.Volume{
							{
								Name: claimName,
								VolumeSource: apiv1.VolumeSource{
									PersistentVolumeClaim: &apiv1.PersistentVolumeClaimVolumeSource{
										ClaimName: claimName,
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
		if _, err = deploymentsClient.Get(context.TODO(), deploymentName, metav1.GetOptions{}); err != nil {
			if !errors.IsNotFound(err) {
				common.CheckError(err)
				return
			}
			//如果不存在则创建deployment
			common.SmartIDELog.Info(i18nInstance.Start.Info_k8s_creating)
			if result, err = deploymentsClient.Create(context.TODO(), deployment, metav1.CreateOptions{}); err != nil {
				common.CheckError(err)
				return
			}
			common.SmartIDELog.Info(i18nInstance.Start.Info_k8s_created)
			common.SmartIDELog.Debug(fmt.Sprintf("Created deployment %q.\n", result.GetObjectMeta().GetName()))
			err, pod = getPod(serviceName, clientset, workspaceInfo.K8sInfo.Namespace)
			common.CheckError(err)
			err = copySshDir(kubeConfig, clientset, pod, workspaceInfo.K8sInfo.Namespace, sshPath, serviceName)
			common.CheckError(err)
			gitcmd := fmt.Sprint("su smartide -c ' cd /home/project && ", "git clone ", workspaceInfo.GitCloneRepoUrl, " && git checkout ", workspaceInfo.Branch, "'")
			common.SmartIDELog.Info(i18nInstance.Start.Info_git_clone)
			out, errmes, _ := kubectl.ExecInPod(kubectl.ExecInPodRequest{
				RestConfig:    kubeConfig,
				K8sCli:        clientset,
				ContainerName: serviceName,
				Command:       gitcmd,
				Namespace:     workspaceInfo.K8sInfo.Namespace,
				Pod:           pod,
			})
			if errmes != "" {
				common.SmartIDELog.Error(fmt.Sprint("git clone failed : ", errmes))
			}
			common.SmartIDELog.Info(out)
			config.GitConfig("true", false, "", nil, &compose.Service{}, common.SSHRemote{}, kubectl.ExecInPodRequest{
				RestConfig:    kubeConfig,
				K8sCli:        clientset,
				ContainerName: serviceName,
				Command:       "",
				Namespace:     workspaceInfo.K8sInfo.Namespace,
				Pod:           pod,
			})

		} else if hasChanged {

			common.SmartIDELog.Info(i18nInstance.Start.Info_k8s_updating)
			retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
				if result, err = deploymentsClient.Update(context.TODO(), deployment, metav1.UpdateOptions{}); err != nil {
					common.CheckError(err)
				}
				return err
			})
			common.CheckError(retryErr)
			//如果存在则更新deployement

			common.SmartIDELog.Info(i18nInstance.Start.Info_k8s_updated)

			common.SmartIDELog.Debug(fmt.Sprintf("Updated deployment %q.\n", result.GetObjectMeta().GetName()))
			err, pod = getPod(serviceName, clientset, workspaceInfo.K8sInfo.Namespace)
			common.CheckError(err)
			//err, pod = getPod(serviceName, clientset, namespace)

		}
		if hasChanged {
			workspaceInfo.Name = serviceName
			workspaceId, err := dal.InsertOrUpdateWorkspace(workspaceInfo)
			common.CheckError(err)
			common.SmartIDELog.InfoF(i18nInstance.Start.Info_workspace_saved, workspaceId)

		}
		// stopCh control the port forwarding lifecycle. When it gets closed the
		// port forward will terminate
		stopCh := make(chan struct{}, 1)

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
		for i, v := range workspaceInfo.Extend.Ports {
			readyCh := make(chan struct{})
			go func(v config.PortMapInfo) {
				// PortForward the pod specified from its port 9090 to the local port
				// 8080
				// readyCh communicate when the port forward is ready to get traffic
				workspaceInfo.Extend.Ports[i].OriginLocalPort, err = common.CheckAndGetAvailableLocalPort(v.OriginLocalPort, 100)
				common.CheckError(err)
				common.SmartIDELog.InfoF(i18nInstance.Start.Info_k8s_port_forward_start, v.ContainerPort)
				err := kubectl.PortForwardAPod(kubectl.PortForwardAPodRequest{
					RestConfig: kubeConfig,
					Pod: apiv1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name:      pod.Name,
							Namespace: workspaceInfo.K8sInfo.Namespace,
						},
					},
					LocalPort: workspaceInfo.Extend.Ports[i].OriginLocalPort,
					PodPort:   v.ContainerPort,
					Streams:   stream,
					StopCh:    stopCh,
					ReadyCh:   readyCh,
				})
				common.CheckError(err)
				common.SmartIDELog.Info(fmt.Sprint(v.LocalPortDesc, ": ", workspaceInfo.Extend.Ports[i].OriginLocalPort))

			}(v)
			select {
			case <-readyCh:
				break
			}

		}
		common.SmartIDELog.Info(i18nInstance.Start.Info_k8s_port_forward_end)
		common.SmartIDELog.Debug("Port forwarding is ready to get traffic. have fun!")
		common.SmartIDELog.Info(i18nInstance.VmStart.Info_warting_for_webide)
		url := getWebIDEUrl(currentConfig, workspaceInfo)
		// common.SmartIDELog.Info(i18nInstance.Start.Info_open_in_brower, url)
		go func(checkUrl string) {
			isUrlReady := false
			for !isUrlReady {
				resp, err := http.Get(checkUrl)
				if (err == nil) && (resp.StatusCode == 200) {
					isUrlReady = true
					common.OpenBrowser(checkUrl)
					common.SmartIDELog.InfoF(i18nInstance.VmStart.Info_open_brower, url)
				}
			}
		}(url)
		common.SmartIDELog.Info(i18nInstance.Start.Info_end)
		prompt()
	}
}

// git init boathouse-calculator && cd boathouse-calculator     //新建仓库并进入文件夹
// git config core.sparsecheckout true //设置允许克隆子目录
// echo '.ide/.ide.yaml' >> .git/info/sparse-checkout //设置要克隆的仓库的子目录路径(更改tt*为你的文件目录)   //空格别漏
// git remote add origin git@github.com:idcf-boat-house/boathouse-calculator.git  //这里换成你要克隆的项目和库
// git pull origin master
//isContainMaster := false
func getIDEYML(repoDir string, repoName string, sshPath string, workspaceInfo *workspace.WorkspaceInfo) string {
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

		sshConfigPath := filepath.Join(sshPath, "config")
		if !common.IsExit(sshConfigPath) {
			f, err := os.Create(sshConfigPath)
			if err != nil {
				common.CheckError(err)
			}
			f.Close()
		}
		b, _ := common.CheckFileContainsStr(sshConfigPath, "StrictHostKeyChecking no")
		if !b {
			common.AppendToFile(sshConfigPath, "StrictHostKeyChecking no")

		}
		cmd := exec.Command("git", "init", repoName)
		cmd.Dir = ideDir
		err := cmd.Run()
		common.CheckError(err)
		cmd = exec.Command("git", "config", "core.sparsecheckout", "true")
		cmd.Dir = repoDir
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
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		common.CheckError(err)
		cmd = exec.Command("git", "fetch")
		cmd.Dir = repoDir
		err = cmd.Run()
		common.CheckError(err)
		if workspaceInfo.Branch != "" {
			cmd = exec.Command("git", "checkout", workspaceInfo.Branch)
			cmd.Dir = repoDir
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

			for _, branch := range branches {
				if strings.Contains(branch, "main") || strings.Contains(branch, "master") {
					workspaceInfo.Branch = strings.ReplaceAll(strings.Trim(branch, " "), "remotes/origin/", "")
					cmd = exec.Command("git", "checkout", workspaceInfo.Branch)
					cmd.Dir = repoDir
					cmd.Stderr = os.Stderr
					err = cmd.Run()
					common.CheckError(err)
				}

			}
		}

	}
	return sshPath
}

func getWebIDEUrl(currentConfig *config.SmartIdeConfig, workspaceInfo workspace.WorkspaceInfo) string {
	webidePord := 6800
	for _, v := range workspaceInfo.Extend.Ports {
		if v.LocalPortDesc == "webide" {
			webidePord = v.OriginLocalPort
		}
	}
	var url string
	switch strings.ToLower(currentConfig.Workspace.DevContainer.IdeType) {
	case "vscode":
		url = fmt.Sprintf("http://localhost:%v/?folder=vscode-remote://localhost:%v/home/project/%v",
			webidePord, webidePord, workspaceInfo.GetProjectDirctoryName())
	case "jb-projector":
		url = fmt.Sprintf(`http://localhost:%v`, webidePord)
	default:
		url = fmt.Sprintf(`http://localhost:%v`, webidePord)
	}
	return url
}

func getPod(lableSelect string, clientset *kubernetes.Clientset, namespace string) (error, apiv1.Pod) {
	podReq, _ := labels.NewRequirement("app", selection.Equals, []string{lableSelect})
	selector := labels.NewSelector()
	selector = selector.Add(*podReq)
	pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: selector.String(),
	})
	if len(pods.Items) > 0 {
		return err, pods.Items[len(pods.Items)-1]

	} else {
		err, pod := getPod(lableSelect, clientset, namespace)
		return err, pod
	}
}

func copySshDir(kubeConfig *rest.Config, clientset *kubernetes.Clientset, pod apiv1.Pod, namespace string, sshPath string, serviceName string) error {
	ioStreams, _, _, _ := genericclioptions.NewTestIOStreams()
	// err := kubectl.CopyToPod(kubeConfig, clientset, pod.Name, namespace, "/home/smartide/.ssh", sshPath)

	// 定义RestClient，用于与k8s API server进行交互
	o := kubectl.NewCopyOptions(ioStreams)
	o.ClientConfig = kubeConfig
	o.Clientset = clientset
	o.Container = serviceName
	dest := fmt.Sprintf("%s/%s%s", pod.Namespace, pod.Name, ":/home/smartide/.ssh")
	erro := kubectl.CopyToPod(o, sshPath, dest)
	if erro != nil {
		return erro
	}
	_, _, erro = kubectl.ExecInPod(kubectl.ExecInPodRequest{
		RestConfig:    kubeConfig,
		K8sCli:        clientset,
		ContainerName: serviceName,
		Command:       "sudo chown -R smartide:smartide /home/smartide/.ssh ",
		Namespace:     namespace,
		Pod:           pod,
	})
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
