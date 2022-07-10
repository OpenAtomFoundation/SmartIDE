package workspace

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	. "github.com/ahmetb/go-linq/v3"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/leansoftX/smartide-cli/pkg/ssh_config"
)

type SSHConfigRecord struct {
	Host         string
	HostName     string
	User         string
	Port         int
	ForwardAgent string
	IdentityFile string
}

// ControlledConfigKeys 定义可进行修改的key的范围，
// 如果1个key在config文件里出现, 而configRecord中对应的key值为空或者没有该key，
// - 如果key在ControlledKeys中，说明要把key从配置中移除
// - 如果key不在ControlledKeys中，说明时用户自定义的配置，不用对key进行任何操作
var ControlledConfigKeys = [6]string{
	"Host",
	"HostName",
	"User",
	"Port",
	"ForwardAgent",
	"IdentityFile",
}

type SSHConfigMap map[string]string

// short alias for log
var logger = common.SmartIDELog

// UpdateSSHConfig 修改SSH文件的入口函数，需要传入workspace对象
func (workspaceInfo WorkspaceInfo) UpdateSSHConfig() {
	// check workspace id
	if workspaceInfo.ID == "" {
		logger.Warning("workspaceID is empty, skip to operate ssh config")
		return
	}

	logger.DebugF("find port map info....\n")
	portMapInfo, err := workspaceInfo.Extend.Ports.Find("tools-ssh")
	common.CheckError(err)
	sshPort := portMapInfo.GetSSHPortAtLocalHost()
	logger.InfoF("workspaceID: %v, ssh port : %v\n", workspaceInfo.ID, sshPort)

	// check config file
	logger.DebugF("find ssh file path....")

	configPath, err := getSSHConfigPath()
	common.CheckError(err)
	logger.DebugF("ssh file path: %v\n", configPath)

	IdentityFile := getIdentityFile()

	logger.DebugF("private key file path: %v\n", IdentityFile)

	// get file/ create file
	logger.Debug("ensure config file exist...")
	err = ensureSSHConfigFileExist(configPath)
	common.CheckError(err)

	// check host is exist?
	logger.Debug("decoding file content to ssh config...")
	file, _ := os.Open(configPath)

	cfg, _ := ssh_config.Decode(file)

	logger.DebugF("decoded config: %v", cfg.String())
	configMap := GenerateConfigMap(workspaceInfo.ID, IdentityFile, sshPort)
	record := configMap.ConvertToRecord()
	logger.Debug("config record:", record.ToString())

	logger.DebugF("check host %v is exist in config file...\n", record.Host)
	isHostExistInConfig := isHostExistInConfig(cfg, record.Host)
	logger.DebugF("check result:%v\n", isHostExistInConfig)

	if !isHostExistInConfig {
		// append to ssh config
		appendRecord(record, configPath)
	} else {
		// update config
		updateRecord(record, cfg, configMap, configPath)
	}
}
func (record SSHConfigRecord) ToString() string {
	// 不要随意修改下面的模板， 里面包含了首尾换行和首行2个空格缩进的格式
	var templateText string

	if record.IdentityFile != "" {
		templateText = `
Host {{.Host}}
  HostName {{.HostName}}
  User {{.User}}
  ForwardAgent {{.ForwardAgent}}
  Port {{.Port}}
  IdentityFile {{.IdentityFile}}
`
	} else {
		templateText = `
Host {{.Host}}
  HostName {{.HostName}}
  User {{.User}}
  ForwardAgent {{.ForwardAgent}}
  Port {{.Port}}
`
	}
	var tpl bytes.Buffer
	parser, _ := template.New("ssh-config").Parse(templateText)
	err := parser.Execute(&tpl, record)
	common.CheckError(err)
	return tpl.String()
}

func (configMap SSHConfigMap) ConvertToRecord() SSHConfigRecord {
	port, err := strconv.Atoi(configMap["Port"])
	record := SSHConfigRecord{
		Host:         configMap["Host"],
		HostName:     configMap["HostName"],
		User:         configMap["User"],
		Port:         port,
		ForwardAgent: configMap["ForwardAgent"],
		IdentityFile: configMap["IdentityFile"],
	}
	common.CheckError(err)
	return record
}

func updateRecord(record SSHConfigRecord, cfg *ssh_config.Config, configMap map[string]string, configPath string) {
	logger.DebugF("search host %v from config file...\n", record.Host)
	host := searchHostFromConfig(cfg, record.Host)
	if host == nil {
		logger.WarningF("can not find host: %v in ssh config file\n", record.Host)
		return
	}

	// patch nodes
	logger.DebugF("patch host config %v ...\n", record.Host)
	patchNodeKeyValue(host, configMap)

	cfgStr := cfg.String()

	writer, err := os.OpenFile(configPath, os.O_TRUNC|os.O_CREATE|os.O_RDWR, 0755)
	common.CheckError(err)
	_, err = writer.WriteString(cfgStr)
	_ = writer.Close()
	common.CheckError(err)
	logger.Info("ssh config update success, your can view it in vscode remote target list")
}

