/*
 * @Date: 2022-03-30 23:10:52
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-11-04 23:22:19
 * @FilePath: /cli/internal/biz/config/config_convert.go
 */

package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/jinzhu/copier"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/leansoftX/smartide-cli/pkg/k8s"
	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
	k8sYaml "sigs.k8s.io/yaml"

	appsV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	networkingV1 "k8s.io/api/networking/v1"
	//metaV1 "k8s.io/apimachinery/pkg/meta/v1"
)

// 转换为 SmartIdeConfig 类型
func (k8sConfig *SmartIdeK8SConfig) ConvertToSmartIdeConfig() *SmartIdeConfig {

	if k8sConfig != nil {
		var smartIdeConfig SmartIdeConfig
		smartIdeConfig.Orchestrator = k8sConfig.Orchestrator
		smartIdeConfig.Version = k8sConfig.Version
		smartIdeConfig.Workspace.DevContainer = k8sConfig.Workspace.DevContainer
		smartIdeConfig.Workspace.KubeDeployFileExpression = k8sConfig.Workspace.KubeDeployFileExpression
		return &smartIdeConfig
	}
	return nil
}

// 转换为 SmartIdeConfig 类型
func (smartideConfig *SmartIdeConfig) ConvertToSmartIdeK8SConfig() *SmartIdeK8SConfig {

	if smartideConfig != nil {
		var k8sConfig SmartIdeK8SConfig
		copier.CopyWithOption(&k8sConfig, &smartideConfig, copier.Option{IgnoreEmpty: true, DeepCopy: true})

		return &k8sConfig
	}
	return nil
}

func (k8sConfig *SmartIdeK8SConfig) ConvertToConfigYaml() (string, error) {
	smartideIdeConfig := k8sConfig.ConvertToSmartIdeConfig()
	bytes, err := yaml.Marshal(smartideIdeConfig)
	return string(bytes), err
}

// TODO, 获取devcontainer的登录用户
func (originK8sConfig SmartIdeK8SConfig) GetSystemUserName() string {
	return "root"
}

