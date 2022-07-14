package workspace

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/leansoftX/smartide-cli/pkg/ssh_config"
)

func TestGetFullPath(t *testing.T) {

	userName := "john lock"
	filePath := fmt.Sprintf(`C:\Users\%v\.ssh\id_rsa`, userName)
	t.Log(filePath)
	slashStr := filepath.FromSlash(filePath)
	t.Logf("after slash: %v", slashStr)

	if trimmed := strings.TrimSpace(filePath); strings.IndexAny(trimmed, " ") > 0 {
		t.Logf("has space")
	} else {
		t.Logf("no space")
	}

}

func TestRecordToString(t *testing.T) {

	t.Run("should not write identity file when not private key", func(t *testing.T) {
		var record SSHConfigRecord = SSHConfigRecord{
			Host:         "smartide-9",
			HostName:     "localhost",
			User:         "smartide",
			Port:         8022,
			ForwardAgent: "yes",
			IdentityFile: "",
		}

		got := record.ToString()
		t.Logf("record string: \n%v", got)
		want := `
Host smartide-9
  HostName localhost
  User smartide
  ForwardAgent yes
  Port 8022
`

		if got != want {
			t.Errorf("got %q want %q", got, want)
		}
	})

	t.Run("should have identity file when have private key", func(t *testing.T) {
		var record SSHConfigRecord = SSHConfigRecord{
			Host:         "smartide-9",
			HostName:     "localhost",
			User:         "smartide",
			Port:         8022,
			ForwardAgent: "yes",
			IdentityFile: "/temp/testuser/id_rsa",
		}

		got := record.ToString()
		t.Logf("record string: \n%v", got)
		want := `
Host smartide-9
  HostName localhost
  User smartide
  ForwardAgent yes
  Port 8022
  IdentityFile /temp/testuser/id_rsa
`

		if got != want {
			t.Errorf("got %q want %q", got, want)
		}
	})

	t.Run("win-should have double quoted file path when path have spaces", func(t *testing.T) {
		var record SSHConfigRecord = SSHConfigRecord{
			Host:         "smartide-9",
			HostName:     "localhost",
			User:         "smartide",
			Port:         8022,
			ForwardAgent: "yes",
			IdentityFile: "\"C:\\test user\\id_rsa\"",
		}

		got := record.ToString()
		t.Logf("record string: \n%v", got)
		want := `
Host smartide-9
  HostName localhost
  User smartide
  ForwardAgent yes
  Port 8022
  IdentityFile "C:\test user\id_rsa"
`

		if got != want {
			t.Errorf("got %q want %q", got, want)
		}
	})

	t.Run("linux-should have double quoted file path when path have spaces", func(t *testing.T) {
		var record SSHConfigRecord = SSHConfigRecord{
			Host:         "smartide-9",
			HostName:     "localhost",
			User:         "smartide",
			Port:         8022,
			ForwardAgent: "yes",
			IdentityFile: "\"/home/test user/id_rsa\"",
		}

		got := record.ToString()
		t.Logf("record string: \n%v", got)
		want := `
Host smartide-9
  HostName localhost
  User smartide
  ForwardAgent yes
  Port 8022
  IdentityFile "/home/test user/id_rsa"
`

		if got != want {
			t.Errorf("got %q want %q", got, want)
		}
	})
}

func TestPatchKeyValue(t *testing.T) {

	configContent := `Host SmartIDE-8
  HostName localhost
  User smartide
  ForwardAgent yes
  Port 7122`

	config, err := ssh_config.Decode(strings.NewReader(configContent))
	if err != nil {
		t.Error(err)
	}
	hostInfo := searchHostFromConfig(config, "SmartIDE-8")
	nodes := filterNodes(hostInfo)

	configMap := GenerateConfigMap("8", "C:\\users\\root\\id_rsa", 7123)
	patchNodeKeyValue(hostInfo, configMap)

	// refresh nodes addr
	nodes = filterNodes(hostInfo)

	for _, node := range nodes {
		valueInMap, ok := configMap[node.Key]
		if !ok {
			t.Errorf("should exist key %v in configMap", node.Key)
		}

		if valueInMap != node.Value {
			t.Error("values in node not update!")
		}
	}

	for k, v := range configMap {
		if strings.EqualFold(k, "Host") {
			continue
		}
		first := First(nodes, k)
		if first == nil {
			t.Errorf("key %v not exist in nodes", k)
		}
		if first.Value != v {
			t.Errorf("value in node (%v) not equas to value in map(%v)", first.Value, v)
		}
	}

}
