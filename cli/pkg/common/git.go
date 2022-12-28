/*
 * @Author: jason chen (jasonchen@leansoftx.com, http://smallidea.cnblogs.com)
 * @Description:
 * @Date: 2021-11
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-12-28 08:18:47
 */
package common

import (
	"errors"
	"fmt"
	"net/url"
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

// 检查git 地址是否正确
func (g gitOperation) CheckGitRemoteUrl(url string) bool {
	pattern := `((git|ssh|http(s)?)|(git@[\w\.]+))(:(//)?)([\w\.@\:/\-~]+)(\.git)(/)?`

	match, _ := regexp.MatchString(pattern, url)
	return match

}

// 获取 repo url
func (g gitOperation) GetRepositoryUrl(actualGitRepoUrl string) string {
	if runtime.GOOS == "windows" {
		return ""
	}
	uri, err := url.Parse(actualGitRepoUrl)
	if err != nil {
		SmartIDELog.Warning(err.Error())
		return ""
	}
	if uri == nil {
		return ""
	}
	if uri.Scheme == "https" || uri.Scheme == "http" {
		if actualGitRepoUrl[len(actualGitRepoUrl)-4:] == ".git" {
			return actualGitRepoUrl[:len(actualGitRepoUrl)-4]
		} else {
			return actualGitRepoUrl
		}
	}
	return ""
}

// 获取检查 git repo 是否存在的command
func (g gitOperation) GetCommand4RepositoryUrl(actualGitRepoUrl string) string {
	return fmt.Sprintf(`curl -s -o /dev/null -I -w "%%{http_code}" %v`, actualGitRepoUrl)
}

// git repository 的可访问状态
type GitRepoAccessStatusEnum string

const (
	GitRepoStatusEnum_NotExists         GitRepoAccessStatusEnum = "not exists"
	GitRepoStatusEnum_PrivateRepository GitRepoAccessStatusEnum = "private repository"
	GitRepoStatusEnum_Other             GitRepoAccessStatusEnum = "not exists or private repository"
)

type GitRepoAccessError struct {
	GitRepoAccessStatus GitRepoAccessStatusEnum
	Err                 error
}

// NameEmtpyError实现了 Error() 方法的对象都可以
func (e *GitRepoAccessError) Error() string {
	if e != nil {
		if e.GitRepoAccessStatus == GitRepoStatusEnum_NotExists {
			return "当前 git 仓库不存在"
		} else if e.GitRepoAccessStatus == GitRepoStatusEnum_PrivateRepository {
			return "当前 git 仓库为私有库"
		} else if e.GitRepoAccessStatus == GitRepoStatusEnum_Other {
			return "当前 git 仓库不存在或者为私有库"
		}
	}
	return ""
}

// 检查错误
func (g gitOperation) CheckError4RepositoryUrl(actualGitRepoUrl string, httpCode int) (err *GitRepoAccessError) {
	if httpCode == 200 {
		return
	}

	uri, _ := url.Parse(actualGitRepoUrl)
	hostName := strings.ToLower(uri.Hostname())
	if strings.Contains(hostName, "github.com") { // 如果是github，不存在或者私有都会返回是404
		if httpCode == 404 {
			err = &GitRepoAccessError{GitRepoAccessStatus: GitRepoStatusEnum_Other}
		} else {
			err = &GitRepoAccessError{GitRepoAccessStatus: GitRepoStatusEnum_PrivateRepository}
		}
	} else if strings.Contains(hostName, "gitee.com") {
		if httpCode == 404 {
			err = &GitRepoAccessError{GitRepoAccessStatus: GitRepoStatusEnum_NotExists}
		} else {
			err = &GitRepoAccessError{GitRepoAccessStatus: GitRepoStatusEnum_PrivateRepository}
		}
	} else {
		err = &GitRepoAccessError{GitRepoAccessStatus: GitRepoStatusEnum_Other}
	}

	return
}

// 使用git下载指定的文件
func (g gitOperation) SparseCheckout(rootDir string, gitCloneUrl string, fileExpression string, branch string) ([]string, error) {
	if gitCloneUrl == "" {
		return []string{}, errors.New("actual git repo url is null")
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
	repoDirPath := filepath.Join(rootDir, repoName)
	if branch == "" {
		branch = g.GetMainBranchName(repoDirPath)
	}
	remoteName := g.GetRemoteName(repoDirPath)
	branchCommand := fmt.Sprintf(`
	git checkout %v 
	git reset --hard %v/%v 
	git pull
	`, branch, remoteName, branch)
	err = EXEC.Realtime(branchCommand, repoDirPath)
	if err != nil {
		SmartIDELog.Warning(err.Error())
	}

	//3. 获取下载的文件列表
	tempExpression := filepath.Join(rootDir, repoName, fileExpression)
	files, err := filepath.Glob(tempExpression)
	if err != nil {
		return []string{}, err
	}

	return files, nil
}

func (g gitOperation) GetRemoteName(repoDirPath string) string {
	output, _ := EXEC.CombinedOutput("git remote show", repoDirPath)
	tmpArray := strings.Split(strings.TrimSpace(output), "\n")
	remoteName := tmpArray[len(tmpArray)-1]
	return remoteName
}

// 获取默认分支
// e.g. git remote show $(git remote show|tail -1)|grep 'HEAD branch'|awk '{print $NF}'
func (g gitOperation) GetMainBranchName(repoDirPath string) string {
	remoteName := g.GetRemoteName(repoDirPath)

	output1, _ := EXEC.CombinedOutput("git remote show "+remoteName, repoDirPath)
	tmpArray1 := strings.Split(strings.TrimSpace(output1), "\n")
	for _, line := range tmpArray1 {
		/*
					* remote origin
			  Fetch URL: https://github.com/idcf-boat-house/boathouse-calculator.git
			  Push  URL: https://github.com/idcf-boat-house/boathouse-calculator.git
			  HEAD branch: master
			  Remote branches:
			    feature-vmlc-arm-improve tracked
			    kaikeba                  tracked
			    master                   tracked
			    new-base                 tracked
			    test-git-config          tracked
			  Local branches configured for 'git pull':
			    feature-vmlc-arm-improve merges with remote feature-vmlc-arm-improve
			    master                   merges with remote master
			  Local refs configured for 'git push':
			    feature-vmlc-arm-improve pushes to feature-vmlc-arm-improve (up to date)
			    master                   pushes to master                   (local out of date)
		*/
		text := "HEAD branch:"
		index := strings.Index(line, text)
		if index > -1 {
			tmp := line[index+len(text):]
			return strings.TrimSpace(tmp)
		}
	}
	return ""
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
	filePathExpression = strings.ReplaceAll(filePathExpression, "\\", "/")
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
