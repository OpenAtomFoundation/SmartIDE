package compose

import (
	"errors"
	"fmt"
	"strings"
)

// 路径(Short Syntax)
type VolumeMapSimple struct {
	Host      string // 外部主机的路径
	Container string // 内部容器的路径
	Mode      string // 权限
}

// 新建一个路径A到路径B的映射
func NewVolumeMapSimple(hostVolumeMap string, containerVolumeMap string) VolumeMapSimple {
	return VolumeMapSimple{
		Host:      hostVolumeMap,
		Container: containerVolumeMap,
	}
}

// 新建一个相同的路径映射
func NewVolumeMapSimpleSame(volumeMap string) VolumeMapSimple {
	volumeMapSimple := NewVolumeMapSimple(volumeMap, volumeMap)
	volumeMapSimple.Mode = VolumeReadOnly
	return volumeMapSimple
}

// 实现公共接口
func (VolumeMapSimple) IsVolumeMap() bool {
	return true
}

func (m VolumeMapSimple) MarshalYAML() (result interface{}, err error) {
	if len(m.Host) == 0 {
		err = errors.New("docker: simple-volume-map host can not be empty")
		return
	}
	tmp := m.Host
	if len(m.Container) > 0 {
		tmp += fmt.Sprintf(":%s", m.Container)
		if len(m.Mode) > 0 {
			tmp += fmt.Sprintf(":%s", m.Mode)
		}
	}
	result = tmp
	return
}

func (m *VolumeMapSimple) UnmarshalYAML(unmarshal func(interface{}) error) (err error) {
	var origin string
	if err = unmarshal(&origin); err != nil {
		return
	}
	// 拆分
	parts := strings.Split(origin, ":")
	if len(parts) > 3 {
		err = errors.New("docker: simple-volume-map format error")
		return
	}
	m.Host = parts[0]
	if len(parts) > 1 {
		m.Container = parts[1]
	}
	if len(parts) > 2 {
		m.Mode = parts[2]
	}
	// 校验
	if len(m.Host) == 0 {
		err = errors.New("docker: simple-volume-map format error")
		return
	}
	if len(m.Container) == 0 && len(m.Mode) > 0 {
		err = errors.New("docker: simple-volume-map format error")
		return
	}
	return
}
