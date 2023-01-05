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

package compose

import (
	"errors"
)

// 密钥(Short Syntax)
type SecretSimple struct {
	Source string // 名称
}

// 新建一个密钥
func NewSecretSimple(source string) SecretSimple {
	return SecretSimple{
		Source: source,
	}
}

// 实现公共接口
func (SecretSimple) IsSecret() bool {
	return true
}

func (m SecretSimple) MarshalYAML() (result interface{}, err error) {
	if len(m.Source) == 0 {
		err = errors.New("docker: simple-secret source can not be empty")
		return
	}
	result = m.Source
	return
}

func (m *SecretSimple) UnmarshalYAML(unmarshal func(interface{}) error) (err error) {
	var origin string
	if err = unmarshal(&origin); err != nil {
		return
	}
	m.Source = origin
	// 校验
	if len(m.Source) == 0 {
		err = errors.New("docker: simple-secret format error")
		return
	}
	return
}
