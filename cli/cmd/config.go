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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/leansoftX/smartide-cli/internal/biz/config"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: i18nInstance.Config.Info_help_short,
	Long:  i18nInstance.Config.Info_help_long,
	Example: `  smartide config list
  smartide config set template-repo=<repourl>
  smartide config set images-registry=<registryurl>`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return nil
		} else if len(args) == 1 && (args[0] == "list" || args[0] == "ls") {

			var configStruct config.GlobalConfig
			configStruct.LoadConfigYaml()

			bytes, err := json.MarshalIndent(configStruct, "", "    ")
			common.CheckError(err)
			str := string(bytes)
			common.SmartIDELog.Console(str)

		} else if len(args) == 2 && args[0] == "set" {
			if args[1] != "" {
				var configStruct config.GlobalConfig
				configStruct.LoadConfigYaml()
				paramArry := strings.Split(args[1], "=")
				if len(paramArry) != 2 {
					common.SmartIDELog.Error(i18nInstance.Config.Err_set_config)
					return nil
				}
				paramKey := paramArry[0]
				paramVal := paramArry[1]
				paramStr := fmt.Sprintf("%v=%v", paramKey, paramVal)
				if paramKey == "template-repo" {
					configStruct.TemplateActualRepoUrl = paramVal
				} else if paramKey == "images-registry" {
					configStruct.ImagesRegistry = paramVal
				} else {
					return nil
				}
				configStruct.SaveConfigYaml()
				common.SmartIDELog.Info(i18nInstance.Config.Info_set_config_success, paramStr)
			}

		}
		return nil
	},
}

func init() {

}
