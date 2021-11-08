package cmd

import (
	"github.com/leansoftX/smartide-cli/lib/i18n"
	"github.com/spf13/cobra"
)

// overwriter help command
var helpCmd = &cobra.Command{
	Use:   "help",
	Short: i18n.GetInstance().Help.Info.Help_short,
	Long:  i18n.GetInstance().Help.Info.Help_long,
	Run: func(cmd *cobra.Command, args []string) {

	},
}
