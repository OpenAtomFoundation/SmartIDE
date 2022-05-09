/*
 * @Date: 2022-03-30 23:10:52
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-04-20 15:35:32
 * @FilePath: /smartide-cli/internal/biz/config/config_convert.go
 */

package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"strings"

	"github.com/jinzhu/copier"
	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/api/resource"
	k8sYaml "sigs.k8s.io/yaml"

	coreV1 "k8s.io/api/core/v1"
)

// 转换为 SmartIdeConfig 类型
func (k8sConfig *SmartIdeK8SConfig) ConvertToSmartIdeConfig() *SmartIdeConfig {

	if k8sConfig != nil {
		var smartIdeConfig SmartIdeConfig
		smartIdeConfig.Orchestrator = k8sConfig.Orchestrator
		smartIdeConfig.Version = k8sConfig.Version
		smartIdeConfig.Workspace.DevContainer = k8sConfig.Workspace.DevContainer
		smartIdeConfig.Workspace.KubeDeployFiles = k8sConfig.Workspace.KubeDeployFiles
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

//
func (k8sConfig *SmartIdeK8SConfig) ConvertToConfigYaml() (string, error) {
	smartideIdeConfig := k8sConfig.ConvertToSmartIdeConfig()
	bytes, err := yaml.Marshal(smartideIdeConfig)
	return string(bytes), err
}

//TODO, 获取devcontainer的登录用户
func (originK8sConfig SmartIdeK8SConfig) GetSystemUserName() string {

	return "root"
}

// 转换为临时的yaml文件
func (originK8sConfig SmartIdeK8SConfig) ConvertToTempYaml(repoName string, namespace string, systemUserName string) SmartIdeK8SConfig {
	k8sConfig := SmartIdeK8SConfig{}
	copier.CopyWithOption(&k8sConfig, &originK8sConfig, copier.Option{IgnoreEmpty: true, DeepCopy: true}) // 把一个对象赋值给另外一个对象

	repoName = strings.ToLower(repoName)
	//1. namespace
	for i := 0; i < len(k8sConfig.Workspace.Deployments); i++ {
		k8sConfig.Workspace.Deployments[i].Namespace = namespace
	}
	for i := 0; i < len(k8sConfig.Workspace.Services); i++ {
		k8sConfig.Workspace.Services[i].Namespace = namespace
	}

	// 一个 pvc
	pvc := coreV1.PersistentVolumeClaim{}
	pvcName := fmt.Sprintf("%v-pvc-claim", repoName)
	storageClassName := "smartide-file-storageclass"
	pvc.ObjectMeta.Name = pvcName
	pvc.Spec.AccessModes = append(pvc.Spec.AccessModes, coreV1.ReadWriteMany) // ReadWriteMany 可以在多个节点 和 多个pod间访问
	pvc.Spec.StorageClassName = &storageClassName
	pvc.Spec.Resources.Requests = coreV1.ResourceList{
		coreV1.ResourceName(coreV1.ResourceStorage): resource.MustParse("2Gi"), // 默认的存储大小为 2G
	}
	pvc.Kind = "PersistentVolumeClaim" // 必须要赋值，否则为空
	pvc.APIVersion = "v1"              // 必须要赋值，否则为空
	k8sConfig.Workspace.PVCS = append(k8sConfig.Workspace.PVCS, pvc)

	//
	addPvcFunc := func(containerName string, containerDirPath string, storageSubPath string) {
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
	// git
	if originK8sConfig.Workspace.DevContainer.Volumes.HasGitConfig.Value() {
		gitDirPath := fmt.Sprintf("/home/%v/.git", systemUserName)
		addPvcFunc(originK8sConfig.Workspace.DevContainer.ServiceName, gitDirPath, gitDirPath)
	}
	// ssh
	if originK8sConfig.Workspace.DevContainer.Volumes.HasSshKey.Value() {
		sshDirPath := fmt.Sprintf("/home/%v/.ssh", systemUserName)
		addPvcFunc(originK8sConfig.Workspace.DevContainer.ServiceName, sshDirPath, sshDirPath)
	}
	// 其他类型
	hasProjectConfig := false
	for containerName, container := range originK8sConfig.Workspace.Containers {

		for _, pv := range container.PersistentVolumes {
			switch pv.DirectoryType {
			case PersistentVolumeDirectoryTypeEnum_Project:
				if containerName == originK8sConfig.Workspace.DevContainer.ServiceName {
					hasProjectConfig = true
					projectSubPath := fmt.Sprintf("/home/%v/project", systemUserName)
					addPvcFunc(containerName, pv.MountPath, projectSubPath)
				}

			case PersistentVolumeDirectoryTypeEnum_DbData:
				dbSubPath := "smartide-db"
				addPvcFunc(containerName, pv.MountPath, dbSubPath) // 当前容器
				//addPvcFunc(originK8sConfig.Workspace.DevContainer.ServiceName, dbSubPath, dbSubPath) // devContainer 中映射db路径

			case PersistentVolumeDirectoryTypeEnum_Other:
				addPvcFunc(containerName, pv.MountPath, pv.MountPath)

			case PersistentVolumeDirectoryTypeEnum_Agent:
				agentSubPath := "smartide-agent"
				addPvcFunc(containerName, pv.MountPath, agentSubPath) // 当前容器
				//addPvcFunc(originK8sConfig.Workspace.DevContainer.ServiceName, agentSubPath, agentSubPath) // devContainer 中映射agent路径

			}
		}
	}
	// project，没有项目路径映射的话，就用默认的
	if !hasProjectConfig {
		projectPath := fmt.Sprintf("/home/%v/project", systemUserName)
		addPvcFunc(originK8sConfig.Workspace.DevContainer.ServiceName, projectPath, projectPath)
	}

	return k8sConfig
}

// 保存k8s 临时yaml文件
func (k8sConfig *SmartIdeK8SConfig) SaveK8STempYaml(gitRepoRootDirPath string) (string, error) {

	k8sYamlContent, err := k8sConfig.ConvertToK8sYaml()
	if err != nil {
		return "", err
	}

	tempConfigFileRelativePath := path.Join(gitRepoRootDirPath, fmt.Sprintf("k8s_deployment_%v_temp.yaml", path.Base(gitRepoRootDirPath)))
	err = ioutil.WriteFile(tempConfigFileRelativePath, []byte(k8sYamlContent), 0777)
	if err != nil {
		return "", err
	}

	tempConfigFileRelativePath = strings.Replace(tempConfigFileRelativePath, gitRepoRootDirPath, "", -1)

	return tempConfigFileRelativePath, nil
}

func (k8sConfig *SmartIdeK8SConfig) ConvertToK8sYaml() (string, error) {

	// 现转换为json，在转换为yaml格式
	var func1 = func(obj interface{}) (string, error) {
		json, err := json.Marshal(obj)
		if err != nil {
			return "", err
		}
		k8sYamlContentBytes, err := k8sYaml.JSONToYAML(json)
		if err != nil {
			return "", err
		}

		return fmt.Sprintln("---") + string(k8sYamlContentBytes), nil
	}

	k8sYamlContent := ""
	for _, deployment := range k8sConfig.Workspace.Deployments {
		content, err := func1(deployment)
		if err != nil {
			return "", err
		}
		k8sYamlContent += string(content)
	}
	for _, service := range k8sConfig.Workspace.Services {
		content, err := func1(service)
		if err != nil {
			return "", err
		}
		k8sYamlContent += string(content)
	}
	for _, pvc := range k8sConfig.Workspace.PVCS {
		content, err := func1(pvc)
		if err != nil {
			return "", err
		}
		k8sYamlContent += string(content)
	}
	for _, networkPolicy := range k8sConfig.Workspace.Networks {
		content, err := func1(networkPolicy)
		if err != nil {
			return "", err
		}
		k8sYamlContent += string(content)
	}
	for _, other := range k8sConfig.Workspace.Others {
		content, err := func1(other)
		if err != nil {
			return "", err
		}
		k8sYamlContent += string(content)
	}

	return k8sYamlContent, nil
}
