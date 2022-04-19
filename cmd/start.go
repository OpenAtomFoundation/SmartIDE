/*
 * @Author: jason chen (jasonchen@leansoftx.com, http://smallidea.cnblogs.com)
 * @Description:
 * @Date: 2021-11
 * @LastEditors: kenan
 * @LastEditTime: 2022-04-13 15:25:44
 */
package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/leansoftX/smartide-cli/cmd/server"
	smartideServer "github.com/leansoftX/smartide-cli/cmd/server"
	"github.com/leansoftX/smartide-cli/cmd/start"
	"github.com/leansoftX/smartide-cli/internal/apk/i18n"
	"github.com/leansoftX/smartide-cli/internal/biz/config"
	"github.com/leansoftX/smartide-cli/internal/biz/workspace"
	"github.com/leansoftX/smartide-cli/internal/dal"
	"github.com/leansoftX/smartide-cli/internal/model"

	"github.com/leansoftX/smartide-cli/internal/apk/appinsight"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	ssh2 "golang.org/x/crypto/ssh"
)

// 远程服务器上的根目录
const REMOTE_REPO_ROOT string = "project"

// var i18nInstance.Start = i18n.GetInstance().Start
var i18nInstance = i18n.GetInstance()

// yaml 文件的相对路径
var configYamlFileRelativePath string = model.CONST_Default_ConfigRelativeFilePath

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:     "start",
	Short:   i18nInstance.Start.Info_help_short,
	Long:    i18nInstance.Start.Info_help_long,
	Aliases: []string{"up"},
	Example: `  smartide start
  smartide start --workspaceid {workspaceid}
  smartide start <workspaceid>
  smartide start <git clone url>
  smartide start --host <host> --username <username> --password <password> --repourl <git clone url> --branch <branch name> --filepath <config file path>
  smartide start --host <host> <git clone url>
  smartide start --k8s <context> --namespace <namespace> --repoUrl <git clone url> --branch master
  smartide start --k8s <context> --namespace <namespace> <git clone url>`,
	RunE: func(cmd *cobra.Command, args []string) error {

		if apiHost, _ := cmd.Flags().GetString(server.Flags_ServerHost); apiHost != "" {
			wsURL := fmt.Sprint(strings.ReplaceAll(strings.ReplaceAll(apiHost, "https", "ws"), "http", "ws"), "/ws/smartide/ws")
			common.WebsocketStart(wsURL)
			token, _ := cmd.Flags().GetString(server.Flags_ServerToken)
			if token != "" {
				if workspaceIdStr := getWorkspaceIdFromFlagsAndArgs(cmd, args); strings.Contains(workspaceIdStr, "SWS") {
					if pid, err := workspace.GetParentId(workspaceIdStr, 1, token, apiHost); err == nil && pid > 0 {
						common.SmartIDELog.Ws_id = workspaceIdStr
						common.SmartIDELog.ParentId = pid
					}
				} else {
					if workspaceIdStr, _ := cmd.Flags().GetString(server.Flags_ServerWorkspaceid); workspaceIdStr != "" {
						if no, _ := workspace.GetWorkspaceNo(workspaceIdStr, token, apiHost); no != "" {
							if pid, err := workspace.GetParentId(no, 1, token, apiHost); err == nil && pid > 0 {
								common.SmartIDELog.Ws_id = no
								common.SmartIDELog.ParentId = pid
							}
						}
					}

				}
			}

		}
		//0. 提示文本
		common.SmartIDELog.Info(i18nInstance.Start.Info_start)

		//0.1. 从参数中获取结构体，并做基本的数据有效性校验
		common.SmartIDELog.Info(i18nInstance.Main.Info_workspace_loading)
		worksapceInfo, err := getWorkspaceFromCmd(cmd, args)
		common.CheckErrorFunc(err, func(err error) {
			mode, _ := cmd.Flags().GetString("mode")
			isModeServer := strings.ToLower(mode) == "server"
			if !isModeServer {
				return
			}
			if err != nil {
				smartideServer.Feedback_Finish(server.FeedbackCommandEnum_Start, cmd, false, 0, workspace.WorkspaceInfo{}, err.Error(), "")
			}
		})

		// 检查是否为服务器端模式
		err = smartideServer.Check(cmd)
		common.CheckError(err)

		//ai记录
		var trackEvent string
		for _, val := range args {
			trackEvent = trackEvent + " " + val
		}

		// 执行命令
		if worksapceInfo.Mode == workspace.WorkingMode_Local { // 本地模式
			executeStartCmdFunc := func(yamlConfig config.SmartIdeConfig) {
				var imageNames []string
				for _, service := range yamlConfig.Workspace.Servcies {
					imageNames = append(imageNames, service.Image)
				}
				appinsight.SetTrack(cmd.Use, Version.TagName, trackEvent, string(worksapceInfo.Mode), strings.Join(imageNames, ","))
			}
			start.ExecuteStartCmd(worksapceInfo, func(v string, d common.Docker) {}, executeStartCmdFunc)

		} else if worksapceInfo.Mode == workspace.WorkingMode_K8s { // k8s 模式
			executeStartCmdFunc := func(yamlConfig config.SmartIdeConfig) {
				var imageNames []string
				for _, service := range yamlConfig.Workspace.Servcies {
					imageNames = append(imageNames, service.Image)
				}
				appinsight.SetTrack(cmd.Use, Version.TagName, trackEvent, string(worksapceInfo.Mode), strings.Join(imageNames, ","))
			}
			start.ExecuteK8sStartCmd(worksapceInfo, executeStartCmdFunc)

			//99. 死循环进行驻守
			for {
				time.Sleep(500)
			}

		} else if worksapceInfo.Mode == workspace.WorkingMode_Server { // server vm 模式
			executeStartCmdFunc := func(yamlConfig config.SmartIdeConfig) {
				var imageNames []string
				for _, service := range yamlConfig.Workspace.Servcies {
					imageNames = append(imageNames, service.Image)
				}
				appinsight.SetTrack(cmd.Use, Version.TagName, trackEvent, string(worksapceInfo.Mode), strings.Join(imageNames, ","))
			}
			start.ExecuteServerVmStartCmd(worksapceInfo, executeStartCmdFunc)

			//99. 死循环进行驻守
			for {
				time.Sleep(500)
			}

		} else { // vm 模式
			executeStartCmdFunc := func(yamlConfig config.SmartIdeConfig) {
				var imageNames []string
				for _, service := range yamlConfig.Workspace.Servcies {
					imageNames = append(imageNames, service.Image)
				}
				appinsight.SetTrack(cmd.Use, Version.TagName, trackEvent, string(worksapceInfo.Mode), strings.Join(imageNames, ","))
			}
			start.ExecuteVmStartCmd(worksapceInfo, executeStartCmdFunc, cmd)

		}

		return nil
	},
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

