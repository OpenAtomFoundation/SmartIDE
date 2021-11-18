package host

import (
	"fmt"
	"strconv"

	"github.com/leansoftX/smartide-cli/cmd/dal"
	"github.com/leansoftX/smartide-cli/lib/common"
	"github.com/leansoftX/smartide-cli/lib/i18n"
	"github.com/spf13/cobra"
)

const flag_hostid = "hostid"

var i18nInstance = i18n.GetInstance()

// initCmd represents the init command
var HostGetCmd = &cobra.Command{
	Use:   "get",
	Short: i18nInstance.Host.Info.Help_get_short,
	Long:  i18nInstance.Host.Info.Help_get_long,
	Example: ` smartide host get --hostid <hostid>
  smartide host get <hostid>`,
	Run: func(cmd *cobra.Command, args []string) {

		hostId := getHostIdFromFlagsAndArgs(cmd, args)
		if hostId <= 0 {
			common.SmartIDELog.WarningF(i18nInstance.Common.Warn.Warn_param_is_null, flag_hostid)
			cmd.Help()
			return
		}

		remote, err := dal.GetRemoteById(hostId)
		common.CheckError(err)
		createTime := remote.CreatedTime.Format("2006-01-02 15:04:05")

		print := fmt.Sprintf(i18nInstance.Host.Info.Info_host_detail_template,
			remote.ID, remote.Addr, remote.AuthType, remote.UserName, createTime)
		common.SmartIDELog.Console(print)

		dal.GetRemoteById(hostId)
	},
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
	HostGetCmd.Flags().Int32P("hostid", "r", 0, i18nInstance.Host.Info.Help_flag_hostid)

}
