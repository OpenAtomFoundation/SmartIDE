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
func (target SelectedTemplateTypeBo) GetTemplateLocalRootDirAbsolutePath() string {
	return filepath.Join(config.SmartIdeHome, golbalModel.TMEPLATE_DIR_NAME)
}

// 获取模板库本地绝对目录
func (target SelectedTemplateTypeBo) GetTemplateLocalDirAbsolutePath() string {
	return filepath.Join(target.GetTemplateLocalRootDirAbsolutePath(), target.GetTemplateDirRelativePath())
}

// 获取模板文件所在相对目录
func (target SelectedTemplateTypeBo) GetTemplateDirRelativePath() string {
	return filepath.Join(target.TypeName, target.SubType)
}

func (target SelectedTemplateTypeBo) IsNil() bool {
	return target.TypeName == "" || target.SubType == "" || target.TemplateActualRepoUrl == ""
}