var (
	flag_workspaceid = "workspaceid"
	flag_host        = "host"
	flag_port        = "port"
	flag_username    = "username"
	flag_password    = "password"
	flag_filepath    = "filepath"
	flag_repourl     = "repourl"
	flag_branch      = "branch"
	flag_k8s         = "k8s"
	flag_namespace   = "namespace"

//	flag_loginurl    = "login_url"
)

// 获取工作区id
func getWorkspaceIdFromFlagsAndArgs(cmd *cobra.Command, args []string) string {
	fflags := cmd.Flags()

	// 从args 或者 flag 中获取值
	var workspaceId string
	if len(args) > 0 { // 从args中加载
		workspaceId = args[0]

		checkFlagUnnecessary(fflags, flag_workspaceid, workspaceId)
	} else if fflags.Changed(flag_workspaceid) { // 从flag中加载
		fflags.GetString(flag_workspaceid)
		tmpWorkspaceId, err := fflags.GetString(flag_workspaceid)
		common.CheckError(err)
		if tmpWorkspaceId != "" {
			workspaceId = tmpWorkspaceId
		}
	}

	return workspaceId
}

// 从start command的flag、args中获取workspace
func getWorkspaceFromCmd(cmd *cobra.Command, args []string) (workspaceInfo workspace.WorkspaceInfo, err error) {
	fflags := cmd.Flags()
	//从args或flags中获取workspaceid，如 : smartide start 1
	workspaceIdStr := getWorkspaceIdFromFlagsAndArgs(cmd, args)
	//从args或flags中获取giturl，如 : smartide start https://github.com/idcf-boat-house/boathouse-calculator.git
	gitUrl := getGitUrlFromArgs(cmd, args)
	//1. 加载workspace
	workspaceInfo = workspace.WorkspaceInfo{}

	// server 模式从api 获取ws
	if strings.Contains(strings.ToUpper(workspaceIdStr), "SW") {
		auth := model.Auth{}
		auth, err = workspace.GetCurrentUser()
		if auth != (model.Auth{}) && auth.Token != "" {
			// 从api 获取workspace
			workspaceInfo, err = workspace.GetWorkspaceFromServer(auth, workspaceIdStr)
			common.CheckError(err)

			// 使用是否关联 server workspace 进行判断
			if (workspaceInfo.ServerWorkSpace == model.ServerWorkspace{}) {
				common.SmartIDELog.Error("没有查询到对应的数据！")
			}
		}

	} else {
		workspaceId, _ := strconv.Atoi(workspaceIdStr)
		common.CheckError(err)
		//1.1. 指定了从workspaceid，从sqlite中读取
		if workspaceId > 0 {
			checkFlagUnnecessary(fflags, flag_host, flag_workspaceid)
			checkFlagUnnecessary(fflags, flag_username, flag_workspaceid)
			checkFlagUnnecessary(fflags, flag_password, flag_workspaceid)

			workspaceInfo, err = getWorkspaceWithDbAndValid(workspaceId) // 从数据库中加载
			if err != nil {
				return workspace.WorkspaceInfo{}, err
			}
			if workspaceInfo.IsNil() {
				return workspace.WorkspaceInfo{}, errors.New(i18nInstance.Main.Err_workspace_none)
			}

		} else { //1.2. 没有指定 workspaceid 的情况
			workingMode := workspace.WorkingMode_Local

			// 当前目录
			pwd, err := os.Getwd()
			if err != nil {
				return workspace.WorkspaceInfo{}, err
			}

			// 模式
			if fflags.Changed(flag_host) { // vm 模式
				workingMode = workspace.WorkingMode_Remote
				hostInfo, err := getRemoteAndValid(fflags)
				if err != nil {
					return workspace.WorkspaceInfo{}, err
				}

				workspaceInfo.Remote = hostInfo
				if gitUrl != "" {
					workspaceInfo.GitCloneRepoUrl = gitUrl
				} else {
					workspaceInfo.GitCloneRepoUrl = getFlagValue(fflags, flag_repourl)
				}
				if strings.Index(workspaceInfo.GitCloneRepoUrl, "git") == 0 {
					workspaceInfo.GitRepoAuthType = workspace.GitRepoAuthType_SSH
				} else if strings.Index(workspaceInfo.GitCloneRepoUrl, "https") == 0 {
					workspaceInfo.GitRepoAuthType = workspace.GitRepoAuthType_HTTPS
				} else if strings.Index(workspaceInfo.GitCloneRepoUrl, "http") == 0 {
					workspaceInfo.GitRepoAuthType = workspace.GitRepoAuthType_HTTP
				}

				repoName := getRepoName(workspaceInfo.GitCloneRepoUrl)
				repoWorkspaceDir := filepath.Join("~", REMOTE_REPO_ROOT, repoName)
				workspaceInfo.WorkingDirectoryPath = repoWorkspaceDir

			} else if fflags.Changed("k8s") { // k8s模式
				workspaceInfo.GitCloneRepoUrl = getFlagValue(fflags, flag_repourl)
				if gitUrl != "" {
					workspaceInfo.GitCloneRepoUrl = gitUrl
				} else {
					workspaceInfo.GitCloneRepoUrl = getFlagValue(fflags, flag_repourl)
				}

				workingMode = workspace.WorkingMode_K8s
				workspaceInfo.K8sInfo.Namespace = getFlagValue(fflags, flag_namespace)
				if workspaceInfo.K8sInfo.Namespace == "" {
					workspaceInfo.K8sInfo.Namespace = "default"
				}
				workspaceInfo.K8sInfo.Context = getFlagValue(fflags, flag_k8s)

			} else { // 本地模式
				// 本地模式下，不需要录入git库的克隆地址、分支
				checkFlagUnnecessary(fflags, flag_repourl, "mode=local")
				checkFlagUnnecessary(fflags, flag_branch, "mode=local")
				//本地模式下需要clone repo的情况：smartide start https://gitee.com/idcf-boat-house/boathouse-calculator.git
				if gitUrl != "" {
					common.SmartIDELog.Info(i18nInstance.Start.Info_git_clone)
					clonedRepoDir, err := cloneRepo4LocalWithCommand(pwd, gitUrl) //cloneRepo4Local(pwd, gitUrl)
					common.CheckError(err)
					common.SmartIDELog.Info(i18nInstance.Common.Info_gitrepo_clone_done)

					workspaceInfo.WorkingDirectoryPath = clonedRepoDir
					workspaceInfo.GitCloneRepoUrl, _ = getLocalGitRepoUrl(clonedRepoDir) // 获取本地关联的repo url
					os.Chdir(clonedRepoDir)
				} else {
					workspaceInfo.WorkingDirectoryPath = pwd
					workspaceInfo.GitCloneRepoUrl, _ = getLocalGitRepoUrl(pwd) // 获取本地关联的repo url
				}
			}

			// 运行模式
			workspaceInfo.Mode = workingMode

			// 从数据库中查找是否存在对应的workspace信息，主要是针对在当前目录再次start的情况
			workspaceInfoDb, err := dal.GetSingleWorkspaceByParams(workspaceInfo.Mode, workspaceInfo.WorkingDirectoryPath,
				workspaceInfo.GitCloneRepoUrl, workspaceInfo.Remote.ID, workspaceInfo.Remote.Addr)
			common.CheckError(err)

			if workspaceInfoDb.ID != "" && workspaceInfoDb.IsNotNil() {
				workspaceInfo = workspaceInfoDb

			} else {
				if workspaceInfo.Mode != workspace.WorkingMode_K8s {
					// docker-compose 文件路径
					workspaceInfo.TempDockerComposeFilePath = workspaceInfo.GetTempDockerComposeFilePath()
				}

			}
		}
	}

	/* // 项目名称
	workspaceInfo.GetProjectDirctoryName() = getProjectName(workspaceInfo.Mode, workspaceInfo.GitCloneRepoUrl) */

	// 涉及到配置文件的改变
	if fflags.Changed(flag_filepath) {
		configFilePath, err := fflags.GetString(flag_filepath)
		common.CheckError(err)

		if configFilePath != "" {
			workspaceInfo.ConfigFilePath = configFilePath
		}
	}
	if workspaceInfo.ConfigFilePath == "" { // 避免配置文件的路径为空
		workspaceInfo.ConfigFilePath = model.CONST_Default_ConfigRelativeFilePath
	}
	if fflags.Changed(flag_branch) {
		branch, err := fflags.GetString(flag_branch)
		common.CheckError(err)

		if branch != "" {
			workspaceInfo.Branch = branch
		}
	}

	// 验证
	if workspaceInfo.Mode == workspace.WorkingMode_Remote {
		// path change
		workspaceInfo.ConfigFilePath = common.FilePahtJoin4Linux(workspaceInfo.ConfigFilePath)
		workspaceInfo.TempDockerComposeFilePath = common.FilePahtJoin4Linux(workspaceInfo.TempDockerComposeFilePath)
		workspaceInfo.WorkingDirectoryPath = common.FilePahtJoin4Linux(workspaceInfo.WorkingDirectoryPath)

		// 在远程模式下，首先验证远程服务器是否可以登录
		ssmRemote := common.SSHRemote{}
		common.SmartIDELog.InfoF(i18nInstance.Main.Info_ssh_connect_check, workspaceInfo.Remote.Addr, workspaceInfo.Remote.SSHPort)
		err = ssmRemote.CheckDail(workspaceInfo.Remote.Addr, workspaceInfo.Remote.SSHPort, workspaceInfo.Remote.UserName, workspaceInfo.Remote.Password)
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

//
func getLocalGitRepoUrl(pwd string) (gitRemmoteUrl, pathName string) {
	// current directory
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

//
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

// 友好的错误
type FriendlyError struct {
	Err error
}

//
func (e *FriendlyError) Error() string {
	return e.Err.Error()
}

// 从数据库中加载工作区信息，并进行验证
func getWorkspaceWithDbAndValid(workspaceId int) (workspaceInfo workspace.WorkspaceInfo, err error) {

	workspaceInfo, err = dal.GetSingleWorkspace(int(workspaceId))
	common.CheckError(err)

	// 验证在workspace中是否存在
	if workspaceInfo.IsNil() {
		msg := fmt.Sprintf(i18nInstance.Main.Err_workspace_none)
		err = errors.New(msg)
		return workspaceInfo, err
	}

	// 如果扩展信息为空（通常是旧项目导致的），就从docker-compose文件中加载扩展信息
	if workspaceInfo.Extend.IsNil() &&
		workspaceInfo.ConfigYaml.IsNotNil() && workspaceInfo.TempDockerCompose.IsNotNil() {
		workspaceInfo.Extend = workspaceInfo.GetWorkspaceExtend()

	}

	// 当前目录下不需要再次录入workspaceid
	twd, _ := os.Getwd()
	if workspaceInfo.Mode == workspace.WorkingMode_Local && twd == workspaceInfo.WorkingDirectoryPath {
		common.SmartIDELog.Warning(i18nInstance.Main.Err_flag_value_invalid2)

	}

	// 临时compose文件路径
	if workspaceInfo.TempDockerComposeFilePath == "" {
		workspaceInfo.TempDockerComposeFilePath = workspaceInfo.GetTempDockerComposeFilePath()
	}

	// 验证数据
	err = workspaceInfo.Valid()

	return workspaceInfo, err
}

// 根据参数，从数据库或者其他参数中加载远程服务器的信息
func getRemoteAndValid(fflags *pflag.FlagSet) (remoteInfo workspace.RemoteInfo, err error) {

	host, _ := fflags.GetString(flag_host)
	remoteInfo = workspace.RemoteInfo{}

	// 指定了host信息，尝试从数据库中加载
	if common.IsNumber(host) {
		remoteId, err := strconv.Atoi(host)
		common.CheckError(err)
		remoteInfo, err = dal.GetRemoteById(remoteId)
		common.CheckError(err)

		if (workspace.RemoteInfo{} == remoteInfo) {
			common.SmartIDELog.Warning(i18nInstance.Host.Err_host_data_not_exit)
		}
	} else {
		remoteInfo, err = dal.GetRemoteByHost(host)

		// 如果在sqlite中有缓存数据，就不需要用户名、密码
		if (workspace.RemoteInfo{} != remoteInfo) {
			checkFlagUnnecessary(fflags, flag_username, flag_host)
			checkFlagUnnecessary(fflags, flag_password, flag_host)
		}
	}

	// 从参数中加载
	if (workspace.RemoteInfo{} == remoteInfo) {
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

func init() {

	startCmd.Flags().Int32P("workspaceid", "w", 0, i18nInstance.Remove.Info_flag_workspaceid)

	startCmd.Flags().StringP("host", "o", "", i18nInstance.Start.Info_help_flag_host)
	startCmd.Flags().IntP("port", "p", 22, i18nInstance.Start.Info_help_flag_port)
	startCmd.Flags().StringP("username", "u", "", i18nInstance.Start.Info_help_flag_username)
	startCmd.Flags().StringP("password", "t", "", i18nInstance.Start.Info_help_flag_password)
	startCmd.Flags().StringP("repourl", "r", "", i18nInstance.Start.Info_help_flag_repourl)

	startCmd.Flags().StringVarP(&configYamlFileRelativePath, "filepath", "f", "", i18nInstance.Start.Info_help_flag_filepath)
	startCmd.Flags().StringP("branch", "b", "", i18nInstance.Start.Info_help_flag_branch)
	startCmd.Flags().StringP("k8s", "k", "", i18nInstance.Start.Info_help_flag_k8s)
	startCmd.Flags().StringP("namespace", "n", "", i18nInstance.Start.Info_help_flag_k8s_namespace)

}

// get repo name
func getRepoName(repoUrl string) string {

	index := strings.LastIndex(repoUrl, "/")
	return strings.Replace(repoUrl[index+1:], ".git", "", -1)
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
	return gitUrl
}

// 直接调用git命令进行git clone
func cloneRepo4LocalWithCommand(rootDir string, gitUrl string) (string, error) {
	repoName := getRepoName(gitUrl)
	repoPath := filepath.Join(rootDir, repoName)

	var execCommand *exec.Cmd
	command := "git clone " + gitUrl
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

// clone repos for local mode
func cloneRepo4Local(rootDir string, gitUrl string) string {
	repoName := getRepoName(gitUrl)
	repoPath := filepath.Join(rootDir, repoName)
	options := &git.CloneOptions{
		URL:      gitUrl,
		Progress: os.Stdout,
	}
	if strings.Index(gitUrl, "git@") == 0 {
		var publicKey *ssh.PublicKeys

		// current user directory -> id_rsa file path
		currentUser, err := user.Current()
		if err != nil {
			log.Fatalf(err.Error())
		}
		homeDirectory := currentUser.HomeDir
		idRsaFilePath := ""
		if runtime.GOOS == "windows" {
			idRsaFilePath = path.Join(homeDirectory, "\\.ssh\\id_rsa")
		} else {
			idRsaFilePath = path.Join(homeDirectory, "/.ssh/id_rsa")
		}

		//
		publicKey, keyError := ssh.NewPublicKeysFromFile("git", idRsaFilePath, "")
		if keyError != nil {
			common.CheckError(keyError)
		}
		//ignore known_hosts check to fix gitee.com known_hosts lack issue
		publicKey.HostKeyCallback = ssh2.InsecureIgnoreHostKey()
		options.Auth = publicKey
	}
	_, err := git.PlainClone(repoPath, false, options)
	if err != nil {
		if err.Error() == "repository already exists" {
			message := fmt.Sprintf(i18nInstance.Main.Err_git_clone_folder_exist, repoName)
			common.SmartIDELog.Error(message)
		} else {
			common.CheckError(err)
		}
	}
	return repoPath
}
