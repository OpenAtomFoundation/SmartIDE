package compose

// 暴露的映射的公共接口
type YmlSecret struct {
	File     string `yaml:"file,omitempty"`     // 文件路径
	External string `yaml:"external,omitempty"` // 是否已存在，存在不需要再创建。
	Name     string `yaml:"name,omitempty"`     // (v3.5+) 名称
}
