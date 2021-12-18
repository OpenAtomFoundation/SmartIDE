package compose

// 暴露的映射的公共接口
type Secret interface {
	IsSecret() bool
}
