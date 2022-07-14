package start

import (
	"fmt"

	"github.com/leansoftX/smartide-cli/internal/biz/config"
	"github.com/leansoftX/smartide-cli/internal/biz/workspace"
	"github.com/leansoftX/smartide-cli/pkg/docker/compose"
)

// addon 启用-加入workspace
func AddonEnable(workspaceInfo workspace.WorkspaceInfo) workspace.WorkspaceInfo {
	if workspaceInfo.Addon.Type == "webterminal" {
		workspaceInfo = addWebTerminalToDockerCompose(workspaceInfo)
	}
	return workspaceInfo
}

// 初始化addon webterminal
func addWebTerminalToDockerCompose(workspaceInfo workspace.WorkspaceInfo) workspace.WorkspaceInfo {
	if len(workspaceInfo.TempDockerCompose.Networks) == 0 {
		workspaceInfo.TempDockerCompose.Networks = make(map[string]compose.Network)
		workspaceInfo.TempDockerCompose.Networks["smartide-network"] = compose.Network{External: true}
	} else {
		if _, ok := workspaceInfo.TempDockerCompose.Networks["smartide-network"]; !ok {
			workspaceInfo.TempDockerCompose.Networks["smartide-network"] = compose.Network{External: true}
		}
	}
	if len(workspaceInfo.TempDockerCompose.Services) == 0 {
		workspaceInfo.TempDockerCompose.Services = make(map[string]compose.Service)
	}
	webterminServiceName := fmt.Sprintf("%v_smartide-webterminal", workspaceInfo.Name)
	workspaceInfo.TempDockerCompose.Services[webterminServiceName] = config.GetWebTerminalCompose(webterminServiceName, workspaceInfo.WorkingDirectoryPath)
	return workspaceInfo
}

func init() {

}
