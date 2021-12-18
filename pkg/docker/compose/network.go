package compose

// 网络配置
type Network struct {
	// TODO driver
	// TODO driver_opts
	// TODO attachable
	// TODO enable_ipv6
	// TODO ipam
	// TODO internal
	// TODO labels
	External bool `yaml:"external,omitempty"` // 是否是外部创建
	// TODO name
}
