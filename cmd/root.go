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
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/leansoftX/smartide-cli/lib/common"
	"github.com/leansoftX/smartide-cli/lib/i18n"
	"github.com/spf13/cobra"

	"github.com/spf13/viper"
)

var cfgFile string
var isDebug bool = false

var instanceI18nMain = i18n.GetInstance().Main

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "smartide",
	Short: instanceI18nMain.Info.Help_short,
	Long:  instanceI18nMain.Info.Help_long, // logo only show in init
	PersistentPreRun: func(cmd *cobra.Command, args []string) {

		// 初始化
		logLevel := ""
		if isDebug {
			logLevel = "debug"
		}
		common.SmartIDELog.InitLogger(logLevel)
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
	rootCmd.Flags().BoolP("help", "h", false, i18n.GetInstance().Help.Info.Help_short)
	rootCmd.PersistentFlags().BoolVarP(&isDebug, "debug", "d", false, i18n.GetInstance().Main.Info.Help_flag_debug)

	// disable completion command
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	// override help command
	rootCmd.SetHelpCommand(helpCmd)

	// usage template
	usage_tempalte := strings.ReplaceAll(i18n.GetInstance().Main.Info.Usage_template, "\\n", "\n")
	rootCmd.SetUsageTemplate(usage_tempalte)

	// custom command
	//rootCmd.AddCommand(initCmd) //屏蔽
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(versionCmd, udpateCmd)
	//rootCmd.AddCommand(restartCmd)
	//rootCmd.AddCommand(vmCmd)
	rootCmd.AddCommand(newCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(hostCmd)
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
