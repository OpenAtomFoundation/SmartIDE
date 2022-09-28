/*
 * @Author: jason chen
 * @Date: 2021-11-08
 * @Description: sqlite data access layer
 */

package workspace

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/leansoftX/smartide-cli/internal/biz/config"
	"github.com/leansoftX/smartide-cli/internal/model"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/leansoftX/smartide-cli/pkg/docker/compose"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/yaml.v2"
)

// 工作模式
type WorkingModeEnum string

const (
	WorkingMode_Remote WorkingModeEnum = "remote"
	WorkingMode_Local  WorkingModeEnum = "local"
	WorkingMode_K8s    WorkingModeEnum = "k8s"
	//WorkingMode_Server WorkingModeEnum = "server"
)

type CliRunningEvnEnum string

const (
	// mode=server会调用我们的server的各种api，mode=pipeline不会这样做，而是通过通用的方法，比如传参和回调 的方式和其他系统集成
	CliRunningEvnEnum_Pipeline CliRunningEvnEnum = "pipeline"
	CliRunningEvnEnum_Server   CliRunningEvnEnum = "server"
	CliRunningEnvEnum_Client   CliRunningEvnEnum = "client"
)

type CacheEnvEnum string

const (
	CacheEnvEnum_Server CacheEnvEnum = "server"
	CacheEnvEnum_Local  CacheEnvEnum = "local"
)

// 远程连接的类型
type RemoteAuthType string

const (
	RemoteAuthType_SSH      RemoteAuthType = "ssh"
	RemoteAuthType_Password RemoteAuthType = "password"
)

// git库的连接方式
type GitRepoAuthType string

const (
	GitRepoAuthType_SSH   GitRepoAuthType = "ssh"
	GitRepoAuthType_HTTPS GitRepoAuthType = "https"
	GitRepoAuthType_HTTP  GitRepoAuthType = "http"
)

type WorkspaceType string

const (
	WorkspaceType_Server GitRepoAuthType = "server"
	WorkspaceType_Local  GitRepoAuthType = "local"
)

// addon
type Addon struct {
	IsEnable bool
	Type     string
}

type WorkspaceInfo struct {
	ID string
	// 文件夹名称，project
	Name string
	// addon
	Addon Addon
	// 即repo clone到本地时的文件夹路径
	WorkingDirectoryPath string
	// 配置文件相对路径，相对于 WorkingDirectoryPath
	ConfigFileRelativePath string
	// 临时文件（docker-compose 或者 k8s yaml）生成后的保存路径
	TempYamlFileAbsolutePath string
	// 模式，local 本地、remote 远程、k8s
	Mode WorkingModeEnum
	// CLI运行环境
	CliRunningEnv CliRunningEvnEnum
	// 缓存环境
	CacheEnv CacheEnvEnum
	// host信息
	Remote RemoteInfo
	// git 库的克隆地址
	GitCloneRepoUrl string
	// WebIDE中文件所在根目录名称 （git库名称 或者 当前目录名）
	projectDirctoryName string
	// git 库的认证方式
	GitRepoAuthType GitRepoAuthType

	// 指定的分支
	Branch string

	// 配置文件
	ConfigYaml config.SmartIdeConfig

	//
	K8sInfo K8sInfo

	// 临时的docker-compose文件
	TempDockerCompose compose.DockerComposeYml

	// 链接的docker-compose文件
	//LinkDockerCompose compose.DockerComposeYml

	// 扩展信息
	Extend WorkspaceExtend

	// 创建时间
	CreatedTime time.Time

	// 关联的服务端workspace
	ServerWorkSpace *model.ServerWorkspace

	ResourceID int
}

// 获取工作区状态值对应的label
var _workspaceStatusMap map[int]string

