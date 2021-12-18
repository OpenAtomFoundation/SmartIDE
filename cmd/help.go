/*
 * @Author: jason chen (jasonchen@leansoftx.com, http://smallidea.cnblogs.com)
 * @Description:
 * @Date: 2021-11
 * @LastEditors:
 * @LastEditTime:
 */
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
