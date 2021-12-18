package tunnel

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/leansoftX/smartide-cli/internal/apk/i18n"
	"github.com/leansoftX/smartide-cli/internal/model"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"golang.org/x/crypto/ssh"
)

// 转发的端口
type AutoTunnelMultipleOptions struct {
	PortMappings []PortMapping
}

//
type PortMapping struct {
	LocalPortDesc    string
	OriginLocalPort  int
	CurrentLocalPort int
	ContainerPort    int
	PortMapType      PortMapTypeEnum
}

//
func (instance *AutoTunnelMultipleOptions) AppendPortMapping(
	mapType PortMapTypeEnum, orginLocalPort int, currentLocalPort int, localPortDesc string, containerPort int) {
	portMapping := PortMapping{
		OriginLocalPort:  orginLocalPort,
		CurrentLocalPort: currentLocalPort,
		LocalPortDesc:    localPortDesc,
		ContainerPort:    containerPort,
		PortMapType:      mapType,
	}
	instance.PortMappings = append(instance.PortMappings, portMapping)
}

//
type PortMapTypeEnum string

//
const (
	PortMapInfo_None        PortMapTypeEnum = ""
	PortMapInfo_Full        PortMapTypeEnum = "full"
	PortMapInfo_OnlyLabel   PortMapTypeEnum = "label"
	PortMapInfo_OnlyCompose PortMapTypeEnum = "compose"
)

// 自动端口转发
func AutoTunnel(clientConn *ssh.Client, options AutoTunnelMultipleOptions) {

	var tunneledContainerPorts []int // 已打通隧道的端口列表

	go func() {
		for {
			//TOOD: 完善端口被删除的情况，又开放的情况
			// 获取开放的端口
			scanContainerPorts := scanServerPorts(clientConn)       // 扫描出来的端口
			var mapping map[string]string = make(map[string]string) //

			// 发现新的端口
			for _, scanContainerPort := range scanContainerPorts {
				// 是否已经映射
				if common.Contains4Int(tunneledContainerPorts, scanContainerPort) {
					continue
				}

				// 当前端口已经在compose中定义
				localPortStr := fmt.Sprint(scanContainerPort)
				hasMappingWithCompose := false
				for _, item := range options.PortMappings {
					if scanContainerPort == item.ContainerPort {
						hasMappingWithCompose = true
						if item.LocalPortDesc != "" {
							localPortStr = fmt.Sprintf("%v %v", item.CurrentLocalPort, item.LocalPortDesc)
						}
						if item.PortMapType == PortMapInfo_Full || item.PortMapType == PortMapInfo_OnlyCompose {
							common.SmartIDELog.InfoF(i18n.GetInstance().Common.Info_port_is_binding, localPortStr)

							if item.CurrentLocalPort != item.OriginLocalPort {
								common.SmartIDELog.InfoF(i18n.GetInstance().Common.Info_port_binding_result2, localPortStr, item.CurrentLocalPort, item.ContainerPort)
							} else {
								common.SmartIDELog.InfoF(i18n.GetInstance().Common.Info_port_binding_result, localPortStr, item.ContainerPort)
							}
						}

						break
					}
				}

				// 如果没有在compose文件中定义，那么就是要映射端口
				if !hasMappingWithCompose {
					dynamicContainerPort := scanContainerPort // 动态的端口
					localPort := common.CheckAndGetAvailableLocalPort(scanContainerPort, 100)

					if localPort != dynamicContainerPort {
						common.SmartIDELog.InfoF(i18n.GetInstance().Common.Info_port_binding_result2, localPortStr, localPort, dynamicContainerPort)
					} else {
						common.SmartIDELog.InfoF(i18n.GetInstance().Common.Info_port_binding_result, localPortStr, dynamicContainerPort)
					}

					mapping["localhost:"+strconv.Itoa(localPort)] = "localhost:" + strconv.Itoa(dynamicContainerPort)
				}

				// 记录到已经映射
				tunneledContainerPorts = append(tunneledContainerPorts, scanContainerPort)
			}

			// 端口转发（只转发新增的端口）
			if len(mapping) > 0 {

				err := TunnelMultiple(clientConn, mapping) // 执行绑定
				if err != nil {
					common.SmartIDELog.Fatal(err)
				}

			}

			// 避免太多频繁
			time.Sleep(time.Second * 3)
		}
	}()

}

// 通过ssh通道去扫描容器内都除了22端口，都开放了哪些
func scanServerPorts(clientConn *ssh.Client) []int {
	session, err := clientConn.NewSession()
	if err != nil {
		clientConn.Close()
	}

	// 在ssh主机上执行命令
	/* 	tmpSesssion, _ := clientConn.NewSession()
	   	tmpSesssion.Run(`apt update
	   					apt install net-tools`) */
	// https://www.tecmint.com/install-netstat-in-linux/
	cmd := `netstat -tunlp`
	out, err := session.CombinedOutput(cmd)
	outContent := string(out)
	common.CheckError(err)

	// 分析命令行打印的结果，提取端口
	var ports []int
	//common.SmartIDELog.Debug(outContent)
	netstatInfoArray := parseCmdOutput(outContent)
	for _, netstatInfo := range netstatInfoArray {
		if netstatInfo.Port != model.CONST_Container_WebIDEPort &&
			netstatInfo.Host != "127.0.0.1" &&
			netstatInfo.Port != model.CONST_Container_SSHPort &&
			netstatInfo.PID != "" &&
			!common.Contains4Int(ports, netstatInfo.Port) {

			// 防止重复的debug
			if statOutput != outContent {
				statOutput = outContent
				common.SmartIDELog.Debug(outContent)
			}

			ports = append(ports, netstatInfo.Port)
		}
	}

	return ports
}

