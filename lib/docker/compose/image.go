package compose

import (
	"errors"
	"fmt"
	"strings"
)

// 镜像
type Image struct {
	Name string
	Tag  string
}

// 新建一个镜像
func NewImage(name string, tag string) Image {
	return Image{
		Name: name,
		Tag:  tag,
	}
}

func (m Image) MarshalYAML() (result interface{}, err error) {
	if len(m.Name) == 0 {
		err = errors.New("docker: image name can not be empty")
		return
	}
	if len(m.Tag) > 0 {
		result = fmt.Sprintf("%s:%s", m.Name, m.Tag)
	} else {
		result = m.Name
	}
	return
}

func (m *Image) UnmarshalYAML(unmarshal func(interface{}) error) (err error) {
	var origin string
	if err = unmarshal(&origin); err != nil {
		return
	}
	parts := strings.Split(origin, ":")
	if len(parts) > 2 {
		err = errors.New("docker: image format error")
		return
	}
	m.Name = parts[0]
	if len(parts) > 1 {
		m.Tag = parts[1]
	}
	return
}
