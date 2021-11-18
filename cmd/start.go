package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/leansoftX/smartide-cli/cmd/dal"
	"github.com/leansoftX/smartide-cli/cmd/lib"
	"github.com/leansoftX/smartide-cli/cmd/start"
	"github.com/leansoftX/smartide-cli/lib/common"
	"github.com/leansoftX/smartide-cli/lib/i18n"
	"gopkg.in/src-d/go-git.v4"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// 远程服务器上的根目录
const REMOTE_REPO_ROOT string = "project"

// var i18nInstance.Start = i18n.GetInstance().Start
var i18nInstance = i18n.GetInstance()

// yaml 文件的相对路径
var configYamlFileRelativePath string = lib.ConfigRelativeFilePath

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: i18nInstance.Start.Info.Help_short,
	Long:  i18nInstance.Start.Info.Help_long,
	Example: `  smartide start --host <host> --username <username> --password <password> --repourl <git clone url> --branch <branch name> --filepath <config file path>
  smartide start --workspaceid <workspaceid>
  smartide get <workspaceid>`,
	RunE: func(cmd *cobra.Command, args []string) error {

		//0. 提示文本
		common.SmartIDELog.Info(i18n.GetInstance().Start.Info.Info_start)

		//0.1. 从参数中获取结构体，并做基本的数据有效性校验
		common.SmartIDELog.Info("加载工作区信息...")
		worksapce, err := getWorkspace4Start(cmd, args)
		common.CheckError(err)
		/* if validErr != nil {
			return validErr // 采用return的方式，可以显示flag列表 //TODO 根据错误的类型，如果是参数格式错误就是return，其他直接抛错
		} */

		// 执行命令
		if worksapce.Mode == dal.WorkingMode_Local {
			start.ExecuteStartCmd(worksapce, func(v string, d common.Docker) {})
		} else {
			start.ExecuteVmStartCmd(worksapce)
		}

		return nil
	},
}

//
func checkFlagNotRequired(fflags *pflag.FlagSet, flagName string, headers ...string) {
	if fflags.Changed(flagName) {
		common.SmartIDELog.WarningF(strings.Join(headers, " ")+"设置%v无效", flagName)
	}
}