func getWorkspaceStatusDescFromServer(workspaceStatus model.WorkspaceStatusEnum) (string, error) {

	if len(_workspaceStatusMap) == 0 { // 如果缓存中没有，就从服务器去取
		auth, err := GetCurrentUser()
		if err != nil {
			return "", err
		}

		url, _ := common.UrlJoin(auth.LoginUrl, "/api/sysDictionary/findSysDictionary")
		headers := map[string]string{
			"Content-Type": "application/json",
		}
		if auth.Token != nil {
			headers["x-token"] = auth.Token.(string)
		}
		httpClient := common.CreateHttpClientEnableRetry()
		response, err := httpClient.Get(url.String(),
			map[string]string{"type": "smartide_workspace_status"}, headers) //
		//	response, err := common.Get(url.String(), map[string]string{"type": "smartide_workspace_status"}, headers)
		if err != nil {
			return "", err
		}

		workspaceStatusDictionaryResponse := &model.WorkspaceStatusDictionaryResponse{}
		err = json.Unmarshal([]byte(response), workspaceStatusDictionaryResponse)
		if err != nil {
			return "", err
		}

		for _, item := range workspaceStatusDictionaryResponse.Data.ResysDictionary.SysDictionaryDetails {
			if item.Value == int(workspaceStatus) {
				return item.Label, nil
			}
		}
	}

	if _, ok := _workspaceStatusMap[int(workspaceStatus)]; ok {
		return _workspaceStatusMap[int(workspaceStatus)], nil
	}

	return "", nil
}

func CreateWorkspaceInfoFromServer(serverWorkSpace model.ServerWorkspace) (WorkspaceInfo, error) {
	projectName := serverWorkSpace.Name
	if projectName == "" {
		projectName = getRepoName(serverWorkSpace.GitRepoUrl)
	}

	// 基本信息
	workspaceInfo := WorkspaceInfo{
		ID:                     serverWorkSpace.NO,
		Name:                   serverWorkSpace.Name, //+ fmt.Sprintf(" (%v)", label),
		ConfigFileRelativePath: serverWorkSpace.ConfigFilePath,
		GitCloneRepoUrl:        serverWorkSpace.GitRepoUrl,
		Branch:                 serverWorkSpace.Branch,
		Mode:                   WorkingMode_Remote,
		CacheEnv:               CacheEnvEnum_Server,
		CreatedTime:            serverWorkSpace.CreatedAt,
		WorkingDirectoryPath:   common.PathJoin("~", model.CONST_REMOTE_REPO_ROOT, projectName),
		ResourceID:             serverWorkSpace.ResourceID,
	}

	// 关联的资源信息
	switch serverWorkSpace.Resource.Type {
	case model.ReourceTypeEnum_Remote:
		workspaceInfo.Mode = WorkingMode_Remote
		workspaceInfo.Remote = RemoteInfo{
			ID:       int(serverWorkSpace.Resource.ID),
			Addr:     serverWorkSpace.Resource.IP,
			UserName: serverWorkSpace.Resource.UserName,
			Password: serverWorkSpace.Resource.Password,
			SSHPort:  serverWorkSpace.Resource.Port,
		}
		workspaceInfo.TempYamlFileAbsolutePath = workspaceInfo.GetTempDockerComposeFilePath()
	case model.ReourceTypeEnum_K8S:
		workspaceInfo.Mode = WorkingMode_K8s
		workspaceInfo.K8sInfo = K8sInfo{
			KubeConfigContent: serverWorkSpace.Resource.KubeConfigContent,
			Context:           serverWorkSpace.Resource.KubeContext,
			Namespace:         serverWorkSpace.KubeNamespace,

			IngressBaseDnsName:   serverWorkSpace.Resource.KubeBaseDNS,
			IngressName:          serverWorkSpace.Resource.Name,
			IngressAuthType:      serverWorkSpace.KubeIngressAuthenticationType,
			IngressLoginUserName: serverWorkSpace.KubeIngressLoginUserName,
			IngressLoginPassword: serverWorkSpace.KubeIngressLoginPassword,
		}

	}

	switch serverWorkSpace.Resource.AuthenticationType {
	case model.AuthenticationTypeEnum_Password:
		workspaceInfo.Remote.AuthType = RemoteAuthType_Password
	case model.AuthenticationTypeEnum_SSH:
		workspaceInfo.Remote.AuthType = RemoteAuthType_SSH
		/* 	case model.AuthenticationTypeEnum_KubeConfig:
		workspaceInfo.Remote.AuthType = RemoteAuthType_Password */
	}

	if workspaceInfo.ConfigFileRelativePath == "" {
		workspaceInfo.ConfigFileRelativePath = model.CONST_Default_ConfigRelativeFilePath

	}

	workspaceInfo.ServerWorkSpace = &serverWorkSpace

	if serverWorkSpace.Extend != "" {
		err := json.Unmarshal([]byte(serverWorkSpace.Extend), &workspaceInfo.Extend)
		if err != nil {
			return WorkspaceInfo{}, err
		}
	}
	if serverWorkSpace.ConfigFileContent != "" {
		tmpConfig, _, err := config.NewComposeConfigFromContent(serverWorkSpace.ConfigFileContent, "")
		if err != nil {
			return WorkspaceInfo{}, err
		}
		workspaceInfo.ConfigYaml = *tmpConfig
	}

	// 配置文件
	if workspaceInfo.Mode == WorkingMode_Remote {
		if serverWorkSpace.LinkDockerCompose != "" {
			err := yaml.Unmarshal([]byte(serverWorkSpace.LinkDockerCompose), &workspaceInfo.ConfigYaml.Workspace.LinkCompose)
			if err != nil {
				return WorkspaceInfo{}, err
			}
		}
		if serverWorkSpace.TempDockerComposeContent != "" {
			err := yaml.Unmarshal([]byte(serverWorkSpace.TempDockerComposeContent), &workspaceInfo.TempDockerCompose)
			if err != nil {
				return WorkspaceInfo{}, err
			}
		}
	} else if workspaceInfo.Mode == WorkingMode_K8s {
		if serverWorkSpace.LinkDockerCompose != "" {
			originK8sYaml, err := config.NewK8sConfigFromContent(serverWorkSpace.ConfigFileContent, serverWorkSpace.LinkDockerCompose)
			if err != nil {
				return WorkspaceInfo{}, err
			}
			workspaceInfo.K8sInfo.OriginK8sYaml = *originK8sYaml
		}
		if serverWorkSpace.TempDockerComposeContent != "" {
			tempK8sYaml, err := config.NewK8sConfigFromContent(serverWorkSpace.ConfigFileContent, serverWorkSpace.TempDockerComposeContent)
			if err != nil {
				return WorkspaceInfo{}, err
			}
			workspaceInfo.K8sInfo.TempK8sConfig = *tempK8sYaml
			// workspaceInfo.K8sInfo.Namespace = (*tempK8sYaml).Workspace.Services[0].Namespace

			if serverWorkSpace.KubeNamespace != "" {
				workspaceInfo.K8sInfo.Namespace = serverWorkSpace.KubeNamespace
			} else if len((*tempK8sYaml).Workspace.Services) > 0 {
				workspaceInfo.K8sInfo.Namespace = (*tempK8sYaml).Workspace.Services[0].Namespace
			} else {
				return WorkspaceInfo{}, errors.New("namespace is nil!")
			}
		}
	} else {
		return WorkspaceInfo{}, errors.New("所选模式不支持！")
	}

	workspaceInfo.ServerWorkSpace = &serverWorkSpace

	return workspaceInfo, nil
}

