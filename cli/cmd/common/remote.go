/*
SmartIDE - CLI
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

package common

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/leansoftX/smartide-cli/internal/apk/i18n"
	"github.com/leansoftX/smartide-cli/internal/biz/workspace"
	"github.com/leansoftX/smartide-cli/internal/dal"
	"github.com/leansoftX/smartide-cli/internal/model"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/spf13/pflag"
)

var (
	flag_host     = "host"
	flag_port     = "port"
	flag_username = "username"
	flag_password = "password"
)

var i18nInstance = i18n.GetInstance()

// 根据参数，从数据库或者其他参数中加载远程服务器的信息
func GetRemoteAndValid(fflags *pflag.FlagSet) (remoteInfo *workspace.RemoteInfo, err error) {

	host, _ := fflags.GetString(flag_host)
	userName, _ := fflags.GetString(flag_username)

	// 指定了host信息，尝试从数据库中加载
	if common.IsNumber(host) {
		remoteId, err := strconv.Atoi(host)
		common.CheckError(err)
		remoteInfo, err = dal.GetRemoteById(remoteId)
		common.CheckError(err)

		if remoteInfo == nil {
			common.SmartIDELog.Warning(i18nInstance.Host.Err_host_data_not_exit)
		}
	} else {
		remoteInfo, err = dal.GetRemoteByHost(host, userName)

		// 如果在sqlite中有缓存数据，就不需要用户名、密码
		if remoteInfo != nil && flag_password != remoteInfo.Password {
			common.SmartIDELog.Importance(fmt.Sprintf("可以使用 start --host %v 引用该服务器资源", remoteInfo.ID))
		}
	}

	// 从参数中加载
	if remoteInfo == nil {
		//  必填字段验证
		err = CheckFlagRequired(fflags, flag_host)
		if err != nil {
			return remoteInfo, &FriendlyError{Err: err}
		}
		err = CheckFlagRequired(fflags, flag_username)
		if err != nil {
			return remoteInfo, &FriendlyError{Err: err}
		}

		remoteInfo.Addr = host
		remoteInfo.UserName = GetFlagValue(fflags, flag_username)
		remoteInfo.SSHPort, err = fflags.GetInt(flag_port) //strconv.Atoi(getFlagValue(fflags, flag_port))
		common.CheckError(err)
		if remoteInfo.SSHPort <= 0 {
			remoteInfo.SSHPort = model.CONST_Container_SSHPort
		}
		// 认证类型
		if fflags.Changed(flag_password) {
			remoteInfo.Password = GetFlagValue(fflags, flag_password)
			remoteInfo.AuthType = workspace.RemoteAuthType_Password
		} else {
			remoteInfo.AuthType = workspace.RemoteAuthType_SSH
		}

	}

	return remoteInfo, err
}

func GetFlagValue(fflags *pflag.FlagSet, flag string) string {
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

// 检查参数是否填写
func CheckFlagRequired(fflags *pflag.FlagSet, flagName string) error {
	if !fflags.Changed(flagName) {
		return fmt.Errorf(i18nInstance.Main.Err_flag_value_required, flagName)
	}
	return nil
}

// 在某些情况下，参数填了也没有意义，比如指定了workspaceid，就不需要再填host
func CheckFlagUnnecessary(fflags *pflag.FlagSet, flagName string, preFlagName string) {
	if fflags.Changed(flagName) {
		common.SmartIDELog.WarningF(i18nInstance.Main.Err_flag_value_invalid, preFlagName, flagName)
	}
}

// 友好的错误
type FriendlyError struct {
	Err error
}

func (e *FriendlyError) Error() string {
	return e.Err.Error()
}
