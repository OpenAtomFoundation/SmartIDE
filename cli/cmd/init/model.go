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

package init

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
