/*
 * @Date: 2022-07-11 15:38:06
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-10-24 14:47:18
 * @FilePath: /cli/cmd/config.go
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
					configStruct.TemplateRepo = paramVal
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
