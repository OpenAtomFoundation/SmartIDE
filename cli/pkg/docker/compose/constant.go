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

const (
	// 端口协议
	ProtocolTCP = "tcp" // TCP协议
	ProtocolUDP = "udp" // UDP协议

	// 重启策略
	RestartNo            = "no"             // 默认重启策略，不重启容器
	RestartAlways        = "always"         // 总是重启容器
	RestartOnFailure     = "on-failure"     // 如果退出时的错误码是错误的话，重启容器
	RestartUnlessStopped = "unless-stopped" // 只有停止了要重启容器

	// 镜像标签
	TagDefault = "latest" // 镜像默认Tag

	// 挂载卷的读写模式
	VolumeReadOnly  = "ro" // 只读模式
	VolumeReadWrite = "rw" // 读写模式
)
