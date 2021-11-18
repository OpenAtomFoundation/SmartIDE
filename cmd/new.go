package cmd

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/leansoftX/smartide-cli/cmd/dal"
	"github.com/leansoftX/smartide-cli/cmd/start"
	"github.com/leansoftX/smartide-cli/lib/common"

	"github.com/leansoftX/smartide-cli/lib/i18n"
	"github.com/spf13/cobra"
)

var instanceI18nNew = i18n.GetInstance().New

var newProjectType string
var newProjectFolder string
var newTypeStruct []NewType

var newCmd = &cobra.Command{
	Use:   "new",
	Short: instanceI18nNew.Info.Help_short,
	Long:  instanceI18nNew.Info.Help_long,
	RunE: func(cmd *cobra.Command, args []string) error {
		if newProjectType == "" {
			fmt.Println(instanceI18nNew.Info.Help_info)
			for i := 0; i < len(newTypeStruct); i++ {
				fmt.Println(newTypeStruct[i].TypeName)
			}
			fmt.Println("")
			fmt.Println(instanceI18nNew.Info.Help_info_operation)
			fmt.Println(cmd.Flags().FlagUsages())
		} else {
			var yamlUrl string
			for i := 0; i < len(newTypeStruct); i++ {
				if newTypeStruct[i].TypeName == newProjectType {
					yamlUrl = newTypeStruct[i].TypeYamlUrl
					break
				}
			}
			if yamlUrl == "" {
				fmt.Println(instanceI18nNew.Info.Info_type_no_exist)
				return nil
			}
			isIdeYaml := checkFileIsExist(".ide/.ide.yaml")
			if isIdeYaml {
				fmt.Println(instanceI18nNew.Info.Info_yaml_exist)
				return nil
			}
			folderPath, _ := os.Getwd()
			isEmpty, _ := folderEmpty(folderPath)
			if isEmpty {
				//创建ide.yaml
				downloadFile(yamlUrl, "")
			} else {
				var s string
				fmt.Print(instanceI18nNew.Info.Info_noempty_is_comfirm)
				fmt.Scanln(&s)
				if s == "y" {
					//创建ide.yaml
					downloadFile(yamlUrl, "")
				} else {
					return nil
				}
			}
			//创建.gitignore
			var d1 = []byte("node_modules/")
			_ = ioutil.WriteFile(folderPath+"/.gitignore", d1, 0666)
		}

		//执行start
		if newProjectType != "" {
			// if newProjectFolder != "" {
			// 	fmt.Println(newProjectFolder)
			// }
			//0. 提示文本
			common.SmartIDELog.Info(i18n.GetInstance().Start.Info.Info_start)

			//0.1. 校验是否能正常执行docker
			start.CheckLocalEnv()

			//0.1. 从参数中获取结构体，并做基本的数据有效性校验
			worksapce, validErr := getWorkspace4Start(cmd, args)
			if validErr != nil {
				return validErr // 采用return的方式，可以显示flag列表 //TODO 根据错误的类型，如果是参数格式错误就是return，其他直接抛错
			}

			// 执行命令
			if worksapce.Mode == dal.WorkingMode_Local {
				start.ExecuteStartCmd(worksapce, func(dockerContainerName string, docker common.Docker) {
					if dockerContainerName != "" {
						common.SmartIDELog.Info(instanceI18nNew.Info.Info_creating_project)
						out, err := docker.Exec(context.Background(), dockerContainerName, "/home/project", []string{"npm", "install", "-g", "express-generator"}, []string{})
						out, err = docker.Exec(context.Background(), dockerContainerName, "/home/project", []string{"express", "-f"}, []string{})
						out, err = docker.Exec(context.Background(), dockerContainerName, "/home/project", []string{"sed", "-i", "s/3000/3001/", "/home/project/bin/www"}, []string{})
						common.CheckError(err)
						common.SmartIDELog.Debug(out)
					}
				})
			}
		}
		return nil
	},
}

//下载yaml文件到指定目录
func downloadFile(url, filePath string) {
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	if filePath == "" {
		_, err := folderEmpty(".ide")
		if err != nil {
			os.Mkdir(".ide", os.ModePerm)
		}
		filePath = ".ide/.ide.yaml"
	} else {
		_, err := folderEmpty(filePath + "/.ide")
		if err != nil {
			os.Mkdir(filePath+"/.ide", os.ModePerm)
		}
		filePath = filePath + "/.ide/.ide.yaml"
	}
	ioutil.WriteFile(filePath, data, 0644)
}

//判断文件夹是否为空
//空为true
func folderEmpty(dirname string) (bool, error) {
	dir, err := ioutil.ReadDir(dirname)
	if err != nil {
		return true, err
	}
	if len(dir) == 0 {
		return true, nil
	} else {
		return false, nil
	}
}

// 判断文件是否存在  存在返回 true 不存在返回false
func checkFileIsExist(filename string) bool {
	var exist = true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}

//获取new的类型
func SetNewType(newType []NewType) {
	newTypeStruct = newType
}

func init() {
	newCmd.Flags().StringVarP(&newProjectType, "type", "t", "", instanceI18nNew.Info.Help_flag_type)
}

type NewType struct {
	TypeName    string `json:"type_name"`
	TypeYamlUrl string `json:"type_yaml_url"`
}
