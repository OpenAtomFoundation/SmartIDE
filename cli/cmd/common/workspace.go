/*
 * @Date: 2022-10-28 16:49:11
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-11-04 15:04:16
 * @FilePath: /cli/cmd/common/workspace.go
 */

package common

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/jinzhu/copier"
	"github.com/leansoftX/smartide-cli/cmd/server"
	"github.com/leansoftX/smartide-cli/internal/biz/config"
	templateModel "github.com/leansoftX/smartide-cli/internal/biz/template/model"
	"github.com/leansoftX/smartide-cli/internal/biz/workspace"
	"github.com/leansoftX/smartide-cli/internal/dal"
	"github.com/leansoftX/smartide-cli/internal/model"

	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"gopkg.in/src-d/go-git.v4"
)

var (
	flag_workspaceid = "workspaceid"
	/* 	flag_host        = "host"
	   	flag_port        = "port"
	   	flag_username    = "username"
	   	flag_password    = "password" */
	flag_filepath    = "filepath"
	flag_repourl     = "repourl"
	flag_branch      = "branch"
	flag_k8s         = "k8s"
	flag_kubeconfig  = "kubeconfig"
	flag_gitpassword = "gitpassword"
)

// 获取工作区id
func GetWorkspaceIdFromFlagsOrArgs(cmd *cobra.Command, args []string) string {
	fflags := cmd.Flags()

	// 从args 或者 flag 中获取值
	if len(args) > 0 { // 从args中加载
		tmpWorkspaceId := args[0]
		checkFlagUnnecessary(fflags, flag_workspaceid, tmpWorkspaceId)

		// 是否为数字，或者包含sw
		if common.IsNumber(tmpWorkspaceId) || strings.Index(strings.ToLower(tmpWorkspaceId), "ws") == 1 {
			return tmpWorkspaceId
		}
	}

	// 从 workspaceid 参数中加载
	if fflags.Changed(flag_workspaceid) { // 从flag中加载
		tmpWorkspaceId, err := fflags.GetString(flag_workspaceid)
		common.CheckError(err)
		tmpWorkspaceId = strings.TrimSpace(tmpWorkspaceId)

		// 是否为数字，或者包含sw
		if common.IsNumber(tmpWorkspaceId) || strings.Index(strings.ToLower(tmpWorkspaceId), "ws") == 1 {
			return tmpWorkspaceId
		}
	}

	// 从 serverworkspaceid 参数中获取
	if fflags.Changed("serverworkspaceid") { // 从flag中加载
		serverWorkspaceId, err := fflags.GetString("serverworkspaceid")
		common.CheckError(err)
		serverWorkspaceId = strings.TrimSpace(serverWorkspaceId)

		// 是否为数字，或者包含sw
		if common.IsNumber(serverWorkspaceId) || strings.Index(strings.ToLower(serverWorkspaceId), "ws") == 1 {
			return serverWorkspaceId
		}
	}

	return ""
}

