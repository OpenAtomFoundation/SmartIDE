/*
 * @Date: 2022-04-20 10:46:56
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-11-04 15:04:36
 * @FilePath: /cli/cmd/new/newLocal.go
 */
package new

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	cmdCommon "github.com/leansoftX/smartide-cli/cmd/common"
	"github.com/leansoftX/smartide-cli/cmd/start"
	"github.com/leansoftX/smartide-cli/internal/biz/config"
	"github.com/leansoftX/smartide-cli/internal/biz/workspace"
	golbalModel "github.com/leansoftX/smartide-cli/internal/model"

	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/spf13/cobra"
)

func LocalNew(cmd *cobra.Command, args []string, workspaceInfo workspace.WorkspaceInfo,
	yamlExecuteFun func(yamlConfig config.SmartIdeConfig)) {

	// 环境监测
	err := common.CheckLocalGitEnv() //检测git环境
	common.CheckError(err)
	err = common.CheckLocalEnv() //检测docker环境
	common.CheckError(err)

	// 获取command中的配置
	selectedTemplateType, err := cmdCommon.GetTemplateSetting(cmd, args)
	common.CheckError(err)
	if selectedTemplateType == nil { // 未指定模板类型的时候，提示用户后退出
		return // 退出
	}

	// 检测当前文件夹是否有.ide.yaml，有了返回
	hasIdeConfigYaml := common.IsExist(".ide/.ide.yaml")
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

	// 复制template 到当前文件夹
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
		start.ExecuteStartCmd(workspaceInfo, isUnforward, func1, yamlExecuteFun, args, cmd)
	}

}

// 复制templates
func copyTemplateToCurrentDir(modelType, newProjectType string) {
	if newProjectType == "" {
		newProjectType = "_default"
	}
	templatePath := common.PathJoin(config.SmartIdeHome, golbalModel.TMEPLATE_DIR_NAME, modelType, newProjectType)
	templatesFolderIsExist := common.IsExist(templatePath)
	if !templatesFolderIsExist {
		common.SmartIDELog.Error(i18nInstance.New.Info_type_no_exist)
	}
	folderPath, err := os.Getwd()
	common.CheckError(err)
	copyerr := copyDir(templatePath, folderPath)
	common.CheckError(copyerr)
}

// 判断文件夹是坦为空
// 空为true
func folderEmpty(dirPth string) (bool, error) {
	fis, err := os.ReadDir(dirPth)
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
