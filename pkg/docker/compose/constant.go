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
