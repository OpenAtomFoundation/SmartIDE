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
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/spf13/cobra"
	"golang.org/x/text/encoding/simplifiedchinese"
)

// initCmd represents the init command
var versionCmd = &cobra.Command{
	Use:     "version",
	Short:   i18nInstance.Version.Info_help_short,
	Long:    i18nInstance.Version.Info_help_long,
	Aliases: []string{"v"},
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

	systemInfo := getOsInformation()

	json := fmt.Sprintf(i18nInstance.Version.Info_template,
		smartVersion.VersionNumber, systemInfo, smartVersion.BuildNumber,
		smartVersion.BuildTime.Format("2006-01-02 15:04:05"), smartVersion.TargetCommitish, smartVersion.Company)
	return json
}

func init() {

}

// 获取系统版本
func getOsInformation() string {
	var execCommand *exec.Cmd

	output := runtime.GOOS + " " + runtime.GOARCH

	switch runtime.GOOS {
	case "windows":

		execCommand = exec.Command("powershell", "/c", "Get-WmiObject -Class Win32_OperatingSystem | Select-Object -ExpandProperty Caption")
		outputBytes, err := execCommand.Output()

		if err == nil {

			decodeBytes, _ := simplifiedchinese.GB18030.NewDecoder().Bytes(outputBytes)
			output = string(decodeBytes)
		}

	case "darwin":

		execCommand = exec.Command("bash", "-c", "sw_vers")
		outputBytes, err := execCommand.Output()
		if err == nil {
			tmp := string(outputBytes)
			productName := ""
			productVersion := ""
			for _, str := range strings.Split(tmp, "\n") {
				if strings.Contains(strings.ToLower(str), "productname") {
					productName = strings.ReplaceAll(strings.Split(str, ":")[1], "\t", "")
				} else if strings.Contains(strings.ToLower(str), "productversion") {
					productVersion = strings.ReplaceAll(strings.Split(str, ":")[1], "\t", "")
				}
			}
			output = productName + "\t" + productVersion
		}

	case "linux":

		execCommand = exec.Command("bash", "-c", "lsb_release -a")
		outputBytes, err := execCommand.Output()
		if err == nil {
			tmp := string(outputBytes)
			for _, str := range strings.Split(tmp, "\n") {
				if strings.Contains(strings.ToLower(str), "description") {
					output = strings.ReplaceAll(strings.Split(str, ":")[1], "\t", "")
					break
				}
			}
		}

	default:

	}

	//output = strings.ReplaceAll(output, "\t", "")

	return output

}

/*
sw_vers on macOS

ProductName: Mac OS X

ProductVersion: 10.14.5

BuildVersion: 18F132

---
lsb_release -a on Ubuntu

Distributor ID: Ubuntu

Description: Ubuntu 14.04.5 LTS

Release: 14.04

Codename: trusty

-----
ver on Windows

Microsoft Windows [Version 10.0.17134.829]
*/