func appendRecord(record SSHConfigRecord, configPath string) {

	logger.InfoF("append host: %v to .ssh/config...\n", record.Host)
	file, err := os.OpenFile(configPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		logger.WarningF("error when write setting to ssh config file, %v\n", err.Error())
		return
	}
	str := record.ToString()
	_, err = file.WriteString(str)
	common.CheckError(err)
	_ = file.Close()

	logger.Info("update config success, your can view it in VSCode's remote SSH target list")
}

func GenerateConfigMap(workspaceId string, IdentityFile string, sshPort int) SSHConfigMap {

	sshUserName := "smartide" //  默认值：smartide
	host := "localhost"       // host 默认就是localhost，所有模式下，要配置的ssh端口号都是针对本机
	configMap := SSHConfigMap{}
	configMap["Host"] = fmt.Sprintf("SmartIDE-%v", workspaceId)
	configMap["HostName"] = host
	configMap["User"] = sshUserName
	configMap["Port"] = strconv.Itoa(sshPort)
	configMap["ForwardAgent"] = "yes"
	// 带有空格的路径，需要在两侧用双引号括起来
	IdentityFile = normalizePathString(IdentityFile)
	configMap["IdentityFile"] = IdentityFile
	return configMap
}

func normalizePathString(path string) string {
	if path == "" {
		return path
	}

	if isPathContainSpace(path) {
		return fmt.Sprintf(`"%v"`, path)
	}

	return strings.TrimSpace(path)

}

func ensureSSHConfigFileExist(configPath string) error {
	_, err := os.Stat(configPath)

	if err != nil {
		logger.WarningF("config file: %v not exist, will init a new config file\n", configPath)
		//_, err := os.Create(configPath)
		err := common.FS.CreateOrOverWrite(configPath, "")
		if err != nil {
			common.CheckError(err)
		}
		logger.InfoF("config file: %v created\n", configPath)
	}
	return nil
}

func filterNodes(host *ssh_config.Host) []*ssh_config.KV {

	nodesKv := []*ssh_config.KV{}

	for _, node := range host.Nodes {
		switch t := node.(type) {
		case *ssh_config.KV:
			nodesKv = append(nodesKv, t)
		default:
			continue
		}
	}
	return nodesKv
}

func patchNodeKeyValue(host *ssh_config.Host, recordMap SSHConfigMap) {

	nodesKv := []*ssh_config.KV{}

	for _, node := range host.Nodes {
		switch t := node.(type) {
		case *ssh_config.KV:
			nodesKv = append(nodesKv, t)
		default:
			continue
		}
	}

	// patch key/value from record map
	for k, v := range recordMap {
		if strings.EqualFold(k, "Host") {
			continue
		}
		first := First(nodesKv, k)
		if first == nil {
			// append node
			newNode := &ssh_config.KV{
				Key:          k,
				Value:        v,
				LeadingSpace: 2,
			}
			logger.DebugF("append key %v to host %v...\n", k, recordMap["Host"])
			host.Nodes = append(host.Nodes, newNode)
		} else {

			if v == "" {
				// value为空字符串时 直接用空字符串会导致VSCode解析出现问题
				// 所以用双引号包起来 key: "",
				first.Value = `""`
			} else {
				first.Value = v
			}
			logger.DebugF("update key %v to new value %v...\n", k, first.Value)
		}

	}
}

func First(nodes []*ssh_config.KV, key string) *ssh_config.KV {
	for _, node := range nodes {
		if !strings.EqualFold(key, node.Key) {
			continue
		}
		return node
	}
	return nil
}

func getSSHConfigPath() (string, error) {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	fullPath := filepath.Join(userHomeDir, ".ssh", "config")

	return fullPath, nil
}

// 获取用户home下的 id_rsa 文件路径
func getIdentityFile() string {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	fullPath := filepath.Join(userHomeDir, ".ssh", "id_rsa")
	_, err = os.Stat(fullPath)
	if err != nil {
		logger.DebugF("user key file not found at: %v, will skip write identity file config.", fullPath)
		return ""
	}

	return fullPath
}

func isHostExistInConfig(cfg *ssh_config.Config, host string) bool {
	configKey := "HostName"
	hostValue, err := cfg.Get(host, configKey)
	if err != nil {
		common.CheckError(err)
		return false
	}
	return hostValue != ""
}

func searchHostFromConfig(cfg *ssh_config.Config, hostStr string) *ssh_config.Host {

	count := From(cfg.Hosts).CountWith(func(h interface{}) bool {
		hst := h.(*ssh_config.Host)
		return hst.Matches(hostStr) && len(hst.Nodes) > 1

	})

	if count <= 0 {
		return nil
	}

	first := From(cfg.Hosts).FirstWith(func(h interface{}) bool {
		hst := h.(*ssh_config.Host)
		return hst.Matches(hostStr) && len(hst.Nodes) > 1
	})

	switch t := first.(type) {
	case *ssh_config.Host:
		return t
	default:
		return nil
	}

}

// 路径中是否包含空格，for: "/user/super root/.ssh" => true
func isPathContainSpace(pathString string) bool {
	if pathString == "" {
		return false
	}
	if trimmed := strings.TrimSpace(pathString); strings.IndexAny(trimmed, " ") > 0 {
		return true
	}
	return false
}
