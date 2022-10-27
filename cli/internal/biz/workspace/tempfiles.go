package workspace

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/leansoftX/smartide-cli/internal/apk/i18n"

	"github.com/leansoftX/smartide-cli/internal/model"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"gopkg.in/yaml.v2"
)

var i18nInstance = i18n.GetInstance()

// 获取生成的临时 docker-compose 文件路径
func (workspace WorkspaceInfo) GetTempDockerComposeFilePath() string {
	dockerComposeFileName := fmt.Sprintf("docker-compose-%s.yaml", workspace.GetProjectDirctoryName()) // docker-compose 文件的名称
	//yamlFileDirPath := common.PathJoin(workspace.WorkingDirectoryPath, model.CONST_TempDirPath)        //
	yamlFilePath := filepath.Join(workspace.WorkingDirectoryPath, model.CONST_GlobalTempDirPath, dockerComposeFileName)

	return yamlFilePath
}

// 保存docker compose、smartide 配置文件到 临时文件夹下
func (workspace WorkspaceInfo) SaveTempFiles() (err error) { // dockerComposeFilePath string
	// 检查临时文件 是否写入到	.gitignore
	checkLocalGitignoreContainTmpDir(workspace.WorkingDirectoryPath)

	// 临时文件夹
	projectName := workspace.GetProjectDirctoryName()
	tempDockerComposeFilePath := workspace.GetTempDockerComposeFilePath() // docker-compose 的临时文件
	//workingDirectoryPath := workspace.ConfigYaml.GetWorkingDirectoryPath()
	tempConfigFilePath := getTempConfigFilePath(workspace.WorkingDirectoryPath, projectName) // 临时配置文件的存放路径

	// 创建 或者 清空文件夹
	tempDirPath := filepath.Dir(tempDockerComposeFilePath) // 临时文件所在的目录
	if !common.IsExist(tempDirPath) {                      // 创建文件夹
		os.MkdirAll(tempDirPath, os.ModePerm)
		common.SmartIDELog.InfoF(i18nInstance.Common.Info_temp_create_directory, tempDirPath)
	} else { // 清空文件夹
		dir, err := ioutil.ReadDir(tempDirPath)
		common.SmartIDELog.Error(err)
		for _, d := range dir {
			os.RemoveAll(common.PathJoin([]string{tempDirPath, d.Name()}...))
		}
	}

	// create docker-compose file
	if !common.IsExist(tempDockerComposeFilePath) {
		os.Create(tempDockerComposeFilePath)
	}
	sYaml, err := workspace.TempDockerCompose.ToYaml()
	common.CheckError(err)
	if sYaml == "" {
		common.SmartIDELog.Error("docker-compose 为空")
	}
	err = ioutil.WriteFile(tempDockerComposeFilePath, []byte(sYaml), 0)
	if err != nil {
		common.SmartIDELog.Fatal(err)
	}

	// create config file
	if !common.IsExist(tempConfigFilePath) {
		os.Create(tempConfigFilePath)
	}
	sConfig, err := workspace.ConfigYaml.ToYaml()
	common.CheckError(err)
	if sConfig == "" {
		common.SmartIDELog.Error("配置文件 为空")
	}
	err = ioutil.WriteFile(tempConfigFilePath, []byte(sConfig), 0)
	if err != nil {
		common.SmartIDELog.Fatal(err)
	}

	return err
}

