/*
 * @Author: jason chen (jasonchen@leansoftx.com, http://smallidea.cnblogs.com)
 * @Description:
 * @Date: 2021-11
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-09-27 12:19:13
 */
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/leansoftX/smartide-cli/internal/biz/workspace"
	"github.com/leansoftX/smartide-cli/internal/dal"
	"github.com/leansoftX/smartide-cli/internal/model"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var listCmd = &cobra.Command{
	Use:     "list",
	Short:   i18nInstance.List.Info_help_short,
	Long:    i18nInstance.List.Info_help_long,
	Aliases: []string{"ls"},
	Example: `  smartide list`,
	Run: func(cmd *cobra.Command, args []string) {

		common.SmartIDELog.Info(i18nInstance.List.Info_start)
		cliRunningEnv := workspace.CliRunningEnvEnum_Client
		if value, _ := cmd.Flags().GetString("mode"); strings.ToLower(value) == "server" {
			cliRunningEnv = workspace.CliRunningEvnEnum_Server
		}
		printWorkspaces(cliRunningEnv)
		common.SmartIDELog.Info(i18nInstance.List.Info_end)
	},
}

// 打印 service 列表
func printWorkspaces(cliRunningEnv workspace.CliRunningEvnEnum) {
	workspaceInfos, err := dal.GetWorkspaceList()
	common.CheckError(err)

	auth, err := workspace.GetCurrentUser()
	common.CheckError(err)
	if auth != (model.Auth{}) && auth.Token != "" {
		// 从api 获取workspace
		serverWorkSpaces, err := workspace.GetServerWorkspaceList(auth, cliRunningEnv)

		if err != nil { // 有错误仅给警告
			common.SmartIDELog.Importance("从服务器获取工作区列表失败，" + err.Error())
		} else { //
			workspaceInfos = append(workspaceInfos, serverWorkSpaces...)
		}
	}
	if len(workspaceInfos) <= 0 {
		common.SmartIDELog.Info(i18nInstance.List.Info_dal_none)
		return
	}
	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
	fmt.Fprintln(w, i18nInstance.List.Info_workspace_list_header)

	// 等于标题字符的长度
	tmpArray := strings.Split(i18nInstance.List.Info_workspace_list_header, "\t")
	outputArray := []string{}
	for _, str := range tmpArray {
		chars := ""
		for i := 0; i < len(str); i++ {
			chars += "-"
		}
		outputArray = append(outputArray, chars)
	}
	fmt.Fprintln(w, strings.Join(outputArray, "\t"))

	// 内容
	for _, worksapceInfo := range workspaceInfos {
		dir := worksapceInfo.WorkingDirectoryPath
		if len(dir) <= 0 {
			dir = "-"
		}
		configFile := worksapceInfo.ConfigFileRelativePath
		if len(configFile) <= 0 {
			configFile = "-"
		} else {
			if strings.Index(configFile, ".ide") == 0 ||
				strings.Index(configFile, "/.ide") == 0 ||
				strings.Index(configFile, "\\.ide") == 0 {
				configFile = filepath.Base(configFile)
			}
		}
		createTime := ""
		if worksapceInfo.Mode != workspace.WorkingMode_Local {
			local, _ := time.LoadLocation("Local")                                         // 北京时区
			createTime = worksapceInfo.CreatedTime.In(local).Format("2006-01-02 15:04:05") // 格式化输出
		} else {
			createTime = worksapceInfo.CreatedTime.Format("2006-01-02 15:04:05") // 格式化输出
		}
		host := "-"
		if (worksapceInfo.Remote != workspace.RemoteInfo{}) {
			host = fmt.Sprint(worksapceInfo.Remote.UserName, "@", worksapceInfo.Remote.Addr, ":", worksapceInfo.Remote.SSHPort)
		}
		workspaceName := worksapceInfo.Name
		if worksapceInfo.ServerWorkSpace != nil {
			label := worksapceInfo.ServerWorkSpace.Status.GetDesc()
			workspaceName = fmt.Sprintf("%v (%v)", workspaceName, label)
		}
		gitBranch := worksapceInfo.GitBranch
		if gitBranch == "" {
			gitBranch = "master"
		}

		line := fmt.Sprintf("%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v", worksapceInfo.ID, workspaceName, worksapceInfo.Mode,
			worksapceInfo.GitCloneRepoUrl, gitBranch, configFile, host, createTime)
		fmt.Fprintln(w, line)
	}
	w.Flush()
}

func init() {

}
