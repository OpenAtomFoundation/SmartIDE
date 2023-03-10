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

// 挂载卷
type VolumeMapComplex struct {
	Type     string `yaml:"type"`                // 挂载类型
	Source   string `yaml:"source"`              // 外部的源地址
	Target   string `yaml:"target"`              // 容器内的目标地址
	ReadOnly string `yaml:"read_only,omitempty"` // 只读标志
	// TODO bind
	// TODO volume
	// TODO tmpfs
	// TODO consistency
}

// 实现公共接口
func (VolumeMapComplex) IsVolumeMap() bool {
	return true
}
