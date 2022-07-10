package compose

import (
	"errors"
	"fmt"
	"strings"
)

// 端口(Short Syntax)
type PortSimple struct {
	Host      string // 外部主机的端口号
	Container string // 内部容器的端口号
	Protocol  string // 协议
}

// 新建一个端口A到端口B的映射
func NewPortSimple(hostPort int, containerPort int) PortSimple {
	return PortSimple{
		Host:      fmt.Sprintf("%d", hostPort),
		Container: fmt.Sprintf("%d", containerPort),
	}
}

// 新建一个相同端口映射
func NewPortSimpleSame(port int) PortSimple {
	return NewPortSimple(port, port)
}

// 实现公共接口
func (PortSimple) IsPort() bool {
	return true
}

func (m PortSimple) MarshalYAML() (result interface{}, err error) {
	if len(m.Host) == 0 {
		err = errors.New("docker: simple-port host can not be empty")
		return
	}
	tmp := m.Host
	if len(m.Container) > 0 {
		tmp += fmt.Sprintf(":%s", m.Container)
		if len(m.Protocol) > 0 {
			tmp += fmt.Sprintf("/%s", m.Protocol)
		}
	}
	result = tmp
	return
}

func (m *PortSimple) UnmarshalYAML(unmarshal func(interface{}) error) (err error) {
	var origin string
	if err = unmarshal(&origin); err != nil {
		return
	}
	// 拆分协议部分
	parts, remain := strings.Split(origin, "/"), ""
	if len(parts) > 2 {
		err = errors.New("docker: simple-port format error")
		return
	}
	if len(parts) > 1 {
		m.Protocol = parts[1]
	}
	remain = parts[0]
	// 拆分主机和容器端口
	loc := strings.LastIndex(remain, ":")
	if loc < 0 {
		m.Host, m.Container = remain, ""
	} else if loc != len(remain)-1 {
		m.Host, m.Container = remain[0:loc], remain[loc+1:]
	} else {
		m.Host, m.Container = remain[0:loc], ""
	}
	// 校验
	if len(m.Host) == 0 {
		err = errors.New("docker: simple-port format error")
		return
	}
	if len(m.Container) == 0 && len(m.Protocol) > 0 {
		err = errors.New("docker: simple-port format error")
		return
	}
	return
}
