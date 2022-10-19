/*
 * @Date: 2022-04-20 10:46:40
 * @LastEditors: kenan
 * @LastEditTime: 2022-10-19 12:40:46
 * @FilePath: /cli/cmd/new/newVm.go
 */

package new

import (
	"errors"
	"fmt"
	"strings"

	"github.com/leansoftX/smartide-cli/internal/biz/config"
	"github.com/leansoftX/smartide-cli/internal/biz/workspace"
	"github.com/leansoftX/smartide-cli/internal/model"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/spf13/cobra"

	smartideServer "github.com/leansoftX/smartide-cli/cmd/server"
	"github.com/leansoftX/smartide-cli/cmd/start"
)

func VmNew(cmd *cobra.Command, args []string, workspaceInfo workspace.WorkspaceInfo,
	yamlExecuteFun func(yamlConfig config.SmartIdeConfig)) {

	mode, _ := cmd.Flags().GetString("mode")
	isModeServer := strings.ToLower(mode) == "server"
	// 错误反馈
	serverFeedback := func(err error) {
		if !isModeServer {
			return
		}
		if err != nil {
			smartideServer.Feedback_Finish(smartideServer.FeedbackCommandEnum_New, cmd, false, nil, workspaceInfo, err.Error(), "")
			common.CheckError(err)
		}
	}

	if apiHost, _ := cmd.Flags().GetString(smartideServer.Flags_ServerHost); apiHost != "" {
		wsURL := fmt.Sprint(strings.ReplaceAll(strings.ReplaceAll(apiHost, "https", "ws"), "http", "ws"), "/ws/smartide/ws")
		common.WebsocketStart(wsURL)
		token, _ := cmd.Flags().GetString(smartideServer.Flags_ServerToken)
		if token != "" {
			if workspaceIdStr, _ := cmd.Flags().GetString(smartideServer.Flags_ServerWorkspaceid); workspaceIdStr != "" {
				if no, _ := workspace.GetWorkspaceNo(workspaceIdStr, token, apiHost); no != "" {
					if pid, err := workspace.GetParentId(no, workspace.ActionEnum_Workspace_Start, token, apiHost); err == nil && pid > 0 {
						common.SmartIDELog.Ws_id = no
						common.SmartIDELog.ParentId = pid
					}
				}
			}

		}
	}

	//0. 连接到远程主机
	msg := fmt.Sprintf(" %v@%v:%v ...", workspaceInfo.Remote.UserName, workspaceInfo.Remote.Addr, workspaceInfo.Remote.SSHPort)
	common.SmartIDELog.Info(i18nInstance.VmStart.Info_connect_remote + msg)
	idRsa := ""
	if workspaceInfo.Remote.Password == "" && common.Mode == "server" {
		_, idRsa = common.GetSSHkeyPolicyIdRsa(common.SmartIDELog.Ws_id, common.ServerHost, common.ServerToken, common.ServerUserGuid)
	}
	sshRemote, err := common.NewSSHRemote(workspaceInfo.Remote.Addr, workspaceInfo.Remote.SSHPort, workspaceInfo.Remote.UserName, workspaceInfo.Remote.Password, idRsa)
	common.CheckErrorFunc(err, serverFeedback)

	//1. 检查远程主机是否有docker、docker-compose、git
	err = sshRemote.CheckRemoteEnv()
	common.CheckErrorFunc(err, serverFeedback)

	// 获取command中的配置
	selectedTemplateSettings, err := getTemplateSetting(cmd, args)
	common.CheckError(err)
	if selectedTemplateSettings == nil { // 未指定模板类型的时候，提示用户后退出
		return // 退出
	}

	// 文件夹检查
	workspaceDirName, _ := cmd.Flags().GetString("workspacename") // 指定的项目名称
	if workspaceDirName == "" {
		common.CheckErrorFunc(errors.New("参数 workspacename 不能为空！"), serverFeedback)
	}
	err = checkRemoteDir(sshRemote, workspaceInfo.WorkingDirectoryPath, cmd)
	common.CheckErrorFunc(err, serverFeedback)

	// 复制 template 到远程主机的文件夹中
	if selectedTemplateSettings.SubType == "" {
		selectedTemplateSettings.SubType = "_default"
	}
	projectDir := common.FilePahtJoin4Linux("~", model.CONST_REMOTE_REPO_ROOT, workspaceDirName)
	err = gitCloneTemplateRepo4Remote(sshRemote, projectDir, config.GlobalSmartIdeConfig.TemplateRepo,
		selectedTemplateSettings.TypeName, selectedTemplateSettings.SubType)
	common.CheckErrorFunc(err, serverFeedback)

	// 执行vm start命令
	isUnforward, _ := cmd.Flags().GetBool("unforward")
	start.ExecuteVmStartCmd(workspaceInfo, isUnforward, yamlExecuteFun, cmd, args, true)
}