//
func checkFlagRequired(fflags *pflag.FlagSet, flagName string) error {
	if !fflags.Changed(flagName) {
		return fmt.Errorf("%v 参数必填", flagName)
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

		checkFlagNotRequired(fflags, flag_workspaceid)
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
func getWorkspace4Start(cmd *cobra.Command, args []string) (workspaceInfo dal.WorkspaceInfo, err error) {
	fflags := cmd.Flags()
	workspaceId := getWorkspaceIdFromFlagsAndArgs(cmd, args)

	//1. 加载workspace
	workspaceInfo = dal.WorkspaceInfo{}
	//1.1. 指定了从workspaceid，从sqlite中读取
	if workspaceId > 0 {
		checkFlagNotRequired(fflags, flag_host)
		checkFlagNotRequired(fflags, flag_username)
		checkFlagNotRequired(fflags, flag_password)

		workspaceInfo, err = getWorkspaceWithDbAndValid(workspaceId)
		if err != nil {
			return dal.WorkspaceInfo{}, err
		}
		if (workspaceInfo == dal.WorkspaceInfo{}) {
			return dal.WorkspaceInfo{}, errors.New("查找不到对应的workspace信息")
		}

	} else { //1.2. 没有指定 workspaceid 的情况
		workingMode := dal.WorkingMode_Local

		// 当前目录
		pwd, err := os.Getwd()
		if err != nil {
			return dal.WorkspaceInfo{}, err
		}

		// 1.2.1. 远程模式
		if fflags.Changed(flag_host) {
			workingMode = dal.WorkingMode_Remote
			hostInfo, err := getRemoteAndValid(fflags)
			if err != nil {
				return dal.WorkspaceInfo{}, err
			}

			workspaceInfo.Remote = hostInfo
			workspaceInfo.GitCloneRepoUrl = getFlagValue(fflags, flag_repourl)
			if strings.Index(workspaceInfo.GitCloneRepoUrl, "git") == 0 {
				workspaceInfo.GitRepoAuthType = dal.GitRepoAuthType_SSH
			} else if strings.Index(workspaceInfo.GitCloneRepoUrl, "https") == 0 {
				workspaceInfo.GitRepoAuthType = dal.GitRepoAuthType_HTTPS
			}

			repoName := getRepoName(workspaceInfo.GitCloneRepoUrl)
			workspaceInfo.ProjectName = repoName

			repoWorkspaceDir := filepath.Join("~", REMOTE_REPO_ROOT, repoName)
			workspaceInfo.WorkingDirectoryPath = repoWorkspaceDir

			//TODO git 连接验证
		} else { //1.2.2. 本地模式
			// 本地模式下，不需要录入git库的克隆地址、分支
			checkFlagNotRequired(fflags, flag_repourl)
			checkFlagNotRequired(fflags, flag_branch)

			workspaceInfo.WorkingDirectoryPath = pwd

			repoUrl := getLocalGitRepoUrl()
			repoName := getRepoName(repoUrl)
			workspaceInfo.GitCloneRepoUrl = repoUrl
			workspaceInfo.ProjectName = repoName
		}

		workspaceInfo.Mode = workingMode
		workspaceInfo.Branch = getFlagValue(fflags, flag_branch)

		if fflags.Changed(flag_filepath) {
			workspaceInfo.ConfigFilePath = getFlagValue(fflags, flag_filepath)
		}

	}

	// 避免为空
	if workspaceInfo.ConfigFilePath == "" {
		workspaceInfo.ConfigFilePath = lib.ConfigRelativeFilePath
	}

	// 验证
	if workspaceInfo.Mode == dal.WorkingMode_Remote {
		// path change
		workspaceInfo.ConfigFilePath = common.FilePahtJoin4Linux(workspaceInfo.ConfigFilePath)
		workspaceInfo.TempDockerComposeFilePath = common.FilePahtJoin4Linux(workspaceInfo.TempDockerComposeFilePath)
		workspaceInfo.WorkingDirectoryPath = common.FilePahtJoin4Linux(workspaceInfo.WorkingDirectoryPath)

		// 在远程模式下，首先验证远程服务器是否可以登录
		ssmRemote := common.SSHRemote{}
		common.SmartIDELog.Info(fmt.Sprintf("检查ssh连接是否正常 %v:%v ...", workspaceInfo.Remote.Addr, workspaceInfo.Remote.SSHPort))
		err = ssmRemote.CheckDail(workspaceInfo.Remote.Addr, workspaceInfo.Remote.SSHPort, workspaceInfo.Remote.UserName, workspaceInfo.Remote.Password)
		if err != nil {
			return workspaceInfo, err
		}

	} else {
		//TODO 配置文件是否存在
	}

	return workspaceInfo, err
}

//
func getLocalGitRepoUrl() (gitRemmoteUrl string) {
	// current directory
	pwd, err := os.Getwd()
	common.CheckError(err)

	// git remote url
	gitRepo, err := git.PlainOpen(pwd)
	common.CheckError(err)
	gitRemote, err := gitRepo.Remote("origin")
	common.CheckError(err)
	gitRemmoteUrl = gitRemote.Config().URLs[0]

	return gitRemmoteUrl
}

//
func getFlagValue(fflags *pflag.FlagSet, flag string) string {
	value, err := fflags.GetString(flag)
	if err != nil {
		if strings.Contains(err.Error(), "flag accessed but not defined:") {
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

func (e *FriendlyError) Error() string {
	return e.Err.Error()
}

//
func getWorkspaceWithDbAndValid(workspaceId int) (workspaceInfo dal.WorkspaceInfo, err error) {

	workspaceInfo, err = dal.GetSingleWorkspace(int(workspaceId))
	common.CheckError(err)

	// 验证在workspace中是否存在
	if (workspaceInfo == dal.WorkspaceInfo{}) {
		msg := fmt.Sprintf("指定的的workspaceid（%v）无效", workspaceId)
		err = errors.New(msg)
		return workspaceInfo, err
	}

	twd, _ := os.Getwd()
	if workspaceInfo.Mode == dal.WorkingMode_Local && twd == workspaceInfo.WorkingDirectoryPath {
		common.SmartIDELog.Warning("当前目录下不需要录入workspaceid")
	}

	repoName := getRepoName(workspaceInfo.GitCloneRepoUrl)
	workspaceInfo.ProjectName = repoName

	return workspaceInfo, err
}

// 根据参数，从数据库或者其他参数中加载远程服务器的信息
func getRemoteAndValid(fflags *pflag.FlagSet) (remoteInfo dal.RemoteInfo, err error) {

	host, _ := fflags.GetString(flag_host)
	remoteInfo = dal.RemoteInfo{}

	// 指定了host信息，尝试从数据库中加载
	if common.IsNumber(host) {
		remoteId, err := strconv.Atoi(host)
		common.CheckError(err)
		remoteInfo, err = dal.GetRemoteById(remoteId)
		common.CheckError(err)

		if (dal.RemoteInfo{} == remoteInfo) {
			common.SmartIDELog.Warning("没有在缓存中查找到关联的host信息")
		}
	} else {
		remoteInfo, err = dal.GetRemoteByHost(host)

		// 如果在sqlite中有缓存数据，就不需要用户名、密码
		if (dal.RemoteInfo{} != remoteInfo) {
			header := "host信息已经缓存，"
			checkFlagNotRequired(fflags, flag_username, header)
			checkFlagNotRequired(fflags, flag_password, header)
		}
	}

	// 从参数中加载
	if (dal.RemoteInfo{} == remoteInfo) {
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
			remoteInfo.SSHPort = lib.CONST_DefaultSSHPort
		}
		// 认证类型
		if fflags.Changed(flag_password) {
			remoteInfo.Password = getFlagValue(fflags, flag_password)
			remoteInfo.AuthType = dal.RemoteAuthType_Password
		} else {
			remoteInfo.AuthType = dal.RemoteAuthType_SSH
		}

	}

	//common.CheckError(err)

	return remoteInfo, err
}

func init() {

	startCmd.Flags().Int32P("workspaceid", "w", 0, "设置后，可以使用本地保存的信息环境信息，直接启动web ide环境")

	startCmd.Flags().StringP("host", "o", "", "可以指定host，或者hostid")
	startCmd.Flags().IntP("port", "p", 22, "SSH 端口，默认为22")
	startCmd.Flags().StringP("username", "u", "", "SSH 登录用户")
	//vmStartCmd.MarkFlagRequired("username")
	startCmd.Flags().StringP("password", "t", "", "SSH 用户密码")
	startCmd.Flags().StringP("repourl", "r", "", "远程代码仓库的克隆地址")
	//vmStartCmd.MarkFlagRequired("repourl")

	startCmd.Flags().StringVarP(&configYamlFileRelativePath, "filepath", "f", "", i18nInstance.Start.Info.Help_flag_filepath)
	startCmd.Flags().StringP("branch", "b", "", "指定git分支")

}

// get repo name
func getRepoName(repoUrl string) string {
	/* _, err := url.ParseRequestURI(repoUrl)
	if err != nil {
		panic(err)
	} */
	index := strings.LastIndex(repoUrl, "/")
	return strings.Replace(repoUrl[index+1:], ".git", "", -1)
}
