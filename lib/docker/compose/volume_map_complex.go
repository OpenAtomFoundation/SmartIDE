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
