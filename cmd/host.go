/*
 * @Author: jason chen (jasonchen@leansoftx.com, http://smallidea.cnblogs.com)
 * @Description:
 * @Date: 2021-11
 * @LastEditors:
 * @LastEditTime:
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
  smartide host get <hostid>`,
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func init() {
	hostCmd.AddCommand(host.HostGetCmd)
	hostCmd.AddCommand(host.HostListCmd)

}
