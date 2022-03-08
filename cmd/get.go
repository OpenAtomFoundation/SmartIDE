/*
 * @Author: jason chen (jasonchen@leansoftx.com, http://smallidea.cnblogs.com)
 * @Description:
 * @Date: 2021-11
 * @LastEditors:
 * @LastEditTime:
 */
package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/leansoftX/smartide-cli/internal/biz/workspace"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/spf13/cobra"
)

//
var getCmd = &cobra.Command{
	Use:   "get",
	Short: i18nInstance.Get.Info_help_short,
	Long:  i18nInstance.Get.Info_help_long,
	Example: `  smartide get --workspaceid {workspaceid}
  smartide get {workspaceid}`,
	Run: func(cmd *cobra.Command, args []string) {

		workspaceId := getWorkspaceIdFromFlagsAndArgs(cmd, args)
		if workspaceId <= 0 {
			common.SmartIDELog.Warning(i18nInstance.Get.Warn_flag_workspaceid_none)
			cmd.Help()
			return
		}

		// 从数据库中查询
		workspaceInfo, err := getWorkspaceWithDbAndValid(workspaceId)
		common.CheckError(err)

		// 打印
		print := fmt.Sprintf(i18nInstance.Get.Info_workspace_detail_template,
			workspaceInfo.ID, workspaceInfo.Name, workspaceInfo.Mode, workspaceInfo.ConfigFilePath, workspaceInfo.WorkingDirectoryPath,
			workspaceInfo.GitCloneRepoUrl, workspaceInfo.GitRepoAuthType)
		common.SmartIDELog.Console(print)

		// 显示全部
		if all, err := cmd.Flags().GetBool("all"); all && err == nil {
			// 端口绑定信息
			if workspaceInfo.Extend.IsNotNil() {
				w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
				fmt.Fprintln(w, "Ports:")
				fmt.Fprintln(w, "Service\t| Label\t| Current Local Port\t| Local Port\t| Container Port")
				for _, portInfo := range workspaceInfo.Extend.Ports {
					line := fmt.Sprintf("%v\t| %v\t| %v\t| %v\t| %v",
						portInfo.ServiceName, portInfo.HostPortDesc, portInfo.CurrentHostPort, portInfo.OriginHostPort, portInfo.ContainerPort)
					fmt.Fprintln(w, line)
				}
				fmt.Fprintln(w)
				w.Flush()
			}

			// 配置文件
			configYamlStr, err := workspaceInfo.ConfigYaml.ToYaml()
			common.CheckError(err)
			console := fmt.Sprintf("-- Configration file ---------\n%v\n%v", workspaceInfo.TempDockerComposeFilePath, configYamlStr)
			common.SmartIDELog.Console(console)

			// 链接的 docker-compose
			if workspaceInfo.ConfigYaml.IsLinkDockerComposeFile() {
				linkDockerYamlStr, err := workspaceInfo.LinkDockerCompose.ToYaml()
				common.CheckError(err)
				console = fmt.Sprintf("-- link Docker-Compose ---------\n%v\n%v", workspaceInfo.ConfigYaml.Workspace.DockerComposeFile, linkDockerYamlStr)
				common.SmartIDELog.Console(console)
			}

			// 生成的docker-compose
			dockerYamlStr, err := workspaceInfo.TempDockerCompose.ToYaml()
			common.CheckError(err)
			console = fmt.Sprintf("-- Docker-Compose ---------\n%v\n%v", workspaceInfo.TempDockerComposeFilePath, dockerYamlStr)
			common.SmartIDELog.Console(console)
		}

		// 远程连接模式 的信息
		if workspaceInfo.Mode == workspace.WorkingMode_Remote {
			print = fmt.Sprintf(i18nInstance.Get.Info_workspace_host_detail_template,
				workspaceInfo.Remote.ID, workspaceInfo.Remote.Addr, workspaceInfo.Remote.AuthType)
			common.SmartIDELog.Console(print)

		}

	},
}

func init() {
	getCmd.Flags().Int32P("workspaceid", "w", 0, i18nInstance.Get.Info_help_flag_workspaceid)
	getCmd.Flags().BoolP("all", "a", false, i18nInstance.Get.Info_help_flag_all)

}
