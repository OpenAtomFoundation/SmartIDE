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

package compose

// 密钥(Long Syntax)
type SecretComplex struct {
	Source string `yaml:"source"`           // 名称
	Target string `yaml:"target,omitempty"` // 文件名
	Uid    string `yaml:"uid,omitempty"`    // 文件UID
	Gid    string `yaml:"gid,omitempty"`    // 文件GID
	Mode   string `yaml:"mode,omitempty"`   // 文件权限
}

// 实现公共接口
func (SecretComplex) IsSecret() bool {
	return true
}