// 转换为临时的yaml文件
func (originK8sConfig SmartIdeK8SConfig) ConvertToTempK8SYaml(workspaceName string, namespace string, systemUserName string,
	labels map[string]string,
	portConfigs map[string]uint, usedCpu float32, usedMemory float32) SmartIdeK8SConfig {
	//0.
	k8sConfig := SmartIdeK8SConfig{}
	copier.CopyWithOption(&k8sConfig, originK8sConfig, copier.Option{IgnoreEmpty: true, DeepCopy: true}) // 把一个对象赋值给另外一个对象

	//1. namespace
	//1.1. 创建kind
	namespaceKind := coreV1.Namespace{}
	namespaceKind.Kind = "Namespace" // 必须要赋值，否则为空
	namespaceKind.APIVersion = "v1"  // 必须要赋值，否则为空
	namespaceKind.ObjectMeta.Name = namespace
	namespaceKind = k8s.AddLabels(namespaceKind, labels).(coreV1.Namespace)
	k8sConfig.Workspace.Others = append(k8sConfig.Workspace.Others, namespaceKind)

	//1.2. 挂载到这个namespace上
	for i := 0; i < len(k8sConfig.Workspace.Deployments); i++ {
		k8sConfig.Workspace.Deployments[i].ObjectMeta.Namespace = namespace // namespace
		k8sConfig.Workspace.Deployments[i] = k8s.AddLabels(k8sConfig.Workspace.Deployments[i], labels).(appsV1.Deployment)

		// 附加的信息
		deployment := k8sConfig.Workspace.Deployments[i]
		for index, container := range deployment.Spec.Template.Spec.Containers {
			// port
			if container.Name == originK8sConfig.Workspace.DevContainer.ServiceName {
				// container
				for portLabel, port := range portConfigs {
					container.Ports = append(container.Ports, coreV1.ContainerPort{Name: portLabel, ContainerPort: int32(port)})
				}

				// 关联的 service
				for i := 0; i < len(k8sConfig.Workspace.Services); i++ {
					currentService := k8sConfig.Workspace.Services[i]
					for selectorKey, selectorValue := range currentService.Spec.Selector {
						if _, ok := deployment.Spec.Template.Labels[selectorKey]; ok &&
							deployment.Spec.Template.Labels[selectorKey] == selectorValue {
							for portLabel, port := range portConfigs {
								servicePort := coreV1.ServicePort{}
								servicePort.Port = int32(port)
								servicePort.TargetPort = intstr.FromInt(int(port))
								servicePort.Name = portLabel
								currentService.Spec.Ports = append(currentService.Spec.Ports, servicePort)
								k8sConfig.Workspace.Services[i] = currentService
							}

						}
					}

				}
			}

			// 配额
			if usedCpu > 0 {
				container.Resources.Limits["cpu"] = resource.MustParse(fmt.Sprint(usedCpu))
				container.Resources.Requests["cpu"] = resource.MustParse(fmt.Sprint(usedCpu))
			}
			if usedMemory > 0 {
				container.Resources.Limits["memory"] = resource.MustParse(fmt.Sprint(usedMemory))
				container.Resources.Requests["memory"] = resource.MustParse(fmt.Sprint(usedMemory))
			}

			deployment.Spec.Template.Spec.Containers[index] = container
		}
		k8sConfig.Workspace.Deployments[i] = deployment

	}
	for i := 0; i < len(k8sConfig.Workspace.Services); i++ {
		currentService := k8sConfig.Workspace.Services[i]
		currentService.ObjectMeta.Namespace = namespace                         // namespace
		currentService = k8s.AddLabels(currentService, labels).(coreV1.Service) // labels

		k8sConfig.Workspace.Services[i] = currentService
	}
	for i := 0; i < len(k8sConfig.Workspace.Networks); i++ {
		k8sConfig.Workspace.Networks[i].ObjectMeta.Namespace = namespace
		k8sConfig.Workspace.Networks[i] = k8s.AddLabels(k8sConfig.Workspace.Networks[i], labels).(networkingV1.NetworkPolicy)
	}
	for i := 0; i < len(k8sConfig.Workspace.Others); i++ {
		other := k8sConfig.Workspace.Others[i]

		re := reflect.ValueOf(other)
		kindName := ""
		if re.Kind() == reflect.Ptr {
			re = re.Elem()
		}
		kindName = fmt.Sprint(re.FieldByName("Kind"))
		if kindName != "Namespace" {
			tmp := reflect.New(re.Type()).Elem()
			tmp.FieldByName("ObjectMeta").FieldByName("Namespace").SetString(namespace)
			k8sConfig.Workspace.Others[i] = tmp.Interface()
		} else if kindName == "Pod" { // pod 也需要加上端口映射 和 配额
			pod := other.(coreV1.Pod)

			for index, container := range pod.Spec.Containers {
				if container.Name == originK8sConfig.Workspace.DevContainer.ServiceName {
					// container
					for portLabel, port := range portConfigs {
						container.Ports = append(container.Ports, coreV1.ContainerPort{Name: portLabel, ContainerPort: int32(port)})
					}

					// 关联的 service
					for i := 0; i < len(k8sConfig.Workspace.Services); i++ {
						currentService := k8sConfig.Workspace.Services[i]
						for selectorKey, selectorValue := range currentService.Spec.Selector {
							if _, ok := pod.Labels[selectorKey]; ok &&
								pod.Labels[selectorKey] == selectorValue {
								for portLabel, port := range portConfigs {
									servicePort := coreV1.ServicePort{}
									servicePort.Port = int32(port)
									servicePort.TargetPort = intstr.FromInt(int(port))
									servicePort.Name = portLabel
									currentService.Spec.Ports = append(currentService.Spec.Ports, servicePort)
									k8sConfig.Workspace.Services[i] = currentService
								}

							}
						}

					}

					// 配额
					if usedCpu > 0 {
						container.Resources.Limits["cpu"] = resource.MustParse(fmt.Sprint(usedCpu))
						container.Resources.Requests["cpu"] = resource.MustParse(fmt.Sprint(usedCpu))
					}
					if usedMemory > 0 {
						container.Resources.Limits["memory"] = resource.MustParse(fmt.Sprint(usedMemory))
						container.Resources.Requests["memory"] = resource.MustParse(fmt.Sprint(usedMemory))
					}

					pod.Spec.Containers[index] = container
				}

			}

			other = pod
		}

		// labels
		k8sConfig.Workspace.Others[i] = k8s.AddLabels(other, labels)

	}

	return k8sConfig

	//2. 创建 一个pvc
	/* workspaceName = strings.TrimSpace(strings.ToLower(workspaceName))
	pvc := coreV1.PersistentVolumeClaim{}
	pvcName := fmt.Sprintf("%v-pvc-claim", workspaceName)
	storageClassName := "smartide-file-storageclass" //TODO: const
	pvc.ObjectMeta.Name = pvcName
	pvc.Spec.AccessModes = append(pvc.Spec.AccessModes, coreV1.ReadWriteMany) // ReadWriteMany 可以在多个节点 和 多个pod间访问
	pvc.Spec.StorageClassName = &storageClassName
	pvc.Spec.Resources.Requests = coreV1.ResourceList{
		coreV1.ResourceName(coreV1.ResourceStorage): resource.MustParse("2Gi"), // 默认的存储大小为 2G
	}
	pvc.ObjectMeta.Namespace = namespace
	pvc.Kind = "PersistentVolumeClaim" // 必须要赋值，否则为空
	pvc.APIVersion = "v1"              // 必须要赋值，否则为空
	k8sConfig.Workspace.PVCS = append(k8sConfig.Workspace.PVCS, pvc)

	//
	boundPvcFunc := func(containerName string, containerDirPath string, storageSubPath string) {
		if storageSubPath[0:1] == "/" {
			storageSubPath = storageSubPath[1:]
		}

		for index, deployment := range k8sConfig.Workspace.Deployments {
			volumeName := pvcName + "-storage"

			isCotain := false
			for _, item := range deployment.Spec.Template.Spec.Volumes {
				if item.Name == volumeName {
					isCotain = true
					break
				}
			}
			if !isCotain {
				volume := coreV1.Volume{}
				volume.Name = volumeName
				volume.PersistentVolumeClaim = &coreV1.PersistentVolumeClaimVolumeSource{}
				volume.PersistentVolumeClaim.ClaimName = pvcName
				deployment.Spec.Template.Spec.Volumes = append(deployment.Spec.Template.Spec.Volumes, volume)
			}

			for indexContainer, container := range deployment.Spec.Template.Spec.Containers {
				if container.Name == containerName {

					volumeMount := coreV1.VolumeMount{}
					volumeMount.MountPath = containerDirPath
					volumeMount.Name = volumeName
					volumeMount.SubPath = storageSubPath
					container.VolumeMounts = append(container.VolumeMounts, volumeMount)
					deployment.Spec.Template.Spec.Containers[indexContainer] = container

					k8sConfig.Workspace.Deployments[index] = deployment

					break
				}
			}

		}
	}

	//3. 把pvc写入到deployment中
	// 直接把用户目录进行映射，其中就包含了用户的的 “.gitconfig" 文件
	if originK8sConfig.Workspace.DevContainer.Volumes.HasGitConfig.Value() {
		gitContainerDirPath := fmt.Sprintf("/%v", systemUserName) //TODO 直接把当前用户目录进行映射太暴力
		gitPvcSubDirPath := fmt.Sprintf("/home/%v", systemUserName)
		boundPvcFunc(originK8sConfig.Workspace.DevContainer.ServiceName, gitContainerDirPath, gitPvcSubDirPath)
	}
	// ssh
	if originK8sConfig.Workspace.DevContainer.Volumes.HasSshKey.Value() {
		sshContainerDirPath := fmt.Sprintf("/%v/.ssh", systemUserName)
		sshPvcSubDirPath := fmt.Sprintf("/home/%v/.ssh", systemUserName)
		boundPvcFunc(originK8sConfig.Workspace.DevContainer.ServiceName, sshContainerDirPath, sshPvcSubDirPath)
	}
	// 其他类型
	hasProjectConfig := false
	for containerName, container := range originK8sConfig.Workspace.Containers {
		for _, pv := range container.PersistentVolumes {
			switch pv.DirectoryType {
			case PersistentVolumeDirectoryTypeEnum_Project:
				if containerName == originK8sConfig.Workspace.DevContainer.ServiceName {
					hasProjectConfig = true
					projectPvcSubPath := fmt.Sprintf("/home/%v/project", systemUserName)
					boundPvcFunc(containerName, pv.MountPath, projectPvcSubPath)
				}

			case PersistentVolumeDirectoryTypeEnum_DbData:
				dbSubPath := "smartide-db"
				boundPvcFunc(containerName, pv.MountPath, dbSubPath) // 当前容器

			case PersistentVolumeDirectoryTypeEnum_Other:
				boundPvcFunc(containerName, pv.MountPath, pv.MountPath)

			case PersistentVolumeDirectoryTypeEnum_Agent:
				agentSubPath := "smartide-agent"
				boundPvcFunc(containerName, pv.MountPath, agentSubPath) // 当前容器

			}
		}
	}
	// project，没有项目路径映射的话，就用默认的
	if !hasProjectConfig {
		subProjectPath := fmt.Sprintf("/home/%v/project", systemUserName)
		boundPvcFunc(originK8sConfig.Workspace.DevContainer.ServiceName, "/home/project", subProjectPath)
	}

	return k8sConfig */
}

