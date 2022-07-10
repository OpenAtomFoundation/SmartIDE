/*
 * @Date: 2022-04-21 14:42:12
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-04-21 15:05:52
 * @FilePath: /smartide-cli/cmd/new/commandServerVm.go
 */

package new

import (
	"github.com/leansoftX/smartide-cli/internal/biz/config"
	"github.com/leansoftX/smartide-cli/internal/biz/workspace"
	"github.com/spf13/cobra"
)

func VmNew_Server(cmd *cobra.Command, args []string, workspaceInfo workspace.WorkspaceInfo,
	yamlExecuteFun func(yamlConfig config.SmartIdeConfig)) {

}
