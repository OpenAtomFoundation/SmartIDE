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

// 构建时的配置项
type Build struct {
	Context    string                 `yaml:"context"`              // 一个包含Dockerfile的目录，或者一个Git仓库的URL
	Dockerfile string                 `yaml:"dockerfile,omitempty"` // 替代的Dockerfile
	Args       map[string]interface{} `yaml:"args,omitempty"`       // Dockerfile中定义的ARG参数的值
	CacheFrom  []Image                `yaml:"cache_from,omitempty"` // (v3.2+) 缓存的镜像列表
	Labels     map[string]interface{} `yaml:"labels,omitempty"`     // (v3.3+) 目标镜像的元数据信息
	ShmSize    interface{}            `yaml:"shm_size,omitempty"`   // (v3.5+) 构建镜像的/dev/shm分区大小，可以是整数字节数，也可以是字符串
	Target     string                 `yaml:"target,omitempty"`     // (v3.4+) 构建Dockerfile中的指定Stage
}
