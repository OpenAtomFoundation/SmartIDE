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
	"time"

	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: i18nInstance.Version.Info_help_short,
	Long:  i18nInstance.Version.Info_help_long,
	Run: func(cmd *cobra.Command, args []string) {

		common.SmartIDELog.Console(Version.ConvertToJson())
	},
}

type SmartVersion struct {
	VersionNumber        string
	TagName              string `json:"tag_name"`
	BuildNumber          string `json:"build_number"`
	TargetCommitish      string `json:"target_commitish"`
	TargetCommitishShort string
	BuildQuqueTime       string `json:"build_ququeTime"`
	Company              string `json:"company"`
	// 编译时间
	BuildTime time.Time
}

//
func (smartVersion *SmartVersion) ConvertToJson() string {
	json := fmt.Sprintf(i18nInstance.Version.Info_template,
		smartVersion.VersionNumber, smartVersion.BuildNumber, smartVersion.BuildTime.Format("2006-01-02 15:04:05"), smartVersion.TargetCommitish, smartVersion.Company)
	return json
}

func init() {

}
