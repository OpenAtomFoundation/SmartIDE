/*
 * @Author: jason chen (jasonchen@leansoftx.com, http://smallidea.cnblogs.com)
 * @Description:
 * @Date: 2021-11
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-10-24 14:40:56
 */
package host

import (
	"fmt"
	"strconv"

	"github.com/leansoftX/smartide-cli/internal/apk/i18n"
	"github.com/leansoftX/smartide-cli/internal/biz/workspace"
	"github.com/leansoftX/smartide-cli/internal/dal"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/spf13/cobra"
)

const flag_hostid = "hostid"

var i18nInstance = i18n.GetInstance()

// initCmd represents the init command
var HostGetCmd = &cobra.Command{
	Use:   "get",
	Short: i18nInstance.Host.Info_help_get_short,
	Long:  i18nInstance.Host.Info_help_get_long,
	Example: ` smartide host get --hostid <hostid>
  smartide host get <hostid>`,
	Run: func(cmd *cobra.Command, args []string) {

		hostId := getHostIdFromFlagsAndArgs(cmd, args)
		if hostId <= 0 {
			common.SmartIDELog.WarningF(i18nInstance.Common.Warn_param_is_null, flag_hostid)
			cmd.Help()
			return
		}

		remoteInfo, err := dal.GetRemoteById(hostId)
		entryptionKey4Host(*remoteInfo)
		common.CheckError(err)
		createTime := remoteInfo.CreatedTime.Format("2006-01-02 15:04:05")

		print := fmt.Sprintf(i18nInstance.Host.Info_host_detail_template,
			remoteInfo.ID, remoteInfo.Addr, remoteInfo.AuthType, remoteInfo.UserName, createTime)
		common.SmartIDELog.Console(print)

	},
}

func entryptionKey4Host(remoteInfo workspace.RemoteInfo) {
	if remoteInfo.Password != "" {
		common.SmartIDELog.AddEntryptionKey(remoteInfo.Password)
	}
	if remoteInfo.SSHKey != "" {
		common.SmartIDELog.AddEntryptionKeyWithReservePart(remoteInfo.SSHKey)
	}
}

// 获取工作区id
func getHostIdFromFlagsAndArgs(cmd *cobra.Command, args []string) int {
	fflags := cmd.Flags()

	// 从args 或者 flag 中获取值
	var hostId int
	if len(args) > 0 { // 从args中加载
		str := args[0]
		tmpHostId, err := strconv.Atoi(str)
		if err == nil && tmpHostId > 0 {
			hostId = tmpHostId
		}

	} else if fflags.Changed(flag_hostid) { // 从flag中加载
		tmpHostId, err := fflags.GetInt32(flag_hostid)
		common.CheckError(err)
		if tmpHostId > 0 {
			hostId = int(tmpHostId)
		}
	}

	return hostId
}

func init() {
	HostGetCmd.Flags().Int32P("hostid", "r", 0, i18nInstance.Host.Info_help_flag_hostid)

}
