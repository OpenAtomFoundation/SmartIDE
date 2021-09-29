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

	"github.com/spf13/cobra"
)

// initCmd represents the init command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Version", //TODO: 国际化
	Long:  "Version", //TODO: 国际化
	Run: func(cmd *cobra.Command, args []string) {

		fmt.Println(Version.ConvertToJson())
	},
}

type SmartVersion struct {
	VersionNumber   string
	TagName         string `json:"tag_name"`
	BuildNumber     string `json:"build_number"`
	TargetCommitish string `json:"target_commitish"`
	BuildQuqueTime  string `json:"build_ququeTime"`
	Company         string `json:"company"`
}

//
func (smartVersion *SmartVersion) ConvertToJson() string {
	json := fmt.Sprintf(`
版本号：%v
构建号：%v
提交记录：%v
构建时间：%v
公司：%v`,
		smartVersion.VersionNumber, smartVersion.BuildNumber, smartVersion.TargetCommitish, smartVersion.BuildQuqueTime, smartVersion.Company)
	return json
}

func init() {
	//rootCmd.AddCommand(initCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