// 工作区数据为空
func (w WorkspaceInfo) IsNil() bool {

	return w.ID == "" || w.Name == "" ||
		w.WorkingDirectoryPath == "" ||
		w.ConfigFileRelativePath == "" ||
		w.Mode == "" || w.CliRunningEnv == "" || w.CacheEnv == "" // || w.ProjectName == "" len(w.Extend.Ports) == 0 ||
}

// 工作区数据不为空
func (w WorkspaceInfo) IsNotNil() bool {
	return !w.IsNil()
}

// 验证
func (w WorkspaceInfo) Valid() error {
	/* if w.GetProjectDirctoryName() == "" {
		return errors.New("[Workspace] 项目名不能为空")
	} */

	if w.Mode == "" {
		return errors.New(i18nInstance.Main.Err_workspace_mode_none)

	}

	if w.ConfigFileRelativePath == "" {
		return errors.New(i18nInstance.Main.Err_workspace_config_filepath_none)

	}

	if w.WorkingDirectoryPath == "" {
		return errors.New(i18nInstance.Main.Err_workspace_workingdir_none)

	}

	/* if w.GitCloneRepoUrl != "" {
		if !common.CheckGitRemoteUrl(w.GitCloneRepoUrl) {
			msg := fmt.Sprintf(i18nInstance.Main.Err_workspace_giturl_valid, w.GitCloneRepoUrl)
			return errors.New(msg)

		}
	} */

	return nil
}

