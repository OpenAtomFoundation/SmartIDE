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

package init

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/leansoftX/smartide-cli/internal/apk/i18n"
	"github.com/leansoftX/smartide-cli/internal/biz/config"
	"github.com/leansoftX/smartide-cli/pkg/common"
)

var i18nInstance = i18n.GetInstance()

const TMEPLATE_DIR_NAME = "templates"

// 打印 service 列表
func PrintTemplates(newType []NewTypeBO) {
	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
	for i := 0; i < len(newType); i++ {
		line := fmt.Sprintf("%v %v", i, newType[i].TypeName)
		fmt.Fprintln(w, line)
	}
	w.Flush()

}

// 加载templates索引json
func LoadTemplatesJson() (templateTypes []NewTypeBO, err error) {
	// new type转换为结构体
	templatesPath := common.PathJoin(config.SmartIdeHome, TMEPLATE_DIR_NAME, "templates.json")
	templatesByte, err := os.ReadFile(templatesPath)
	if err != nil {
		return templateTypes, errors.New(i18nInstance.New.Err_read_templates + templatesPath + err.Error())
	}

	err = json.Unmarshal(templatesByte, &templateTypes)
	return templateTypes, err
}
func init() {

}