// 从start command的flag、args中获取workspace
func GetWorkspaceFromCmd(cmd *cobra.Command, args []string) (workspaceInfo workspace.WorkspaceInfo, err error) {
	fflags := cmd.Flags()
	//1.
	// 从args或flags中获取workspaceid，如 : smartide start 1
	workspaceIdStr := GetWorkspaceIdFromFlagsOrArgs(cmd, args)

	// 加载workspace
	workspaceInfo = workspace.WorkspaceInfo{
		CliRunningEnv: workspace.CliRunningEnvEnum_Client,
		CacheEnv:      workspace.CacheEnvEnum_Local,
	}
	//1.1. 运行环境
	if value, _ := fflags.GetString("mode"); value != "" {
		if strings.ToLower(value) == "server" {
			workspaceInfo.CliRunningEnv = workspace.CliRunningEvnEnum_Server
		} else if strings.ToLower(value) == "pipeline" {
			workspaceInfo.CliRunningEnv = workspace.CliRunningEvnEnum_Pipeline
		}

	}
	if strings.Index(strings.ToLower(workspaceIdStr), "ws") == 1 {
		workspaceInfo.CacheEnv = workspace.CacheEnvEnum_Server
	} else if value, _ := fflags.GetString("serverworkspaceid"); value != "" {
		workspaceInfo.CacheEnv = workspace.CacheEnvEnum_Server
	}

	//1.2. 涉及到配置文件的改变
	if tmp, _ := fflags.GetString(flag_filepath); tmp != "" {
		configFilePath, err := fflags.GetString(flag_filepath)
		common.CheckError(err)

		if configFilePath != "" {
			if strings.Contains(configFilePath, "/") || strings.Contains(configFilePath, "\\") {
				workspaceInfo.ConfigFileRelativePath = configFilePath
			} else {
				workspaceInfo.ConfigFileRelativePath = filepath.Join(".ide", configFilePath)
			}

		}
	}
	if workspaceInfo.ConfigFileRelativePath == "" { // 避免配置文件的路径为空
		workspaceInfo.ConfigFileRelativePath = model.CONST_Default_ConfigRelativeFilePath
	}
	if tmp, _ := fflags.GetString(flag_branch); tmp != "" {
		branch, err := fflags.GetString(flag_branch)
		common.CheckError(err)

		if branch != "" {
			workspaceInfo.GitBranch = branch
		}
	}

	//2. 获取基本的信息
	//2.1. 存储在 server 的工作区
	if workspaceInfo.CacheEnv == workspace.CacheEnvEnum_Server {
		auth := model.Auth{}
		if workspaceInfo.CliRunningEnv == workspace.CliRunningEvnEnum_Server {
			auth.Token, _ = fflags.GetString(server.Flags_ServerToken)
			auth.LoginUrl, _ = fflags.GetString(server.Flags_ServerHost)
			auth.UserName, _ = fflags.GetString(server.Flags_ServerUsername)
		} else {
			auth, err = workspace.GetCurrentUser()
			if err != nil {
				return
			}
		}

		if auth.IsNotNil() {
			// 从 api 获取 workspace
			var workspaceInfo_ *workspace.WorkspaceInfo
			workspaceInfo_, err = workspace.GetWorkspaceFromServer(auth, workspaceIdStr, workspaceInfo.CliRunningEnv)
			if workspaceInfo_ == nil || workspaceInfo.ServerWorkSpace == nil {
				return workspace.WorkspaceInfo{}, fmt.Errorf("get workspace (%v) is null", workspaceIdStr)
			}

			// 模板信息
			if cmd.Use == "new" {
				// 和服务器上的模板 actual repo url 保持一致
				config.GlobalSmartIdeConfig.TemplateActualRepoUrl = workspaceInfo.ServerWorkSpace.TemplateGitUrl
				workspaceInfo.SelectedTemplate, _ = getSeletedTemplate(cmd, args) // 模板相关
			}
			selectedTemplate := workspaceInfo.SelectedTemplate
			workspaceInfo = *workspaceInfo_
			if workspaceInfo.SelectedTemplate == nil && selectedTemplate != nil {
				workspaceInfo.SelectedTemplate = selectedTemplate
			}

			// 在 server 上运行的时候，git 用户名密码会从cmd中传递过来
			loadGitInfo4Workspace(&workspaceInfo, cmd, args)

			// 使用是否关联 server workspace 进行判断
			if workspaceInfo.ServerWorkSpace == nil {
				errMsg := ""
				if err != nil {
					errMsg = err.Error()
				}
				err = fmt.Errorf("没有查询到 (%v) 对应的工作区数据！"+errMsg, workspaceIdStr)
			} else {
				// 避免namespace为空
				if workspaceInfo.Mode == workspace.WorkingMode_K8s {
					if workspaceInfo.K8sInfo.Namespace == "" &&
						len(workspaceInfo.K8sInfo.TempK8sConfig.Workspace.Deployments) > 0 {
						workspaceInfo.K8sInfo.Namespace = workspaceInfo.K8sInfo.TempK8sConfig.Workspace.Deployments[0].Namespace
					}
				}
			}

		} else {
			err = fmt.Errorf("请运行smartide login命令登录！")
		}
		if err != nil {
			return
		}

	} else { //2.2. 存储在本地的工作区
		// 模板信息
		if cmd.Use == "new" {
			workspaceInfo.SelectedTemplate, _ = getSeletedTemplate(cmd, args) // 模板相关
		}

		workspaceId, _ := strconv.Atoi(workspaceIdStr)
		common.CheckError(err)
		//2.2.1. 指定了从workspaceid，从sqlite中读取
		if workspaceId > 0 {
			checkFlagUnnecessary(fflags, flag_host, flag_workspaceid)
			checkFlagUnnecessary(fflags, flag_username, flag_workspaceid)
			checkFlagUnnecessary(fflags, flag_password, flag_workspaceid)

			err := getWorkspaceWithDbAndValid(workspaceId, &workspaceInfo) // 从数据库中加载

			if err != nil {
				return workspace.WorkspaceInfo{}, err
			}
			if workspaceInfo.IsNil() {
				return workspace.WorkspaceInfo{}, errors.New(i18nInstance.Main.Err_workspace_none)
			}

		} else { //2.2.2. 没有指定 workspaceid 的情况

			//1.2.0.1. 工作区类型
			workspaceInfo.Mode = workspace.WorkingMode_Local
			if getFlagValue(fflags, flag_host) != "" {
				workspaceInfo.Mode = workspace.WorkingMode_Remote
			} else if getFlagValue(fflags, "k8s") != "" {
				workspaceInfo.Mode = workspace.WorkingMode_K8s
			}

			//1.2.0.2. git 相关
			loadGitInfo4Workspace(&workspaceInfo, cmd, args)

			// 模式
			if workspaceInfo.Mode == workspace.WorkingMode_Remote { //1.2.1. vm 模式
				hostInfo, err := getRemoteAndValid(fflags)
				if err != nil {
					return workspace.WorkspaceInfo{}, err
				}
				workspaceInfo.Remote = *hostInfo

				// new 命令时，项目名称
				if cmd.Name() == "new" {
					if tmp, er := fflags.GetString("workspacename"); tmp == "" || er != nil {
						return workspace.WorkspaceInfo{}, errors.New("参数 workspacename 不为空！")
					}
					workspaceInfo.Name, _ = fflags.GetString("workspacename")
				}

				if workspaceInfo.Name == "" {
					workspaceInfo.Name = common.GetRepoName(workspaceInfo.GitCloneRepoUrl) + "-" + common.RandLowStr(3)
				}
				workspaceInfo.WorkingDirectoryPath = common.FilePahtJoin4Linux("~", model.CONST_REMOTE_REPO_ROOT, workspaceInfo.Name)

			} else if workspaceInfo.Mode == workspace.WorkingMode_K8s { //1.2.2. k8s模式
				workspaceInfo.K8sInfo.KubeConfigFilePath = getFlagValue(fflags, flag_kubeconfig)
				workspaceInfo.K8sInfo.Context = getFlagValue(fflags, flag_k8s)

			} else { //1.2.3. 本地模式
				// 当前目录
				pwd, err := os.Getwd()
				if err != nil {
					return workspace.WorkspaceInfo{}, err
				}
				workspaceInfo.Name = filepath.Base(pwd) + "-" + common.RandLowStr(3) // 工作区名称
				workspaceInfo.WorkingDirectoryPath = pwd                             // 工作目录默认使用当前文件夹

				// 本地模式下，不需要录入git库的克隆地址、分支
				checkFlagUnnecessary(fflags, flag_branch, "mode=local")

				// 本地模式下需要clone repo的情况：smartide start https://gitee.com/idcf-boat-house/boathouse-calculator.git
				if workspaceInfo.GitCloneRepoUrl != "" {
					common.SmartIDELog.Info(i18nInstance.Start.Info_git_clone)
					actualGtiRepoUrl := workspaceInfo.GitCloneRepoUrl
					if workspaceInfo.GitRepoAuthType == workspace.GitRepoAuthType_Basic {
						actualGtiRepoUrl, err =
							common.AddUsernamePassword4ActualGitRpoUrl(actualGtiRepoUrl, workspaceInfo.GitUserName, workspaceInfo.GitPassword)
						if err != nil {
							common.SmartIDELog.Warning(err.Error())
						}
					}
					clonedRepoDir, err := cloneRepo4LocalWithCommand(pwd, actualGtiRepoUrl) //
					common.CheckError(err)
					common.SmartIDELog.Info(i18nInstance.Common.Info_gitrepo_clone_done)

					workspaceInfo.WorkingDirectoryPath = clonedRepoDir
					os.Chdir(clonedRepoDir)
				} else {
					// 非 new 命令，才会从本地获取 git remote url
					if cmd.Use != "new" {
						workspaceInfo.GitCloneRepoUrl, _, err = getLocalGitRepoUrl(pwd) // 获取本地关联的repo url
						common.CheckError(err)
					}

				}
			}

			// 当为 new command 时，再次处理 workspace name
			if workspaceInfo.Name == "" {
				if cmd.Use == "new" {
					workspaceInfo.Name = fmt.Sprintf("%v_%v-%v",
						workspaceInfo.SelectedTemplate.TypeName, workspaceInfo.SelectedTemplate.SubType, common.RandLowStr(3))
				}
			}

			// 从数据库中查找是否存在对应的workspace信息
			workspaceInfoDb, err := dal.GetSingleWorkspaceByParams(workspaceInfo.Mode, workspaceInfo.WorkingDirectoryPath,
				workspaceInfo.GitCloneRepoUrl, workspaceInfo.GitBranch, workspaceInfo.ConfigFileRelativePath,
				workspaceInfo.Remote.ID, workspaceInfo.Remote.Addr, workspaceInfo.Remote.UserName)
			common.CheckError(err)
			isNewWorkspace := true
			if workspaceInfoDb.ID != "" {
				if workspaceInfo.CliRunningEnv != workspace.CliRunningEnvEnum_Client {
					common.SmartIDELog.Error("Not running on the server")
				}
				if workspaceInfo.Mode == workspace.WorkingMode_Remote {
					isEnable := ""
					fmt.Printf("远程工作区重复（可通过smartide list查看），是否创建新的工作区？（y | n）")
					fmt.Scanln(&isEnable)
					if strings.ToLower(isEnable) != "y" {
						isNewWorkspace = false
						os.Exit(1)
					}

				} else {
					isNewWorkspace = false
					selectedTemplate := workspaceInfo.SelectedTemplate
					copier.CopyWithOption(&workspaceInfo, workspaceInfoDb,
						copier.Option{
							IgnoreEmpty: false,
							DeepCopy:    true,
						})
					// 防止拷贝后，模板信息为空
					if workspaceInfo.SelectedTemplate == nil && selectedTemplate != nil {
						workspaceInfo.SelectedTemplate = selectedTemplate
					}

				}

			}

			if isNewWorkspace {
				if workspaceInfo.Mode != workspace.WorkingMode_K8s {
					// docker-compose 文件路径
					workspaceInfo.TempYamlFileAbsolutePath = workspaceInfo.GetTempDockerComposeFilePath()
				}

			}
		}
	}
	// 防止工作区名称为空的情况
	if workspaceInfo.Name == "" {
		workspaceInfo.Name = common.GetRepoName(workspaceInfo.GitCloneRepoUrl) + "-" + common.RandLowStr(3)
	}
	// addon,处理已经addon webterminal后再次运行不加aadon
	addon, _ := fflags.GetString("addon")
	if addon != "" {
		workspaceInfo.Addon = workspace.Addon{
			Type:     addon,
			IsEnable: true,
		}
	} else {
		// 仅对本地模式有效
		webterminServiceName := fmt.Sprintf("%v_smartide-webterminal", workspaceInfo.Name)
		if _, ok := workspaceInfo.TempDockerCompose.Services[webterminServiceName]; ok {
			workspaceInfo.Addon = workspace.Addon{
				Type:     "webterminal",
				IsEnable: true,
			}
		}
	}

	// 验证
	if workspaceInfo.Mode == workspace.WorkingMode_Remote {
		// path change
		workspaceInfo.ConfigFileRelativePath = common.FilePahtJoin4Linux(workspaceInfo.ConfigFileRelativePath)
		workspaceInfo.TempYamlFileAbsolutePath = common.FilePahtJoin4Linux(workspaceInfo.TempYamlFileAbsolutePath)
		workspaceInfo.WorkingDirectoryPath = common.FilePahtJoin4Linux(workspaceInfo.WorkingDirectoryPath)

		// 在远程模式下，首先验证远程服务器是否可以登录
		ssmRemote := common.SSHRemote{}
		common.SmartIDELog.InfoF(i18nInstance.Main.Info_ssh_connect_check, workspaceInfo.Remote.Addr, workspaceInfo.Remote.SSHPort)

		err = ssmRemote.CheckDail(workspaceInfo.Remote.Addr, workspaceInfo.Remote.SSHPort, workspaceInfo.Remote.UserName, workspaceInfo.Remote.Password, workspaceInfo.Remote.SSHKey)
		if err != nil {
			return workspaceInfo, err
		}
	}

	// 提示 加载对应的workspace
	if workspaceInfo.ID != "" && workspaceInfo.IsNotNil() {
		common.SmartIDELog.InfoF(i18nInstance.Start.Info_workspace_record_load, workspaceInfo.ID)

	}

	return workspaceInfo, err
}

