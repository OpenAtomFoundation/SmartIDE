package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/leansoftX/smartide-cli/lib/common"
	"github.com/spf13/cobra"
	"gopkg.in/src-d/go-git.v4"

	"github.com/leansoftX/smartide-cli/cmd/dal"
	"github.com/leansoftX/smartide-cli/cmd/start"
)

var removeCmdFlag struct {
	// 是否仅删除本地的工作区
	IsOnlyRemoveLocalWorkspace bool

	// 是否仅删除远程的容器
	IsOnlyRemoveContainer bool

	// 是否确定删除
	IsContinue bool

	// 是否删除远程主机上的文件夹
	IsRemoveRemoteDirectory bool

	// 删除compose对应的所有镜像
	IsRemoveAllComposeImages bool
}

// removeCmd represents the remove command
var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: i18nInstance.Remove.Info_help_short,
	Long:  i18nInstance.Remove.Info_help_long,
	Example: `
	smartide remove [--workspaceid] {workspaceid} [-y] [-s] [-i] 
	smartide remove [--workspaceid] {workspaceid} [-y] [-f] [-c] [-i]`,
	Run: func(cmd *cobra.Command, args []string) {

		common.SmartIDELog.Info(i18nInstance.Remove.Info_start)

		// 获取 workspace 信息
		common.SmartIDELog.Info(i18nInstance.Main.Info_workspace_loading) //
		workspaceInfo := loadWorkspaceWithDb(cmd, args)

		// 验证
		if removeCmdFlag.IsOnlyRemoveContainer && removeCmdFlag.IsOnlyRemoveLocalWorkspace {
			common.SmartIDELog.Error("参数 workspace 和 container 不能同时存在！")
		}
		if workspaceInfo.Mode == dal.WorkingMode_Local && removeCmdFlag.IsOnlyRemoveContainer {
			common.SmartIDELog.Error("在远程主机模式下，container 参数无效！")
		}

		// 提示 是否确认删除
		if !removeCmdFlag.IsContinue { // 如果设置了参数yes，那么默认就是确认删除
			isEnableRemove := ""
			common.SmartIDELog.Console(i18nInstance.Remove.Info_is_confirm_remove)
			fmt.Scanln(&isEnableRemove)
			if strings.ToLower(isEnableRemove) != "y" {
				return
			}
		}

		// 执行删除动作
		if workspaceInfo.IsNil() {
			common.SmartIDELog.Error(i18nInstance.Main.Err_workspace_none)
		}
		if !removeCmdFlag.IsOnlyRemoveLocalWorkspace { // 仅删除容器的话，就不去远程主机上进行操作
			if workspaceInfo.Mode == dal.WorkingMode_Local {
				removeLocalMode(workspaceInfo)
			} else {
				removeRemoteMode(workspaceInfo)
			}

		}

		// remote workspace in db
		if !removeCmdFlag.IsOnlyRemoveContainer { // 在仅删除容器的模式下，不删除工作区
			common.SmartIDELog.Info("删除工作区数据...")
			err := dal.RemoveWorkspace(workspaceInfo.ID)
			common.CheckError(err)
		}

		// log
		common.SmartIDELog.Info(i18nInstance.Remove.Info_end)
	},
}

// 从flag、args中获取参数信息，然后再去数据库中读取相关数据
func loadWorkspaceWithDb(cmd *cobra.Command, args []string) dal.WorkspaceInfo {
	workspaceInfo := dal.WorkspaceInfo{}
	workspaceId := getWorkspaceIdFromFlagsAndArgs(cmd, args)
	if workspaceId > 0 { // 从db中获取workspace的信息
		var err2 error
		workspaceInfo, err2 = getWorkspaceWithDbAndValid(workspaceId)
		common.CheckError(err2)

	} else { // 当没有workspace id 的时候，只能是本地模式 + 当前目录对应workspace
		// current directory
		pwd, err := os.Getwd()
		common.CheckError(err)

		// git remote url
		gitRepo, err := git.PlainOpen(pwd)
		common.CheckError(err)
		gitRemote, err := gitRepo.Remote("origin")
		common.CheckError(err)
		gitRemmoteUrl := gitRemote.Config().URLs[0]

		workspaceInfo, err = dal.GetSingleWorkspaceByParams(dal.WorkingMode_Local, pwd, gitRemmoteUrl, -1, "")
		common.CheckError(err)
		if workspaceInfo.IsNil() {
			common.SmartIDELog.Error(i18nInstance.Remove.Err_workspace_not_exit)
		}
	}

	return workspaceInfo
}

