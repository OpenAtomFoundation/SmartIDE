/*
 * @Date: 2022-04-20 10:46:56
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-05-15 23:08:52
 * @FilePath: /smartide-cli/cmd/new/newLocal.go
 */
package new

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"text/tabwriter"

	"github.com/leansoftX/smartide-cli/cmd/start"
	"github.com/leansoftX/smartide-cli/internal/biz/config"
	"github.com/leansoftX/smartide-cli/internal/biz/workspace"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/spf13/cobra"
)

// 模板文件夹的名称
const TMEPLATE_DIR_NAME = "templates"

func LocalNew(cmd *cobra.Command, args []string, workspaceInfo workspace.WorkspaceInfo,
	yamlExecuteFun func(yamlConfig config.SmartIdeConfig)) {

	// 环境监测
	err := common.CheckLocalGitEnv() //检测git环境
	common.CheckError(err)
	err = common.CheckLocalEnv() //检测docker环境
	common.CheckError(err)

	// 获取command中的配置
	selectedTemplateType, err := getTemplateSetting(cmd, args)
	common.CheckError(err)
	if selectedTemplateType == nil { // 未指定模板类型的时候，提示用户后退出
		return // 退出
	}

	// 检测当前文件夹是否有.ide.yaml，有了返回
	hasIdeConfigYaml := common.IsExit(".ide/.ide.yaml")
	if hasIdeConfigYaml {
		common.SmartIDELog.Info("当前目录已经完成初始化，无须再次进行！")
	}

	// 检测并阻断
	folderPath, _ := os.Getwd()
	isEmpty, _ := folderEmpty(folderPath) // 检测当前文件夹是否为空
	if !isEmpty {
		isContinue, _ := cmd.Flags().GetBool("yes")
		if !isContinue { // 如果没有设置yes，那么就要给出提示
			var s string
			common.SmartIDELog.Importance(i18nInstance.New.Info_noempty_is_comfirm)
			fmt.Scanln(&s)
			if s != "y" {
				return
			}
		} else {
			common.SmartIDELog.Importance("当前文件夹不为空，当前文件夹内数据将被重置。")
		}
	}

	// 复制template 下到当前文件夹
	copyTemplateToCurrentDir(selectedTemplateType.TypeName, selectedTemplateType.SubType)

	// 执行start命令
	common.SmartIDELog.Info(i18nInstance.Start.Info_start) // 执行start
	if workspaceInfo.Mode == workspace.WorkingMode_Local {
		func1 := func(dockerContainerName string, docker common.Docker) {
			if dockerContainerName != "" {
				common.SmartIDELog.Info(i18nInstance.New.Info_creating_project)
				for i := 0; i < len(selectedTemplateType.Commands); i++ {
					workFolder := fmt.Sprintf("/home/project/%v", workspaceInfo.GetProjectDirctoryName())
					cmdarr := strings.Split(selectedTemplateType.Commands[i], " ")
					out, err := docker.Exec(context.Background(), dockerContainerName, workFolder, cmdarr, []string{})
					common.CheckError(err)
					common.SmartIDELog.Debug(out)
				}
			}
		}
		isUnforward, _ := cmd.Flags().GetBool("unforward")
		start.ExecuteStartCmd(workspaceInfo, isUnforward, func1, yamlExecuteFun)
	}

}

// 从command的参数中获取模板设置信息
func getTemplateSetting(cmd *cobra.Command, args []string) (*TemplateTypeBo, error) {
	common.SmartIDELog.Info(i18nInstance.New.Info_loading_templates)

	// git clone
	err := templatesClone() //
	if err != nil {
		return nil, err
	}

	templateTypes, err := loadTemplatesJson() // 解析json文件
	if err != nil {
		return nil, err
	}

	if len(args) == 0 {
		// print
		fmt.Println(i18nInstance.New.Info_help_info)
		printTemplates(templateTypes) // 打印支持的模版列表
		fmt.Println(i18nInstance.New.Info_help_info_operation)
		fmt.Println(cmd.Flags().FlagUsages())
		return nil, nil
	}

	//2.
	selectedTemplateTypeName := ""
	if len(args) > 0 {
		selectedTemplateTypeName = args[0]
	}
	selectedTemplateSubTypeName, err := cmd.Flags().GetString("type")
	if err != nil {
		return nil, err
	}
	selectedTemplateTypeName = strings.TrimSpace(selectedTemplateTypeName)
	selectedTemplateSubTypeName = strings.TrimSpace(selectedTemplateSubTypeName)

	//3. 遍历进行查找
	var selectedTemplate *TemplateTypeBo
	for _, currentTemplateType := range templateTypes {
		if currentTemplateType.TypeName == selectedTemplateTypeName {

			isSelected := false
			if selectedTemplateSubTypeName == "_default" {
				isSelected = true

			} else {
				for _, currentSubTemplateType := range currentTemplateType.SubTypes {
					if currentSubTemplateType.Name == selectedTemplateSubTypeName {
						isSelected = true
						break
					}

				}
			}

			if isSelected {
				tmp := TemplateTypeBo{
					TypeName: selectedTemplateTypeName,
					SubType:  selectedTemplateSubTypeName,
					Commands: currentTemplateType.Commands,
				}
				selectedTemplate = &tmp

				break
			}

		}
	}
	if selectedTemplate == nil {
		return nil, errors.New(i18nInstance.New.Info_type_no_exist)
	}
	return selectedTemplate, nil
}

