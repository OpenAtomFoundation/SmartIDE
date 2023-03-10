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

package remove

import (
	"context"
	"errors"
	"os"
	"os/exec"

	"github.com/docker/docker/client"
	"github.com/leansoftX/smartide-cli/cmd/start"
	"github.com/leansoftX/smartide-cli/internal/biz/workspace"
	"github.com/leansoftX/smartide-cli/pkg/common"
)

// 本地删除工作去对应的环境
func RemoveLocal(workspaceInfo workspace.WorkspaceInfo, isRemoveAllComposeImages bool, isForce bool) error {
	// 校验是否能正常执行docker
	err := common.CheckLocalEnv()
	if err != nil {
		return err
	}

	if !common.IsExist(workspaceInfo.WorkingDirectoryPath) {
		if isForce {
			common.SmartIDELog.Importance(i18nInstance.Remove.Warn_workspace_dir_not_exit)
			// 中断，不再执行后续的步骤
			return nil
		} else {
			return errors.New(i18nInstance.Remove.Err_workspace_dir_not_exit)
		}
	}

	// 保存临时文件
	if !common.IsExist(workspaceInfo.TempYamlFileAbsolutePath) || !common.IsExist(workspaceInfo.ConfigFileRelativePath) {
		workspaceInfo.SaveTempFiles()

	}

	// 关联的容器
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	containers := start.GetLocalContainersWithServices(ctx, cli,
		workspaceInfo.WorkingDirectoryPath, workspaceInfo.ConfigYaml.GetServiceNames())
	if len(containers) <= 0 {
		common.SmartIDELog.Importance(i18nInstance.Start.Warn_docker_container_getnone)
	}

	// docker-compose 删除容器
	if len(containers) > 0 {
		common.SmartIDELog.Info(i18nInstance.Remove.Info_docker_removing)
		composeCmd := exec.Command("docker-compose", "-f", workspaceInfo.TempYamlFileAbsolutePath, "--project-directory", workspaceInfo.WorkingDirectoryPath, "down", "-v")
		composeCmd.Stdout = os.Stdout
		composeCmd.Stderr = os.Stderr
		if composeCmdErr := composeCmd.Run(); composeCmdErr != nil {
			//common.SmartIDELog.Fatal(composeCmdErr)
			if composeCmdErr != nil {
				return composeCmdErr
			}

		}
	}

	// remove images
	if isRemoveAllComposeImages {
		common.SmartIDELog.Info(i18nInstance.Remove.Info_docker_rmi_removing)

		for _, service := range workspaceInfo.TempDockerCompose.Services {
			if service.Image != "" { // 镜像信息不为空
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

	//remove config note from .ssh/config file
	workspaceInfo.RemoveSSHConfig()

	return nil
}
