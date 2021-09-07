package i18n

import (
	"encoding/json"
	"path/filepath"

	//"log"
	//"os"

	"strings"

	"github.com/jeandeaual/go-locale"
	"github.com/leansoftX/i18n"

	// 这里不要忘记引入驱动,如引入默认的json驱动
	_ "github.com/leansoftX/i18n/parser_json"
)

//
type Language struct {
	Main struct {
		Info struct {
			Help_short string `json:"help_short"`
			Help_long  string `json:"help_long"`
		} `json:"info"`
	} `json:"main"`

	Init struct {
		Info struct {
			Help_short string `json:"help_short"`
			Help_long  string `json:"help_long"`
			Info_start string `json:"info_start"`
			Info_end   string `json:"info_end"`
		} `json:"info"`
	} `json:"init"`

	Start struct {
		Info struct {
			Help_short              string `json:"help_short"`
			Help_long               string `json:"help_long"`
			Info_start              string `json:"info_start"`
			Info_end                string `json:"info_end"`
			Info_running_container  string `json:"info_running_container"`
			Info_running_openbrower string `json:"info_running_openbrower"`
		} `json:"info"`
		Error struct {
		} `json:"error"`
	} `json:"start"`

	Stop struct {
		Info struct {
			Help_short string `json:"help_short"`
			Help_long  string `json:"help_long"`
			Info_start string `json:"info_start"`
			Info_end   string `json:"info_end"`
		} `json:"info"`
		Error struct {
		} `json:"error"`
	} `json:"Stop"`

	Remove struct {
		Info struct {
			Help_short string `json:"help_short"`
			Help_long  string `json:"help_long"`
			Info_start string `json:"info_start"`
			Info_end   string `json:"info_end"`
		} `json:"info"`
		Error struct {
		} `json:"error"`
	} `json:"Remove"`
}

var instance *Language

func GetInstance() *Language {
	if instance == nil {
		/* exePath, err := os.Getwd()
		if err != nil {
			log.Println(err)
			panic(err)
		} */

		// https://github.com/leansoftX/i18n
		languageDir := filepath.ToSlash("lib/i18n/language")
		currentLang, _ := locale.GetLocale()
		if strings.Index(currentLang, "zh-") == 0 { // 如果不是简体中文，就是英文
			currentLang = "zh_cn"
		} else {
			currentLang = "en_us"
		}
		lang := i18n.NewI18n(
			// 这里指定语言文件路径
			i18n.LangDirectory(languageDir),

			// 这里如果不i设置, 则默认使用zh_cn
			i18n.DefaultLang(strings.ToLower(currentLang)),

			// 这里如果不i设置, 则默认使用 json,可以自定义解析器和配置文件格式
			//i18n.DefaultParser("json"),
		)

		//instance := Language{}

		l := lang.LoadWithDefault("")
		bodyBytes, _ := json.Marshal(l)
		json.Unmarshal(bodyBytes, &instance)

	}
	return instance
}

/*
func main() {
	lang := i18n.NewI18n(
		// 这里指定语言文件路径
		i18n.LangDirectory("/language"),

		// 这里如果不i设置, 则默认使用zh_cn
		//i18n.DefaultLang("zh_cn"),

		// 这里如果不i设置, 则默认使用 json,可以自定义解析器和配置文件格式
		//i18n.DefaultParser("json"),
	)

	// 加载error.json文件内的具体配置项, 多级加载, 使用.连接
	test := lang.Load("test")
	test2 := lang.Load("err2.bb.cc")

	fmt.Println(test)
	fmt.Println(test2)
}
*/
