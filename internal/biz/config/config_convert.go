/*
 * @Date: 2022-03-30 23:10:52
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-04-11 10:37:47
 * @FilePath: /smartide-cli/internal/biz/config/config_convert.go
 */

package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"

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
		var smartIdeConfig SmartIdeK8SConfig
		smartIdeConfig.Orchestrator = smartideConfig.Orchestrator
		smartIdeConfig.Version = smartideConfig.Version
		smartIdeConfig.Workspace.DevContainer = smartideConfig.Workspace.DevContainer
		smartIdeConfig.Workspace.KubeDeployFiles = smartideConfig.Workspace.KubeDeployFiles
		return &smartIdeConfig
	}
	return nil
}

//
func (k8sConfig *SmartIdeK8SConfig) ConvertToConfigYaml() (string, error) {
	smartideIdeConfig := k8sConfig.ConvertToSmartIdeConfig()
	bytes, err := yaml.Marshal(smartideIdeConfig)
	return string(bytes), err
}

// 转换为临时的yaml文件
func (originK8sConfig SmartIdeK8SConfig) ConvertToTempYaml(repoName string, namespace string) SmartIdeK8SConfig {
	k8sConfig := SmartIdeK8SConfig{}
	copier.CopyWithOption(&k8sConfig, &originK8sConfig, copier.Option{IgnoreEmpty: true, DeepCopy: true})

	//1. namespace
	for i := 0; i < len(k8sConfig.Workspace.Deployments); i++ {
		k8sConfig.Workspace.Deployments[i].Namespace = namespace
	}
	for i := 0; i < len(k8sConfig.Workspace.Services); i++ {
		k8sConfig.Workspace.Services[i].Namespace = namespace
	}

	//2. pvc
	//2.1. git 库
	pvcRepo := coreV1.PersistentVolumeClaim{}
	pvcRepo.ObjectMeta.Name = fmt.Sprintf("%v-claim-repo", repoName)
	pvcRepo.Spec.AccessModes = append(pvcRepo.Spec.AccessModes, coreV1.ReadWriteOnce)
	pvcRepo.Spec.Resources.Requests = coreV1.ResourceList{
		coreV1.ResourceName(coreV1.ResourceStorage): resource.MustParse("100Mi"),
	}
	pvcRepo.Kind = "PersistentVolumeClaim"
	pvcRepo.APIVersion = "v1"
	k8sConfig.Workspace.PVCS = append(k8sConfig.Workspace.PVCS, pvcRepo)

	//2.2. git 配置
	pvcGitConfig := coreV1.PersistentVolumeClaim{}
	pvcGitConfig.ObjectMeta.Name = fmt.Sprintf("%v-claim-gitconfig", repoName)
	pvcGitConfig.Spec.AccessModes = append(pvcGitConfig.Spec.AccessModes, coreV1.ReadWriteOnce)
	pvcGitConfig.Spec.Resources.Requests = coreV1.ResourceList{
		coreV1.ResourceName(coreV1.ResourceStorage): resource.MustParse("100Mi"),
	}
	pvcGitConfig.Kind = "PersistentVolumeClaim"
	pvcGitConfig.APIVersion = "v1"
	k8sConfig.Workspace.PVCS = append(k8sConfig.Workspace.PVCS, pvcGitConfig)

	//2.3. ssh 配置
	pvcSSHConfig := coreV1.PersistentVolumeClaim{}
	pvcSSHConfig.ObjectMeta.Name = fmt.Sprintf("%v-claim-ssh", repoName)
	pvcSSHConfig.Spec.AccessModes = append(pvcSSHConfig.Spec.AccessModes, coreV1.ReadWriteOnce)
	pvcSSHConfig.Spec.Resources.Requests = coreV1.ResourceList{
		coreV1.ResourceName(coreV1.ResourceStorage): resource.MustParse("100Mi"),
	}
	pvcSSHConfig.Kind = "PersistentVolumeClaim"
	pvcSSHConfig.APIVersion = "v1"
	k8sConfig.Workspace.PVCS = append(k8sConfig.Workspace.PVCS, pvcSSHConfig)

	// 需要添加pvc
	pvcMap := map[string]string{
		pvcRepo.ObjectMeta.Name:      "/home/project",
		pvcGitConfig.ObjectMeta.Name: "/home/smartide/.git",
		pvcSSHConfig.ObjectMeta.Name: "/home/smartide/.ssh",
	}

	//3. 把pvc写入到deployment中
	for index, deployment := range k8sConfig.Workspace.Deployments {

		for pvcName := range pvcMap {
			volume := coreV1.Volume{}
			volume.Name = pvcName
			volume.PersistentVolumeClaim = &coreV1.PersistentVolumeClaimVolumeSource{}
			volume.PersistentVolumeClaim.ClaimName = pvcName
			deployment.Spec.Template.Spec.Volumes = append(deployment.Spec.Template.Spec.Volumes, volume)
		}

		for indexContainer, container := range deployment.Spec.Template.Spec.Containers {
			if container.Name == k8sConfig.Workspace.DevContainer.ServiceName {

				for pvcName, dirPath := range pvcMap {
					volumeMount := coreV1.VolumeMount{}
					volumeMount.MountPath = dirPath
					volumeMount.Name = pvcName
					container.VolumeMounts = append(container.VolumeMounts, volumeMount)
					deployment.Spec.Template.Spec.Containers[indexContainer] = container
				}

				k8sConfig.Workspace.Deployments[index] = deployment

				break
			}
		}

	}

	return k8sConfig
}

//
func (k8sConfig *SmartIdeK8SConfig) SaveK8STempYaml(dir string, repoName string) (string, error) {

	k8sYamlContent, err := k8sConfig.ConvertToK8sYaml()
	if err != nil {
		return "", err
	}
	tempConfigFilePath := path.Join(dir, fmt.Sprintf("k8s_deployment_%v_temp.yaml", repoName))
	err = ioutil.WriteFile(tempConfigFilePath, []byte(k8sYamlContent), 0777)
	if err != nil {
		return "", err
	}

	return tempConfigFilePath, nil
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
