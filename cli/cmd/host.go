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
package cmd

import (
	"github.com/leansoftX/smartide-cli/cmd/host"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var hostCmd = &cobra.Command{
	Use:   "host",
	Short: i18nInstance.Host.Info_help_short,
	Long:  i18nInstance.Host.Info_help_long,
	Example: `  smartide host list
  smartide host get <hostid>
  smartide host add <host> --username <username> --password <password> --port <port>
  smartide host remove <hostid>`,
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func init() {
	hostCmd.AddCommand(host.HostGetCmd)
	hostCmd.AddCommand(host.HostListCmd)
	hostCmd.AddCommand(host.HostAddCmd)
	hostCmd.AddCommand(host.HostRemoveCmd)
}
