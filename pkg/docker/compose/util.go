package compose

import (
	yaml "gopkg.in/yaml.v2"
)

// 编码为Yaml格式
func MarshalYaml(obj interface{}) string {
	v, _ := yaml.Marshal(obj)
	return string(v)
}

// 解码Yaml文本
func UnmarshalYaml(content string, obj interface{}) (err error) {
	err = yaml.Unmarshal([]byte(content), obj)
	return
}
