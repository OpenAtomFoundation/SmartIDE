package cmd

import (
	"fmt"

	"github.com/leansoftX/smartide-cli/cmd/dal"
	"github.com/leansoftX/smartide-cli/lib/common"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "",
	Long:  "",
	Example: `  smartide get --workspaceid <workspaceid>
  smartide get <workspaceid>`,
	Run: func(cmd *cobra.Command, args []string) {

		workspaceId := getWorkspaceIdFromFlagsAndArgs(cmd, args)
		if workspaceId <= 0 {
			common.SmartIDELog.Warning("参数 workspaceid 为空！")
			cmd.Help()
			return
		}

		workspace, err := dal.GetSingleWorkspace(workspaceId)
		common.CheckError(err)

		print := fmt.Sprintf("工作区信息\nID:\t%v\n名称：\t%v\n运行模式：\t%v\n配置文件路径：\t%v\n工作目录：\t%v\nGit库克隆地址：\t%v\nGit库验证方式：\t%v\n",
			workspace.ID, workspace.Name, workspace.Mode, workspace.ConfigFilePath, workspace.WorkingDirectoryPath, workspace.GitCloneRepoUrl, workspace.GitRepoAuthType)
		common.SmartIDELog.Console(print)

		// 远程连接模式 的信息
		if workspace.Mode == dal.WorkingMode_Remote {
			print = fmt.Sprintf("HOST\nHost ID:\t%v\n地址：\t%v\n验证模式：\t%v\n",
				workspace.Remote.ID, workspace.Remote.Addr, workspace.Remote.AuthType)
			common.SmartIDELog.Console(print)
		}

	},
}

func init() {
	getCmd.Flags().Int32P("workspaceid", "w", 0, "设置后，可以使用本地保存的信息环境信息，直接启动web ide环境")

}
