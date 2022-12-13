/*
 * @Author: Bo Dai (daibo@leansoftx.com)
 * @Description:
 * @Date: 2022-07
 * @LastEditors: Bo Dai
 * @LastEditTime: 2022年8月10日 10点34分
 */

package cmd

import (
	initExtended "github.com/leansoftX/smartide-cli/cmd/init"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: i18nInstance.Init.Info_help_short,
	Long:  i18nInstance.Init.Info_help_long,
	Example: ` smartide init
	 smartide init <templatetype> -T {typename}`,
	Run: func(cmd *cobra.Command, args []string) {


		// 环境监测
		err := common.CheckLocalGitEnv() //检测git环境
		common.CheckError(err)
		err = common.CheckLocalEnv() //检测docker环境
		common.CheckError(err)
		initExtended.InitLocalConfig(cmd, args)

	},
}

// 打印 service 列表

func init() {
	initCmd.Flags().StringP("type", "T", "", i18nInstance.New.Info_help_flag_type)
}
