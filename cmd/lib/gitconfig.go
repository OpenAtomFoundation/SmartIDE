/*
 * @Author: kenan
 * @Date: 2021-10-13 15:31:52
 * @LastEditors: kenan
 * @LastEditTime: 2021-10-15 14:56:20
 * @Description: file content
 */

package lib

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/leansoftX/smartide-cli/lib/common"
	"github.com/leansoftX/smartide-cli/lib/i18n"
)

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

func ConfigGit() {

}