func (w WorkspaceInfo) GetProjectDirctoryName() string {
	if w.projectDirctoryName == "" {
		if w.Mode == WorkingMode_Remote { // 远程模式
			/* if w.GitCloneRepoUrl == "" { // 当前模式下，不可能git库为空
				common.SmartIDELog.Error(i18nInstance.Common.Err_sshremote_param_repourl_none)
			} */

			if strings.TrimSpace(w.GitCloneRepoUrl) == "" {
				w.projectDirctoryName = w.Name
			} else {
				w.projectDirctoryName = getRepoName(w.GitCloneRepoUrl)
			}

		} else if w.CliRunningEnv == CliRunningEvnEnum_Server {
			if w.GitCloneRepoUrl == "" {
				w.projectDirctoryName = w.Name
			} else {
				w.projectDirctoryName = common.GetRepoName(w.GitCloneRepoUrl)
			}

		} else { // 本地模式
			//
			if w.GitCloneRepoUrl == "" && w.WorkingDirectoryPath == "" {
				common.SmartIDELog.Error(i18nInstance.Main.Err_workspace_property_urlandworkingdir_none)
			}

			if w.GitCloneRepoUrl == "" { // 从工作目录中获取
				fileInfo, err := os.Stat(w.WorkingDirectoryPath)
				common.CheckError(err)
				w.projectDirctoryName = fileInfo.Name()
			} else { // 从git url中获取
				w.projectDirctoryName = getRepoName(w.GitCloneRepoUrl)
			}
		}

	}

	return w.projectDirctoryName
}

// 从 volumes 中获取容器工作目录
func (c *WorkspaceInfo) GetContainerWorkingPathWithVolumes() string {
	projectPath := ""

	service := c.TempDockerCompose.Services[c.ConfigYaml.Workspace.DevContainer.ServiceName]
	for _, volume := range service.Volumes {
		if strings.Contains(volume, ":/home/project") {
			tmp := strings.ReplaceAll(volume, "\\'", "")
			index := strings.Index(tmp, ":/home/project")
			projectPath = tmp[index+1:]
			break
		}
	}

	return projectPath
}

// get repo name
func getRepoName(repoUrl string) string {

	index := strings.LastIndex(repoUrl, "/")
	return strings.Replace(repoUrl[index+1:], ".git", "", -1)
}

func getLocalGitRepoUrl() (gitRemmoteUrl, pathName string) {
	// current directory
	pwd, err := os.Getwd()
	common.CheckError(err)
	fileInfo, err := os.Stat(pwd)
	common.CheckError(err)
	pathName = fileInfo.Name()

	// git remote url
	gitRepo, err := git.PlainOpen(pwd)
	//common.CheckError(err)
	if err == nil {
		gitRemote, err := gitRepo.Remote("origin")
		if err == nil {
			//common.CheckError(err)
			gitRemmoteUrl = gitRemote.Config().URLs[0]
		}
	}
	return gitRemmoteUrl, pathName
}

// 改变配置文件
func (w *WorkspaceInfo) IsChangeConfig(currentConfigContent, linkDockerComposeContent string) (hasChanged bool) {
	// 参数检查
	if currentConfigContent == "" {
		msg := fmt.Sprintf(i18nInstance.Common.Warn_param_is_null, "configContent")
		common.SmartIDELog.Error(msg)
	}

	// 如果临时compose文件为空，那么肯定是改变
	if w.TempDockerCompose.IsNil() {
		return true
	}

	// 改变
	hasChanged = false // 默认为false
	ogriginConfigYamlContent, err := w.ConfigYaml.ToYaml()
	common.CheckError(err)
	if strings.ReplaceAll(currentConfigContent, " ", "") != strings.ReplaceAll(ogriginConfigYamlContent, " ", "") {
		var configYaml config.SmartIdeConfig
		err := yaml.Unmarshal([]byte(currentConfigContent), &configYaml)
		w.ConfigYaml = configYaml
		common.CheckError(err)
		hasChanged = true

	}
	originLinkComposeYamlContent, err := w.ConfigYaml.Workspace.LinkCompose.ToYaml()
	common.CheckError(err)
	if strings.ReplaceAll(linkDockerComposeContent, " ", "") != strings.ReplaceAll(originLinkComposeYamlContent, " ", "") {
		var linkDockerCompose compose.DockerComposeYml
		err := yaml.Unmarshal([]byte(linkDockerComposeContent), &linkDockerCompose)
		w.ConfigYaml.Workspace.LinkCompose = &linkDockerCompose
		common.CheckError(err)
		hasChanged = true

	}

	return hasChanged
}

