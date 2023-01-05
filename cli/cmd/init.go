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
