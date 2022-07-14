/*
 * @Date: 2022-04-22 10:22:50
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-04-22 10:26:14
 * @FilePath: /smartide-cli/cmd/common/remote.go
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
	//flag_project  = "project"
	//flag_branch   = "branch"
	//flag_k8s         = "k8s"
	//flag_namespace   = "namespace"
)

var i18nInstance = i18n.GetInstance()

// 根据参数，从数据库或者其他参数中加载远程服务器的信息
func GetRemoteAndValid(fflags *pflag.FlagSet) (remoteInfo workspace.RemoteInfo, err error) {

	host, _ := fflags.GetString(flag_host)
	remoteInfo = workspace.RemoteInfo{}

	// 指定了host信息，尝试从数据库中加载
	if common.IsNumber(host) {
		remoteId, err := strconv.Atoi(host)
		common.CheckError(err)
		remoteInfo, err = dal.GetRemoteById(remoteId)
		common.CheckError(err)

		if (workspace.RemoteInfo{} == remoteInfo) {
			common.SmartIDELog.Warning(i18nInstance.Host.Err_host_data_not_exit)
		}
	} else {
		remoteInfo, err = dal.GetRemoteByHost(host)

		// 如果在sqlite中有缓存数据，就不需要用户名、密码
		if (workspace.RemoteInfo{} != remoteInfo) {
			CheckFlagUnnecessary(fflags, flag_username, flag_host)
			CheckFlagUnnecessary(fflags, flag_password, flag_host)
		}
	}

	// 从参数中加载
	if (workspace.RemoteInfo{} == remoteInfo) {
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

//
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

//
func (e *FriendlyError) Error() string {
	return e.Err.Error()
}
