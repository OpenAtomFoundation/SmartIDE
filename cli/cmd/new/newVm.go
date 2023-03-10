/*
SmartIDE - CLI
Copyright (C) 2023 leansoftX.com

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package new

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/leansoftX/smartide-cli/internal/biz/config"
	"github.com/leansoftX/smartide-cli/internal/biz/workspace"
	"github.com/leansoftX/smartide-cli/internal/model"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/spf13/cobra"

	//smartideServer "github.com/leansoftX/smartide-cli/cmd/server"
	"github.com/leansoftX/smartide-cli/cmd/start"
)

func VmNew(cmd *cobra.Command, args []string, workspaceInfo workspace.WorkspaceInfo,
	yamlExecuteFun func(yamlConfig config.SmartIdeConfig, workspaceInfo workspace.WorkspaceInfo, cmdtype, userguid, workspaceid string)) {

	// 错误反馈
	serverFeedback := preRun(cmd, workspaceInfo, workspace.ActionEnum_Workspace_Start, nil)

	//1. 连接到远程主机
	//1.1.
	msg := fmt.Sprintf(" %v@%v:%v ...", workspaceInfo.Remote.UserName, workspaceInfo.Remote.Addr, workspaceInfo.Remote.SSHPort)
	common.SmartIDELog.Info(i18nInstance.VmStart.Info_connect_remote + msg)
	sshRemote, err := common.NewSSHRemote(workspaceInfo.Remote.Addr, workspaceInfo.Remote.SSHPort, workspaceInfo.Remote.UserName, workspaceInfo.Remote.Password, workspaceInfo.Remote.SSHKey)
	common.CheckErrorFunc(err, serverFeedback)

	//1.2. 检查远程主机是否有docker、docker-compose、git
	err = sshRemote.CheckRemoteEnv()
	common.CheckErrorFunc(err, serverFeedback)

	//2. 文件夹拷贝
	//2.1. 项目文件夹检查
	workspaceDirName := workspaceInfo.Name
	err = checkRemoteDir(sshRemote, workspaceInfo.WorkingDirectoryPath, cmd)
	common.CheckErrorFunc(err, serverFeedback)
	projectRemoteAbsoluteDir := common.FilePahtJoin4Linux("~", model.CONST_REMOTE_REPO_ROOT, workspaceDirName)

	//2.2. git clone 项目文件夹
	if workspaceInfo.GitCloneRepoUrl != "" {
		err = start.GitCloneAndCheckoutBranch(sshRemote, workspaceInfo, cmd)
		common.CheckErrorFunc(err, serverFeedback)
	}
	//2.2.1. 检查是否包含配置文件，如果有就报错
	isExistConfigFile := sshRemote.IsFileExist(filepath.Join(projectRemoteAbsoluteDir, workspaceInfo.ConfigFileRelativePath))
	if isExistConfigFile {
		errMsg := fmt.Sprintf("模板中已经包含相同配置文件 %v", workspaceInfo.ConfigFileRelativePath)
		feedbackErr := model.CreateFeedbackError(errMsg, false)
		common.CheckErrorFunc(&feedbackErr, serverFeedback)
	}

	//2.3. 复制 template 到远程主机的文件夹中
	err = gitCloneTemplateRepo4Remote(sshRemote, projectRemoteAbsoluteDir,
		config.GlobalSmartIdeConfig.TemplateActualRepoUrl,
		workspaceInfo.SelectedTemplate.TypeName, workspaceInfo.SelectedTemplate.SubType)
	common.CheckErrorFunc(err, serverFeedback)
	//2.3.1. 检查是否包含配置文件，如果没有就报错
	isExistConfigFile = sshRemote.IsFileExist(filepath.Join(projectRemoteAbsoluteDir, workspaceInfo.ConfigFileRelativePath))
	if !isExistConfigFile {
		errMsg := fmt.Sprintf("模板库中没有找到配置文件 %v", workspaceInfo.ConfigFileRelativePath)
		feedbackErr := model.CreateFeedbackError(errMsg, false)
		common.CheckErrorFunc(&feedbackErr, serverFeedback)
	}

	//9. 执行vm start命令
	isUnforward, _ := cmd.Flags().GetBool("unforward")
	start.ExecuteVmStartCmd(workspaceInfo, isUnforward, yamlExecuteFun, cmd, args, true)
}

// 在服务器上使用git下载制定的template文件，完成后删除.git文件
func gitCloneTemplateRepo4Remote(sshRemote common.SSHRemote,
	projectDir string, templateGitCloneUrl string, baseType string, subType string) error {

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
