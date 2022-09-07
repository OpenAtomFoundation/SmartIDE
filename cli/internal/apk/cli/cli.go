/*
 * @Date: 2022-09-03 16:12:56
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-09-05 11:22:48
 * @FilePath: /cli/internal/apk/cli/cli.go
 */

package cli

import (
	"strings"

	"github.com/leansoftX/smartide-cli/pkg/common"
)

// 通过shell获取smartide的版本
func GetCliVersionByShell() string {
	output, _ := common.EXEC.CombinedOutput("smartide version", "")
	return strings.Split(output, "\n")[0]
	return strings.ReplaceAll(output, "\n", "; ")
}