// 把结构化对象转换为string
func (instance *WorkspaceExtend) ToJson() string {

	d, err := json.Marshal(&instance)
	common.CheckError(err)

	return string(d)
}

func (instance *WorkspaceExtend) IsNotNil() bool {
	return !instance.IsNil()
}

func (instance *WorkspaceExtend) IsNil() bool {
	return instance == nil || len(instance.Ports) <= 0
}

type ExtendPorts []config.PortMapInfo

// 工作区扩展字段
type WorkspaceExtend struct {
	// 端口映射情况
	Ports ExtendPorts `json:"Ports"`
}

// 在扩展的端口列表中查找
func (portMaps ExtendPorts) Find(portLabel string) (*config.PortMapInfo, error) {
	for _, info := range portMaps {
		if strings.EqualFold(info.HostPortDesc, portLabel) {
			return &info, nil
		}
	}
	return nil, fmt.Errorf(fmt.Sprintf("没有查找到 %v 对应的端口信息", portLabel))
}

func (portMaps ExtendPorts) IsExit(portMapInfo *config.PortMapInfo) bool {
	if portMapInfo == nil {
		return false
	}

	isContain := false
	for _, originPortMapInfo := range portMaps {
		if portMapInfo.HostPortDesc != "" {
			if strings.EqualFold(originPortMapInfo.HostPortDesc, portMapInfo.HostPortDesc) {
				isContain = true
				break
			}
		} else {
			if originPortMapInfo.ServiceName == portMapInfo.ServiceName && originPortMapInfo.ContainerPort == portMapInfo.ContainerPort {
				isContain = true
				break
			}
		}

	}

	return isContain
}

func (portMaps ExtendPorts) AppendOrUpdate(portMapInfo *config.PortMapInfo) ExtendPorts {
	if portMapInfo == nil {
		panic("obj is nil")
	}

	isContain := false
	for index, originPortMapInfo := range portMaps {
		if portMapInfo.HostPortDesc != "" {
			if strings.EqualFold(originPortMapInfo.HostPortDesc, portMapInfo.HostPortDesc) {
				isContain = true
				// portMaps[index] = *portMapInfo
				break
			}
		} else {
			if originPortMapInfo.ServiceName == portMapInfo.ServiceName && originPortMapInfo.ContainerPort == portMapInfo.ContainerPort {
				isContain = true
				// portMaps[index] = *portMapInfo
				break
			}
		}

		if isContain {
			tmp := portMaps[index]
			tmp.CurrentHostPort = portMapInfo.CurrentHostPort
			tmp.OriginHostPort = portMapInfo.OriginHostPort
			tmp.ContainerPort = portMapInfo.ContainerPort
			portMaps[index] = tmp
		}

	}

	if !isContain {
		portMaps = append(portMaps, *portMapInfo)
	}

	return portMaps
}

// 远程主机信息
type RemoteInfo struct {
	ID int
	// dns 或者 ip
	Addr        string
	UserName    string
	AuthType    RemoteAuthType
	Password    string
	SSHPort     int
	CreatedTime time.Time
}

type K8sInfo struct {
	ID          int
	CreatedTime time.Time

	// cluster + auth
	Context string
	// 对应集合的名称
	ClusterName string

	Namespace      string
	DeploymentName string
	PVCName        string

	IngressAuthType      model.KubeIngressAuthenticationTypeEnum
	IngressLoginUserName string
	IngressLoginPassword string

	// server绑定的域名，比如 xxx.com
	IngressBaseDnsName string
	// 别名，与 “IngressBaseDnsName” 组合在一起，比如 custom_name.xxx.com
	IngressName string

	// kubeconfig 文件路径，工作区缓存在本地的时候有效
	KubeConfigFilePath string
	// kubeconfig 的文件内容
	KubeConfigContent string

	// 原始的k8s yaml
	OriginK8sYaml config.SmartIdeK8SConfig
	// 增加注入后的 k8s yaml文件
	TempK8sConfig config.SmartIdeK8SConfig
}

func (r RemoteInfo) IsNil() bool {
	return r.ID <= 0 || r.Addr == "" || r.UserName == "" || r.AuthType == ""
}

func (w RemoteInfo) IsNotNil() bool {
	return !w.IsNil()
}
