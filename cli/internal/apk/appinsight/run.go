package appinsight

import (
	_ "embed"
	"flag"
	"os"

	"github.com/microsoft/ApplicationInsights-Go/appinsights"
)

var (
	telemetryClient    appinsights.TelemetryClient
	instrumentationKey string
)

func SetTrack(cmd, version, args, workModel, imageName string) {
	event := appinsights.NewEventTelemetry("smartide-cli-" + cmd)
	/*
		cmd- smartide 启动事件
		version- 当前的smartide-cli版本
		- 本地模式启动次数
		- 远程模式启动次数
		- 所使用的镜像名称和次数
		- 用户本地对外的ip地址，如果比较容易拿到的话
		- 用户使用的操作系统和本地环境语言
	*/
	event.Properties["cmd"] = cmd
	event.Properties["version"] = version
	event.Properties["args"] = args
	event.Properties["workmodel"] = workModel
	event.Properties["image"] = imageName
	hostname, _ := os.Hostname()
	event.Tags.User().SetId(hostname)
	event.Tags.Application().SetVer(version)
	telemetryClient.Track(event)
}

//初始化
func init() {
	flag.StringVar(&instrumentationKey, "instrumentationKey", "$(instrumentationKey)", "set instrumentation key from azure portal")
	telemetryConfig := appinsights.NewTelemetryConfiguration(instrumentationKey)
	telemetryConfig.EndpointUrl = "https://dc.applicationinsights.azure.cn/v2/track"
	telemetryClient = appinsights.NewTelemetryClientFromConfig(telemetryConfig)

	appinsights.NewDiagnosticsMessageListener(func(msg string) error {
		//common.SmartIDELog.Debug(msg)
		return nil
	})
}