// 打印 service 列表
func printTemplates(newType []NewTypeBO) {
	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
	fmt.Fprintln(w, i18nInstance.New.Info_templates_list_header)
	for i := 0; i < len(newType); i++ {
		line := fmt.Sprintf("%v\t%v", newType[i].TypeName, "_default")
		fmt.Fprintln(w, line)
		for j := 0; j < len(newType[i].SubTypes); j++ {
			subTypeName := newType[i].SubTypes[j]
			if subTypeName != (SubType{}) && subTypeName.Name != "" {
				line := fmt.Sprintf("%v\t%v", newType[i].TypeName, subTypeName.Name)
				fmt.Fprintln(w, line)
			}
		}
	}
	w.Flush()
	fmt.Println("")
}

//复制templates
func copyTemplateToCurrentDir(modelType, newProjectType string) {
	if newProjectType == "" {
		newProjectType = "_default"
	}
	templatePath := common.PathJoin(config.SmartIdeHome, TMEPLATE_DIR_NAME, modelType, newProjectType)
	templatesFolderIsExist := common.IsExit(templatePath)
	if !templatesFolderIsExist {
		common.SmartIDELog.Error(i18nInstance.New.Info_type_no_exist)
	}
	folderPath, err := os.Getwd()
	common.CheckError(err)
	copyerr := copyDir(templatePath, folderPath)
	common.CheckError(copyerr)
}

//判断文件夹是坦为空
//空为true
func folderEmpty(dirPth string) (bool, error) {
	fis, err := ioutil.ReadDir(dirPth)
	if err != nil {
		return false, err
	}

	isEmpty := true
	for _, f := range fis {
		if !f.IsDir() && runtime.GOOS == "darwin" && strings.Contains(f.Name(), ".DS_Store") {
			continue
		}

		isEmpty = false
		break
	}

	return isEmpty, nil
}

// clone模版repo
func templatesClone() error {
	templatePath := common.PathJoin(config.SmartIdeHome, TMEPLATE_DIR_NAME)
	templateGitPath := common.PathJoin(templatePath, ".git")
	templatesGitIsExist := common.IsExit(templateGitPath)

	// 通过判断.git目录存在，执行git pull，保持最新
	if templatesGitIsExist {
		err := common.EXEC.Realtime(`
git checkout -- * 
git pull
		`, templatePath)
		if err != nil {
			return err
		}

	} else {
		err := os.RemoveAll(templatePath)
		if err != nil {
			return err
		}

		command := fmt.Sprintf("git clone %v %v", config.GlobalSmartIdeConfig.TemplateRepo, templatePath)
		err = common.EXEC.Realtime(command, "")
		if err != nil {
			return err
		}

	}

	return nil
}

//强制获取templates
func forceTemplatesPull(gitFolder string) (errArry []string) {
	gitCmd := *exec.Command("git", "fetch", "--all")
	gitCmd.Dir = gitFolder
	gitErr := gitCmd.Run()
	if gitErr != nil {
		errArry = append(errArry, "git fetch --all")
	}
	gitCmd = *exec.Command("git", "reset", "--hard", "origin/master")
	gitCmd.Dir = gitFolder
	gitErr = gitCmd.Run()
	if gitErr != nil {
		errArry = append(errArry, "git reset --hard origin/master")
	}
	gitCmd = *exec.Command("git", "pull")
	gitCmd.Dir = gitFolder
	gitErr = gitCmd.Run()
	if gitErr != nil {
		errArry = append(errArry, "git pull")
	}
	return errArry
}

/**
 * 拷贝文件夹,同时拷贝文件夹中的文件
 * @param srcPath 需要拷贝的文件夹路径
 * @param destPath 拷贝到的位置
 */
func copyDir(srcPath string, destPath string) error {
	//检测目录正确性
	if srcInfo, err := os.Stat(srcPath); err != nil {
		return err
	} else {
		if !srcInfo.IsDir() {
			e := common.SmartIDELog.Debug("srcPath不是一个正确的目录！")
			return e
		}
	}
	if destInfo, err := os.Stat(destPath); err != nil {
		return err
	} else {
		if !destInfo.IsDir() {
			e := common.SmartIDELog.Debug("destInfo不是一个正确的目录！")
			return e
		}
	}
	err := filepath.Walk(srcPath, func(path string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		if !f.IsDir() {
			path := strings.Replace(path, "\\", "/", -1)
			srcPath = strings.Replace(srcPath, "\\", "/", -1)
			destPath = strings.Replace(destPath, "\\", "/", -1)
			destNewPath := strings.Replace(path, srcPath, destPath, -1)
			copyFile(path, destNewPath)
		}
		return nil
	})
	return err
}

//生成目录并拷贝文件
func copyFile(src, dest string) (w int64, err error) {
	srcFile, err := os.Open(src)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer srcFile.Close()
	//分割path目录
	destSplitPathDirs := strings.Split(dest, "/")
	//检测时候存在目录
	destSplitPath := ""
	for index, dir := range destSplitPathDirs {
		if index < len(destSplitPathDirs)-1 {
			destSplitPath = destSplitPath + dir + "/"
			b := common.IsExit(destSplitPath)
			if !b {
				//创建目录
				err := os.Mkdir(destSplitPath, os.ModePerm)
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	}
	dstFile, err := os.Create(dest)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer dstFile.Close()
	return io.Copy(dstFile, srcFile)
}

//加载templates索引json
func loadTemplatesJson() (templateTypes []NewTypeBO, err error) {
	// new type转换为结构体
	templatesPath := common.PathJoin(config.SmartIdeHome, TMEPLATE_DIR_NAME, "templates.json")
	templatesByte, err := os.ReadFile(templatesPath)
	if err != nil {
		return templateTypes, errors.New(i18nInstance.New.Err_read_templates + templatesPath + err.Error())
	}

	err = json.Unmarshal(templatesByte, &templateTypes)
	return templateTypes, err
}
