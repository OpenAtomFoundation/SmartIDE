package compose

// 网络
type NetworkMap struct {
	Aliases []string `yaml:"aliases,omitempty"` // 别名列表
	// TODO ipv4_address
	// TODO ipv6_address
	// TODO ipam
}
