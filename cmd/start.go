/*
 * @Author: jason chen (jasonchen@leansoftx.com, http://smallidea.cnblogs.com)
 * @Description:
 * @Date: 2021-11
 * @LastEditors:
 * @LastEditTime:
 */
package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/leansoftX/smartide-cli/cmd/start"
	"github.com/leansoftX/smartide-cli/internal/apk/i18n"
	"github.com/leansoftX/smartide-cli/internal/biz/config"
	"github.com/leansoftX/smartide-cli/internal/biz/workspace"
	"github.com/leansoftX/smartide-cli/internal/dal"
	"github.com/leansoftX/smartide-cli/internal/model"

	"github.com/leansoftX/smartide-cli/internal/apk/appinsight"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"gopkg.in/src-d/go-git.v4"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// 远程服务器上的根目录
const REMOTE_REPO_ROOT string = "project"

// var i18nInstance.Start = i18n.GetInstance().Start
var i18nInstance = i18n.GetInstance()

// yaml 文件的相对路径
var configYamlFileRelativePath string = model.CONST_Default_ConfigRelativeFilePath

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: i18nInstance.Start.Info_help_short,
	Long:  i18nInstance.Start.Info_help_long,
	Example: `  smartide start --host <host> --username <username> --password <password> --repourl <git clone url> --branch <branch name> --filepath <config file path>
  smartide start --workspaceid {workspaceid}
  smartide get {workspaceid}`,
	RunE: func(cmd *cobra.Command, args []string) error {
		//0. 提示文本
		common.SmartIDELog.Info(i18nInstance.Start.Info_start)

		//0.1. 从参数中获取结构体，并做基本的数据有效性校验
		common.SmartIDELog.Info(i18nInstance.Main.Info_workspace_loading)
		worksapceInfo, err := getWorkspace4Start(cmd, args)
		common.CheckError(err)

		//ai记录
		var trackEvent string
		for _, val := range args {
			trackEvent = trackEvent + " " + val
		}

		// 执行命令
		if worksapceInfo.Mode == workspace.WorkingMode_Local {
			executeStartCmdFunc := func(yamlConfig config.SmartIdeConfig) {
				var imageNames string
				for image := range yamlConfig.Workspace.Servcies {
					tag := yamlConfig.Workspace.Servcies[image].Image.Tag
					imageName := yamlConfig.Workspace.Servcies[image].Image.Name
					if tag == "" {
						imageNames = imageNames + imageName + ","
					} else {
						imageNames = imageNames + imageName + ":" + tag + ","
					}
				}
				appinsight.SetTrack(cmd.Use, Version.TagName, trackEvent, string(worksapceInfo.Mode), imageNames)
			}

			start.ExecuteStartCmd(worksapceInfo, func(v string, d common.Docker) {}, executeStartCmdFunc)
		} else {
			executeStartCmdFunc := func(yamlConfig config.SmartIdeConfig) {
				var imageNames string
				for image := range yamlConfig.Workspace.Servcies {
					tag := yamlConfig.Workspace.Servcies[image].Image.Tag
					imageName := yamlConfig.Workspace.Servcies[image].Image.Name
					if tag == "" {
						imageNames = imageNames + imageName + ","
					} else {
						imageNames = imageNames + imageName + ":" + tag + ","
					}
				}
				appinsight.SetTrack(cmd.Use, Version.TagName, trackEvent, string(worksapceInfo.Mode), imageNames)
			}

			start.ExecuteVmStartCmd(worksapceInfo, executeStartCmdFunc)
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
)

// 获取工作区id
func getWorkspaceIdFromFlagsAndArgs(cmd *cobra.Command, args []string) int {
	fflags := cmd.Flags()

	// 从args 或者 flag 中获取值
	var workspaceId int
	if len(args) > 0 { // 从args中加载
		str := args[0]
		tmpWorkspaceId, err := strconv.Atoi(str)
		if err == nil && tmpWorkspaceId > 0 {
			workspaceId = tmpWorkspaceId
		}

		checkFlagUnnecessary(fflags, flag_workspaceid, strconv.Itoa(tmpWorkspaceId))
	} else if fflags.Changed(flag_workspaceid) { // 从flag中加载
		tmpWorkspaceId, err := fflags.GetInt32(flag_workspaceid)
		common.CheckError(err)
		if tmpWorkspaceId > 0 {
			workspaceId = int(tmpWorkspaceId)
		}
	}

	return workspaceId
}

// 从start command的flag、args中获取workspace
func getWorkspace4Start(cmd *cobra.Command, args []string) (workspaceInfo workspace.WorkspaceInfo, err error) {
	fflags := cmd.Flags()
	workspaceId := getWorkspaceIdFromFlagsAndArgs(cmd, args)

	//1. 加载workspace
	workspaceInfo = workspace.WorkspaceInfo{}
	//1.1. 指定了从workspaceid，从sqlite中读取
	if workspaceId > 0 {
		checkFlagUnnecessary(fflags, flag_host, flag_workspaceid)
		checkFlagUnnecessary(fflags, flag_username, flag_workspaceid)
		checkFlagUnnecessary(fflags, flag_password, flag_workspaceid)

		workspaceInfo, err = getWorkspaceWithDbAndValid(workspaceId)
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

		// 1.2.1. 远程模式
		if fflags.Changed(flag_host) {
			workingMode = workspace.WorkingMode_Remote
			hostInfo, err := getRemoteAndValid(fflags)
			if err != nil {
				return workspace.WorkspaceInfo{}, err
			}

			workspaceInfo.Remote = hostInfo
			workspaceInfo.GitCloneRepoUrl = getFlagValue(fflags, flag_repourl)
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

		} else { //1.2.2. 本地模式
			// 本地模式下，不需要录入git库的克隆地址、分支
			checkFlagUnnecessary(fflags, flag_repourl, "mode=local")
			checkFlagUnnecessary(fflags, flag_branch, "mode=local")

			workspaceInfo.WorkingDirectoryPath = pwd
			workspaceInfo.GitCloneRepoUrl, _ = getLocalGitRepoUrl() // 获取本地关联的repo url
		}

		// 运行模式
		workspaceInfo.Mode = workingMode

		// 从数据库中查找是否存在对应的workspace信息，主要是针对在当前目录再次start的情况
		workspaceInfoDb, err := dal.GetSingleWorkspaceByParams(workspaceInfo.Mode, workspaceInfo.WorkingDirectoryPath,
			workspaceInfo.GitCloneRepoUrl, workspaceInfo.Remote.ID, workspaceInfo.Remote.Addr)
		common.CheckError(err)
		if workspaceInfoDb.ID > 0 && workspaceInfoDb.IsNotNil() {
			workspaceInfo = workspaceInfoDb

		} else {
			// docker-compose 文件路径
			workspaceInfo.TempDockerComposeFilePath = workspaceInfo.GetTempDockerComposeFilePath()
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

	} else {

	}

	// 提示 加载对应的workspace
	if workspaceInfo.ID > 0 && workspaceInfo.IsNotNil() {
		common.SmartIDELog.InfoF(i18nInstance.Start.Info_workspace_record_load, workspaceInfo.ID)

	}

	return workspaceInfo, err
}

//
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

}

// get repo name
func getRepoName(repoUrl string) string {

	index := strings.LastIndex(repoUrl, "/")
	return strings.Replace(repoUrl[index+1:], ".git", "", -1)
}
