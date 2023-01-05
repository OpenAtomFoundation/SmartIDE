/*
SmartIDE - Dev Containers
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

package host

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/leansoftX/smartide-cli/internal/biz/workspace"
	"github.com/leansoftX/smartide-cli/internal/dal"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var HostListCmd = &cobra.Command{
	Use:     "list",
	Short:   i18nInstance.Host.Info_help_list_short,
	Long:    i18nInstance.Host.Info_help_list_long,
	Aliases: []string{"ls"},
	Example: `
  smartide host list`,
	Run: func(cmd *cobra.Command, args []string) {
		list, err := dal.GetRemoteList()
		common.CheckError(err)
		printRemotes(list)
	},
}

// 打印 service 列表
func printRemotes(remotes []workspace.RemoteInfo) {
	if len(remotes) <= 0 {
		common.SmartIDELog.Info(i18nInstance.Common.Warn_dal_record_not_exit)
		return
	}
	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
	fmt.Fprintln(w, i18nInstance.Host.Info_host_table_header)
	for _, remoteInfo := range remotes {
		entryptionKey4Host(remoteInfo)

		createTime := remoteInfo.CreatedTime.Format("2006-01-02 15:04:05")
		line := fmt.Sprintf("%v\t%v\t%v\t%v", remoteInfo.ID, remoteInfo.Addr, remoteInfo.SSHPort, createTime)
		fmt.Fprintln(w, line)
	}
	w.Flush()
}

func init() {

}
