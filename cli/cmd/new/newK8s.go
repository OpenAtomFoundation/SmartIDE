/*
 * @Date: 2022-10-27 09:35:51
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-10-27 15:22:32
 * @FilePath: /cli/cmd/new/newK8s.go
 */
/*
 * @Date: 2022-04-20 10:46:40
 * @LastEditors: kenan
 * @LastEditTime: 2022-10-20 10:10:08
 * @FilePath: /cli/cmd/new/newVm.go
 */

package new

import (
	"errors"

	"github.com/leansoftX/smartide-cli/cmd/start"
	"github.com/leansoftX/smartide-cli/internal/biz/config"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/leansoftX/smartide-cli/pkg/k8s"

	// templateModel "github.com/leansoftX/smartide-cli/internal/biz/template/model"
	"github.com/leansoftX/smartide-cli/internal/biz/workspace"
	"github.com/spf13/cobra"
)

func K8sNew_Local(cmd *cobra.Command, args []string,
	k8sUtil *k8s.KubernetesUtil,
	workspaceInfo workspace.WorkspaceInfo,
	//selectedTemplate templateModel.SelectedTemplateTypeBo,
	yamlExecuteFun func(yamlConfig config.SmartIdeConfig)) {

	//1. clone 模板文件到本地
	//1.1. 获取 command 中的配置
	selectedTemplateSettings, err := getTemplateSetting(cmd, args) // 包含了“clone 模板文件到本地”
	common.CheckError(err)
	if selectedTemplateSettings == nil { // 未指定模板类型的时候，提示用户后退出
		common.CheckError(errors.New("模板配置为空！"))
		return // 退出
	}
	workspaceInfo.SelectedTemplate = selectedTemplateSettings

	//1.3. 调用 k8s start 方法，传递项目文件副本所在的本地目录、项目文件所在的相对目录、配置文件名称（包含在workspace对象中）、git clone url
	//1.3.1. 根据 “项目文件副本所在的本地目录、项目文件所在的相对目录、配置文件名称（包含在workspace对象中）” 加载配置文件
	//1.3.2. 根据 “git clone url ” clone 代码到pod的指定目录，根据 “项目文件所在的相对目录” 拷贝文件到 “项目文件夹” 中
	/*
	   git clone {template_actual_repo_url} ~/.ide/template
	   mv ~/.ide/template/{relative_dir_path} ~/projects/{project_name}
	*/
	workspaceInfo, err = start.ExecuteK8s_LocalWS_LocalEnv(cmd, *k8sUtil, workspaceInfo, yamlExecuteFun)
	common.CheckError(err)

}
