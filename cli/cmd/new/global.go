/*
 * @Date: 2022-04-20 11:03:53
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-11-09 15:20:18
 * @FilePath: /cli/cmd/new/global.go
 */

package new

import (
	"encoding/json"
	"fmt"
	"strings"

	smartideServer "github.com/leansoftX/smartide-cli/cmd/server"

	"github.com/leansoftX/smartide-cli/internal/apk/i18n"
	"github.com/leansoftX/smartide-cli/internal/biz/workspace"
	"github.com/leansoftX/smartide-cli/internal/model"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/leansoftX/smartide-cli/pkg/k8s"
	"github.com/spf13/cobra"
)

var i18nInstance = i18n.GetInstance()

func preRun(cmd *cobra.Command, workspaceInfo workspace.WorkspaceInfo, logAction workspace.ActionEnum, k8sUtil *k8s.KubernetesUtil) func(err error) {
	mode, _ := cmd.Flags().GetString("mode")
	isModeServer := strings.ToLower(mode) == "server"
	// 错误反馈
	serverFeedback := func(err error) {
		if !isModeServer {
			return
		}
		if err != nil {
			errFeedback, ok := err.(*model.FeedbackError)
			if !ok {
				tmp := model.CreateFeedbackError(err.Error(), true)
				errFeedback = &tmp
			} else { // 不需要重试的错误，直接删除namespace
				if !errFeedback.IsRetry && k8sUtil != nil && workspaceInfo.Mode == workspace.WorkingMode_K8s {
					k8sUtil.ExecKubectlCommandCombined("delete namespace --force "+workspaceInfo.K8sInfo.Namespace, "")
				}
			}
			errMsgBits, _ := json.Marshal(errFeedback)
			smartideServer.Feedback_Finish(smartideServer.FeedbackCommandEnum_New, cmd, false, nil, workspaceInfo, string(errMsgBits), "")
			common.CheckError(err) // exit
		}
	}

	// 日志
	if apiHost, _ := cmd.Flags().GetString(smartideServer.Flags_ServerHost); apiHost != "" {
		wsURL := fmt.Sprint(strings.ReplaceAll(strings.ReplaceAll(apiHost, "https", "ws"), "http", "ws"), "/ws/smartide/ws")
		common.WebsocketStart(wsURL)
		token, _ := cmd.Flags().GetString(smartideServer.Flags_ServerToken)
		if token != "" {
			if workspaceIdStr, _ := cmd.Flags().GetString(smartideServer.Flags_ServerWorkspaceid); workspaceIdStr != "" {
				if no, _ := workspace.GetWorkspaceNo(workspaceIdStr, token, apiHost); no != "" {
					if pid, err := workspace.GetParentId(no, logAction, token, apiHost); err == nil && pid > 0 {
						common.SmartIDELog.Ws_id = no
						common.SmartIDELog.ParentId = pid
					}
				}
			}

		}
	}

	return serverFeedback
}

/*
// 从command的参数中获取模板设置信息（下载模板库文件到本地）
func GetTemplateSetting(cmd *cobra.Command, args []string) (*templateModel.SelectedTemplateTypeBo, error) {
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
	//2.1.
	selectedTemplateTypeName := ""
	if len(args) > 0 {
		selectedTemplateTypeName = args[0]
	}
	selectedTemplateTypeName = strings.TrimSpace(selectedTemplateTypeName)
	//2.2.
	selectedTemplateSubTypeName, err := cmd.Flags().GetString("type")
	selectedTemplateSubTypeName = strings.TrimSpace(selectedTemplateSubTypeName)
	if err != nil {
		return nil, err
	}
	if selectedTemplateSubTypeName == "" {
		selectedTemplateSubTypeName = "_default"
	}

	//3. 遍历进行查找
	var selectedTemplate *templateModel.SelectedTemplateTypeBo
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
				tmp := templateModel.SelectedTemplateTypeBo{
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
	selectedTemplate.TemplateActualRepoUrl = config.GlobalSmartIdeConfig.TemplateActualRepoUrl
	return selectedTemplate, nil
}

// 加载templates索引json
func loadTemplatesJson() (templateTypes []templateModel.TemplateTypeInfo, err error) {
	// new type转换为结构体
	templatesPath := common.PathJoin(config.SmartIdeHome, golbalModel.TMEPLATE_DIR_NAME, "templates.json")
	templatesByte, err := os.ReadFile(templatesPath)
	if err != nil {
		return templateTypes, errors.New(i18nInstance.New.Err_read_templates + templatesPath + err.Error())
	}

	err = json.Unmarshal(templatesByte, &templateTypes)
	return templateTypes, err
}

// clone模版repo
func templatesClone() error {
	templatePath := filepath.Join(config.SmartIdeHome, golbalModel.TMEPLATE_DIR_NAME)
	templateGitPath := filepath.Join(templatePath, ".git")
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

// 打印 service 列表
func printTemplates(newType []templateModel.TemplateTypeInfo) {
	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
	fmt.Fprintln(w, i18nInstance.New.Info_templates_list_header)
	for i := 0; i < len(newType); i++ {
		line := fmt.Sprintf("%v\t%v", newType[i].TypeName, "_default")
		fmt.Fprintln(w, line)
		for j := 0; j < len(newType[i].SubTypes); j++ {
			subTypeName := newType[i].SubTypes[j]
			if subTypeName != (templateModel.SubType{}) && subTypeName.Name != "" {
				line := fmt.Sprintf("%v\t%v", newType[i].TypeName, subTypeName.Name)
				fmt.Fprintln(w, line)
			}
		}
	}
	w.Flush()
	fmt.Println("")
}
*/
