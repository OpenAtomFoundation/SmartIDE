/*
 * @Author: jason chen (jasonchen@leansoftx.com, http://smallidea.cnblogs.com)
 * @Description:
 * @Date: 2021-11
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-09-06 09:30:32
 */
package common

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
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
	if gitCloneUrl == "" {
		return []string{}, errors.New("git clone url is null!")
	}
	repoName := GetRepoName(gitCloneUrl)

	//1. 配置
	//1.1. command
	sparseCheckout := fmt.Sprintf("echo \"%v\" >> .git/info/sparse-checkout", fileExpression) //TODO 会插入多条
	if runtime.GOOS == "windows" {
		sparseCheckout = fmt.Sprintf(`$content = "%v"
$checkoutFilePath = ".git\\info\\sparse-checkout"		
if (Test-Path $checkoutFilePath) { $content = (Get-Content $checkoutFilePath)+"%v" } 
Set-Content $checkoutFilePath -Value $content -Encoding Ascii`,
			"`n"+fileExpression, "`n"+fileExpression)
	}
	command := fmt.Sprintf(`
	git init %v
	cd %v
	git config core.sparsecheckout true
	%v
	git remote add -f origin %v
	git remote set-url origin %v
	git fetch
`, repoName, repoName, sparseCheckout, gitCloneUrl, gitCloneUrl)

	//1.2. exec
	err := EXEC.Realtime(command, rootDir)
	if err != nil {
		return []string{}, err
	}

	//2. checkout
	repoDirPath := PathJoin(rootDir, repoName)
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
	git reset --hard FETCH_HEAD
	git pull`
	err = EXEC.Realtime(branchCommand, repoDirPath)
	if err != nil {
		return []string{}, err
	}

	//3. 获取下载的文件列表
	tempExpression := PathJoin(rootDir, repoName, fileExpression)
	files, err := filepath.Glob(tempExpression)
	if err != nil {
		return []string{}, err
	}

	return files, nil
}

// 从git下载相关文件
func (g gitOperation) DownloadFilesByGit(workingRootDir string, gitCloneUrl string, branch string, filePathExpression string) (
	gitRepoRootDirPath string, fileRelativePaths []string, err error) {
	// home 目录的路径
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}

	// 文件路径
	//workingRootDir := filepath.Join(home, ".ide", ".k8s") // 工作目录，repo 会clone到当前目录下
	sshPath := filepath.Join(home, ".ssh")
	if !IsExist(workingRootDir) { // 目录如果不存在，就要创建
		os.MkdirAll(workingRootDir, os.ModePerm)
	}

	// 设置不进行host的检查，即clone新的git库时，不会给出提示
	err = FS.SkipStrictHostKeyChecking(sshPath, false)
	if err != nil {
		return
	}

	// 下载指定的文件
	//filePathExpression = common.PathJoin(".ide", filePathExpression)
	fileRelativePaths, err = GIT.SparseCheckout(workingRootDir, gitCloneUrl, filePathExpression, branch)
	if err != nil {
		return
	}

	// 还原.ssh config 的设置
	err = FS.SkipStrictHostKeyChecking(sshPath, true)
	if err != nil {
		return
	}

	gitRepoRootDirPath = filepath.Join(workingRootDir, GetRepoName(gitCloneUrl)) // git repo 的根目录
	for index, _ := range fileRelativePaths {
		fileRelativePaths[index] = strings.Replace(fileRelativePaths[index], gitRepoRootDirPath, "", -1) // 把绝对路径改为相对路径
	}
	return gitRepoRootDirPath, fileRelativePaths, nil
}
