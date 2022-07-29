package init

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/leansoftX/smartide-cli/internal/apk/i18n"
)

var i18nInstance = i18n.GetInstance()

// 打印 service 列表
func printTemplates(newType []NewTypeBO) {
	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
	for i := 0; i < len(newType); i++ {
		line := fmt.Sprintf("%v %v", i, newType[i].TypeName)
		fmt.Fprintln(w, line)
	}
	w.Flush()

}
func init() {

}