func getSeletedTemplate(cmd *cobra.Command, args []string) (*templateModel.SelectedTemplateTypeBo, error) {
	if cmd.Use != "new" {
		return nil, nil
	}
	selectedTemplateSettings, err := GetTemplateSetting(cmd, args) // 包含了“clone 模板文件到本地”
	if err != nil {
		return nil, err
	}
	/* if selectedTemplateSettings == nil { // 未指定模板类型的时候，提示用户后退出
		common.CheckError(errors.New("模板配置为空！"))
	} */
	// workspaceInfo.SelectedTemplate = selectedTemplateSettings
	return selectedTemplateSettings, err
}

func entryptionKey4Workspace(workspaceInfo workspace.WorkspaceInfo) {
	if workspaceInfo.Remote.Password != "" {
		common.SmartIDELog.AddEntryptionKey(workspaceInfo.Remote.Password)
	}
	if workspaceInfo.K8sInfo.KubeConfigContent != "" {
		common.SmartIDELog.AddEntryptionKeyWithReservePart(workspaceInfo.K8sInfo.KubeConfigContent)
	}
}

// 从命令行中加载 git 相关信息
func loadGitInfo4Workspace(workspaceInfo *workspace.WorkspaceInfo, cmd *cobra.Command, args []string) error {
	if workspaceInfo.GitCloneRepoUrl == "" {
		// 从args或flags中获取giturl，如 : smartide start https://github.com/idcf-boat-house/boathouse-calculator.git
		workspaceInfo.GitCloneRepoUrl = getGitUrlFromArgs(cmd, args)
	}

	// 用户名密码
	fflags := cmd.Flags()
	gitUsername := getFlagValue(fflags, "gitusername")
	gitPassword := getFlagValue(fflags, "gitpassword")
	if gitPassword != "" && gitUsername == "" {
		return errors.New("当参数 --gitpassword 不为空时，--gitusername 参数必须设置")
	} else if gitPassword == "" && gitUsername != "" {
		return errors.New("当参数 --gitusername 不为空时，--gitpassword 参数必须设置")
	}
	workspaceInfo.GitUserName = gitUsername
	workspaceInfo.GitPassword = gitPassword

	// 认证类型
	if strings.Index(workspaceInfo.GitCloneRepoUrl, "git") == 0 {
		workspaceInfo.GitRepoAuthType = workspace.GitRepoAuthType_SSH
	} else if workspaceInfo.GitUserName != "" && workspaceInfo.GitPassword != "" {
		workspaceInfo.GitRepoAuthType = workspace.GitRepoAuthType_Basic
	} else {
		workspaceInfo.GitRepoAuthType = workspace.GitRepoAuthType_Public
	}

	return nil
}

