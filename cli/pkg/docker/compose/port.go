package compose

// 暴露的端口映射的公共接口
type Port interface {
	IsPort() bool
}
