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
