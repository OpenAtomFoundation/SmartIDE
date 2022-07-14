package compose

// 构建时的配置项
type Build struct {
	Context    string                 `yaml:"context"`              // 一个包含Dockerfile的目录，或者一个Git仓库的URL
	Dockerfile string                 `yaml:"dockerfile,omitempty"` // 替代的Dockerfile
	Args       map[string]interface{} `yaml:"args,omitempty"`       // Dockerfile中定义的ARG参数的值
	CacheFrom  []Image                `yaml:"cache_from,omitempty"` // (v3.2+) 缓存的镜像列表
	Labels     map[string]interface{} `yaml:"labels,omitempty"`     // (v3.3+) 目标镜像的元数据信息
	ShmSize    interface{}            `yaml:"shm_size,omitempty"`   // (v3.5+) 构建镜像的/dev/shm分区大小，可以是整数字节数，也可以是字符串
	Target     string                 `yaml:"target,omitempty"`     // (v3.4+) 构建Dockerfile中的指定Stage
}
