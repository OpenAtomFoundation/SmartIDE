/*
 * @Date: 2022-07-22 14:55:03
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-07-22 14:56:20
 * @FilePath: /cli/cmd/init.go
 */

package cmd

import (
	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:     "init",
	Short:   "",
	Long:    "",
	Example: `smartide init`,
	Run: func(cmd *cobra.Command, args []string) {
		common.SmartIDELog.Info("test")
	},
}

func init() {

}
