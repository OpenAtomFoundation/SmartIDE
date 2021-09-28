package compose

// 挂载卷配置
type Volume struct {
	// TODO driver
	// TODO driver_opts
	External bool `yaml:"external,omitempty"` // 是否是外部创建
	// TODO labels
	// TODO name
}