// 在远程服务器上保存 docker-compose、config 文件
func (workspace WorkspaceInfo) SaveTempFilesForRemote(sshRemote common.SSHRemote) (err error) { // dockerComposeFilePath string

	// 检查临时文件 是否写入到	.gitignore
	tempDirName := ""        // 临时文件夹所在目录
	tempParentDirPath := "." // 保存临时文件夹的上级目录
	index := strings.LastIndex(model.CONST_GlobalTempDirPath, "/")
	if index >= 0 {
		tempDirName = model.CONST_GlobalTempDirPath[index+1:]
		tempParentDirPath = model.CONST_GlobalTempDirPath[:index]
	} else {
		tempDirName = model.CONST_GlobalTempDirPath
	}

	// 临时文件夹
	tempDirPath := common.PathJoin(workspace.WorkingDirectoryPath, tempParentDirPath)
	sshRemote.CheckAndCreateDir(tempDirPath)

	// .ignore 文件
	gitignoreFilePath := common.PathJoin(tempDirPath, ".gitignore")
	gitignoreContent := "/" + tempDirName + "/"
	if sshRemote.IsFileExist(gitignoreFilePath) {
		output := sshRemote.GetContent(gitignoreFilePath)
		if !strings.Contains(output, gitignoreContent) {
			sshRemote.CreateFileByEcho(gitignoreFilePath, gitignoreContent)
		}
	} else {
		sshRemote.CreateFileByEcho(gitignoreFilePath, gitignoreContent)
	}

	// 临时文件夹
	var tempRemoteDockerComposeFilePath, tempRemoteConfigFilePath string
	tempRemoteDockerComposeFilePath = common.FilePahtJoin4Linux(workspace.GetTempDockerComposeFilePath())                                           // docker-compose 的临时文件
	tempRemoteConfigFilePath = common.FilePahtJoin4Linux(getTempConfigFilePath(workspace.WorkingDirectoryPath, workspace.GetProjectDirctoryName())) // 临时配置文件的存放路径
	pwd, err := sshRemote.GetRemotePwd()
	common.CheckError(err)
	if strings.Index(tempRemoteDockerComposeFilePath, "~") == 0 {
		tempRemoteDockerComposeFilePath = strings.Replace(tempRemoteDockerComposeFilePath, "~", pwd, -1)
	}
	if strings.Index(tempRemoteConfigFilePath, "~") == 0 {
		tempRemoteConfigFilePath = strings.Replace(tempRemoteConfigFilePath, "~", pwd, -1)
	}

	// 创建 或者 清空文件夹中的内容
	remoteTempDirPath := common.FilePahtJoin4Linux(workspace.WorkingDirectoryPath, model.CONST_GlobalTempDirPath)
	if strings.Index(remoteTempDirPath, "~") == 0 {
		remoteTempDirPath = strings.Replace(remoteTempDirPath, "~", pwd, -1)
	}
	command := fmt.Sprintf(`[[ -d "%v" ]] && rm -rf %v || mkdir -p %v `, remoteTempDirPath, remoteTempDirPath+"/*", remoteTempDirPath)
	output, err := sshRemote.ExeSSHCommand(command)
	common.CheckError(err, output)

	// create docker-compose file
	common.SmartIDELog.InfoF(i18nInstance.Common.Info_temp_created_docker_compose, tempRemoteDockerComposeFilePath)

	dCompose, err := workspace.TempDockerCompose.ToYaml()
	if err != nil {
		common.SmartIDELog.Fatal(err)
	}
	tmpComposeContent := strings.ReplaceAll(string(dCompose), "\"", "\\\"") // 防止在远程主机上双引号被剔除
	command = fmt.Sprintf(`echo "%v" >> %v`, tmpComposeContent, tempRemoteDockerComposeFilePath)
	output, err = sshRemote.ExeSSHCommand(command)
	common.CheckError(err, output)

	// create config file
	common.SmartIDELog.InfoF(i18nInstance.Common.Info_temp_created_config, tempRemoteConfigFilePath)
	dConfig, err := yaml.Marshal(&workspace.ConfigYaml)
	if err != nil {
		common.SmartIDELog.Fatal(err)
	}
	command = fmt.Sprintf(`echo "%v" >> %v`, string(dConfig), tempRemoteConfigFilePath)
	output, err = sshRemote.ExeSSHCommand(command)
	common.CheckError(err, output)

	return err
}

// 获取配置文件的临时存放路径
func getTempConfigFilePath(localWorkingDir string, projectName string) string {

	workingDir := ""

	// 确定当前工作目录
	if localWorkingDir == "" {
		dirName, err := os.Getwd()
		if err != nil {
			common.SmartIDELog.Fatal(err)
		}
		workingDir = dirName
	} else {
		workingDir = localWorkingDir
	}

	dockerComposeFileName := fmt.Sprintf("config-%s.yaml", projectName)         // docker-compose 文件的名称
	yamlFileDirPath := filepath.Join(workingDir, model.CONST_GlobalTempDirPath) //
	yamlFilePath := filepath.Join(yamlFileDirPath, dockerComposeFileName)

	return yamlFilePath
}

// 检测是否存在包含 tmp 的 .gitignore文件
func checkLocalGitignoreContainTmpDir(workingDir string) {
	/* dirname, err := os.Getwd()
	if err != nil {
		common.SmartIDELog.Fatal(err)
	} */

	// 临时文件夹的名称，以及临时文件夹的上级目录
	var tempDirName, tempParentDir string
	index := strings.LastIndex(model.CONST_GlobalTempDirPath, "/")
	if index >= 0 {
		tempDirName = model.CONST_GlobalTempDirPath[index+1:]
		tempParentDir = model.CONST_GlobalTempDirPath[:index]
	} else {
		tempDirName = model.CONST_GlobalTempDirPath
	}

	// ignore
	gitignorePath := common.PathJoin(workingDir, tempParentDir, ".gitignore")
	if !common.IsExist(gitignorePath) {
		dirPath := filepath.Dir(gitignorePath)

		if !common.IsExist(dirPath) {
			os.Create(dirPath)
		}

		err := ioutil.WriteFile(gitignorePath, []byte("/"+tempDirName+"/"), 0666)
		common.CheckError(err)
	} else {
		bytes, _ := ioutil.ReadFile(gitignorePath)
		content := string(bytes)
		if !strings.Contains(content, "/"+tempDirName+"/") { // 不包含时，要附加
			err := ioutil.WriteFile(gitignorePath, []byte(content+"\n"+"/"+tempDirName+"/"), 0666)
			common.CheckError(err)
		}
	}
}
