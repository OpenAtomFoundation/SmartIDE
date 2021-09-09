package i18n

import (
	"embed"
	"encoding/json"

	//"log"
	//"os"

	"strings"

	"github.com/jeandeaual/go-locale"

	// 这里不要忘记引入驱动,如引入默认的json驱动
	_ "github.com/leansoftX/i18n/parser_json"
)

//
type I18nSource struct {
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

var instance *I18nSource

var I18nSource_EN string
var I18nSource_ZH string

//go:embed language/*
var f embed.FS

// get internationalization source
// 获取当前系统的语言，动态加载对应的json文件并解析成结构体，方便在代码中调用
// 1. 新增，首先在“lib/i18n/language”的对应节点下新增，并同步在“lib/i18n/language/translate.go”中的I18nSource增加相应的属性；
// 2. 在代码中使用
//    var instanceI18nStart = i18n.GetInstance().Start
//    fmt.println(instanceI18nStart.Help_short)
func GetInstance() *I18nSource {
	if instance == nil {
		/* exePath, err := os.Getwd()
		if err != nil {
			log.Println(err)
			panic(err)
		} */

		// https://github.com/leansoftX/i18n
		/* languageDir := filepath.ToSlash("lib/i18n/language") */
		currentLang, _ := locale.GetLocale()
		if strings.Index(currentLang, "zh-") == 0 { // 如果不是简体中文，就是英文
			currentLang = "zh_cn"
		} else {
			currentLang = "en_us"
		}

		data, _ := f.ReadFile("language/" + currentLang + "/info.json")
		json.Unmarshal(data, &instance)
	}
	return instance
}
