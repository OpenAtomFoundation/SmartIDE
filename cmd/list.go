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
var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "",
	Long:    "",
	Example: `  smartide list`,
	Run: func(cmd *cobra.Command, args []string) {

		list, err := dal.GetWorkspaceList()
		common.CheckError(err)

		common.SmartIDELog.Debug("工作区数据查询中...")
		printWorkspaces(list)
		common.SmartIDELog.Debug("查询结束")
	},
}

// 打印 service 列表
func printWorkspaces(workspaces []dal.WorkspaceInfo) {
	if len(workspaces) <= 0 {
		common.SmartIDELog.Info("没有查询到数据！")
		return
	}
	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
	fmt.Fprintln(w, "ID\tName\tMode\tWoring Dir\tConfig File\tHost\tCreate Time")
	for _, worksapce := range workspaces {
		dir := worksapce.WorkingDirectoryPath
		if len(dir) <= 0 {
			dir = "-"
		}
		config := worksapce.ConfigFilePath
		if len(config) <= 0 {
			config = "-"
		}
		createTime := worksapce.CreatedTime.Format("2006-01-02 15:04:05")
		host := "-"
		if (worksapce.Remote != dal.RemoteInfo{}) {
			host = fmt.Sprint(worksapce.Remote.Addr, ":", worksapce.Remote.SSHPort)
		}
		line := fmt.Sprintf("%v\t%v\t%v\t%v\t%v\t%v\t%v", worksapce.ID, worksapce.Name, worksapce.Mode, dir, config, host, createTime)
		fmt.Fprintln(w, line)
	}
	w.Flush()
}

func init() {

}
