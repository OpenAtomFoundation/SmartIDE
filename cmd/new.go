package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/leansoftX/smartide-cli/cmd/start"
	"github.com/leansoftX/smartide-cli/internal/apk/appinsight"
	"github.com/leansoftX/smartide-cli/internal/biz/config"
	"github.com/leansoftX/smartide-cli/internal/biz/workspace"
	"github.com/leansoftX/smartide-cli/pkg/common"

	"github.com/spf13/cobra"
)

var newProjectType string
var newTypeStruct []NewType

const templateFolder = "templates"

var newCmd = &cobra.Command{
	Use:   "new",
	Short: i18nInstance.New.Info_help_short,
	Long:  i18nInstance.New.Info_help_long,
	Example: `  smartide new
  smartide new <templatetype> -t {typename}`,
	RunE: func(cmd *cobra.Command, args []string) error {

		if len(args) == 0 {
			common.SmartIDELog.Info(i18nInstance.New.Info_loading_templates)
			templateGitPath := filepath.Join(config.SmartIdeHome, templateFolder, ".git")
			templatesGitIsExist := common.IsExit(templateGitPath)
			if !templatesGitIsExist {
				templatesClone()
			}
			//加载templates索引json
			loadTemplatesJson()
			fmt.Println(i18nInstance.New.Info_help_info)
			printTemplates(newTypeStruct)
			fmt.Println(i18nInstance.New.Info_help_info_operation)
			fmt.Println(cmd.Flags().FlagUsages())
		} else if len(args) == 1 {
			//检测docker环境
			err := start.CheckLocalEnv()
			common.CheckError(err)

			//检测git环境
			err = config.CheckLocalGitEnv()
			common.CheckError(err)

			//1.检测当前文件夹是否有.ide.yaml
			//有了返回
			isIdeYaml := common.IsExit(".ide/.ide.yaml")
			if isIdeYaml {
				common.SmartIDELog.Info(i18nInstance.New.Info_yaml_exist)
				return nil
			}

			//将最新template克隆到userfolder/.ide/templates
			common.SmartIDELog.Info(i18nInstance.New.Info_loading_templates)
			templatesClone()

			//加载templates索引json
			loadTemplatesJson()

			//2.加载type类型的command
			var typeName string
			var typeCommand, subType []string
			for i := 0; i < len(newTypeStruct); i++ {
				if newTypeStruct[i].TypeName == args[0] {
					typeName = newTypeStruct[i].TypeName
					subType = newTypeStruct[i].SubType
					typeCommand = newTypeStruct[i].Command
					break
				}
			}

			if typeName == "" {
				common.SmartIDELog.Info(i18nInstance.New.Info_type_no_exist)
				return nil
			}
			if newProjectType != "" {
				var isExist bool
				newProjectType = strings.Replace(newProjectType, " ", "", -1)
				for i := 0; i < len(subType); i++ {
					if subType[i] == newProjectType {
						isExist = true
						break
					}
				}
				if !isExist {
					common.SmartIDELog.Info(i18nInstance.New.Info_type_no_exist)
					return nil
				}
			}
			//3.检测当前文件夹是否为空
			folderPath, _ := os.Getwd()
			isEmpty, _ := folderEmpty(folderPath)
			if !isEmpty {
				var s string
				fmt.Print(i18nInstance.New.Info_noempty_is_comfirm)
				fmt.Scanln(&s)
				if s != "y" {
					return nil
				}
			}
			//复制templates下到当前文件夹
			copyTemplate(args[0], newProjectType)

			/*------------------------执行start----------------------------------*/

			//执行start
			//0. 杝示文本
			common.SmartIDELog.Info(i18nInstance.Start.Info_start)

			//0.1. 从坂数中获坖结构体，并坚基本的数杮有效性校验
			common.SmartIDELog.Info(i18nInstance.Main.Info_workspace_loading)

			worksapceInfo, err := getWorkspace4Start(cmd, args)
			common.CheckError(err)

			//ai记录
			var trackEvent string
			for _, val := range args {
				trackEvent = trackEvent + " " + val
			}

			// 执行命令
			if worksapceInfo.Mode == workspace.WorkingMode_Local {
				start.ExecuteStartCmd(worksapceInfo, func(dockerContainerName string, docker common.Docker) {
					if dockerContainerName != "" {
						common.SmartIDELog.Info(i18nInstance.New.Info_creating_project)
						for i := 0; i < len(typeCommand); i++ {
							workFolder := fmt.Sprintf("/home/project/%v", worksapceInfo.GetProjectDirctoryName())
							var cmdarr []string
							cmdarr = strings.Split(typeCommand[i], " ")
							out, err := docker.Exec(context.Background(), dockerContainerName, workFolder, cmdarr, []string{})
							common.CheckError(err)
							common.SmartIDELog.Debug(out)
						}
					}
				}, func(yamlConfig config.SmartIdeConfig) {
					var imageNames string
					for image := range yamlConfig.Workspace.Servcies {
						tag := yamlConfig.Workspace.Servcies[image].Image.Tag
						imageName := yamlConfig.Workspace.Servcies[image].Image.Name
						if tag == "" {
							imageNames = imageNames + imageName + ","
						} else {
							imageNames = imageNames + imageName + ":" + tag + ","
						}
					}
					appinsight.SetTrack(cmd.Use, Version.TagName, trackEvent, string(worksapceInfo.Mode), imageNames)
				})
			}
		}
		return nil
	},
}

