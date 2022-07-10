/*
 * @Date: 2022-03-08 17:27:09
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-07-06 21:56:55
 * @FilePath: /smartide-cli/cmd/server/args.go
 */
package server

import (
	"fmt"
	"strings"

	"github.com/leansoftX/smartide-cli/internal/apk/i18n"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var i18nInstance = i18n.GetInstance()

const (
	Flags_Mode = "mode"

	Flags_ServerWorkspaceid = "serverworkspaceid"
	Flags_ServerToken       = "servertoken"
	Flags_ServerUsername    = "serverusername"
	Flags_ServerHost        = "serverhost"
)

// 获取服务器模式下的cmd参数
func GetServerModeInfo(cmd *cobra.Command) (serverModeInfo ServerModeInfo, err error) {
	err = Check(cmd)
	if err != nil {
		return
	}

	fflags := cmd.Flags()

	serverModeInfo.ServerWorkspaceid, _ = fflags.GetString(Flags_ServerWorkspaceid)
	serverModeInfo.ServerToken, _ = fflags.GetString(Flags_ServerToken)
	serverModeInfo.ServerUsername, _ = fflags.GetString(Flags_ServerUsername)
	serverModeInfo.ServerHost, _ = fflags.GetString(Flags_ServerHost)

	return
}

// 验证server模式下，flag是否有录入
func Check(cmd *cobra.Command) (err error) {

	fflags := cmd.Flags()

	// 如果不是 server 模式不需要验证
	mode, _ := fflags.GetString(Flags_Mode)
	if strings.ToLower(mode) != "server" {
		return nil
	}

	/* 	// server workspace id 不能为空
	   	err = checkFlagRequired(fflags, Flags_ServerWorkspaceid)
	   	if err != nil {
	   		return err
	   	} */

	// 当为start时
	if strings.EqualFold(cmd.Name(), "start") {
		// token 不能为空；
		err = checkFlagRequired(fflags, Flags_ServerToken)
		if err != nil {
			return err
		}

		// username、user guid不能为空；
		err = checkFlagRequired(fflags, Flags_ServerUsername)
		if err != nil {
			return err
		}
		/* err = checkFlagRequired(fflags, Flags_ServerUserGUID)
		if err != nil {
			return err
		} */

		// feedback 地址不能为空
		err = checkFlagRequired(fflags, Flags_ServerHost)
		if err != nil {
			return err
		}
	}

	common.SmartIDELog.Info("Mode server params validation passed.")

	return nil
}

// 检查参数是否填写
func checkFlagRequired(fflags *pflag.FlagSet, flagName string) error {
	flagValue, _ := fflags.GetString(flagName)
	if !fflags.Changed(flagName) || flagValue == "" {
		return fmt.Errorf(i18nInstance.Main.Err_flag_value_required, flagName)
	}
	return nil
}
