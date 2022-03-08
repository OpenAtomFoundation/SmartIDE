/*
 * @Author: jason chen (jasonchen@leansoftx.com, http://smallidea.cnblogs.com)
 * @Description:
 * @Date: 2021-11
 * @LastEditors: kenan
 * @LastEditTime: 2022-02-16 10:20:03
 */
package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/docker/docker/client"
	"github.com/leansoftX/smartide-cli/internal/biz/workspace"
	"github.com/leansoftX/smartide-cli/internal/dal"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/leansoftX/smartide-cli/pkg/kubectl"
	"github.com/spf13/cobra"
	"gopkg.in/src-d/go-git.v4"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/leansoftX/smartide-cli/cmd/start"
)

var removeCmdFlag struct {
	// 是否仅删除本地的工作区
	IsOnlyRemoveWorkspaceDataRecord bool

	// 是否仅删除远程的容器
	IsOnlyRemoveContainer bool

	// 是否确定删除
	IsContinue bool

	// 是否删除远程主机上的文件夹
	IsRemoveRemoteDirectory bool

	// 强制删除
	IsForce bool

	// 删除compose对应的所有镜像
	IsRemoveAllComposeImages bool
}

// 删除的模式
type RemoveMode string

const (
	RemoteMode_None                          RemoveMode = "none"
	RemoteMode_OnlyRemoveContainer           RemoveMode = "only_container"
	RemoteMode_OnlyRemoveWorkspaceDataRecord RemoveMode = "only_data_record"
)

