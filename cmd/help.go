package cmd

import (
	"github.com/spf13/cobra"
)

// overwriter help command
var helpCmd = &cobra.Command{
	Use:   "help",
	Short: i18nInstance.Help.Info_help_short,
	Long:  i18nInstance.Help.Info_help_long,
	Run: func(cmd *cobra.Command, args []string) {

	},
}
