/*
 * @Author: kenan
 * @Date: 2022-02-10 16:51:36
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-05-05 17:37:23
 * @FilePath: /smartide-cli/cmd/login.go
 * @Description:
 *
 * Copyright (c) 2022 by kenanlu@leansoftx.com, All Rights Reserved.
 */

package cmd

import (
	"fmt"
	"strings"

	"github.com/howeyc/gopass"
	"github.com/leansoftX/smartide-cli/internal/biz/config"
	"github.com/leansoftX/smartide-cli/internal/biz/workspace"
	"github.com/leansoftX/smartide-cli/internal/model"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/spf13/cobra"
	"github.com/thedevsaddam/gojsonq"
)

// initCmd represents the init command
var loginCmd = &cobra.Command{
	Use:     "login",
	Short:   i18nInstance.Login.Info_help_short,
	Long:    i18nInstance.Login.Info_help_long,
	Example: `  smartide login <loginurl> --username <username> --password <password>`,
	Aliases: []string{"l"},
	Run: func(cmd *cobra.Command, args []string) {

		//1. 准备参数
		loginUrl := "" // e.g. http://test-dev.smartide.cn:8888/smartide/base/cliLogin
		if len(args) > 0 {
			loginUrl = args[0]
		} else {
			loginUrl = config.GlobalSmartIdeConfig.DefaultLoginUrl
		}
		/* for loginUrl == "" {
			fmt.Print("登录地址：")
			fmt.Scanln(&loginUrl)
		} */
		common.SmartIDELog.Info("login : " + loginUrl)

		fflags := cmd.Flags()
		userName, _ := fflags.GetString(flag_username)
		for userName == "" {
			fmt.Print("用户名：")
			fmt.Scanln(&userName)
			if userName == "" {
				fmt.Print("\r")
			}
		}

		userPassword, _ := fflags.GetString(flag_password)
		if userPassword == "" {
			fmt.Print("密码：")
			passwordBytes, _ := gopass.GetPasswdMasked()
			userPassword = string(passwordBytes)
			if userPassword == "" {
				fmt.Print("\r")
			}
		}
		//TODO: 如果密码错误，可以重新录入再试

		//2. 登录
		cliRunningEnv := workspace.CliRunningEnvEnum_Client
		if value, _ := fflags.GetString("mode"); strings.ToLower(value) == "server" {
			cliRunningEnv = workspace.CliRunningEvnEnum_Server
		}
		errPassword := loginAndSaveToken(loginUrl, userName, userPassword) // 使用密码登录
		if errPassword != nil {
			// 尝试使用token登录
			errToken := loginWithTokenAndSaveToken(loginUrl, userName, userPassword, cliRunningEnv)
			if errToken != nil { // 如果token 也登录不成功，就返回用户密码登录方式的error
				common.CheckError(errPassword)
			}
		}

		common.SmartIDELog.Info(loginUrl + " 登录成功！")
	},
}

func loginWithTokenAndSaveToken(loginUrl, userName, token string, cliRunningEnv workspace.CliRunningEvnEnum) error {

	// 请求
	auth := model.Auth{UserName: userName, Token: token, LoginUrl: loginUrl}
	_, err := workspace.GetServerWorkspaceList(auth, cliRunningEnv)
	if err != nil {
		return err
	}

	saveToken(loginUrl, userName, token)

	return nil
}

// 登录
func loginAndSaveToken(loginUrl, userName, userPassword string) error {
	url := fmt.Sprint(loginUrl, "/api/smartide/base/cliLogin")
	response, err := common.PostJson(url, map[string]interface{}{"username": userName, "password": userPassword}, map[string]string{"Content-Type": "application/json"})
	if err != nil {
		return err
	}
	code := gojsonq.New().JSONString(response).Find("code").(float64)
	if code != 0 {
		msg := gojsonq.New().JSONString(response).Find("msg")
		return fmt.Errorf("login fail %q", msg)
	} else {
		token := gojsonq.New().JSONString(response).Find("data.token")
		saveToken(loginUrl, userName, token)
	}

	return nil
}

func saveToken(loginUrl, userName string, token interface{}) {
	c := &config.GlobalSmartIdeConfig
	if !userIsExit(c.Auths, userName, loginUrl) {
		for i := range c.Auths {
			c.Auths[i].CurrentUse = false
		}
		c.Auths = append(c.Auths, model.Auth{
			UserName:   userName,
			Token:      token,
			LoginUrl:   loginUrl,
			CurrentUse: true,
		})
	} else {
		for i, a := range c.Auths {
			if a.UserName == userName && a.LoginUrl == loginUrl {
				c.Auths[i].Token = token
				c.Auths[i].CurrentUse = true
			} else {
				c.Auths[i].CurrentUse = false
			}
		}
	}
	c.SaveConfigYaml()
}

func userIsExit(auths []model.Auth, username string, loginurl string) bool {
	for _, a := range auths {
		if a.UserName == username && a.LoginUrl == loginurl {
			return true
		}
	}
	return false
}

func init() {
	loginCmd.Flags().StringP("username", "u", "", i18nInstance.Login.Info_help_flag_username)
	loginCmd.Flags().StringP("password", "t", "", i18nInstance.Login.Info_help_flag_password)
}
