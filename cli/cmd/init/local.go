package init

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/leansoftX/smartide-cli/internal/biz/config"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/leansoftX/smartide-cli/internal/apk/appinsight"
	"github.com/spf13/cobra"
)

func InitLocalConfig(cmd *cobra.Command, args []string) {

	appinsight.SetCliTrack(appinsight.Cli_Local_Init,args)
	
	// 检测当前文件夹是否有.ide.yaml，有了返回
	hasIdeConfigYaml := common.IsExist(".ide/.ide.yaml")
	if hasIdeConfigYaml {
		common.SmartIDELog.Info(i18nInstance.Init.Info_exist_template)
		return
	}

	if !hasIdeConfigYaml {
		//获取command中的配置
		selectedTemplateType, err := getTemplateSetting(cmd, args)
		common.CheckError(err)
		if selectedTemplateType == nil { // 未指定模板类型的时候，提示用户后退出
			return // 退出
		}
		// 复制template 下到当前文件夹
		CopyTemplateToCurrentDir(selectedTemplateType.TypeName, selectedTemplateType.SubType)
	}

	common.SmartIDELog.Info(i18nInstance.Init.Info_Init_Complete)
}

// 复制templates
func CopyTemplateToCurrentDir(modelType, newProjectType string) {
	if newProjectType == "" {
		newProjectType = "_default"
	}
	templatePath := common.PathJoin(config.SmartIdeHome, TMEPLATE_DIR_NAME, modelType, newProjectType)
	templatesFolderIsExist := common.IsExist(templatePath)
	if !templatesFolderIsExist {
		common.SmartIDELog.Error(i18nInstance.New.Info_type_no_exist)
	}
	folderPath, err := os.Getwd()
	common.CheckError(err)
	copyerr := copyDir(templatePath, folderPath)
	common.CheckError(copyerr)
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
			common.SmartIDELog.Debug("srcPath不是一个正确的目录！")
			return errors.New("srcPath不是一个正确的目录！")
		}
	}
	if destInfo, err := os.Stat(destPath); err != nil {
		return err
	} else {
		if !destInfo.IsDir() {
			common.SmartIDELog.Debug("destInfo不是一个正确的目录！")
			return errors.New("destInfo不是一个正确的目录！")
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

// 生成目录并拷贝文件
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
			b := common.IsExist(destSplitPath)
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

// 从command的参数中获取模板设置信息
func getTemplateSetting(cmd *cobra.Command, args []string) (*TemplateTypeBo, error) {
	common.SmartIDELog.Info(i18nInstance.New.Info_loading_templates)
	// git clone
	err := templatesClone() //
	if err != nil {
		return nil, err
	}

	templateTypes, err := LoadTemplatesJson() // 解析json文件
	if err != nil {
		return nil, err
	}
	selectedTemplateTypeName := ""

	selectedTemplateSubTypeName := ""
	if len(args) > 0 {
		argsTemplateTypeName := ""
		argsTemplateSubTypeName := ""
		common.SmartIDELog.Info(i18nInstance.Init.Info_check_cmdtemplate)
		if cmd.Name() == "init" && len(cmd.Flags().Args()) == 1 {
			argsTemplateTypeName = args[0]
		}
		if cmd.Name() == "start" && len(cmd.Flags().Args()) == 2 {
			argsTemplateTypeName = args[1]
		}
		if cmd.Name() == "start" && len(cmd.Flags().Args()) == 1 {
			argsTemplateTypeName = args[0]
		}
		argsTemplateSubTypeName, err = cmd.Flags().GetString("type")
		if err != nil {
			return nil, err
		}
		for _, currentTemplateType := range templateTypes {
			if currentTemplateType.TypeName == argsTemplateTypeName {
				selectedTemplateTypeName = argsTemplateTypeName
			}
			for _, currentTemplateTypeSubType := range currentTemplateType.SubTypes {
				if currentTemplateTypeSubType.Name == argsTemplateSubTypeName {
					selectedTemplateSubTypeName = currentTemplateTypeSubType.Name
				}
			}
		}
		if selectedTemplateTypeName == "" || selectedTemplateSubTypeName == "" {
			common.SmartIDELog.Info(i18nInstance.Init.Info_noexist_cmdtemplate)
		}
	}

	if len(args) == 0 || selectedTemplateTypeName == "" || selectedTemplateSubTypeName == "" {
		// print
		fmt.Println(i18nInstance.Init.Info_available_templates)
		PrintTemplates(templateTypes) // 打印支持的模版列表

		fmt.Print(i18nInstance.Init.Info_choose_templatetype)

		var indexChar string
		fmt.Scanln(&indexChar)
		index, err := strconv.Atoi(indexChar)
		if err != nil {
			return nil, err
		}
		if index < 0 || index >= len(templateTypes) {

			return nil, err
		}
		selectedTypeName := templateTypes[index].TypeName

		fmt.Println(i18nInstance.Init.Info_available_ides)
		var subTypes = []string{"_default"}
		fmt.Println(0, subTypes[0])
		for i := 0; i < len(templateTypes[index].SubTypes); i++ {
			fmt.Println(i+1, templateTypes[index].SubTypes[i].Name)
			subTypes = append(subTypes, templateTypes[index].SubTypes[i].Name)
		}
		fmt.Print(i18nInstance.Init.Info_choose_idetype)
		var indexIdeStr string
		fmt.Scanln(&indexIdeStr)
		indexIde, err := strconv.Atoi(indexIdeStr)
		if err != nil {
			return nil, err
		}
		if indexIde < 0 || indexIde >= len(subTypes) {
			return nil, err
		}
		fmt.Println("您选择的模板为：", selectedTypeName, subTypes[indexIde])
		selectedTemplateTypeName = selectedTypeName

		selectedTemplateSubTypeName = subTypes[indexIde]

		selectedTemplateTypeName = strings.TrimSpace(selectedTemplateTypeName)
		selectedTemplateSubTypeName = strings.TrimSpace(selectedTemplateSubTypeName)
	}

	//2.

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

// clone模版repo
func templatesClone() error {
	templatePath := common.PathJoin(config.SmartIdeHome, TMEPLATE_DIR_NAME)
	templateGitPath := common.PathJoin(templatePath, ".git")
	templatesGitIsExist := common.IsExist(templateGitPath)

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

		command := fmt.Sprintf("git clone %v %v", config.GlobalSmartIdeConfig.TemplateActualRepoUrl, templatePath)
		err = common.EXEC.Realtime(command, "")
		if err != nil {
			return err
		}

	}

	return nil
}
