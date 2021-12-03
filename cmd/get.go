package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/leansoftX/smartide-cli/cmd/dal"
	"github.com/leansoftX/smartide-cli/lib/common"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "",
	Long:  "",
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
		workspace, err := getWorkspaceWithDbAndValid(workspaceId)
		common.CheckError(err)

		print := fmt.Sprintf(i18nInstance.Get.Info_workspace_detail_template,
			workspace.ID, workspace.Name, workspace.Mode, workspace.ConfigFilePath, workspace.WorkingDirectoryPath, workspace.GitCloneRepoUrl, workspace.GitRepoAuthType)
		common.SmartIDELog.Console(print)

		// 显示全部
		if all, err := cmd.Flags().GetBool("all"); all && err == nil {
			// 端口绑定信息
			if workspace.Extend.IsNotNil() {
				w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
				fmt.Fprintln(w, "Ports:")
				fmt.Fprintln(w, "Service\t| Label\t| Current Local Port\t| Local Port\t| Container Port")
				for _, portInfo := range workspace.Extend.Ports {
					line := fmt.Sprintf("%v\t| %v\t| %v\t| %v\t| %v",
						portInfo.ServiceName, portInfo.LocalPortDesc, portInfo.CurrentLocalPort, portInfo.OriginLocalPort, portInfo.ContainerPort)
					fmt.Fprintln(w, line)
				}
				fmt.Fprintln(w)
				w.Flush()
			}

			// 配置文件
			template := "--配置文件 i18n---------\n%v\n%v\n--Docker-Compose---------\n%v\n%v"
			console := fmt.Sprintf(template, workspace.ConfigFilePath, workspace.ConfigYaml.ToYaml(),
				workspace.TempDockerComposeFilePath, workspace.TempDockerCompose.ToString())
			common.SmartIDELog.Console(console)
		}

		// 远程连接模式 的信息
		if workspace.Mode == dal.WorkingMode_Remote {
			print = fmt.Sprintf(i18nInstance.Get.Info_workspace_host_detail_template,
				workspace.Remote.ID, workspace.Remote.Addr, workspace.Remote.AuthType)
			common.SmartIDELog.Console(print)

		}

	},
}

func init() {
	getCmd.Flags().Int32P("workspaceid", "w", 0, i18nInstance.Get.Info_help_flag_workspaceid)

	getCmd.Flags().BoolP("all", "a", false, "to do i18n")
}