// 保存k8s 临时yaml文件
func (k8sConfig *SmartIdeK8SConfig) SaveK8STempYaml(gitRepoRootDirPath string) (string, error) {

	k8sYamlContent, err := k8sConfig.ConvertToK8sYaml()
	if err != nil {
		return "", err
	}

	tempConfigFileRelativePath := common.PathJoin(gitRepoRootDirPath, fmt.Sprintf("k8s_deployment_%v_temp.yaml", filepath.Base(gitRepoRootDirPath)))
	err = os.WriteFile(tempConfigFileRelativePath, []byte(k8sYamlContent), 0777)
	if err != nil {
		return "", err
	}

	tempConfigFileRelativePath = strings.Replace(tempConfigFileRelativePath, gitRepoRootDirPath, "", -1)

	return tempConfigFileRelativePath, nil
}

// 先转换为json，在转换为yaml格式
func ConvertK8sKindToString(kind interface{}) (string, error) {
	json, err := json.Marshal(kind)
	if err != nil {
		return "", err
	}
	k8sYamlContentBytes, err := k8sYaml.JSONToYAML(json)
	if err != nil {
		return "", err
	}
	result := string(k8sYamlContentBytes)
	result = strings.ReplaceAll(result, "\\\"", "\"")

	return result, err
}

