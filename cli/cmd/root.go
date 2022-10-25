/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
/*
 * @Author: jason chen (jasonchen@leansoftx.com, http://smallidea.cnblogs.com)
 * @Description:
 * @Date: 2021-11
 * @LastEditors:
 * @LastEditTime:
 */
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/leansoftX/smartide-cli/internal/apk/appinsight"
	"github.com/leansoftX/smartide-cli/internal/apk/i18n"
	"github.com/leansoftX/smartide-cli/internal/biz/config"
	"github.com/leansoftX/smartide-cli/internal/model"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/spf13/cobra"

	"github.com/spf13/viper"
)

var (
	serverEventID    string = "servereventid"
	serverUserName   string = "serverusername"
	serverToken      string = "servertoken"
	serverUserGuid   string = "serverownerguid"
	serverHost       string = "serverhost"
	serverMode       string = "mode"
	instanceI18nMain        = i18n.GetInstance().Main
	isDebug          bool   = false
	cfgFile          string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "smartide",
	Short: instanceI18nMain.Info_help_short,
	Long:  instanceI18nMain.Info_help_long, // logo only show in init
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		fflags := cmd.Flags()

		// appInsight 是否开启收集smartide
		mode, _ := cmd.Flags().GetString("mode")
		if common.Contains([]string{"pipeline", "server"}, strings.ToLower(mode)) {
			isInsightDisabled, _ := cmd.Flags().GetString("isInsightDisabled")
			if isInsightDisabled == "" {
				common.SmartIDELog.Error("--isInsightDisabled [true|false] 在 --mode server|pipeline 时必须设置")
			} else {
				isInsightDisabled = strings.ToLower(isInsightDisabled)
				if isInsightDisabled == "false" || isInsightDisabled == "no" || isInsightDisabled == "n" || isInsightDisabled == "0" {
					config.GlobalSmartIdeConfig.IsInsightEnabled = config.IsInsightEnabledEnum_Enabled
				} else {
					config.GlobalSmartIdeConfig.IsInsightEnabled = config.IsInsightEnabledEnum_Disabled
				}
				config.GlobalSmartIdeConfig.SaveConfigYaml()
			}
		} else {
			isInsightDisabled, _ := cmd.Flags().GetString("isInsightDisabled")
			if isInsightDisabled != "" && cmd.Flags().Changed("isInsightDisabled") {
				common.SmartIDELog.Importance("isInsightDisabled 参数仅在 mode = server|pipeline 下生效")
			}

			if config.GlobalSmartIdeConfig.IsInsightEnabled == config.IsInsightEnabledEnum_None {
				var isInsightEnabled string
				fmt.Print("SmartIDE会收集部分运行信息用于改进产品，您可以通过 https://smartide.cn/zh/docs/eula 了解我们的信息收集策略或者直接通过开源的源码查看更多细节。请确认是否允许发送（y/n）？")
				fmt.Scanln(&isInsightEnabled)
				isInsightEnabled = strings.ToLower(isInsightEnabled)
				if isInsightEnabled == "y" || isInsightEnabled == "yes" {
					config.GlobalSmartIdeConfig.IsInsightEnabled = config.IsInsightEnabledEnum_Enabled
				} else {
					config.GlobalSmartIdeConfig.IsInsightEnabled = config.IsInsightEnabledEnum_Disabled
				}
				config.GlobalSmartIdeConfig.SaveConfigYaml()
			}
		}

		// appInsight
		if cmd.Use == "start" || cmd.Use == "new" {
		} else {
			if config.GlobalSmartIdeConfig.IsInsightEnabled == config.IsInsightEnabledEnum_Enabled {
				//ai记录
				var trackEvent string
				for _, val := range args {
					trackEvent = trackEvent + " " + val
				}
				appinsight.SetTrack(cmd.Use, Version.TagName, trackEvent, "no", "no")
			} else {
				common.SmartIDELog.Debug("Application Insights disabled")
			}

		}

		// 初始化 log
		logLevel := ""
		if isDebug {
			logLevel = "debug"
		}
		common.SmartIDELog.InitLogger(logLevel)
		common.SmartIDELog.TekEventId, _ = fflags.GetString(serverEventID)
		common.ServerUserName, _ = fflags.GetString(serverUserName)
		common.ServerHost, _ = fflags.GetString(serverHost)
		common.ServerToken, _ = fflags.GetString(serverToken)
		common.ServerUserGuid, _ = fflags.GetString(serverUserGuid)
		common.Mode, _ = fflags.GetString(serverMode)

		// 加密
		servertoken, _ := fflags.GetString("servertoken")
		if servertoken != "" {
			common.SmartIDELog.AddEntryptionKeyWithReservePart(servertoken)
		}

	},
}

var Version SmartVersion

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(smartVersion SmartVersion) {

	Version = smartVersion
	common.SmartIDELog.Error(rootCmd.Execute())

}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	//rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.smartide-cli.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	//rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	// help command short
	rootCmd.Flags().BoolP("help", "h", false, i18n.GetInstance().Help.Info_help_short)
	rootCmd.PersistentFlags().BoolVarP(&isDebug, "debug", "d", false, i18n.GetInstance().Main.Info_help_flag_debug)
	rootCmd.PersistentFlags().StringP("mode", "m", string(model.RuntimeModeEnum_Client), i18n.GetInstance().Main.Info_help_flag_mode)
	rootCmd.PersistentFlags().StringP("isInsightDisabled", "", "true", "在mode = server|pipeline 模式下是否禁用“收集部分运行信息用于改进产品”")

	rootCmd.PersistentFlags().StringP("serverworkspaceid", "", "", i18n.GetInstance().Main.Info_help_flag_server_workspace_id)
	rootCmd.PersistentFlags().StringP("servertoken", "", "", i18n.GetInstance().Main.Info_help_flag_server_token)
	rootCmd.PersistentFlags().StringP("serverusername", "", "", i18n.GetInstance().Main.Info_help_flag_server_username)
	rootCmd.PersistentFlags().StringP("serveruserguid", "", "", i18n.GetInstance().Main.Info_help_flag_server_userguid)
	rootCmd.PersistentFlags().StringP("serverhost", "", "", i18n.GetInstance().Main.Info_help_flag_server_host)
	rootCmd.PersistentFlags().StringP("servereventid", "", "", "trigger event id")
	rootCmd.PersistentFlags().StringP("serverownerguid", "", "", "serverownerguid")

	// disable completion command
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	// override help command
	rootCmd.SetHelpCommand(helpCmd)

	// usage template
	usage_tempalte := strings.ReplaceAll(i18n.GetInstance().Main.Info_Usage_template, "\\n", "\n")
	rootCmd.SetUsageTemplate(usage_tempalte)

	// custom command
	rootCmd.AddCommand(initCmd)

	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(newCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(versionCmd)

	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(hostCmd)

	rootCmd.AddCommand(resetCmd)
	rootCmd.AddCommand(udpateCmd)
	rootCmd.AddCommand(configCmd)

	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(logoutCmd)
	rootCmd.AddCommand(connectCmd)

	rootCmd.AddCommand(k8sCmd)

	// 不允许命令直接按照名称排序
	cobra.EnableCommandSorting = false
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".smartide-cli" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".smartide-cli")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
