package cmd

import (
	"github.com/leansoftX/smartide-cli/cmd/host"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var hostCmd = &cobra.Command{
	Use:   "host",
	Short: "host list | get",
	Long:  "",
	Example: `  smartide host list
  smartide host get <hostid>`,
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func init() {
	hostCmd.AddCommand(host.HostGetCmd)
	hostCmd.AddCommand(host.HostListCmd)

}
