/*
 * @Author: kenan
 * @Date: 2021-10-13 15:31:52
 * @LastEditors: kenan
 * @LastEditTime: 2021-11-17 10:43:59
 * @Description: file content
 */

package lib

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/docker/docker/client"
	"github.com/leansoftX/smartide-cli/lib/common"
	"github.com/leansoftX/smartide-cli/lib/docker/compose"
	"github.com/leansoftX/smartide-cli/lib/i18n"
)

var PassPhrase string

func ConfigGitByDockerExec() {

	//打开文件io流
	cmd := exec.Command("git", "config", "--list")
	cmd.Stderr = os.Stderr
	out, cmdErr := cmd.Output()
	common.CheckError(cmdErr)

	gitconfigs := string(out)
	if gitconfigs == "" {
		common.SmartIDELog.Error(i18n.GetInstance().Config.Error.Gitconfig_not_exit)
	}
	s := bufio.NewScanner(strings.NewReader(gitconfigs))
	for s.Scan() {
		//以=分割,前面为key,后面为value
		var str = s.Text()
		var index = strings.Index(str, "=")
		var key = str[0:index]
		var value = str[index+1:]

		var yamlFileCongfig YamlFileConfig
		yamlFileCongfig.GetConfig()

		var servicename = yamlFileCongfig.Workspace.DevContainer.ServiceName
		cmdStr := fmt.Sprint("docker exec ", servicename, " git config --global ", key, " ", value)
		cmd := exec.Command("/bin/sh", "-c", cmdStr)

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if cmdErr := cmd.Run(); cmdErr != nil {
			common.SmartIDELog.Fatal(cmdErr)
		}
	}
	err := s.Err()
	if err != nil {
		log.Fatal(err)
	}

}

func SSHVolumesConfig(sshKey string, isVmCommand bool, service compose.Service) (ser compose.Service) {
	var configPaths []string
	if sshKey == "true" {

		if isVmCommand || runtime.GOOS != "windows" {
			configPaths = []string{"$HOME/.ssh:/root/.ssh"}
		} else {
			configPaths = []string{fmt.Sprint(os.Getenv("USERPROFILE"), "/.ssh:/root/.ssh")}

		}
		if configPaths != nil {
			service.Volumes = append(service.Volumes, configPaths...)
		}
	}
	return service
}

func GitConfig(configGit string, isVmCommand bool, containerName string, cli *client.Client, service compose.Service, sshRemote common.SSHRemote) (ser compose.Service) {
	// 获取本机git config 内容
	// git config --list --show-origin
	if configGit == "true" {
		var configStr string
		var err error
		cmd := exec.Command("git", "config", "--list")
		cmd.Stderr = os.Stderr
		var out []byte
		out, err = cmd.Output()
		configStr = string(out)
		var isConfig bool

		if err != nil {
			common.SmartIDELog.Fatal(err)
		}

		if configStr == "" {
			return
		}
		gitconfigs := strings.ReplaceAll(configStr, "file:", "")
		s := bufio.NewScanner(strings.NewReader(gitconfigs))
		for s.Scan() {
			//以=分割,前面为key,后面为value
			var str = s.Text()
			var index = strings.Index(str, "=")
			var key = str[0:index]
			var value = str[index+1:]
			if strings.Contains(key, "user.name") || strings.Contains(key, "user.email") {
				gitConfigCmd := fmt.Sprint("git config --global ", key, " ", "\"", value, "\"")
				if isVmCommand {
					err = sshRemote.ExecSSHCommandRealTime(gitConfigCmd)
					isConfig = true
					common.CheckError(err)
				} else if cli != nil {
					docker := *common.NewDocker(cli)
					out := ""
					out, err = docker.Exec(context.Background(), strings.ReplaceAll(containerName, "/", ""), "/bin", []string{"git", "config", "--global", key, value}, []string{})
					common.CheckError(err)
					common.SmartIDELog.Debug(out)
				}
			}

		}
		if isConfig {
			configPaths := []string{"$HOME/.gitconfig:/root/.gitconfig"}
			if configPaths != nil {
				service.Volumes = append(service.Volumes, configPaths...)
			}
		}

	}
	return service
}
