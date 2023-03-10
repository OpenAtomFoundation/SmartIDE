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

package host

import (
	"fmt"
	"time"

	"github.com/leansoftX/smartide-cli/internal/apk/appinsight"
	"github.com/leansoftX/smartide-cli/internal/dal"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var HostRemoveCmd = &cobra.Command{
	Use:     "remove",
	Short:   i18nInstance.Host.Info_help_host_remove_short,
	Long:    i18nInstance.Host.Info_help_host_remove_long,
	Aliases: []string{"rm"},
	Example: `  smartide host remove --hostid <hostid>
	smartide host remove <hostid>`,
	Run: func(cmd *cobra.Command, args []string) {
		common.SmartIDELog.Info(i18nInstance.Host.Remove_start)
		appinsight.SetCliTrack(appinsight.Cli_Remove_Host, args)
		time.Sleep(time.Duration(1) * time.Second) //延迟1s确保发送成功
		hostId := getHostIdFromFlagsAndArgs(cmd, args)
		if hostId <= 0 {
			common.SmartIDELog.WarningF(i18nInstance.Common.Warn_param_is_null, flag_hostid)
			cmd.Help()
			return
		}
		err := dal.RemoveRemote(hostId, "", "")
		common.CheckError(err)

		common.SmartIDELog.Info(fmt.Sprintf(i18nInstance.Host.Info_host_remove_success, hostId))
	},
}

func init() {
	HostRemoveCmd.Flags().Int32P("hostid", "r", 0, i18nInstance.Host.Info_help_flag_hostid)
}
