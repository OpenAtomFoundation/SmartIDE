/*
 * @Author: kenan
 * @Date: 2022-02-10 16:51:36
 * @LastEditors: kenan
 * @LastEditTime: 2022-02-18 16:11:37
 * @FilePath: /smartide-cli/cmd/login.go
 * @Description:
 *
 * Copyright (c) 2022 by kenanlu@leansoftx.com, All Rights Reserved.
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