// removeCmd represents the remove command
var removeCmd = &cobra.Command{
	Use:     "remove",
	Short:   i18nInstance.Remove.Info_help_short,
	Long:    i18nInstance.Remove.Info_help_long,
	Aliases: []string{"rm"},
	Example: `
	smartide remove [--workspaceid] {workspaceid} [-y] [-w] [-i] [-f] 
	smartide remove [--workspaceid] {workspaceid} [-y] [-s] [-c] [-i] [-f]`,
	Run: func(cmd *cobra.Command, args []string) {

		common.SmartIDELog.Info(i18nInstance.Remove.Info_start)

		// 获取 workspace 信息
		common.SmartIDELog.Info(i18nInstance.Main.Info_workspace_loading) //
		workspaceInfo := loadWorkspaceWithDb(cmd, args)

		// 验证
		if removeCmdFlag.IsOnlyRemoveContainer && removeCmdFlag.IsOnlyRemoveWorkspaceDataRecord { // 仅删除容器 和 仅删除工作区，不能同时存在
			common.SmartIDELog.Error(i18nInstance.Remove.Err_flag_workspace_container)
		}
		if workspaceInfo.Mode == workspace.WorkingMode_Local && removeCmdFlag.IsOnlyRemoveContainer { // 本地模式下，
			common.SmartIDELog.Error(i18nInstance.Remove.Err_flag_container_valid)
		}

		//
		var removeMode RemoveMode = RemoteMode_None
		if removeCmdFlag.IsOnlyRemoveContainer {
			removeMode = RemoteMode_OnlyRemoveContainer
		} else if removeCmdFlag.IsOnlyRemoveWorkspaceDataRecord {
			removeMode = RemoteMode_OnlyRemoveWorkspaceDataRecord
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
		if removeMode == RemoteMode_None || removeMode == RemoteMode_OnlyRemoveContainer { // 仅删除容器的话，就不去远程主机上进行操作
			if workspaceInfo.Mode == workspace.WorkingMode_Local {
				err := removeLocalMode(workspaceInfo, removeCmdFlag.IsRemoveAllComposeImages, removeCmdFlag.IsForce)
				common.CheckError(err)
			} else if workspaceInfo.Mode == workspace.WorkingMode_Remote {
				err := removeRemoteMode(workspaceInfo, removeCmdFlag.IsRemoveAllComposeImages, removeCmdFlag.IsRemoveRemoteDirectory, removeCmdFlag.IsForce)
				common.CheckError(err)
			} else {
				err := removeK8sMode(workspaceInfo, removeCmdFlag.IsRemoveAllComposeImages, removeCmdFlag.IsForce)
				common.CheckError(err)
			}

		}

		// remote workspace in db
		if removeMode == RemoteMode_None || removeMode == RemoteMode_OnlyRemoveWorkspaceDataRecord { // 在仅删除容器的模式下，不删除工作区
			common.SmartIDELog.Info(i18nInstance.Remove.Info_workspace_removing)
			i, err := strconv.Atoi(workspaceInfo.ID)
			common.CheckError(err)
			err = dal.RemoveWorkspace(i)
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
func removeLocalMode(workspaceInfo workspace.WorkspaceInfo, isRemoveAllComposeImages bool, isForce bool) error {
	// 校验是否能正常执行docker
	err := start.CheckLocalEnv()
	common.CheckError(err)

	if !common.IsExit(workspaceInfo.WorkingDirectoryPath) {
		if isForce {
			common.SmartIDELog.Importance(i18nInstance.Remove.Warn_workspace_dir_not_exit)
			// 中断，不再执行后续的步骤
			return nil
		} else {
			return errors.New(i18nInstance.Remove.Err_workspace_dir_not_exit)
		}
	}

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
			if service.Image != "" { // 镜像信息不为空
				//imageNameAndTag := fmt.Sprintf("%v:%v", service.Image.Name, service.Image.Tag)

				force := ""
				if isForce {
					force = "-f"
				}
				removeImagesCmd := exec.Command("docker", "rmi", force, service.Image)
				removeImagesCmd.Stdout = os.Stdout
				removeImagesCmd.Stderr = os.Stderr
				if removeImagesCmdErr := removeImagesCmd.Run(); removeImagesCmdErr != nil {
					common.SmartIDELog.Importance(removeImagesCmdErr.Error())
				} else {
					common.SmartIDELog.InfoF(i18nInstance.Remove.Info_docker_rmi_image_removed, service.Image)
				}

			}
		}
	}

	return nil
}

func removeK8sMode(workspaceInfo workspace.WorkspaceInfo, isRemoveAllComposeImages bool, isForce bool) error {
	if !common.IsExit(workspaceInfo.WorkingDirectoryPath) {
		if isForce {
			common.SmartIDELog.Importance(i18nInstance.Remove.Warn_workspace_dir_not_exit)
			// 中断，不再执行后续的步骤
			return nil
		} else {
			return errors.New(i18nInstance.Remove.Err_workspace_dir_not_exit)
		}
	}

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

	// 删除deployment
	if workspaceInfo.K8sInfo.DeploymentName != "" {
		kubeConfig, err := kubectl.InitKubeConfig(false, workspaceInfo.K8sInfo.Context)
		common.CheckError(err)
		clientset, err := kubectl.NewClientSet(kubeConfig)
		deploymentsClient := clientset.AppsV1().Deployments(workspaceInfo.K8sInfo.Namespace)
		err = deploymentsClient.Delete(context.TODO(), workspaceInfo.K8sInfo.DeploymentName, metav1.DeleteOptions{})
		return err
	}

	return nil
}

// 在远程主机上运行删除命令
func removeRemoteMode(workspaceInfo workspace.WorkspaceInfo, isRemoveAllComposeImages bool, isRemoveRemoteDirectory bool, isForce bool) error {
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
				return nil // 退出当前远程主机上的相关操作
			} else {
				common.CheckError(err)
				return nil
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

		force := ""
		if isForce {
			force = "-f"
		}

		for _, service := range workspaceInfo.TempDockerCompose.Services {
			if service.Image != "" { // 镜像信息不为空
				//imageNameAndTag := fmt.Sprintf("%v:%v", service.Image.Name, service.Image.Tag)
				command := fmt.Sprintf("docker rmi %v %v", force, service.Image)
				_, err = sshRemote.ExeSSHCommand(command)
				if err != nil {
					common.SmartIDELog.Importance(err.Error())
				} else {
					common.SmartIDELog.InfoF(i18nInstance.Remove.Info_docker_rmi_image_removed, service.Image)
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

	return nil
}

// 初始化
func init() {
	removeCmd.Flags().BoolVarP(&removeCmdFlag.IsContinue, "yes", "y", false, i18nInstance.Remove.Info_flag_yes)

	removeCmd.Flags().BoolVarP(&removeCmdFlag.IsOnlyRemoveWorkspaceDataRecord, "workspace", "w", false, i18nInstance.Remove.Info_flag_workspace)
	removeCmd.Flags().BoolVarP(&removeCmdFlag.IsOnlyRemoveContainer, "container", "c", false, i18nInstance.Remove.Info_flag_container)

	removeCmd.Flags().BoolVarP(&removeCmdFlag.IsRemoveRemoteDirectory, "project", "p", false, i18nInstance.Remove.Info_flag_project)
	removeCmd.Flags().BoolVarP(&removeCmdFlag.IsRemoveAllComposeImages, "image", "i", false, i18nInstance.Remove.Info_flag_image)

	removeCmd.Flags().BoolVarP(&removeCmdFlag.IsForce, "force", "f", false, i18nInstance.Remove.Info_flag_force)
}
