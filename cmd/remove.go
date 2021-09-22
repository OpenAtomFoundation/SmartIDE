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
	"context"
	"fmt"

	"github.com/leansoftX/smartide-cli/lib/i18n"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
)

var instanceI18nRemove = i18n.GetInstance().Remove

// removeCmd represents the remove command
var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: instanceI18nRemove.Info.Help_short,
	Long:  instanceI18nRemove.Info.Help_long,
	Run: func(cmd *cobra.Command, args []string) {
		var SmartIDEName = "smartide"

		fmt.Println(instanceI18nRemove.Info.Info_start)
		ctx := context.Background()
		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			panic(err)
		}
		cli.ContainerRemove(ctx, SmartIDEName, types.ContainerRemoveOptions{})
		fmt.Println(instanceI18nRemove.Info.Info_end)
	},
}

func init() {
	//rootCmd.AddCommand(removeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// removeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// removeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
