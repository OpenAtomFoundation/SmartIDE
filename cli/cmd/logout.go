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
	"time"

	"github.com/leansoftX/smartide-cli/internal/apk/appinsight"
	"github.com/leansoftX/smartide-cli/internal/biz/config"
	"github.com/leansoftX/smartide-cli/internal/model"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:     "logout",
	Short:   i18nInstance.Login.Info_help_short,
	Long:    i18nInstance.Login.Info_help_long,
	Example: `  smartide logout`,
	Run: func(cmd *cobra.Command, args []string) {

		appinsight.SetCliTrack(appinsight.Cli_Server_Logout, args)
		time.Sleep(time.Duration(1) * time.Second) //延迟1s确保发送成功
		clearAuths()
		common.SmartIDELog.Info("登录信息已清空！")

	},
}

func clearAuths() {
	c := &config.GlobalSmartIdeConfig
	c.Auths = []model.Auth{}
	c.SaveConfigYaml()
}

func init() {

}
