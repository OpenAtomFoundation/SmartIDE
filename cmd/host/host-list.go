package host

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/leansoftX/smartide-cli/cmd/dal"
	"github.com/leansoftX/smartide-cli/lib/common"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var HostListCmd = &cobra.Command{
	Use:   "list",
	Short: i18nInstance.Host.Info_help_list_short,
	Long:  i18nInstance.Host.Info_help_list_long,
	Example: `
  smartide host list`,
	Run: func(cmd *cobra.Command, args []string) {

		list, err := dal.GetRemoteList()
		common.CheckError(err)
		printRemotes(list)
	},
}

// 打印 service 列表
func printRemotes(remotes []dal.RemoteInfo) {
	if len(remotes) <= 0 {
		common.SmartIDELog.Info(i18nInstance.Common.Warn_dal_record_not_exit)
		return
	}
	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
	fmt.Fprintln(w, i18nInstance.Host.Info_host_table_header)
	for _, worksapce := range remotes {

		createTime := worksapce.CreatedTime.Format("2006-01-02 15:04:05")
		line := fmt.Sprintf("%v\t%v\t%v\t%v", worksapce.ID, worksapce.Addr, worksapce.SSHPort, createTime)
		fmt.Fprintln(w, line)
	}
	w.Flush()
}

func init() {

}
