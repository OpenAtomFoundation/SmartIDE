/*
 * @Author: kenan
 * @Date: 2021-10-13 15:31:52
 * @LastEditors: kenan
 * @LastEditTime: 2021-12-20 12:02:00
 * @Description: file content
 */

package config

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/docker/docker/client"
	"github.com/leansoftX/smartide-cli/internal/apk/i18n"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/leansoftX/smartide-cli/pkg/docker/compose"
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
		common.SmartIDELog.Error(i18n.GetInstance().Config.Err_Gitconfig_not_exit)
	}
	s := bufio.NewScanner(strings.NewReader(gitconfigs))
	for s.Scan() {
		//以=分割,前面为key,后面为value
		var str = s.Text()
		var index = strings.Index(str, "=")
		var key = str[0:index]
		var value = str[index+1:]

		yamlFileCongfig := NewConfig("", "", "")

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

func SSHVolumesConfig(sshKey string, isVmCommand bool, service compose.Service, sshRemote common.SSHRemote) (ser compose.Service) {
	var configPaths []string
	if sshKey == "true" {
		// volumes
		if runtime.GOOS == "windows" && !isVmCommand {
			configPaths = []string{fmt.Sprintf("\\'%v\\.ssh:/home/smartide/.ssh\\'", os.Getenv("USERPROFILE"))}
		} else {
			configPaths = []string{"$HOME/.ssh:/home/smartide/.ssh"}
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
		// git config 默认设置

		gitconfigs := strings.ReplaceAll(configStr, "file:", "")
		s := bufio.NewScanner(strings.NewReader(gitconfigs))
		for s.Scan() {
			//以=分割,前面为key,后面为value
			var str = s.Text()
			var index = strings.Index(str, "=")
			var key = str[0:index]
			var value = str[index+1:]
			if strings.Contains(key, "user.name") || strings.Contains(key, "user.email") || strings.Contains(key, "core.autocrlf") {
				gitConfigCmd := fmt.Sprint("git config --global --replace-all ", key, " ", "\"", value, "\"")
				if isVmCommand {
					output, err := sshRemote.ExeSSHCommand(gitConfigCmd)
					isConfig = true
					common.CheckError(err, output)
				} else if cli != nil {
					docker := *common.NewDocker(cli)
					out := ""
					out, err = docker.Exec(context.Background(), strings.ReplaceAll(containerName, "/", ""), "/bin", []string{"git", "config", "--global", "--replace-all", key, value}, []string{})
					common.CheckError(err)
					common.SmartIDELog.Debug(out)
				}
			}

		}
		if isConfig {
			configPaths := []string{"$HOME/.gitconfig:/home/smartide/.gitconfig"}
			if configPaths != nil {
				service.Volumes = append(service.Volumes, configPaths...)
			}
		}

	}
	return service
}

// 检查本地环境，是否安装git
func CheckLocalGitEnv() error {
	var errMsgArray []string

	// 校验是否能正常执行 git
	gitErr := exec.Command("git", "version").Run()
	if gitErr != nil {
		errMsgArray = append(errMsgArray, i18n.GetInstance().Main.Err_env_git_check)
	}

	// 错误判断
	if len(errMsgArray) > 0 {
		return errors.New(strings.Join(errMsgArray, "; "))
	}

	return nil
}

func LocalContainerGitSet(docker common.Docker, dockerContainerName string) {
	out, err := docker.Exec(context.Background(), dockerContainerName, "/usr/bin", []string{"sudo", "chown", "-R", "smartide:smartide", "/home/smartide/.ssh"}, []string{})
	common.CheckError(err)
	common.SmartIDELog.Debug(out)

	out, err = docker.Exec(context.Background(), dockerContainerName, "/usr/bin", []string{"sudo", "chmod", "755", "/home/smartide/.ssh"}, []string{})
	common.CheckError(err)
	common.SmartIDELog.Debug(out)

	out, err = docker.Exec(context.Background(), dockerContainerName, "/usr/bin", []string{"sudo", "chmod", "644", "/home/smartide/.ssh/authorized_keys"}, []string{})
	common.CheckError(err)
	common.SmartIDELog.Debug(out)

	out, err = docker.Exec(context.Background(), dockerContainerName, "/usr/bin", []string{"sudo", "chmod", "644", "/home/smartide/.ssh/id_rsa.pub"}, []string{})
	common.CheckError(err)
	common.SmartIDELog.Debug(out)

	out, err = docker.Exec(context.Background(), dockerContainerName, "/usr/bin", []string{"sudo", "chmod", "600", "/home/smartide/.ssh/id_rsa"}, []string{})
	common.CheckError(err)
	common.SmartIDELog.Debug(out)
}
