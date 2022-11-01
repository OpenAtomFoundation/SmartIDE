/*
 * @Date: 2022-04-20 11:03:53
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-11-01 17:01:39
 * @FilePath: /cli/cmd/new/global.go
 */

package new

import (
	"errors"
	"fmt"
	"strings"

	smartideServer "github.com/leansoftX/smartide-cli/cmd/server"
	templateModel "github.com/leansoftX/smartide-cli/internal/biz/template/model"

	"github.com/leansoftX/smartide-cli/internal/apk/i18n"
	"github.com/leansoftX/smartide-cli/internal/biz/config"
	"github.com/leansoftX/smartide-cli/internal/biz/workspace"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/spf13/cobra"
)

var i18nInstance = i18n.GetInstance()

func preRun(cmd *cobra.Command, workspaceInfo workspace.WorkspaceInfo) func(err error) {
	mode, _ := cmd.Flags().GetString("mode")
	isModeServer := strings.ToLower(mode) == "server"
	// 错误反馈
	serverFeedback := func(err error) {
		if !isModeServer {
			return
		}
		if err != nil {
			smartideServer.Feedback_Finish(smartideServer.FeedbackCommandEnum_New, cmd, false, nil, workspaceInfo, err.Error(), "")
			common.CheckError(err)
		}
	}

	if apiHost, _ := cmd.Flags().GetString(smartideServer.Flags_ServerHost); apiHost != "" {
		wsURL := fmt.Sprint(strings.ReplaceAll(strings.ReplaceAll(apiHost, "https", "ws"), "http", "ws"), "/ws/smartide/ws")
		common.WebsocketStart(wsURL)
		token, _ := cmd.Flags().GetString(smartideServer.Flags_ServerToken)
		if token != "" {
			if workspaceIdStr, _ := cmd.Flags().GetString(smartideServer.Flags_ServerWorkspaceid); workspaceIdStr != "" {
				if no, _ := workspace.GetWorkspaceNo(workspaceIdStr, token, apiHost); no != "" {
					if pid, err := workspace.GetParentId(no, workspace.ActionEnum_Workspace_Start, token, apiHost); err == nil && pid > 0 {
						common.SmartIDELog.Ws_id = no
						common.SmartIDELog.ParentId = pid
					}
				}
			}

		}
	}

	return serverFeedback
}

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
