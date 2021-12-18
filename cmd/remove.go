/*
 * @Author: jason chen (jasonchen@leansoftx.com, http://smallidea.cnblogs.com)
 * @Description:
 * @Date: 2021-11
 * @LastEditors:
 * @LastEditTime:
 */
package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/docker/docker/client"
	"github.com/leansoftX/smartide-cli/internal/biz/workspace"
	"github.com/leansoftX/smartide-cli/internal/dal"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/leansoftX/smartide-cli/pkg/docker/compose"

	"github.com/spf13/cobra"
	"gopkg.in/src-d/go-git.v4"

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
			common.SmartIDELog.Error(i18nInstance.Remove.Err_flag_workspace_container)
		}
		if workspaceInfo.Mode == workspace.WorkingMode_Local && removeCmdFlag.IsOnlyRemoveContainer {
			common.SmartIDELog.Error(i18nInstance.Remove.Err_flag_container_valid)
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
			if workspaceInfo.Mode == workspace.WorkingMode_Local {
				removeLocalMode(workspaceInfo, removeCmdFlag.IsRemoveAllComposeImages)
			} else {
				removeRemoteMode(workspaceInfo, removeCmdFlag.IsRemoveAllComposeImages, removeCmdFlag.IsRemoveRemoteDirectory)
			}

		}

		// remote workspace in db
		if !removeCmdFlag.IsOnlyRemoveContainer { // 在仅删除容器的模式下，不删除工作区
			common.SmartIDELog.Info(i18nInstance.Remove.Info_workspace_removing)

			err := dal.RemoveWorkspace(workspaceInfo.ID)
			common.CheckError(err)
		}

		// log
		common.SmartIDELog.Info(i18nInstance.Remove.Info_end)
	},
}

// 从flag、args中获取参数信息，然后再去数据库中读取相关数据
func loadWorkspaceWithDb(cmd *cobra.Command, args []string) workspace.WorkspaceInfo {
	workspaceInfo := workspace.WorkspaceInfo{}
	workspaceId := getWorkspaceIdFromFlagsAndArgs(cmd, args)
	if workspaceId > 0 { // 从db中获取workspace的信息
		var err2 error
		workspaceInfo, err2 = getWorkspaceWithDbAndValid(workspaceId)
		common.CheckError(err2)

		// 旧版本会导致这个问题
		if workspaceInfo.ConfigYaml.IsNil() {
			msg := fmt.Sprintf(i18nInstance.Main.Err_workspace_version_old, workspaceId)
			common.SmartIDELog.Error(msg)
		}

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

		workspaceInfo, err = dal.GetSingleWorkspaceByParams(workspace.WorkingMode_Local, pwd, gitRemmoteUrl, -1, "")
		common.CheckError(err)
		if workspaceInfo.IsNil() {
			common.SmartIDELog.Error(i18nInstance.Remove.Err_workspace_not_exit)
		}
	}

	return workspaceInfo
}

// 本地删除工作去对应的环境
func removeLocalMode(workspaceInfo workspace.WorkspaceInfo, isRemoveAllComposeImages bool) error {
	// 校验是否能正常执行docker
	err := start.CheckLocalEnv()
	common.CheckError(err)

	// 保存临时文件
	if !common.IsExit(workspaceInfo.TempDockerComposeFilePath) || !common.IsExit(workspaceInfo.ConfigFilePath) {
		workspaceInfo.SaveTempFiles()

	}

	// 关联的容器
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	common.CheckError(err)
	containers := start.GetLocalContainersWithServices(ctx, cli, workspaceInfo.ConfigYaml.GetServiceNames())
	if len(containers) <= 0 {
		common.SmartIDELog.Importance(i18nInstance.Start.Warn_docker_container_getnone)
	}

	// docker-compose 删除容器
	if len(containers) > 0 {
		common.SmartIDELog.Info(i18nInstance.Remove.Info_docker_removing)
		composeCmd := exec.Command("docker-compose", "-f", workspaceInfo.TempDockerComposeFilePath, "--project-directory", workspaceInfo.WorkingDirectoryPath, "down", "-v")
		composeCmd.Stdout = os.Stdout
		composeCmd.Stderr = os.Stderr
		if composeCmdErr := composeCmd.Run(); composeCmdErr != nil {
			common.SmartIDELog.Fatal(composeCmdErr)
		}
	}

	// remove images
	if isRemoveAllComposeImages {
		common.SmartIDELog.Info(i18nInstance.Remove.Info_docker_rmi_removing)

		for _, service := range workspaceInfo.TempDockerCompose.Services {
			if (service.Image != compose.Image{}) && service.Image.Name != "" { // 镜像信息不为空
				imageNameAndTag := fmt.Sprintf("%v:%v", service.Image.Name, service.Image.Tag)

				removeImagesCmd := exec.Command("docker", "rmi", imageNameAndTag)
				removeImagesCmd.Stdout = os.Stdout
				removeImagesCmd.Stderr = os.Stderr
				if removeImagesCmdErr := removeImagesCmd.Run(); removeImagesCmdErr != nil {
					common.SmartIDELog.Importance(removeImagesCmdErr.Error())
				} else {
					common.SmartIDELog.InfoF(i18nInstance.Remove.Info_docker_rmi_image_removed, imageNameAndTag)
				}

			}
		}
	}

	return nil
}

