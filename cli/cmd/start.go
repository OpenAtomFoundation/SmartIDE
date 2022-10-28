/*
 * @Author: jason chen (jasonchen@leansoftx.com, http://smallidea.cnblogs.com)
 * @Description:
 * @Date: 2021-11
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-10-28 17:15:38
 */
package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	cmdCommon "github.com/leansoftX/smartide-cli/cmd/common"
	"github.com/leansoftX/smartide-cli/cmd/server"
	smartideServer "github.com/leansoftX/smartide-cli/cmd/server"
	"github.com/leansoftX/smartide-cli/cmd/start"

	"github.com/leansoftX/smartide-cli/internal/apk/i18n"
	"github.com/leansoftX/smartide-cli/internal/biz/config"
	"github.com/leansoftX/smartide-cli/internal/biz/workspace"
	"github.com/leansoftX/smartide-cli/internal/model"

	"github.com/leansoftX/smartide-cli/internal/apk/appinsight"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/leansoftX/smartide-cli/pkg/k8s"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

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
  smartide start <actual git repo url>
  smartide start <actual git repo url> <templatetype> -T {typename}
  smartide start --host <host> --username <username> --password <password> --repourl <actual git repo url> --branch <branch name> --filepath <config file path>
  smartide start --host <host> --username <username> --password <password> --repourl <actual git repo url> --branch <branch name> --filepath <config file path> <templatetype> -T {typename}
  smartide start --host <hostid> <actual git repo url> 
  smartide start --host <hostid> <actual git repo url> <templatetype> -T {typename}
  smartide start --k8s <context> --repoUrl <actual git repo url> --branch master
  smartide start --k8s <context> <actual git repo url>`,
	PreRunE: preRun,
	RunE: func(cmd *cobra.Command, args []string) error {

		if apiHost, _ := cmd.Flags().GetString(server.Flags_ServerHost); apiHost != "" {
			wsURL := fmt.Sprint(strings.ReplaceAll(strings.ReplaceAll(apiHost, "https", "ws"), "http", "ws"), "/ws/smartide/ws")
			common.WebsocketStart(wsURL)
			token, _ := cmd.Flags().GetString(server.Flags_ServerToken)
			if token != "" {
				if workspaceIdStr := cmdCommon.GetWorkspaceIdFromFlagsOrArgs(cmd, args); strings.Contains(workspaceIdStr, "SWS") {
					if pid, err := workspace.GetParentId(workspaceIdStr, workspace.ActionEnum_Workspace_Start, token, apiHost); err == nil && pid > 0 {
						common.SmartIDELog.Ws_id = workspaceIdStr
						common.SmartIDELog.ParentId = pid
					}
				} else {
					if workspaceIdStr, _ := cmd.Flags().GetString(server.Flags_ServerWorkspaceid); workspaceIdStr != "" {
						if no, _ := workspace.GetWorkspaceNo(workspaceIdStr, token, apiHost); no != "" {
							if pid, err := workspace.GetParentId(no, workspace.ActionEnum_Workspace_Start, token, apiHost); err == nil && pid > 0 {
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
		workspaceInfo, err := cmdCommon.GetWorkspaceFromCmd(cmd, args) // 获取 workspace 对象 ★★★★★
		entryptionKey4Workspace(workspaceInfo)                         // 申明需要加密的文本
		common.CheckErrorFunc(err, func(err error) {
			mode, _ := cmd.Flags().GetString("mode")
			isModeServer := strings.ToLower(mode) == "server"
			if !isModeServer {
				return
			}
			if err != nil {
				common.SmartIDELog.Importance(err.Error())
				smartideServer.Feedback_Finish(server.FeedbackCommandEnum_Start, cmd, false, nil, workspaceInfo, err.Error(), "")
			}
		})

		// ai记录
		var trackEvent string
		for _, val := range args {
			trackEvent = trackEvent + " " + val
		}

		isUnforward, _ := cmd.Flags().GetBool("unforward")

		executeStartCmdFunc := func(yamlConfig config.SmartIdeConfig) {
			if config.GlobalSmartIdeConfig.IsInsightEnabled != config.IsInsightEnabledEnum_Enabled {
				common.SmartIDELog.Debug("Application Insights disabled")
				return
			}
			var imageNames []string
			for _, service := range yamlConfig.Workspace.Servcies {
				imageNames = append(imageNames, service.Image)
			}
			appinsight.SetTrack(cmd.Use, Version.TagName, trackEvent, string(workspaceInfo.Mode), strings.Join(imageNames, ","))
		}

		//1. 执行命令
		if workspaceInfo.Mode == workspace.WorkingMode_Local { //1.1. 本地模式
			workspaceInfo, err = start.ExecuteStartCmd(workspaceInfo, isUnforward, func(v string, d common.Docker) {}, executeStartCmdFunc, args, cmd)
			common.CheckError(err)

		} else if workspaceInfo.Mode == workspace.WorkingMode_K8s { //1.2. k8s 模式
			if workspaceInfo.CliRunningEnv == workspace.CliRunningEvnEnum_Server { //1.2.1. cli 在服务端运行
				k8sUtil, err := k8s.NewK8sUtilWithContent(workspaceInfo.K8sInfo.KubeConfigContent,
					workspaceInfo.K8sInfo.Context,
					workspaceInfo.K8sInfo.Namespace)
				common.CheckError(err)

				workspaceInfo, err = start.ExecuteK8s_ServerWS_ServerEnv(cmd, *k8sUtil, workspaceInfo, executeStartCmdFunc)
				common.CheckError(err)

			} else { //1.2.2. cli 在客户端运行
				k8sUtil, err := k8s.NewK8sUtil(workspaceInfo.K8sInfo.KubeConfigFilePath,
					workspaceInfo.K8sInfo.Context,
					workspaceInfo.K8sInfo.Namespace)
				common.CheckError(err)

				if workspaceInfo.CacheEnv == workspace.CacheEnvEnum_Server { //1.2.2.1. 远程工作区 本地加载
					workspaceInfo, err = start.ExecuteK8s_ServerWS_LocalEnv(workspaceInfo, executeStartCmdFunc)
					common.CheckError(err)

				} else { //1.2.2.2. 本地工作区，本地启动
					workspaceInfo, err = start.ExecuteK8s_LocalWS_LocalEnv(cmd, *k8sUtil, workspaceInfo, executeStartCmdFunc)
					common.CheckError(err)
				}

			}

		} else if workspaceInfo.Mode == workspace.WorkingMode_Remote { //1.3. 远程主机 模式

			if workspaceInfo.CliRunningEnv == workspace.CliRunningEvnEnum_Server { //1.3.1. cli 在服务端运行
				disabelGitClone := false
				if workspaceInfo.GitCloneRepoUrl == "" {
					disabelGitClone = true
				}
				workspaceInfo, err = start.ExecuteVmStartCmd(workspaceInfo, isUnforward, executeStartCmdFunc, cmd, args, disabelGitClone)
				common.CheckError(err)

			} else { //1.3.2. cli 在客户端运行
				if workspaceInfo.CacheEnv == workspace.CacheEnvEnum_Server { //1.3.2.1. 远程工作区 本地加载
					workspaceInfo, err = start.ExecuteServerVmStartByClientEnvCmd(workspaceInfo, executeStartCmdFunc)
					common.CheckError(err)

				} else { //1.3.2.2. 本地工作区，本地启动
					disabelGitClone := false
					if workspaceInfo.GitCloneRepoUrl == "" {
						disabelGitClone = true
					}
					workspaceInfo, err = start.ExecuteVmStartCmd(workspaceInfo, isUnforward, executeStartCmdFunc, cmd, args, disabelGitClone)
					common.CheckError(err)
				}

			}

		} else {
			return errors.New("暂不支持当前模式")
		}
		common.CheckError(err)

		//99. 结束
		//99.1. 文本
		common.SmartIDELog.Info(i18nInstance.Start.Info_end)
		if workspaceInfo.ConfigYaml.Workspace.DevContainer.IdeType == config.IdeTypeEnum_SDKOnly {
			common.SmartIDELog.Info("当前IDE环境没有提供WebIDE入口，请使用ssh连接工作区")
		}
		//99.2. 死循环进行驻守，允许端口转发 && 是在本地运行
		if !isUnforward && workspaceInfo.CliRunningEnv == workspace.CliRunningEnvEnum_Client {
			for {
				time.Sleep(time.Millisecond * 300)
			}

		}

		return nil
	},
}

// 运行前
func preRun(cmd *cobra.Command, args []string) error {
	kubeconfig, _ := cmd.Flags().GetString(flag_kubeconfig)
	context, _ := cmd.Flags().GetString(flag_k8s)
	mode, _ := cmd.Flags().GetString("mode")

	// 参数验证
	if mode == "server" {
		if kubeconfig != "" {
			common.SmartIDELog.Importance("server 模式下，--kubeconfig参数无效")
		}
	}
	if kubeconfig != "" && context == "" {
		return errors.New("k8s 参数为空！")
	}

	// 密钥加密显示
	gitPassword, _ := cmd.Flags().GetString(flag_gitpassword)
	if gitPassword != "" {
		common.SmartIDELog.AddEntryptionKeyWithReservePart(gitPassword)
	}
	remotePassword, _ := cmd.Flags().GetString(flag_password)
	if remotePassword != "" {
		common.SmartIDELog.AddEntryptionKey(remotePassword)
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
	flag_kubeconfig  = "kubeconfig"
	flag_gitpassword = "gitpassword"
)

func entryptionKey4Workspace(workspaceInfo workspace.WorkspaceInfo) {
	if workspaceInfo.Remote.Password != "" {
		common.SmartIDELog.AddEntryptionKey(workspaceInfo.Remote.Password)
	}
	if workspaceInfo.K8sInfo.KubeConfigContent != "" {
		common.SmartIDELog.AddEntryptionKeyWithReservePart(workspaceInfo.K8sInfo.KubeConfigContent)
	}
}

// 友好的错误
type FriendlyError struct {
	Err error
}

func (e *FriendlyError) Error() string {
	return e.Err.Error()
}

func init() {

	startCmd.Flags().Int32P("workspaceid", "w", 0, i18nInstance.Remove.Info_flag_workspaceid)
	startCmd.Flags().BoolP("unforward", "", false, "是否禁止端口转发")

	startCmd.Flags().StringP("host", "o", "", i18nInstance.Start.Info_help_flag_host)
	startCmd.Flags().IntP("port", "p", 22, i18nInstance.Start.Info_help_flag_port)
	startCmd.Flags().StringP("username", "u", "", i18nInstance.Start.Info_help_flag_username)
	startCmd.Flags().StringP("password", "t", "", i18nInstance.Start.Info_help_flag_password)

	startCmd.Flags().StringP("repourl", "r", "", i18nInstance.Start.Info_help_flag_repourl)
	startCmd.Flags().StringP("branch", "b", "", i18nInstance.Start.Info_help_flag_branch)
	startCmd.Flags().StringP("gitusername", "", "", "访问当前git库的用户信息")
	startCmd.Flags().StringP("gitpassword", "", "", "对当前git库拥有访问权限的令牌")

	startCmd.Flags().StringP("callback-api-address", "", "", i18nInstance.Start.Info_help_flag_callback_api_address)
	startCmd.Flags().StringVarP(&configYamlFileRelativePath, "filepath", "f", "", i18nInstance.Start.Info_help_flag_filepath)

	startCmd.Flags().StringP("k8s", "k", "", i18nInstance.Start.Info_help_flag_k8s)
	startCmd.Flags().StringP("kubeconfig", "", "", "自定义 kube config 文件的本地路径")
	// startCmd.Flags().StringP("namespace", "n", "", i18nInstance.Start.Info_help_flag_k8s_namespace)
	startCmd.Flags().StringP("serverownerguid", "g", "", i18nInstance.Start.Info_help_flag_ownerguid)
	startCmd.Flags().StringP("addon", "", "", "addon webterminal")
	startCmd.Flags().StringP("type", "T", "", i18nInstance.New.Info_help_flag_type)
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
