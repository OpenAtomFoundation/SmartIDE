package k8s

import (
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/leansoftX/smartide-cli/cmd/server"
	"github.com/leansoftX/smartide-cli/internal/biz/config"
	"github.com/leansoftX/smartide-cli/internal/model"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/leansoftX/smartide-cli/pkg/k8s"
	"github.com/spf13/cobra"
)

var (
	k8s_init_flag_resourceid  = "resourceid"
	k8s_init_flag_mode        = "mode"
	k8s_init_flag_serverhost  = "serverhost"
	k8s_init_flag_servertoken = "servertoken"
)

// initCmd represents the init command
var K8sInitCmd = &cobra.Command{
	Use:     "init",
	Short:   i18nInstance.K8sInit.Info_help_short,
	Long:    i18nInstance.K8sInit.Info_help_long,
	Aliases: []string{"init"},
	Example: `  smartide k8s init --resourceid <resourceid> --mode <mode> --serverhost <serverhost>  --servertoken <servertoken>`,
	Run: func(cmd *cobra.Command, args []string) {
		common.SmartIDELog.Info(i18nInstance.K8sInit.Info_start)

		// 获取参数
		fflags := cmd.Flags()
		checkFlagErr := checkFlag(fflags, k8s_init_flag_resourceid)
		if checkFlagErr != nil {
			common.SmartIDELog.Error(checkFlagErr)
		}
		checkFlagErr = checkFlag(fflags, k8s_init_flag_serverhost)
		if checkFlagErr != nil {
			common.SmartIDELog.Error(checkFlagErr)
		}
		checkFlagErr = checkFlag(fflags, k8s_init_flag_servertoken)
		if checkFlagErr != nil {
			common.SmartIDELog.Error(checkFlagErr)
		}
		checkFlagErr = checkFlag(fflags, k8s_init_flag_mode)
		if checkFlagErr != nil {
			common.SmartIDELog.Error(checkFlagErr)
		}

		resourceid, _ := fflags.GetString(k8s_init_flag_resourceid)
		serverHost, _ := fflags.GetString(k8s_init_flag_serverhost)
		serverToken, _ := fflags.GetString(k8s_init_flag_servertoken)
		defaultNamespace := "default"

		currentAuth := model.Auth{
			LoginUrl: serverHost,
			Token:    serverToken,
		}

		//1. Get K8s Resource
		resourceInfo, err := server.GetResourceByID(currentAuth, resourceid)
		common.CheckError(err)
		if resourceInfo == nil {
			common.SmartIDELog.Error(fmt.Sprintf("根据ID（%v）未找到资源数据！", resourceid))
			return
		}

		//2. Save temp k8s config file
		tempK8sConfigFileAbsolutePath := common.PathJoin(config.SmartIdeHome, "tempconfig")
		err = ioutil.WriteFile(tempK8sConfigFileAbsolutePath, []byte(resourceInfo.KubeConfig), 0777)
		k8sUtil, err := k8s.NewK8sUtilWithFile(tempK8sConfigFileAbsolutePath,
			resourceInfo.KubeContext,
			defaultNamespace)
		k8sUtil.Commands = strings.Split(k8sUtil.Commands, "--namespace")[0]

		//3. Kubectl apply ingress controller
		common.SmartIDELog.Info(i18nInstance.K8sInit.Info_log_apply_ingress_controller_start)
		ingressControllerYamlPath := fmt.Sprintf("%v/api/smartide/ingress-controller.yaml", serverHost)
		if resourceInfo.CertType == 2 {
			ingressControllerYamlPath = fmt.Sprintf("%v/api/smartide/ingress-controller-default-cert.yaml", serverHost)
		}
		err = k8sUtil.ExecKubectlCommandRealtime(fmt.Sprintf("apply -f %v", ingressControllerYamlPath), "", false)
		common.SmartIDELog.Info(i18nInstance.K8sInit.Info_log_apply_ingress_controller_success)

		//4. Https static certificate, Kubectl create certficate secret
		if resourceInfo.CertType == 2 {
			common.SmartIDELog.Info(i18nInstance.K8sInit.Info_log_create_certificate_secret_start)
			crtFileAbsolutePath := common.PathJoin(config.SmartIdeHome, "ssl_cert.crt")
			err = ioutil.WriteFile(crtFileAbsolutePath, []byte(resourceInfo.CertCrt), 0777)
			keyFileAbsolutePath := common.PathJoin(config.SmartIdeHome, "ssl_key.key")
			err = ioutil.WriteFile(keyFileAbsolutePath, []byte(resourceInfo.CertKey), 0777)
			err = k8sUtil.ExecKubectlCommandRealtime(fmt.Sprintf("create secret tls ws.smartide.cn --key %v --cert %v", keyFileAbsolutePath, crtFileAbsolutePath), "", false)
			common.SmartIDELog.Info(i18nInstance.K8sInit.Info_log_create_certificate_secret_success)
		}

		//5. Https dynamic certificate, Kubectl apply cert-manager.yaml & cluster-issuer.yaml
		if resourceInfo.CertType == 3 {
			common.SmartIDELog.Info(i18nInstance.K8sInit.Info_log_apply_certificate_manager_start)
			certManagerYamlPath := fmt.Sprintf("%v/api/smartide/cert-manager.yaml", serverHost)
			err = k8sUtil.ExecKubectlCommandRealtime(fmt.Sprintf("apply -f %v", certManagerYamlPath), "", false)
			clusterIssuerYamlPath := fmt.Sprintf("%v/api/smartide/cluster-issuer.yaml", serverHost)
			for {
				err = k8sUtil.ExecKubectlCommandRealtime(fmt.Sprintf("apply -f %v", clusterIssuerYamlPath), "", false)
				if err == nil {
					break
				}
				time.Sleep(5 * time.Second)
			}
			common.SmartIDELog.Info(i18nInstance.K8sInit.Info_log_apply_certificate_manager_success)
		}

		//6. Kubectl apply storage class
		common.SmartIDELog.Info(i18nInstance.K8sInit.Info_log_apply_storage_class_start)
		storageClassYamlPath := fmt.Sprintf("%v/api/smartide/smartide-file-storageclass.yaml", serverHost)
		err = k8sUtil.ExecKubectlCommandRealtime(fmt.Sprintf("apply -f %v", storageClassYamlPath), "", false)
		common.SmartIDELog.Info(i18nInstance.K8sInit.Info_log_apply_storage_class_success)

		//7. Get Ingress Controller and Callback
		common.SmartIDELog.Info(i18nInstance.K8sInit.Info_log_feedback_start)
		externalIp, err := k8sUtil.ExecKubectlCommandWithOutputRealtime("get services --namespace ingress-nginx ingress-nginx-controller --output jsonpath='{.status.loadBalancer.ingress[0].ip}'", "")
		resourceInfo.IP = externalIp
		err = server.UpdateResourceByID(currentAuth, resourceInfo)
		common.SmartIDELog.Info(i18nInstance.K8sInit.Info_log_feedback_success)
		common.SmartIDELog.Info(i18nInstance.K8sInit.Info_log_init_success)
	},
}

func init() {
	K8sInitCmd.Flags().StringP(k8s_init_flag_resourceid, "", "", i18nInstance.K8sInit.Info_help_flag_resourceid)
	K8sInitCmd.Flags().StringP(k8s_init_flag_mode, "", "", i18nInstance.K8sInit.Info_help_flag_mode)
	K8sInitCmd.Flags().StringP(k8s_init_flag_serverhost, "", "", i18nInstance.K8sInit.Info_help_flag_serverhost)
	K8sInitCmd.Flags().StringP(k8s_init_flag_servertoken, "", "", i18nInstance.K8sInit.Info_help_flag_servertoken)
}
