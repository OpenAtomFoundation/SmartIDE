/*
 * @Date: 2022-04-20 10:54:55
 * @LastEditors: kenan
 * @LastEditTime: 2022-04-26 09:41:20
 * @FilePath: /smartide-cli/cmd/new/model.go
 */

package new

type NewTypeBO struct {
	TypeName string    `json:"typename"`
	SubTypes []SubType `json:"subtype"`
	Commands []string  `json:"command"`
}

type TemplateTypeBo struct {
	TypeName string
	SubType  string
	Commands []string
}

type SubType struct {
	Name  string `json:"name"`
	Title string `json:"title"`
	Desc  string `json:"desc"`
	Logo  string `json:"logo"`
}
