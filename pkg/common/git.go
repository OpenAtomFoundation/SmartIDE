/*
 * @Author: jason chen (jasonchen@leansoftx.com, http://smallidea.cnblogs.com)
 * @Description:
 * @Date: 2021-11
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-03-29 16:38:04
 */
package common

import (
	"fmt"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

type gitOperation struct{}

// 文件操作相关
var GIT gitOperation

func init() {
	GIT = gitOperation{}

}

func (g gitOperation) CheckGitRemoteUrl(url string) bool {
	pattern := `((git|ssh|http(s)?)|(git@[\w\.]+))(:(//)?)([\w\.@\:/\-~]+)(\.git)(/)?`

	match, _ := regexp.MatchString(pattern, url)
	return match

}

// 使用git下载指定的文件
func (g gitOperation) SparseCheckout(rootDir string, gitCloneUrl string, fileExpression string, branch string) ([]string, error) {
	repoName := GetRepoName(gitCloneUrl)

	//1. 配置
	//1.1. command
	sparseCheckout := fmt.Sprintf("echo \"%v\" >> .git/info/sparse-checkout", fileExpression) //TODO 会插入多条
	command := fmt.Sprintf(`
	git init %v
	cd %v
	git config core.sparsecheckout true
	%v
	git remote add -f origin %v
	git fetch
`, repoName, repoName, sparseCheckout, gitCloneUrl)

	//1.2. exec
	err := EXEC.Realtime(command, rootDir)
	if err != nil {
		return []string{}, err
	}

	//2. checkout
	repoDirPath := path.Join(rootDir, repoName)
	output, err := EXEC.CombinedOutput("git branch -a", repoDirPath)
	if err != nil {
		return []string{}, err
	}
	branchCommand := "git checkout master"
	if branch != "" {
		branchCommand = fmt.Sprintf("git checkout %v", branch)
	} else {
		if !strings.Contains(output, "origin/master") { // 如果主分支不master，就切换为main
			command = "git checkout main"
		}
	}
	branchCommand += `
	git pull`
	err = EXEC.Realtime(branchCommand, repoDirPath)
	if err != nil {
		return []string{}, err
	}

	//3. 获取下载的文件列表
	tempExpression := path.Join(rootDir, repoName, fileExpression)
	files, err := filepath.Glob(tempExpression)
	if err != nil {
		return []string{}, err
	}

	return files, nil
}