func (k8sConfig *SmartIdeK8SConfig) ConvertToK8sYaml() (string, error) {
	//
	kinds := []interface{}{}
	kinds = append(kinds, k8sConfig.Workspace.Others...)
	for _, deployment := range k8sConfig.Workspace.Deployments {
		kinds = append(kinds, deployment)
	}
	for _, service := range k8sConfig.Workspace.Services {
		kinds = append(kinds, service)
	}
	for _, pvc := range k8sConfig.Workspace.PVCS {
		kinds = append(kinds, pvc)
	}
	for _, networkPolicy := range k8sConfig.Workspace.Networks {
		kinds = append(kinds, networkPolicy)
	}

	// 排序
	sortIndex := func(kind interface{}) int {
		kindName := ""
		re := reflect.ValueOf(kind)
		if re.Kind() == reflect.Ptr {
			re = re.Elem()
		}
		kindName = fmt.Sprint(re.FieldByName("Kind"))

		kindNameArray := []string{"Namespace", "PersistentVolumeClaim", "NetworkPolicy", "Deployment", "Service"}
		for index, item := range kindNameArray {
			if item == kindName {
				return index
			}
		}
		return 999
	}
	sort.Slice(kinds, func(i, j int) bool {
		return sortIndex(kinds[i]) < sortIndex(kinds[j])
	})

	//
	var k8sYamlContents []string
	for _, kind := range kinds {
		content, err := ConvertK8sKindToString(kind)
		if err != nil {
			return "", err
		}
		k8sYamlContents = append(k8sYamlContents, string(content))
	}

	return strings.Join(k8sYamlContents, "\n---\n"), nil
}
