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

// 端口(Long Syntax)
type PortComplex struct {
	Target    string `yaml:"target"`             // 容器内部端口号
	Published string `yaml:"published"`          // 暴露的端口号
	Protocol  string `yaml:"protocol,omitempty"` // 传输协议
	// TODO mode
}

// 实现公共接口
func (PortComplex) IsPort() bool {
	return true
}
