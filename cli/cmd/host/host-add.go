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

package host

import (
	"fmt"
	"strings"

	"github.com/leansoftX/smartide-cli/internal/apk/appinsight"
	"github.com/leansoftX/smartide-cli/internal/biz/workspace"
	"github.com/leansoftX/smartide-cli/internal/dal"
	"github.com/leansoftX/smartide-cli/internal/model"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// initCmd represents the init command
var HostAddCmd = &cobra.Command{
	Use:     "add",
	Short:   i18nInstance.Host.Info_help_host_add_short,
	Long:    i18nInstance.Host.Info_help_host_add_long,
	Example: `  smartide host add <host> --username <username> --password <password> --port <port>`,
	Run: func(cmd *cobra.Command, args []string) {
		common.SmartIDELog.Info(i18nInstance.Host.Add_start)
		appinsight.SetCliTrack(appinsight.Cli_Add_Host, args)
		fflags := cmd.Flags()
		remoteInfo := workspace.RemoteInfo{}
		host := ""
		if len(args) > 0 { // 从args中加载
			str := args[0]
			if str != "" {
				host = str
			} else {
				common.SmartIDELog.Error(fmt.Sprintf(i18nInstance.Host.Err_host_add_addr_required))
			}
		} else {
			common.SmartIDELog.Error(fmt.Sprintf(i18nInstance.Host.Err_host_add_addr_required))
		}
		err := checkFlagRequired(fflags, flag_username)
		if err != nil {
			common.SmartIDELog.Error(fmt.Sprintf(i18nInstance.Host.Err_host_add_username_required))
		}

		remoteInfo.Addr = host
		remoteInfo.UserName = getFlagValue(fflags, flag_username)
		remoteInfo.SSHPort, err = fflags.GetInt(flag_port)
		common.CheckError(err)
		if remoteInfo.SSHPort <= 0 {
			remoteInfo.SSHPort = model.CONST_Container_SSHPort
		}
		// 认证类型
		if fflags.Changed(flag_password) {
			remoteInfo.Password = getFlagValue(fflags, flag_password)
			remoteInfo.AuthType = workspace.RemoteAuthType_Password
		} else {
			remoteInfo.AuthType = workspace.RemoteAuthType_SSH
		}
		// 在远程模式下，首先验证远程服务器是否可以登录
		ssmRemote := common.SSHRemote{}
		common.SmartIDELog.InfoF(i18nInstance.Main.Info_ssh_connect_check, remoteInfo.Addr, remoteInfo.SSHPort)

		err = ssmRemote.CheckDail(remoteInfo.Addr, remoteInfo.SSHPort, remoteInfo.UserName, remoteInfo.Password, "")
		if err != nil {
			common.CheckError(err)
		}
		hostId, err := dal.InsertOrUpdateRemote(remoteInfo)
		common.CheckError(err)
		common.SmartIDELog.Info(fmt.Sprintf(i18nInstance.Host.Info_host_add_success, host, hostId))
	},
	PreRun: func(cmd *cobra.Command, args []string) {
		password := getFlagValue(cmd.Flags(), flag_password)
		if password != "" {
			common.SmartIDELog.AddEntryptionKey(password)
		}
	},
}

func init() {
	HostAddCmd.Flags().StringP("username", "u", "", i18nInstance.Start.Info_help_flag_username)
	HostAddCmd.Flags().StringP("password", "t", "", i18nInstance.Start.Info_help_flag_password)
	HostAddCmd.Flags().IntP("port", "p", 22, i18nInstance.Start.Info_help_flag_port)
}

var (
	flag_host     = "host"
	flag_port     = "port"
	flag_username = "username"
	flag_password = "password"
)

// 检查参数是否填写
func checkFlagRequired(fflags *pflag.FlagSet, flagName string) error {
	if !fflags.Changed(flagName) {
		return fmt.Errorf(i18nInstance.Main.Err_flag_value_required, flagName)
	}
	return nil
}

// 获取Flag值
func getFlagValue(fflags *pflag.FlagSet, flag string) string {
	value, err := fflags.GetString(flag)
	if err != nil {
		if strings.Contains(err.Error(), "flag accessed but not defined:") { // 错误判断，不需要双语
			common.SmartIDELog.Debug(err.Error())
		} else {
			common.SmartIDELog.Error(err)
		}
	}
	return value
}
