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

	"github.com/leansoftX/smartide-cli/lib/i18n"
	"github.com/spf13/cobra"

	"github.com/spf13/viper"
)

var cfgFile string

var instanceI18n = i18n.GetInstance().Main

var preMainLongDescription = `
 _____                      _     ___________ _____ 
/  ___|                    | |   |_   _|  _  \  ___|
\ ` + "`" + `--. _ __ ___   __ _ _ __| |_    | | | | | | |__
 ` + "`" + `--. \ '_ ` + "`" + ` _ \ / _` + "`" + ` | '__| __|   | | | | | |  __|
/\__/ / | | | | | (_| | |  | |_   _| |_| |/ /| |___
\____/|_| |_| |_|\__,_|_|   \__|  \___/|___/ \____/ 

`

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "smartide-cli",
	Short: instanceI18n.Info.Help_short,
	Long:  preMainLongDescription + instanceI18n.Info.Help_long,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {

	cobra.CheckErr(rootCmd.Execute())

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

	// disable completion command
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	// override help command
	rootCmd.SetHelpCommand(helpCmd)

	//
	usage_tempalte := strings.ReplaceAll(i18n.GetInstance().Main.Info.Usage_template, "\\n", "\n")
	rootCmd.SetUsageTemplate(usage_tempalte)

	// custom command
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(removeCmd)

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