// 在远程主机上运行删除命令
func removeRemoteMode(workspaceInfo workspace.WorkspaceInfo, isRemoveAllComposeImages bool, isRemoveRemoteDirectory bool) {
	// ssh 连接
	common.SmartIDELog.Info(i18nInstance.Remove.Info_sshremote_connection_creating)
	sshRemote, err := common.NewSSHRemote(workspaceInfo.Remote.Addr, workspaceInfo.Remote.SSHPort, workspaceInfo.Remote.UserName, workspaceInfo.Remote.Password)
	common.CheckError(err)

	// 检查环境
	err = start.CheckRemoveEnv(sshRemote)
	if err != nil {
		if strings.Contains(err.Error(), "i/o timeout") || strings.Contains(err.Error(), "connect: connection refused") {
			isSkip := ""
			common.SmartIDELog.Importance(err.Error())
			common.SmartIDELog.Console(i18nInstance.Remove.Info_ssh_timeout_confirm_skip)
			fmt.Scanln(&isSkip)
			if strings.ToLower(isSkip) != "y" {
				return // 退出当前远程主机上的相关操作
			} else {
				common.CheckError(err)
				return
			}
		} else {
			common.CheckError(err)
		}
	}

	// 项目文件夹是否存在
	if !sshRemote.IsCloned(workspaceInfo.WorkingDirectoryPath) {
		sshRemote.GitClone(workspaceInfo.GitCloneRepoUrl, workspaceInfo.WorkingDirectoryPath)
		isRemoveRemoteDirectory = true // 创建后就删掉

	}

	// 检查临时文件夹是否存在
	if !sshRemote.IsExit(workspaceInfo.TempDockerComposeFilePath) || !sshRemote.IsExit(workspaceInfo.ConfigYaml.GetConfigYamlFilePath()) {
		workspaceInfo.SaveTempFilesForRemote(sshRemote)

	}

	// 容器列表
	containers, err := start.GetRemoteContainersWithServices(sshRemote, workspaceInfo.ConfigYaml.GetServiceNames())
	common.CheckError(err)
	if len(containers) <= 0 {
		common.SmartIDELog.Importance(i18nInstance.Start.Warn_docker_container_getnone)
	}

	// 远程主机上执行 docker-compose 删除容器
	if len(containers) > 0 {
		common.SmartIDELog.Info(i18nInstance.Remove.Info_docker_removing)
		command := fmt.Sprintf(`docker-compose -f %v --project-directory %v down -v`,
			common.FilePahtJoin4Linux(workspaceInfo.TempDockerComposeFilePath), common.FilePahtJoin4Linux(workspaceInfo.WorkingDirectoryPath))
		err = sshRemote.ExecSSHCommandRealTime(command)
		common.CheckError(err)
	}

	// 删除对应的镜像
	if isRemoveAllComposeImages {
		common.SmartIDELog.Info(i18nInstance.Remove.Info_docker_rmi_removing)

		for _, service := range workspaceInfo.TempDockerCompose.Services {
			if (service.Image != compose.Image{}) && service.Image.Name != "" { // 镜像信息不为空
				imageNameAndTag := fmt.Sprintf("%v:%v", service.Image.Name, service.Image.Tag)
				_, err = sshRemote.ExeSSHCommand("docker rmi " + imageNameAndTag)
				if err != nil {
					common.SmartIDELog.Importance(err.Error())
				} else {
					common.SmartIDELog.InfoF(i18nInstance.Remove.Info_docker_rmi_image_removed, imageNameAndTag)
				}
			}
		}
	}

	// 删除文件夹
	if isRemoveRemoteDirectory { // 在仅删除workspace的模式下，不删除容器
		common.SmartIDELog.Info(i18nInstance.Remove.Info_project_dir_removing)
		workingDirectoryPath := common.FilePahtJoin4Linux(workspaceInfo.WorkingDirectoryPath)
		command := fmt.Sprintf("sudo rm -rf %v", workingDirectoryPath)
		err = sshRemote.ExecSSHCommandRealTime(command)
		common.CheckError(err)

		// 成功后的提示
		common.SmartIDELog.InfoF(i18nInstance.Remove.Info_project_dir_removed, workingDirectoryPath)
	}

}

// 初始化
func init() {
	removeCmd.Flags().Int32P("workspaceid", "w", 0, i18nInstance.Remove.Info_flag_workspaceid)

	removeCmd.Flags().BoolVarP(&removeCmdFlag.IsContinue, "yes", "y", false, i18nInstance.Remove.Info_flag_yes)
	removeCmd.Flags().BoolVarP(&removeCmdFlag.IsRemoveRemoteDirectory, "force", "f", false, i18nInstance.Remove.Info_flag_force)
	removeCmd.Flags().BoolVarP(&removeCmdFlag.IsOnlyRemoveLocalWorkspace, "workspace", "s", false, i18nInstance.Remove.Info_flag_workspace)
	removeCmd.Flags().BoolVarP(&removeCmdFlag.IsOnlyRemoveContainer, "container", "c", false, i18nInstance.Remove.Info_flag_container)
	removeCmd.Flags().BoolVarP(&removeCmdFlag.IsRemoveAllComposeImages, "image", "i", false, i18nInstance.Remove.Info_flag_image)

}