// 从文件夹中获取git repo url
func getLocalGitRepoUrl(pwd string) (gitRemmoteUrl, pathName string, err error) {
	// current directory
	fileInfo, err := os.Stat(pwd)
	if err != nil {
		return
	}
	pathName = fileInfo.Name()

	// git remote url
	gitRepo, err := git.PlainOpen(pwd)
	if err != nil {
		return
	}

	gitRemote, err := gitRepo.Remote("origin")
	if err != nil {
		return
	}
	gitRemmoteUrl = gitRemote.Config().URLs[0]

	return gitRemmoteUrl, pathName, nil
}

// 获取命令行参数的值
func getFlagValue(fflags *pflag.FlagSet, flag string) string {
	value, err := fflags.GetString(flag)
	if err != nil {
		if strings.Contains(err.Error(), "flag accessed but not defined:") { // 错误判断，不需要双语
			common.SmartIDELog.Debug(err.Error())
		} else {
			common.SmartIDELog.Error(err)
		}

	}
	return value
}

/* // 友好的错误
type FriendlyError struct {
	Err error
}

func (e *FriendlyError) Error() string {
	return e.Err.Error()
} */

// 从数据库中加载工作区信息，并进行验证
func getWorkspaceWithDbAndValid(workspaceId int, originWorkspaceInfo *workspace.WorkspaceInfo) error {

	workspaceInfo_, err := dal.GetSingleWorkspace(int(workspaceId))
	if err != nil {
		return err
	}
	workspaceInfo_.CacheEnv = workspace.CacheEnvEnum_Local
	workspaceInfo_.CliRunningEnv = originWorkspaceInfo.CliRunningEnv

	// 验证在workspace中是否存在
	if workspaceInfo_.IsNil() {
		msg := fmt.Sprintf(i18nInstance.Main.Err_workspace_none)
		err = errors.New(msg)
		return err
	}

	// 如果扩展信息为空（通常是旧项目导致的），就从docker-compose文件中加载扩展信息
	if workspaceInfo_.Extend.IsNil() &&
		workspaceInfo_.ConfigYaml.IsNotNil() && workspaceInfo_.TempDockerCompose.IsNotNil() {
		workspaceInfo_.Extend = workspaceInfo_.GetWorkspaceExtend()

	}

	// 当前目录下不需要再次录入workspaceid
	twd, _ := os.Getwd()
	if workspaceInfo_.Mode == workspace.WorkingMode_Local && twd == workspaceInfo_.WorkingDirectoryPath {
		common.SmartIDELog.Warning(i18nInstance.Main.Err_flag_value_invalid2)

	}

	// 临时compose文件路径
	if workspaceInfo_.TempYamlFileAbsolutePath == "" {
		workspaceInfo_.TempYamlFileAbsolutePath = workspaceInfo_.GetTempDockerComposeFilePath()
	}

	// 验证数据
	err = workspaceInfo_.Valid()

	// 赋值
	copier.CopyWithOption(originWorkspaceInfo, &workspaceInfo_, copier.Option{IgnoreEmpty: true, DeepCopy: true})

	return err
}