// 打印 service 列表
func printTemplates(newType []NewType) {
	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
	fmt.Fprintln(w, i18nInstance.New.Info_templates_list_header)
	for i := 0; i < len(newType); i++ {
		line := fmt.Sprintf("%v\t%v", newType[i].TypeName, "_default")
		fmt.Fprintln(w, line)
		for j := 0; j < len(newType[i].SubType); j++ {
			subTypeName := newType[i].SubType[j]
			if subTypeName != "" {
				line := fmt.Sprintf("%v\t%v", newType[i].TypeName, subTypeName)
				fmt.Fprintln(w, line)
			}
		}
	}
	w.Flush()
	fmt.Println("")
}

//复制templates
func copyTemplate(modelType, newProjectType string) {
	if newProjectType == "" {
		newProjectType = "_default"
	}
	templatePath := filepath.Join(config.SmartIdeHome, templateFolder, modelType, newProjectType)
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
func folderEmpty(dirname string) (bool, error) {
	dir, err := ioutil.ReadDir(dirname)
	if err != nil {
		return false, err
	}
	if len(dir) == 0 {
		return true, nil
	} else {
		return false, nil
	}
}

//clone模版
func templatesClone() {
	templatePath := filepath.Join(config.SmartIdeHome, templateFolder)
	templatesFolderIsExist := common.IsExit(templatePath)
	if templatesFolderIsExist {
		var errArry []string
		templateGitPath := filepath.Join(templatePath, ".git")
		templatesGitIsExist := common.IsExit(templateGitPath)
		if templatesGitIsExist {
			//修正git地址不一致
			// gitRepo, err := git.PlainOpen(templatePath)
			// common.CheckError(err)
			// gitRemote, err := gitRepo.Remote("origin")
			// common.CheckError(err)
			// gitRemmoteUrl := gitRemote.Config().URLs[0]
			// if gitRemmoteUrl != config.GlobalSmartIdeConfig.TemplateRepo {
			// 	gitCmd := exec.Command("git", "remote", "set-url", "origin", config.GlobalSmartIdeConfig.TemplateRepo)
			// 	gitCmd.Stdout = os.Stdout
			// 	gitCmd.Stderr = os.Stderr
			// 	gitCmd.Dir = templatePath
			// 	gitErr := gitCmd.Run()
			// 	if gitErr != nil {
			// 		errArry = append(errArry, "git remote set-url")
			// 	}
			// }
			// errArry = forceTemplatesPull(templatePath)
			gitCmd := exec.Command("git", "pull")
			gitCmd.Dir = templatePath
			gitErr := gitCmd.Run()
			if gitErr != nil {
				errArry = append(errArry, "git pull err")
			}
		}
		if len(errArry) != 0 || !templatesGitIsExist {
			err := os.RemoveAll(templatePath)
			common.CheckError(err)
			gitCmd := exec.Command("git", "clone", config.GlobalSmartIdeConfig.TemplateRepo, templatePath)
			gitCmd.Stdout = os.Stdout
			gitCmd.Stderr = os.Stderr
			gitcloneErr := gitCmd.Run()
			common.CheckError(gitcloneErr)
		}

	} else {
		gitCmd := exec.Command("git", "clone", config.GlobalSmartIdeConfig.TemplateRepo, templatePath)
		gitCmd.Stdout = os.Stdout
		gitCmd.Stderr = os.Stderr
		gitcloneErr := gitCmd.Run()
		common.CheckError(gitcloneErr)
	}
}

//强制获取templates
func forceTemplatesPull(gitFolder string) (errArry []string) {
	var gitCmd exec.Cmd
	gitCmd = *exec.Command("git", "fetch", "--all")
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
func loadTemplatesJson() {
	// new type转换为结构体
	templatesPath := filepath.Join(config.SmartIdeHome, templateFolder, "templates.json")
	templatesByte, err := os.ReadFile(templatesPath)
	common.SmartIDELog.Error(err, i18nInstance.New.Err_read_templates, templatesPath)
	err = json.Unmarshal(templatesByte, &newTypeStruct)
	common.SmartIDELog.Error(err)
}

func init() {
	newCmd.Flags().StringVarP(&newProjectType, "type", "t", "", i18nInstance.New.Info_help_flag_type)
}

type NewType struct {
	TypeName string   `json:"typename"`
	SubType  []string `json:"subtype"`
	Command  []string `json:"command"`
}
