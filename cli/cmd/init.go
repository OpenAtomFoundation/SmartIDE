/*
 * @Author: Bo Dai (daibo@leansoftx.com)
 * @Description:
 * @Date: 2022-07
 * @LastEditors: Bo Dai
 * @LastEditTime: 2022年7月21日 09点28分
 */

package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	initExtend "github.com/leansoftX/smartide-cli/cmd/init"
	"github.com/leansoftX/smartide-cli/cmd/start"
	"github.com/leansoftX/smartide-cli/internal/apk/appinsight"
	"github.com/leansoftX/smartide-cli/internal/biz/config"
	"github.com/leansoftX/smartide-cli/internal/biz/workspace"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: i18nInstance.Init.Info_help_short,
	Long:  i18nInstance.Init.Info_help_long,
	Example: ` smartide init
	 smartide init <templatetype> -t {typename}`,
	Run: func(cmd *cobra.Command, args []string) {
		//ai记录
		var trackEvent string
		for _, val := range args {
			trackEvent = trackEvent + " " + val
		}
		// 环境监测
		err := common.CheckLocalGitEnv() //检测git环境
		common.CheckError(err)
		err = common.CheckLocalEnv() //检测docker环境
		common.CheckError(err)

		workspaceInfo, err := getWorkspaceFromCmd(cmd, args)

		executeStartCmdFunc := func(yamlConfig config.SmartIdeConfig) {
			if config.GlobalSmartIdeConfig.IsInsightEnabled != config.IsInsightEnabledEnum_Enabled {
				common.SmartIDELog.Debug("Application Insights disabled")
				return
			}
			var imageNames []string
			for _, service := range yamlConfig.Workspace.Servcies {
				imageNames = append(imageNames, service.Image)
			}
			appinsight.SetTrack(cmd.Use, Version.TagName, trackEvent, string(workspaceInfo.Mode), strings.Join(imageNames, ","))
		}

		selectedTemplateType := new(initExtend.TemplateTypeBo)
		// 检测当前文件夹是否有.ide.yaml，有了返回
		hasIdeConfigYaml := common.IsExist(".ide/.ide.yaml")
		if hasIdeConfigYaml {
			common.SmartIDELog.Info(i18nInstance.Init.Info_exist_template)
		}

		if !hasIdeConfigYaml {
			//获取command中的配置
			selectedTemplateType, err = getTemplateSetting(cmd, args)
			common.CheckError(err)
			if selectedTemplateType == nil { // 未指定模板类型的时候，提示用户后退出
				return // 退出
			}
			// 复制template 下到当前文件夹
			copyTemplateToCurrentDir(selectedTemplateType.TypeName, selectedTemplateType.SubType)
		}
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
			start.ExecuteStartCmd(workspaceInfo, isUnforward, func1, executeStartCmdFunc)
		}
	},
}

// 打印 service 列表

func init() {
	initCmd.Flags().StringP("type", "t", "", i18nInstance.New.Info_help_flag_type)
}

// 从command的参数中获取模板设置信息
func getTemplateSetting(cmd *cobra.Command, args []string) (*initExtend.TemplateTypeBo, error) {
	common.SmartIDELog.Info(i18nInstance.New.Info_loading_templates)

	// git clone
	err := templatesClone() //
	if err != nil {
		return nil, err
	}

	templateTypes, err := initExtend.LoadTemplatesJson() // 解析json文件
	if err != nil {
		return nil, err
	}
	selectedTemplateTypeName := ""

	selectedTemplateSubTypeName := ""
	if len(args) > 0 {
		argsTemplateTypeName := ""
		argsTemplateSubTypeName := ""
		common.SmartIDELog.Info(i18nInstance.Init.Info_check_cmdtemplate)
		argsTemplateTypeName = args[0]
		argsTemplateSubTypeName, err = cmd.Flags().GetString("type")
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
		common.SmartIDELog.Warning(i18nInstance.Init.Info_noexist_cmdtemplate)
	}

	if len(args) == 0 || selectedTemplateTypeName == "" || selectedTemplateSubTypeName == "" {
		// print
		fmt.Println(i18nInstance.Init.Info_available_templates)
		initExtend.PrintTemplates(templateTypes) // 打印支持的模版列表
		var index int
		fmt.Println(i18nInstance.Init.Info_choose_templatetype)
		fmt.Scanln(&index)
		if index < 1 || index >= len(templateTypes) {
			return nil, err
		}
		selectedTypeName := templateTypes[index].TypeName

		fmt.Println(i18nInstance.Init.Info_available_ides)
		for i := 0; i < len(templateTypes[index].SubTypes); i++ {
			fmt.Println(i, templateTypes[index].SubTypes[i].Name)
		}
		fmt.Println(i18nInstance.Init.Info_choose_idetype)
		var indexIde int
		fmt.Scanln(&indexIde)
		if indexIde < 1 || indexIde >= len(templateTypes[index].SubTypes) {
			return nil, err
		}
		fmt.Println("您选择的模板为：", selectedTypeName, templateTypes[index].SubTypes[indexIde].Name)
		selectedTemplateTypeName = selectedTypeName

		selectedTemplateSubTypeName = templateTypes[index].SubTypes[indexIde].Name

		selectedTemplateTypeName = strings.TrimSpace(selectedTemplateTypeName)
		selectedTemplateSubTypeName = strings.TrimSpace(selectedTemplateSubTypeName)
	}

	//2.

	//3. 遍历进行查找
	var selectedTemplate *initExtend.TemplateTypeBo
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
				tmp := initExtend.TemplateTypeBo{
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

//复制templates
func copyTemplateToCurrentDir(modelType, newProjectType string) {
	if newProjectType == "" {
		newProjectType = "_default"
	}
	templatePath := common.PathJoin(config.SmartIdeHome, initExtend.TMEPLATE_DIR_NAME, modelType, newProjectType)
	templatesFolderIsExist := common.IsExist(templatePath)
	if !templatesFolderIsExist {
		common.SmartIDELog.Error(i18nInstance.New.Info_type_no_exist)
	}
	folderPath, err := os.Getwd()
	common.CheckError(err)
	copyerr := copyDir(templatePath, folderPath)
	common.CheckError(copyerr)
}

// clone模版repo
func templatesClone() error {
	templatePath := common.PathJoin(config.SmartIdeHome, initExtend.TMEPLATE_DIR_NAME)
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

		command := fmt.Sprintf("git clone %v %v", config.GlobalSmartIdeConfig.TemplateRepo, templatePath)
		err = common.EXEC.Realtime(command, "")
		if err != nil {
			return err
		}

	}

	return nil
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
