/*
 * @Date: 2022-10-27 10:43:57
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-11-01 17:32:58
 * @FilePath: /cli/internal/biz/template/model/model.go
 */

package model

import (
	"path/filepath"

	"github.com/leansoftX/smartide-cli/internal/biz/config"
	golbalModel "github.com/leansoftX/smartide-cli/internal/model"
)

/*
type NewTypeBO struct {
	TypeName string    `json:"typename"`
	SubTypes []SubType `json:"subtype"`
	Commands []string  `json:"command"`
} */

type TemplateTypeInfo struct {
	TypeName string    `json:"typename"`
	SubTypes []SubType `json:"subtype"`
	Commands []string  `json:"command"`
}

type SubType struct {
	Name  string `json:"name"`
	Title string `json:"title"`
	Desc  string `json:"desc"`
	Logo  string `json:"logo"`
}

type SelectedTemplateTypeBo struct {
	// 模板对应的git库 url
	TemplateActualRepoUrl string
	//
	TypeName string
	SubType  string
	Commands []string
}

// 获取模板库本地根目录
func (target SelectedTemplateTypeBo) GetTemplateRootDirAbsolutePath() string {
	return filepath.Join(config.SmartIdeHome, golbalModel.TMEPLATE_DIR_NAME)
}

// 获取模板文件所在相对目录
func (target SelectedTemplateTypeBo) GetTemplateDirRelativePath() string {
	return filepath.Join(target.TypeName, target.SubType)
}

func (target SelectedTemplateTypeBo) IsNil() bool {
	return target.TypeName == "" || target.SubType == "" || target.TemplateActualRepoUrl == ""
}
