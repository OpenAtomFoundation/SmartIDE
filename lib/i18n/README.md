# i18n
golang i18n, golang实现的多语言解析使用  
internationalization （国际化）简称：i18n，因为在i和n之间还有18个字符，localization（本地化 ），简称L10n。 一般用语言_地区的形式表示一种语言，如：zh_CN表示简体中文。

## 安装
- go mod
```shell script
require github.com/gohouse/i18n master
```
- go get
```shell script
go get github.com/gohouse/i18n
```

## 使用
可以查看包内的示例代码: [https://github.com/gohouse/i18n/blob/master/examples/demo.go](https://github.com/gohouse/i18n/blob/master/examples/demo.go)  

添加语言文件
```shell script
# 创建文件夹
mkdir -p /go/src/gopro/language/zh_cn /go/src/gopro/language/en-us

# 编写中文语言文件
cat >>~/go/src/gopro/language/zh_cn/error.json<<EOF
{
  "test": "测试",
  "params_format_error": "参数格式有误",
  "params_missing": "参数缺失",
  "err2": {
    "aa": "aaxx",
    "bb": {
      "cc": "cc"
    }
  }
}
EOF

# 编写英文语言文件
cat >>~/go/src/gopro/language/en-us/error.json<<EOF
{
  "test": "just a test",
  "params_format_error": "Incorrect parameters format",
  "params_missing": "Missing parameters",
  "err2": {
    "aa": "aaxx",
    "bb": {
      "cc": "cc"
    }
  }
}
EOF
```
编写go代码文件`~/go/src/gopro/demo.go`
```go
package main

import (
	"fmt"
	"github.com/gohouse/i18n"
	// 这里不要忘记引入默认的json驱动
	_ "github.com/gohouse/i18n/parser_json"
)

func main() {
	lang := i18n.NewI18n(
		// 这里指定语言文件路径
		i18n.LangDirectory("/go/src/github.com/gohouse/i18n/examples/language"),

		// 这里如果不i设置, 则默认使用zh_cn
		//i18n.DefaultLang("zh_cn"),

		// 这里如果不i设置, 则默认使用 json,可以自定义解析器和配置文件格式
		//i18n.DefaultParser("json"),
	)

	// 加载error.json文件内的具体配置项, 多级加载, 使用.连接
	test := lang.LoadWithDefault("error.test")
	test2 := lang.LoadWithDefault("error.err2.bb.cc")

	fmt.Println(test)
	fmt.Println(test2)
}
```
结果
```shell script
测试
cc
```

## 额外说明
i18n默认提供了json解析器, 同时, 提供了解析器接口, 可以自由定制其他格式的解析器, 如yml,ini,toml等


## 附件 - 国际化开发的各国语言标识
语言标识|国家地区  
---|:---  
zh_CN  |  简体中文(中国)  
zh_TW  |  繁体中文(台湾地区)  
zh_HK  |  繁体中文(香港)  
en_HK  |  英语(香港)  
en_US  |  英语(美国)  
en_GB  |  英语(英国)  
en_WW  |  英语(全球)  
en_CA  |  英语(加拿大)  
en_AU  |  英语(澳大利亚)  
en_IE  |  英语(爱尔兰)  
en_FI  |  英语(芬兰)  
fi_FI  |  芬兰语(芬兰)  
en_DK  |  英语(丹麦)  
da_DK  |  丹麦语(丹麦)  
en_IL  |  英语(以色列)  
he_IL  |  希伯来语(以色列)  
en_ZA  |  英语(南非)  
en_IN  |  英语(印度)  
en_NO  |  英语(挪威)  
en_SG  |  英语(新加坡)  
en_NZ  |  英语(新西兰)  
en_ID  |  英语(印度尼西亚)  
en_PH  |  英语(菲律宾)  
en_TH  |  英语(泰国)  
en_MY  |  英语(马来西亚)  
en_XA  |  英语(阿拉伯)  
ko_KR  |  韩文(韩国)  
ja_JP  |  日语(日本)  
nl_NL  |  荷兰语(荷兰)  
nl_BE  |  荷兰语(比利时)  
pt_PT  |  葡萄牙语(葡萄牙)  
pt_BR  |  葡萄牙语(巴西)  
fr_FR  |  法语(法国)  
fr_LU  |  法语(卢森堡)  
fr_CH  |  法语(瑞士)  
fr_BE  |  法语(比利时)  
fr_CA  |  法语(加拿大)  
es_LA  |  西班牙语(拉丁美洲)  
es_ES  |  西班牙语(西班牙)  
es_AR  |  西班牙语(阿根廷)  
es_US  |  西班牙语(美国)  
es_MX  |  西班牙语(墨西哥)  
es_CO  |  西班牙语(哥伦比亚)  
es_PR  |  西班牙语(波多黎各)  
de_DE  |  德语(德国)  
de_AT  |  德语(奥地利)  
de_CH  |  德语(瑞士)  
ru_RU  |  俄语(俄罗斯)  
it_IT  |  意大利语(意大利)  
el_GR  |  希腊语(希腊)  
no_NO  |  挪威语(挪威)  
hu_HU  |  匈牙利语(匈牙利)  
tr_TR  |  土耳其语(土耳其)  
cs_CZ  |  捷克语(捷克共和国)  
sl_SL  |  斯洛文尼亚语   
pl_PL  |  波兰语(波兰)  
sv_SE  |  瑞典语(瑞典)  
es_CL  |  西班牙语 (智利)  