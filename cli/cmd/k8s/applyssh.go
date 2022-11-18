package k8s

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"github.com/leansoftX/smartide-cli/internal/apk/appinsight"
	"github.com/leansoftX/smartide-cli/cmd/server"
	"github.com/leansoftX/smartide-cli/internal/apk/i18n"
	"github.com/leansoftX/smartide-cli/internal/biz/config"
	"github.com/leansoftX/smartide-cli/internal/biz/workspace"
	"github.com/leansoftX/smartide-cli/internal/model"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/leansoftX/smartide-cli/pkg/k8s"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v2"
)

var i18nInstance = i18n.GetInstance()

var (
	k8s_applyssh_flag_resourceid  = "resourceid"
	k8s_applyssh_flag_ports       = "ports"
	k8s_applyssh_flag_mode        = "mode"
	k8s_applyssh_flag_serverhost  = "serverhost"
	k8s_applyssh_flag_servertoken = "servertoken"
)

// initCmd represents the init command
var ApplySSHCmd = &cobra.Command{
	Use:     "applyssh",
	Short:   i18nInstance.ApplySSH.Info_help_short,
	Long:    i18nInstance.ApplySSH.Info_help_long,
	Aliases: []string{"ssh"},
	Example: `  smartide k8s applyssh --resourceid <resourceid> --ports <configmap ports string> --mode <mode> --serverhost <serverhost>  --servertoken <servertoken>`,
	Run: func(cmd *cobra.Command, args []string) {
		common.SmartIDELog.Info(i18nInstance.ApplySSH.Info_start)

		// 获取参数
		fflags := cmd.Flags()
		checkFlagErr := checkFlag(fflags, k8s_applyssh_flag_resourceid)
		if checkFlagErr != nil {
			common.SmartIDELog.Error(checkFlagErr)
		}
		checkFlagErr = checkFlag(fflags, k8s_applyssh_flag_ports)
		if checkFlagErr != nil {
			common.SmartIDELog.Error(checkFlagErr)
		}
		checkFlagErr = checkFlag(fflags, k8s_applyssh_flag_serverhost)
		if checkFlagErr != nil {
			common.SmartIDELog.Error(checkFlagErr)
		}
		checkFlagErr = checkFlag(fflags, k8s_applyssh_flag_servertoken)
		if checkFlagErr != nil {
			common.SmartIDELog.Error(checkFlagErr)
		}
		checkFlagErr = checkFlag(fflags, k8s_applyssh_flag_mode)
		if checkFlagErr != nil {
			common.SmartIDELog.Error(checkFlagErr)
		}

		resourceid, _ := fflags.GetString(k8s_applyssh_flag_resourceid)
		ports, _ := fflags.GetString(k8s_applyssh_flag_ports)
		serverHost, _ := fflags.GetString(k8s_applyssh_flag_serverhost)
		serverToken, _ := fflags.GetString(k8s_applyssh_flag_servertoken)
		configMapNamespace := "smartide-ingress-nginx"

		currentAuth := model.Auth{
			LoginUrl: serverHost,
			Token:    serverToken,
		}

		type ApplySshInfo struct {
			PublicPort   int
			ServiceFull  string
			ServiceName  string
			Namespace    string
			InternalPort int
			WorkspaceNo  string
			Action       string // remove, add, empty
		}

		//3. parse ports && Construct Config Map
		configMap := &k8s.ConfigMap{
			APIVersion: "v1",
			Kind:       "ConfigMap",
			Metadata: struct {
				Name      string "yaml:\"name\""
				Namespace string "yaml:\"namespace\""
			}{
				Name:      "ingress-nginx-tcp",
				Namespace: "smartide-ingress-nginx",
			},
			Data: map[string]string{},
		}
		// <外部端口>:<命名空间>/<服务名称>:<内部端口>:<工作区ID>-[<新增或删除的标识>]
		// e.g. 22001:ccdpko/ruoyi-cloud-dev:6822:KWS005-;22002:l494kb/boathouse-calculator-service:6822:KWS006-;22003:g9o07d/ruoyi-cloud-dev:6822:KWS007-add
		applySshArray := []ApplySshInfo{}
		portList := strings.Split(ports, ";")
		for _, port := range portList {
			portInfo := strings.Split(port, ":")

			applySshInfo := ApplySshInfo{}
			applySshInfo.PublicPort, _ = strconv.Atoi(portInfo[0])
			applySshInfo.ServiceFull = portInfo[1]
			applySshInfo.Namespace = strings.Split(applySshInfo.ServiceFull, "/")[0]
			applySshInfo.ServiceName = strings.Split(applySshInfo.ServiceFull, "/")[1]
			applySshInfo.InternalPort, _ = strconv.Atoi(portInfo[2])
			workspaceStr := strings.Split(portInfo[3], "-")
			applySshInfo.WorkspaceNo = workspaceStr[0]
			applySshInfo.Action = workspaceStr[1]

			// 添加到yaml中
			if applySshInfo.Action != "remove" {
				configMap.Data[fmt.Sprint(applySshInfo.PublicPort)] = fmt.Sprintf("%v:%v", applySshInfo.ServiceFull, applySshInfo.InternalPort)
			}

			applySshArray = append(applySshArray, applySshInfo)
		}

		// 反馈错误
		feedbackError := func(feedbackError error) {
			for _, applySsh := range applySshArray {
				if applySsh.Action == "" { // 非新增和删除ssh端口不需要反馈错误
					continue
				}

				workspaceInfo, _ := workspace.GetWorkspaceFromServer(currentAuth, applySsh.WorkspaceNo, workspace.CliRunningEvnEnum_Server)
				if feedbackError != nil {
					server.Feedback_Finish(server.FeedbackCommandEnum_ApplySSH, cmd, false, nil, *workspaceInfo, feedbackError.Error(), "")
					common.CheckError(feedbackError)
				}
			}

		}
		var workID []string
		for _, service := range applySshArray {
			workID = append(workID, service.WorkspaceNo)
		}
		appinsight.SetAllTrack(appinsight.Cli_K8s_Ssh_Apply, args, "", "", strings.Join(workID, ","), "", "", "")

		//1. Get K8s Resource
		auth := model.Auth{}
		auth.LoginUrl = serverHost
		auth.Token = serverToken
		resourceInfo, err := server.GetResourceByID(auth, resourceid)
		common.CheckError(err)
		if resourceInfo == nil {
			common.SmartIDELog.Error(fmt.Sprintf("根据ID（%v）未找到资源数据！", resourceid))
			return
		}

		//2. Save temp k8s config file
		tempK8sConfigFileAbsolutePath := common.PathJoin(config.SmartIdeHome, "tempconfig")
		err = os.WriteFile(tempK8sConfigFileAbsolutePath, []byte(resourceInfo.KubeConfig), 0777)
		feedbackError(err)
		k8sUtil, err := k8s.NewK8sUtilWithFile(tempK8sConfigFileAbsolutePath,
			resourceInfo.KubeContext,
			configMapNamespace)
		feedbackError(err)

		//4. Save Config Map to Temp Yaml
		configMapYamlData, err := yaml.Marshal(&configMap)
		feedbackError(err)
		tempK8sConfigMapYamlFilePath := common.PathJoin(config.SmartIdeHome, "k8s_configmap_temp.yaml")
		err = os.WriteFile(tempK8sConfigMapYamlFilePath, []byte(configMapYamlData), 0777)
		feedbackError(err)

		//5. Kubectl Apply
		common.SmartIDELog.Info(i18nInstance.ApplySSH.Info_log_enable_ssh_start)
		err = k8sUtil.ExecKubectlCommandRealtime(fmt.Sprintf("apply -f %v", tempK8sConfigMapYamlFilePath), "", false)
		feedbackError(err)
		common.SmartIDELog.Info(i18nInstance.ApplySSH.Info_log_enable_ssh_success)

		//6. Callback and log
		wsURL := fmt.Sprint(strings.ReplaceAll(strings.ReplaceAll(serverHost, "https", "ws"), "http", "ws"), "/ws/smartide/ws")
		common.WebsocketStart(wsURL)
		for _, applySsh := range applySshArray {
			if applySsh.Action != "" {
				/* 				sshAction := workspace.ActionEnum_SSH_Disable
				   				if applySsh.Action == "add" {
				   					sshAction = workspace.ActionEnum_SSH_Enable
				   				} */
				workspaceInfo, err := workspace.GetWorkspaceFromServer(currentAuth, applySsh.WorkspaceNo, workspace.CliRunningEvnEnum_Server)
				common.CheckError(err)

				title := "??"
				if applySsh.Action == "add" {
					title = "创建SSH通道"
				} else if applySsh.Action == "remove" {
					title = "删除SSH通道"
				}
				pid, err := workspace.CreateWsLog(workspaceInfo.ServerWorkSpace.NO, currentAuth.Token.(string), currentAuth.LoginUrl, title, "", common.SmartIDELog.TekEventId)
				if err == nil {
					//if pid, err := workspace.GetParentId(workspaceInfo.ServerWorkSpace.NO, sshAction, currentAuth.Token.(string), currentAuth.LoginUrl); err == nil && pid > 0 {
					common.SmartIDELog.Ws_id = workspaceInfo.ServerWorkSpace.NO
					common.SmartIDELog.ParentId = pid
					//}
				}

				// log
				common.SmartIDELog.Info("-----------------------")
				if applySsh.Action == "add" {
					common.SmartIDELog.Info(fmt.Sprintf(i18nInstance.ApplySSH.Info_log_service_enable_ssh_success,
						applySsh.WorkspaceNo, applySsh.ServiceFull, applySsh.PublicPort))
				} else if applySsh.Action == "remove" {
					common.SmartIDELog.Info(fmt.Sprintf(i18nInstance.ApplySSH.Info_log_service_disable_ssh_success,
						applySsh.WorkspaceNo, applySsh.ServiceFull, applySsh.PublicPort))
				}
				common.SmartIDELog.Info("-----------------------")

				// 反馈给
				for index, portDetail := range workspaceInfo.Extend.Ports {
					if portDetail.HostPortDesc == "tools-ssh" && portDetail.ServiceName == applySsh.ServiceName {
						if applySsh.Action == "add" {
							workspaceInfo.Extend.Ports[index].SSHPort = fmt.Sprint(applySsh.PublicPort)
							workspaceInfo.Extend.Ports[index].IsConnected = true
						} else if applySsh.Action == "remove" {
							workspaceInfo.Extend.Ports[index].SSHPort = ""
							workspaceInfo.Extend.Ports[index].IsConnected = false
						}

						err = server.Feedback_Finish(server.FeedbackCommandEnum_ApplySSH, cmd, true, nil, *workspaceInfo, "", "")
						common.CheckError(err)
					} else {
						common.SmartIDELog.Importance("没有找到对应的port信息")
					}
				}

			}
		}

	},
}

// 检查参数是否填写
func checkFlag(fflags *pflag.FlagSet, flagName string) error {
	if !fflags.Changed(flagName) {
		return fmt.Errorf(i18nInstance.Main.Err_flag_value_required, flagName)
	}
	return nil
}

func init() {
	ApplySSHCmd.Flags().StringP(k8s_applyssh_flag_resourceid, "", "", i18nInstance.ApplySSH.Info_help_flag_resourceid)
	ApplySSHCmd.Flags().StringP(k8s_applyssh_flag_ports, "", "", i18nInstance.ApplySSH.Info_help_flag_ports)
	ApplySSHCmd.Flags().StringP(k8s_applyssh_flag_mode, "", "", i18nInstance.ApplySSH.Info_help_flag_mode)
	ApplySSHCmd.Flags().StringP(k8s_applyssh_flag_serverhost, "", "", i18nInstance.ApplySSH.Info_help_flag_serverhost)
	ApplySSHCmd.Flags().StringP(k8s_applyssh_flag_servertoken, "", "", i18nInstance.ApplySSH.Info_help_flag_servertoken)
}
