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

package appinsight

import (
	_ "embed"
	"flag"
	"os"
	"strings"
	"time"

	"github.com/leansoftX/smartide-cli/internal/biz/config"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/microsoft/ApplicationInsights-Go/appinsights"
)

var (
	telemetryClient    appinsights.TelemetryClient
	instrumentationKey string
)

func SetCliTrack(cmd string, args []string) {
	SetAllTrack(cmd, args, "", "", "", "", "", "")
}

func SetCliLoginTrack(cmd, loginurl, username string, args []string) {
	SetAllTrack(cmd, args, loginurl, username, "", "", "", "")
}

func SetCliLocalTrack(cmd string, args []string, clientworkspaceid, images string) {
	clientmachinename, _ := os.Hostname()
	//[]string{""}
	SetAllTrack(cmd, args, "", "", "", clientworkspaceid, clientmachinename, images)
}
func SetWorkSpaceTrack(cmd string, args []string, workspacemode, serveruserguid, serverworkspaceid, clientworkspaceid, clientmachinename, images string) {
	if config.GlobalSmartIdeConfig.IsInsightEnabled == config.IsInsightEnabledEnum_Enabled {
		event := appinsights.NewEventTelemetry(cmd)
		event.Properties["cli-cmd"] = cmd
		var argsstr string
		for _, val := range args {
			argsstr = argsstr + " " + val
		}
		event.Properties["cli-args"] = argsstr
		event.Properties["cli-workspacemode"] = workspacemode
		event.Properties["cli-serveruserguid"] = serveruserguid
		event.Properties["cli-serverworkspaceid"] = serverworkspaceid
		event.Properties["cli-clientworkspaceid"] = clientworkspaceid
		event.Properties["cli-clientmachinename"] = clientmachinename
		event.Properties["cli-images"] = images
		SetTrack(*event)
	}
}
func SetAllTrack(cmd string, args []string, serverhost, serveruserguid, serverworkspaceid, clientworkspaceid, clientmachinename, images string) {
	if config.GlobalSmartIdeConfig.IsInsightEnabled == config.IsInsightEnabledEnum_Enabled {
		event := appinsights.NewEventTelemetry(cmd)
		event.Properties["cli-cmd"] = cmd

		var argsstr string
		for _, val := range args {
			argsstr = argsstr + " " + val
		}
		event.Properties["cli-args"] = argsstr

		if common.Contains([]string{Cli_Server_Connect, Cli_Server_Login, Cli_Server_Logout}, strings.ToLower(cmd)) {
			event.Properties["cli-serverhost"] = serverhost
		} else {
			event.Properties["cli-serverhost"] = Global.Serverhost
		}
		if common.Contains([]string{Cli_K8s_Ingress_Apply, Cli_K8s_Ssh_Apply}, strings.ToLower(cmd)) {
			event.Properties["cli-serverworkspaceid"] = serverworkspaceid
		}

		if Global.Mode != "client" {
			if Global.ServerUserGuid == "" {
				event.Properties["cli-serveruserguid"] = Global.ServerUserName
			} else {
				event.Properties["cli-serveruserguid"] = Global.ServerUserGuid
			}
		} else {
			event.Properties["cli-serveruserguid"] = serveruserguid
			event.Properties["cli-serverworkspaceid"] = serverworkspaceid
		}

		event.Properties["cli-clientworkspaceid"] = clientworkspaceid
		event.Properties["cli-clientmachinename"] = clientmachinename
		event.Properties["cli-images"] = images
		SetTrack(*event)
	}
}

func SetTrack(event appinsights.EventTelemetry) {
	if config.GlobalSmartIdeConfig.IsInsightEnabled == config.IsInsightEnabledEnum_Enabled {
		event.Properties["cli-version"] = Global.Version
		event.Properties["cli-mode"] = Global.Mode

		hostname, _ := os.Hostname()
		event.Tags.User().SetId(hostname)
		event.Tags.Application().SetVer(Global.Version)
		event.Tags.Cloud().SetRole(Global.Cloud_RoleName)
		telemetryClient.Track(&event)
	}
}

// 初始化
func init() {
	flag.StringVar(&instrumentationKey, "instrumentationKey", "$(instrumentationKey)", "set instrumentation key from azure portal")
	telemetryConfig := appinsights.NewTelemetryConfiguration(instrumentationKey)
	telemetryConfig.EndpointUrl = "https://dc.applicationinsights.azure.cn/v2/track"

	// 配置一次调用可以向数据收集器发送多少项目：
	// 每个可以提交中的遥测项的最大数量要求。 如果缓冲了这么多项目，则缓冲区将在 MaxBatchInterval 到期之前被刷新。
	telemetryConfig.MaxBatchSize = 1
	// 在发送排队遥测数据之前配置最大延迟：
	// 发送一批遥测数据之前等待的最长时间。
	telemetryConfig.MaxBatchInterval = 10 * time.Millisecond
	telemetryClient = appinsights.NewTelemetryClientFromConfig(telemetryConfig)

	appinsights.NewDiagnosticsMessageListener(func(msg string) error {
		//common.SmartIDELog.DebugF("[%s] %s\n", time.Now().Format("2006-01-02 15:04:05.000"), msg)
		if strings.Contains(strings.ToLower(msg), "response: 200") {
			common.SmartIDELog.Debug("application insight success!")
		}
		return nil
	})
}
