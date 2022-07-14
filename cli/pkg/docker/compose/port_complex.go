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