var statOutput string = ""

// 解析netstat的运行结果
func parseCmdOutput(cmdOutput string) (result []NetstatInfo) {
	startInex := strings.Index(cmdOutput, "Active Internet connections")
	if startInex < 0 {
		return nil
	}
	strArray := strings.Split(cmdOutput[startInex+1:], "\n")

	header := strArray[1]
	indexs := []int{
		strings.Index(header, "Proto"),
		strings.Index(header, "Recv-Q"),
		strings.Index(header, "Send-Q"),
		strings.Index(header, "Local Address"),
		strings.Index(header, "Foreign Address"),
		strings.Index(header, "State"),
		strings.Index(header, "PID/Program name"),
	}

	for _, index := range indexs {
		if index < 0 {
			common.SmartIDELog.Error(cmdOutput)
		}
	}

	for _, str := range strArray[2:] {

		if strings.Contains(str, "Not all processes could be identified, non-owned process info") ||
			strings.Contains(str, "you would have to be root to see it all.") {
			continue
		}

		if len(strings.ReplaceAll(str, " ", "")) == 0 {
			continue
		}
		if indexs[6] > len(str) {
			common.SmartIDELog.Debug(cmdOutput)
			common.SmartIDELog.Error(str)
		}
		tmp := strings.ReplaceAll(str[indexs[6]:], " ", "")
		pid, programName := "", ""
		if tmp != "-" {
			tmpArray := strings.Split(tmp, "/")
			pid = tmpArray[0]
			programName = tmpArray[1]
		}
		localAddress := strings.ReplaceAll(str[indexs[3]:indexs[4]], " ", "")
		lastColonIndex := strings.LastIndex(localAddress, ":")
		host := localAddress[0:lastColonIndex]
		port, _ := strconv.Atoi(localAddress[lastColonIndex+1:]) // 从最后一个“:”开始截取
		info := NetstatInfo{
			Proto:          strings.ReplaceAll(str[0:indexs[1]], " ", ""),
			RecvQ:          strings.ReplaceAll(str[indexs[1]:indexs[2]], " ", ""),
			SendQ:          strings.ReplaceAll(str[indexs[2]:indexs[3]], " ", ""),
			LocalAddress:   localAddress,
			Host:           host,
			Port:           port,
			ForeignAddress: strings.ReplaceAll(str[indexs[4]:indexs[5]], " ", ""),
			State:          strings.ReplaceAll(str[indexs[5]:indexs[6]], " ", ""),
			PID:            pid,
			ProgramName:    programName,
		}
		result = append(result, info)
	}
	return result
}

//
/*
Proto Recv-Q Send-Q Local Address           Foreign Address         State       PID/Program name
tcp        0      0 127.0.0.11:38053        0.0.0.0:*               LISTEN      -
tcp        0      0 0.0.0.0:22              0.0.0.0:*               LISTEN      13/sshd
tcp        0      0 0.0.0.0:3000            0.0.0.0:*               LISTEN      14/node
tcp        0      0 :::22                   :::*                    LISTEN      13/sshd
udp        0      0 127.0.0.11:56284        0.0.0.0:*                           -
*/
type NetstatInfo struct {
	Host           string
	Port           int
	Proto          string
	RecvQ          string
	SendQ          string
	LocalAddress   string
	ForeignAddress string
	State          string
	PID            string
	ProgramName    string
}

// 转发指定SSH服务器的多个端口 到本地（Localhost）
//
func TunnelMultiple(clientConn *ssh.Client, mapping map[string]string) error {

	pipe := func(writer, reader net.Conn) {
		defer writer.Close()
		defer reader.Close()

		_, err := io.Copy(writer, reader)
		if err != nil {
			common.SmartIDELog.Warning(err.Error())
		}
	}

	//TODO: 这里可能会有问题
	for local, remote := range mapping {

		go func(local, remote string) {

			listener, err := net.Listen("tcp", local)
			if err != nil {
				common.SmartIDELog.Warning("failed to dial to remote: ", err.Error())
			}
			for {
				here, err := listener.Accept()
				if err != nil {
					common.SmartIDELog.Warning("failed to dial to remote: ", err.Error())
				}
				go func(here net.Conn) {
					there, err := clientConn.Dial("tcp", remote)
					if err != nil {
						common.SmartIDELog.Warning(err.Error())
					}
					go pipe(there, here)
					go pipe(here, there)
				}(here)
			}

		}(local, remote)

	}

	return nil
}