// 本地删除工作去对应的环境
func removeLocalMode(workspace dal.WorkspaceInfo) error {
	// 校验是否能正常执行docker
	err := start.CheckLocalEnv()
	common.CheckError(err)

	// docker-compose
	composeCmd := exec.Command("docker-compose", "-f", workspace.TempDockerComposeFilePath, "--project-directory", workspace.WorkingDirectoryPath, "down", "-v")
	composeCmd.Stdout = os.Stdout
	composeCmd.Stderr = os.Stderr
	if composeCmdErr := composeCmd.Run(); composeCmdErr != nil {
		common.SmartIDELog.Fatal(composeCmdErr)
	}

	// remove images
	if removeCmdFlag.IsRemoveAllComposeImages {
		removeImagesCmd := exec.Command("docker", "image", "prune", "-af")
		removeImagesCmd.Stdout = os.Stdout
		removeImagesCmd.Stderr = os.Stderr
		if removeImagesCmdErr := removeImagesCmd.Run(); removeImagesCmdErr != nil {
			common.SmartIDELog.Fatal(removeImagesCmdErr)
		}
	}

	return nil
}

// 在远程主机上运行删除命令
func removeRemoteMode(workspace dal.WorkspaceInfo) {

	// ssh 连接
	common.SmartIDELog.Info(i18nInstance.Remove.Info_sshremote_connection_creating)
	var sshRemote common.SSHRemote
	err := sshRemote.Instance(workspace.Remote.Addr, workspace.Remote.SSHPort, workspace.Remote.UserName, workspace.Remote.Password)
	common.CheckError(err)

	// 检查环境
	err = start.CheckRemoveEnv(sshRemote)
	common.CheckError(err)

	// 检查临时文件夹是否存在
	//workspace.ConfigYaml.SaveTempFilesForRemote(sshRemote, workspace.TempDockerCompose, workspace.ProjectName)

	// 删除容器
	common.SmartIDELog.Info(i18nInstance.Remove.Info_docker_removing)
	composeCmdSub := ""
	if removeCmdFlag.IsRemoveAllComposeImages {
		composeCmdSub = "docker image prune -af"
	}
	command := fmt.Sprintf(`docker-compose -f %v --project-directory %v down -v
	 %v`,
		common.FilePahtJoin4Linux(workspace.TempDockerComposeFilePath), common.FilePahtJoin4Linux(workspace.WorkingDirectoryPath), composeCmdSub)
	err = sshRemote.ExecSSHCommandRealTime(command)
	common.CheckError(err)

	// 删除文件夹
	if removeCmdFlag.IsRemoveRemoteDirectory { // 在仅删除workspace的模式下，不删除容器
		common.SmartIDELog.Info(i18nInstance.Remove.Info_project_dir_removing)
		command := fmt.Sprintf("rm -rf %v", common.FilePahtJoin4Linux(workspace.WorkingDirectoryPath))
		err = sshRemote.ExecSSHCommandRealTime(command)
		common.CheckError(err)

	}
}

func init() {
	removeCmd.Flags().Int32P("workspaceid", "w", 0, i18nInstance.Remove.Info_flag_workspaceid)

	removeCmd.Flags().BoolVarP(&removeCmdFlag.IsContinue, "yes", "y", false, i18nInstance.Remove.Info_flag_yes)
	removeCmd.Flags().BoolVarP(&removeCmdFlag.IsRemoveRemoteDirectory, "force", "f", false, i18nInstance.Remove.Info_flag_force)

	removeCmd.Flags().BoolVarP(&removeCmdFlag.IsOnlyRemoveLocalWorkspace, "workspace", "s", false, "仅删除本地的工作区，不涉及远程主机上的容器 和 文件夹")
	removeCmd.Flags().BoolVarP(&removeCmdFlag.IsOnlyRemoveContainer, "container", "c", false, "仅删除远程主机上的容器，不涉及本地的工作区信息")
	removeCmd.Flags().BoolVarP(&removeCmdFlag.IsRemoveAllComposeImages, "image", "i", false, "删除compose文件关联的所有的镜像")

}