// 根据参数，从数据库或者其他参数中加载远程服务器的信息
func getRemoteAndValid(fflags *pflag.FlagSet) (remoteInfo *workspace.RemoteInfo, err error) {

	host, _ := fflags.GetString(flag_host)
	userName, _ := fflags.GetString(flag_username)

	// 指定了host信息，尝试从数据库中加载
	if common.IsNumber(host) {
		remoteId, err := strconv.Atoi(host)
		if err != nil {
			return nil, err
		}
		remoteInfo, err = dal.GetRemoteById(remoteId)
		if err != nil {
			return nil, err
		}

		if remoteInfo == nil {
			return nil, errors.New(i18nInstance.Host.Err_host_data_not_exit)
		}
	} else {
		remoteInfo, err = dal.GetRemoteByHost(host, userName)

		// 如果在sqlite中有缓存数据，就不需要用户名、密码
		if remoteInfo != nil && remoteInfo.UserName == flag_username {
			checkFlagUnnecessary(fflags, flag_username, flag_host)
			checkFlagUnnecessary(fflags, flag_password, flag_host)
		}
	}

	// 从参数中加载
	if remoteInfo == nil {
		remoteInfo = &workspace.RemoteInfo{}
		//  必填字段验证
		err = checkFlagRequired(fflags, flag_host)
		if err != nil {
			return remoteInfo, &FriendlyError{Err: err}
		}
		err = checkFlagRequired(fflags, flag_username)
		if err != nil {
			return remoteInfo, &FriendlyError{Err: err}
		}

		remoteInfo.Addr = host
		remoteInfo.UserName = getFlagValue(fflags, flag_username)
		remoteInfo.SSHPort, err = fflags.GetInt(flag_port) //strconv.Atoi(getFlagValue(fflags, flag_port))
		common.CheckError(err)
		if remoteInfo.SSHPort <= 0 {
			remoteInfo.SSHPort = model.CONST_Container_SSHPort
		}
		// 认证类型
		if fflags.Changed(flag_password) {
			remoteInfo.Password = getFlagValue(fflags, flag_password)
			remoteInfo.AuthType = workspace.RemoteAuthType_Password
		} else {
			remoteInfo.AuthType = workspace.RemoteAuthType_SSH
		}

	}

	return remoteInfo, err
}