// 在服务器上使用git下载制定的template文件，完成后删除.git文件
func gitCloneTemplateRepo4Remote(sshRemote common.SSHRemote, projectDir string, templateGitCloneUrl string, baseType string, subType string) error {

	// git
	tempDirPath := common.FilePahtJoin4Linux("~", ".ide", "template")
	command := fmt.Sprintf(`
cd %v 
[[ -d .git ]] && (git checkout . && git clean -xdf && git pull) || git clone %v %v
`, tempDirPath, templateGitCloneUrl, tempDirPath)
	err := sshRemote.ExecSSHCommandRealTime(command)
	if err != nil {
		common.SmartIDELog.Importance(err.Error())
		if strings.Contains(err.Error(), "You have not concluded your merge") {
			common.SmartIDELog.Debug("re-pull")
			command = fmt.Sprintf(`cd %v 
git fetch --all && git reset --hard origin/master && git fetch && git pull`,
				tempDirPath)
			err = sshRemote.ExecSSHCommandRealTime(command)
		}

		return err
	}

	// 项目目录如果不存在就创建
	templateDirPath := strings.Join([]string{tempDirPath, baseType, subType, "."}, "/")
	commandCopy := fmt.Sprintf(`
	[[ -d %v ]] && echo '%v directory exist' || mkdir -p %v
	cp -r %v %v 
`, projectDir, projectDir, projectDir, templateDirPath, projectDir)
	err = sshRemote.ExecSSHCommandRealTime(commandCopy)

	return err
}

// 检查远程文件夹
func checkRemoteDir(sshRemote common.SSHRemote, projectDirPath string, cmd *cobra.Command) error {
	// 检测指定的文件夹是否有.ide.yaml，有了返回
	ideFilePath := common.FilePahtJoin4Linux(projectDirPath, ".ide/.ide.yaml")
	hasIdeConfigYaml := sshRemote.IsFileExist(ideFilePath)
	if hasIdeConfigYaml {
		common.SmartIDELog.Info("当前目录已经完成初始化，无须再次进行！")
	}

	//
	if !sshRemote.IsDirExist(projectDirPath) { // 目录如果不存在就创建
		sshRemote.CheckAndCreateDir(projectDirPath)
	} else {
		// 检测并阻断
		isEmpty := sshRemote.IsDirEmpty(projectDirPath) // 检测当前文件夹是否为空
		if !isEmpty {
			isContinue, _ := cmd.Flags().GetBool("yes")
			if !isContinue { // 如果没有设置yes，那么就要给出提示
				var s string
				common.SmartIDELog.Importance(i18nInstance.New.Info_noempty_is_comfirm)
				fmt.Scanln(&s)
				if s != "y" {
					return errors.New("user exit")
				}
			} else {
				common.SmartIDELog.Importance("当前文件夹不为空，当前文件夹内数据将被重置。")
				sshRemote.Clear(projectDirPath)
			}
		}
	}

	return nil
}
