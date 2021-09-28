package compose

// 密钥(Long Syntax)
type SecretComplex struct {
	Source string `yaml:"source"`           // 名称
	Target string `yaml:"target,omitempty"` // 文件名
	Uid    string `yaml:"uid,omitempty"`    // 文件UID
	Gid    string `yaml:"gid,omitempty"`    // 文件GID
	Mode   string `yaml:"mode,omitempty"`   // 文件权限
}

// 实现公共接口
func (SecretComplex) IsSecret() bool {
	return true
}
