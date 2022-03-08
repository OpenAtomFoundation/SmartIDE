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
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/leansoftX/smartide-cli/internal/apk/i18n"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var udpateCmd = &cobra.Command{
	Use:     "update",
	Short:   i18n.GetInstance().Update.Info_help_short,
	Long:    i18n.GetInstance().Update.Info_help_long,
	Aliases: []string{"up"},
	Run: func(cmd *cobra.Command, args []string) {

		common.SmartIDELog.Info("update ...")

		var command string = ""
		var execCommand *exec.Cmd

		// 是否仅更新build版本
		isBuild, err := cmd.Flags().GetBool("build")
		if err != nil {
			common.SmartIDELog.Error(err)
		}

		// 版本号
		version, err := cmd.Flags().GetString("version")
		if err != nil {
			common.SmartIDELog.Error(err)
		}
		if version == "" { // 如果不指定，就是自动升级模式
			versionUrl := ""
			if isBuild {
				versionUrl = "https://smartidedl.blob.core.chinacloudapi.cn/builds/stable.txt"
			} else {
				versionUrl = "https://smartidedl.blob.core.chinacloudapi.cn/releases/stable.txt"
			}

			resp, err := http.Get(versionUrl)
			if err != nil {
				common.SmartIDELog.Error(err)
			}
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				common.SmartIDELog.Error(err)
			}
			version = string(body)

			// 比较版本号，检查是否有必要升级
			if len(strings.Split(Version.BuildNumber, ".")) == 4 && compareVersion(Version.BuildNumber, version) > -1 {
				common.SmartIDELog.WarningF(i18nInstance.Update.Warn_rel_lastest, version)
				return
			}
		}

		// log
		common.SmartIDELog.Info("update to v" + version)

		// 删除文件
		if common.IsExit("smartide") {
			os.Remove("smartide")
			common.SmartIDELog.Info(i18n.GetInstance().Update.Info_remove_repeat)
		}

		// 运行升级脚本
		common.SmartIDELog.Info("updating ... ")
		switch runtime.GOOS {
		case "windows":

			if isBuild {
				command = fmt.Sprintf(`Invoke-WebRequest -Uri ("https://smartidedl.blob.core.chinacloudapi.cn/builds/%v/SetupSmartIDE.msi")  -OutFile "smartide.msi"

				.\smartIDE.msi
				
				`, "")
			} else {
				command = fmt.Sprintf(`Invoke-WebRequest -Uri ("https://smartidedl.blob.core.chinacloudapi.cn/releases/%v/SetupSmartIDE.msi")  -OutFile "smartide.msi"

				.\smartIDE.msi
				
				`, version)
			}
			execCommand = exec.Command("powershell", "/c", command)
		case "darwin":
			if isBuild {
				command = fmt.Sprintf(`curl -OL  "https://smartidedl.blob.core.chinacloudapi.cn/builds/%v/smartide-osx" \
				&& mv -f smartide-osx /usr/local/bin/smartide \
				&& ln -s -f /usr/local/bin/smartide /usr/local/bin/se \
				&& chmod +x /usr/local/bin/smartide`, version)
			} else {
				command = fmt.Sprintf(`curl -OL  "https://smartidedl.blob.core.chinacloudapi.cn/releases/%v/smartide" \
				&& mv -f smartide /usr/local/bin/smartide \
				&& ln -s -f /usr/local/bin/smartide /usr/local/bin/se \
				&& chmod +x /usr/local/bin/smartide`, version)
			}
			execCommand = exec.Command("bash", "-c", command)
		case "linux":
			if isBuild {
				command = fmt.Sprintf(`curl -OL  "https://smartidedl.blob.core.chinacloudapi.cn/builds/%v/smartide-linux" \
				&& sudo mv -f smartide-linux /usr/local/bin/smartide \
				&& sudo ln -s -f /usr/local/bin/smartide /usr/local/bin/se \
				&& chmod +x /usr/local/bin/smartide`, version)
			} else {
				command = fmt.Sprintf(`curl -OL  "https://smartidedl.blob.core.chinacloudapi.cn/releases/%v/smartide-linux" \
				&& sudo mv -f smartide /usr/local/bin/smartide \
				&& sudo ln -s -f /usr/local/bin/smartide /usr/local/bin/se \
				&& chmod +x /usr/local/bin/smartide`, version)
			}
			execCommand = exec.Command("bash", "-c", command)
		default:
			common.SmartIDELog.Error("can not support current os")
		}

		// run
		execCommand.Stdout = os.Stdout
		execCommand.Stderr = os.Stderr
		err = execCommand.Run()
		if err != nil {
			common.SmartIDELog.Error(err)
		}
		common.SmartIDELog.Info("update to v" + version + " complated. ")

		// show current version
		if runtime.GOOS == "darwin" {
			versionCommand := exec.Command("bash", "-c", "smartide version")
			versionCommand.Stdout = os.Stdout
			versionCommand.Stderr = os.Stderr
			versionCommand.Run()
		}
		//TODO windows时，查看“smartide install”进程是否存在，不存在时才运行 smartide version
	},
}

// 比较两个版本号 version1 和 version2。
// 如果 version1 > version2 返回 1，如果 version1 < version2 返回 -1， 除此之外返回 0。
func compareVersion(version1 string, version2 string) int {
	var res int
	ver1Strs := strings.Split(version1, ".")
	ver2Strs := strings.Split(version2, ".")
	ver1Len := len(ver1Strs)
	ver2Len := len(ver2Strs)
	verLen := ver1Len
	if len(ver1Strs) < len(ver2Strs) {
		verLen = ver2Len
	}
	for i := 0; i < verLen; i++ {
		var ver1Int, ver2Int int
		if i < ver1Len {
			ver1Int, _ = strconv.Atoi(ver1Strs[i])
		}
		if i < ver2Len {
			ver2Int, _ = strconv.Atoi(ver2Strs[i])
		}
		if ver1Int < ver2Int {
			res = -1
			break
		}
		if ver1Int > ver2Int {
			res = 1
			break
		}
	}
	return res
}

func init() {
	udpateCmd.Flags().BoolP("build", "b", false, i18n.GetInstance().Update.Info_help_flag_build)
	udpateCmd.Flags().StringP("version", "v", "", i18n.GetInstance().Update.Info_help_flag_version)
}