// 获取参数gitUrl
func getGitUrlFromArgs(cmd *cobra.Command, args []string) string {
	// 从args中获取值
	var gitUrl string = ""
	if len(args) > 0 { // 从args中加载
		str := args[0]
		if strings.Index(str, "git@") == 0 || strings.Index(str, "http://") == 0 || strings.Index(str, "https://") == 0 {
			gitUrl = str
		}
	}

	//
	if gitUrl == "" {
		gitUrl = getFlagValue(cmd.Flags(), flag_repourl)
	}

	return gitUrl
}

// 直接调用git命令进行git clone
func cloneRepo4LocalWithCommand(rootDir string, actualGitRepoUrl string) (string, error) {
	repoName := common.GetRepoName(actualGitRepoUrl)
	repoPath := common.PathJoin(rootDir, repoName)

	var execCommand *exec.Cmd
	command := "git clone " + actualGitRepoUrl
	switch runtime.GOOS {
	case "windows":
		execCommand = exec.Command("powershell", "/c", command)
	case "darwin":
		execCommand = exec.Command("bash", "-c", command)
	case "linux":
		execCommand = exec.Command("bash", "-c", command)
	default:
		common.SmartIDELog.Error("can not support current os")
	}

	// run
	execCommand.Stdout = os.Stdout
	execCommand.Stderr = os.Stderr
	err := execCommand.Run()

	return repoPath, err
}

// 在某些情况下，参数填了也没有意义，比如指定了workspaceid，就不需要再填host
func checkFlagUnnecessary(fflags *pflag.FlagSet, flagName string, preFlagName string) {
	if fflags.Changed(flagName) {
		common.SmartIDELog.WarningF(i18nInstance.Main.Err_flag_value_invalid, preFlagName, flagName)
	}
}

// 检查参数是否填写
func checkFlagRequired(fflags *pflag.FlagSet, flagName string) error {
	if !fflags.Changed(flagName) {
		return fmt.Errorf(i18nInstance.Main.Err_flag_value_required, flagName)
	}
	return nil
}
