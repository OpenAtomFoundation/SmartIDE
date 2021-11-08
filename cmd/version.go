package cmd

import (
	"fmt"

	"github.com/leansoftX/smartide-cli/lib/common"
	"github.com/leansoftX/smartide-cli/lib/i18n"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: i18n.GetInstance().Version.Info.Help_short,
	Long:  i18n.GetInstance().Version.Info.Help_long,
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
}

//
func (smartVersion *SmartVersion) ConvertToJson() string {
	json := fmt.Sprintf(i18n.GetInstance().Version.Info.Template,
		smartVersion.VersionNumber, smartVersion.BuildNumber, smartVersion.TargetCommitish, smartVersion.Company)
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
