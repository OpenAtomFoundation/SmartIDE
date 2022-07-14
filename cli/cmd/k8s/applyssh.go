package k8s

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/leansoftX/smartide-cli/cmd/server"
	"github.com/leansoftX/smartide-cli/internal/apk/i18n"
	"github.com/leansoftX/smartide-cli/internal/biz/config"
	"github.com/leansoftX/smartide-cli/internal/biz/workspace"
	"github.com/leansoftX/smartide-cli/internal/model"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/leansoftX/smartide-cli/pkg/kubectl"
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

		common.SmartIDELog.Info(i18nInstance.ApplySSH.Info_start)

		resourceid, _ := fflags.GetString(k8s_applyssh_flag_resourceid)
		ports, _ := fflags.GetString(k8s_applyssh_flag_ports)
		serverHost, _ := fflags.GetString(k8s_applyssh_flag_serverhost)
		serverToken, _ := fflags.GetString(k8s_applyssh_flag_servertoken)
		configMapNamespace := "ingress-nginx"

		//1. Get K8s Resource
		auth := model.Auth{}
		auth.LoginUrl = serverHost
		auth.Token = serverToken
		resourceInfo, err := server.GetResourceByID(auth, resourceid)
		common.CheckError(err)
		if resourceInfo == nil {
			common.SmartIDELog.Error(fmt.Sprintf("根据ID（%v）未找到资源数据！", resourceid))
		}

		//2. Save temp k8s config file
		tempK8sConfigFileRelativePath := common.PathJoin(config.SmartIdeHome, "tempconfig")
		err = ioutil.WriteFile(tempK8sConfigFileRelativePath, []byte(resourceInfo.KubeConfig), 0777)
		if err != nil {
			common.SmartIDELog.Error(err)
		}
		k8sUtil, err := kubectl.NewK8sUtil(tempK8sConfigFileRelativePath,
			resourceInfo.KubeContext,
			configMapNamespace)
		if err != nil {
			common.SmartIDELog.Error(err)
		}

		//3. Construct Config Map
		configMap := &kubectl.ConfigMap{
			APIVersion: "v1",
			Kind:       "ConfigMap",
			Metadata: struct {
				Name      string "yaml:\"name\""
				Namespace string "yaml:\"namespace\""
			}{
				Name:      "ingress-nginx-tcp",
				Namespace: "ingress-nginx",
			},
			Data: map[string]string{},
		}
		addWorkspaces := []string{}
		removeWorkspaces := []string{}
		portList := strings.Split(ports, ";")
		for _, port := range portList {
			portInfo := strings.Split(port, ":")
			externalport := portInfo[0]
			service := portInfo[1]
			internalport := portInfo[2]
			workspace := portInfo[3]
			if !strings.Contains(workspace, "remove") {
				configMap.Data[externalport] = fmt.Sprintf("%v:%v", service, internalport)
			}
			if strings.Contains(workspace, "add") {
				addWorkspaces = append(addWorkspaces, fmt.Sprintf("%v:%v:%v", strings.Split(workspace, "-")[0], service, externalport))
			}
			if strings.Contains(workspace, "remove") {
				removeWorkspaces = append(removeWorkspaces, fmt.Sprintf("%v:%v:%v", strings.Split(workspace, "-")[0], service, externalport))
			}
		}

		//4. Save Config Map to Temp Yaml
		configMapYamlData, err := yaml.Marshal(&configMap)
		if err != nil {
			common.SmartIDELog.Error(err)
		}
		tempK8sConfigMapYamlFilePath := common.PathJoin(config.SmartIdeHome, "k8s_configmap_temp.yaml")
		err = ioutil.WriteFile(tempK8sConfigMapYamlFilePath, []byte(configMapYamlData), 0777)
		if err != nil {
			common.SmartIDELog.Error(err)
		}

		//5. Kubectl Apply
		common.SmartIDELog.Info(i18nInstance.ApplySSH.Info_log_enable_ssh_start)
		err = k8sUtil.ExecKubectlCommandRealtime(fmt.Sprintf("apply -f %v", tempK8sConfigMapYamlFilePath), "", false)
		if err != nil {
			common.SmartIDELog.Error(err)
		}
		common.SmartIDELog.Info(i18nInstance.ApplySSH.Info_log_enable_ssh_success)

		//6. Callback and log
		wsURL := fmt.Sprint(strings.ReplaceAll(strings.ReplaceAll(serverHost, "https", "ws"), "http", "ws"), "/ws/smartide/ws")
		common.WebsocketStart(wsURL)
		for _, addWorkspaceInfo := range addWorkspaces {
			if strings.Contains(addWorkspaceInfo, "WS") {
				addWorkspaceId := strings.Split(addWorkspaceInfo, ":")[0]
				addWorkspaceService := strings.Split(strings.Split(addWorkspaceInfo, ":")[1], "/")[1]
				addWorkspaceExternalPort := strings.Split(addWorkspaceInfo, ":")[2]
				if pid, err := workspace.GetParentId(addWorkspaceId, 1, serverToken, serverHost); err == nil && pid > 0 {
					common.SmartIDELog.Ws_id = addWorkspaceId
					common.SmartIDELog.ParentId = pid
					common.SmartIDELog.Info("-----------------------")
					common.SmartIDELog.Info(fmt.Sprintf(i18nInstance.ApplySSH.Info_log_service_enable_ssh_success, addWorkspaceId, addWorkspaceService, addWorkspaceExternalPort))
					common.SmartIDELog.Info("-----------------------")
					// feedback
					feedbackMap := make(map[string]interface{})
					feedbackMap["port"] = ""
					feedbackMap["url"] = ""
					currentAuth := model.Auth{
						LoginUrl: serverHost,
						Token:    serverToken,
					}
					workspaceInfo, err := workspace.GetWorkspaceFromServer(currentAuth, addWorkspaceId, workspace.CliRunningEvnEnum_Server)
					common.CheckError(err)
					for index, portDetail := range workspaceInfo.Extend.Ports {
						if portDetail.HostPortDesc == "tools-ssh" && portDetail.ServiceName == addWorkspaceService {
							workspaceInfo.Extend.Ports[index].SSHPort = addWorkspaceExternalPort
						}
					}
					err = server.Feedback_Finish(server.FeedbackCommandEnum_ApplySSH, cmd, true, nil, *workspaceInfo, "", "") //(currentAuth, *workspaceInfo)
					common.CheckError(err)
				}
			}
		}
		for _, removeWorkspaceInfo := range removeWorkspaces {
			if strings.Contains(removeWorkspaceInfo, "WS") {
				removeWorkspaceId := strings.Split(removeWorkspaceInfo, ":")[0]
				removeWorkspaceService := strings.Split(strings.Split(removeWorkspaceInfo, ":")[1], "/")[1]
				removeWorkspaceExternalPort := strings.Split(removeWorkspaceInfo, ":")[2]
				if pid, err := workspace.GetParentId(removeWorkspaceId, 1, serverToken, serverHost); err == nil && pid > 0 {
					common.SmartIDELog.Ws_id = removeWorkspaceId
					common.SmartIDELog.ParentId = pid
					common.SmartIDELog.Info("-----------------------")
					common.SmartIDELog.Info(fmt.Sprintf(i18nInstance.ApplySSH.Info_log_service_disable_ssh_success, removeWorkspaceId, removeWorkspaceService, removeWorkspaceExternalPort))
					common.SmartIDELog.Info("-----------------------")
					// feedback
					feedbackMap := make(map[string]interface{})
					feedbackMap["port"] = ""
					feedbackMap["url"] = ""
					currentAuth := model.Auth{
						LoginUrl: serverHost,
						Token:    serverToken,
					}
					workspaceInfo, err := workspace.GetWorkspaceFromServer(currentAuth, removeWorkspaceId, workspace.CliRunningEvnEnum_Server)
					common.CheckError(err)
					for index, portDetail := range workspaceInfo.Extend.Ports {
						if portDetail.HostPortDesc == "tools-ssh" && portDetail.ServiceName == removeWorkspaceService {
							workspaceInfo.Extend.Ports[index].SSHPort = ""
						}
					}

					err = server.Feedback_Finish(server.FeedbackCommandEnum_ApplySSH, cmd, true, nil, *workspaceInfo, "", "") //(currentAuth, *workspaceInfo)
					common.CheckError(err)
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
