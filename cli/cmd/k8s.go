package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/leansoftX/smartide-cli/cmd/k8s"
	"github.com/leansoftX/smartide-cli/cmd/server"
	"github.com/leansoftX/smartide-cli/internal/biz/config"
	"github.com/leansoftX/smartide-cli/internal/biz/workspace"
	"github.com/leansoftX/smartide-cli/internal/model"
	"github.com/leansoftX/smartide-cli/pkg/common"
	k8sLib "github.com/leansoftX/smartide-cli/pkg/k8s"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v2"
)

var (
	k8s_flag_workspaceid = "serverworkspaceid"
	k8s_flag_public_url  = "public-url"
	k8s_flag_mode        = "mode"
	k8s_flag_serverhost  = "serverhost"
	k8s_flag_servertoken = "servertoken"
)

var k8sCmd = &cobra.Command{
	Use:     "k8s",
	Short:   i18nInstance.K8s.Info_help_short,
	Long:    i18nInstance.K8s.Info_help_long,
	Example: `  smartide k8s --serverworkspaceid <serverworkspaceid> --public-url enable|disable --mode <mode> --serverhost <serverhost> --servertoken <servertoken>`,
	Run: func(cmd *cobra.Command, args []string) {
		fflags := cmd.Flags()
		checkFlagErr := checkFlag(fflags, k8s_flag_workspaceid)
		common.CheckError(checkFlagErr)
		checkFlagErr = checkFlag(fflags, k8s_flag_public_url)
		common.CheckError(checkFlagErr)
		checkFlagErr = checkFlag(fflags, k8s_flag_serverhost)
		common.CheckError(checkFlagErr)
		checkFlagErr = checkFlag(fflags, k8s_flag_servertoken)
		common.CheckError(checkFlagErr)
		checkFlagErr = checkFlag(fflags, k8s_flag_mode)
		common.CheckError(checkFlagErr)
		publicUrl, _ := fflags.GetString(k8s_flag_public_url)
		serverHost, _ := fflags.GetString(k8s_flag_serverhost)
		serverToken, _ := fflags.GetString(k8s_flag_servertoken)
		currentAuth := model.Auth{
			LoginUrl: serverHost,
			Token:    serverToken,
		}

		//0. 初始化Log
		if apiHost, _ := fflags.GetString(k8s_flag_serverhost); apiHost != "" {

			workspaceIngressAction := workspace.ActionEnum_Ingress_Enable
			if strings.ToLower(publicUrl) != "enable" {
				workspaceIngressAction = workspace.ActionEnum_Ingress_Disable
			}

			// ws
			workspaceNo := ""
			if workspaceIdStr, _ := cmd.Flags().GetString(server.Flags_ServerWorkspaceid); workspaceIdStr != "" {
				if no, _ := workspace.GetWorkspaceNo(workspaceIdStr, currentAuth.Token.(string), currentAuth.LoginUrl); no != "" {
					workspaceNo = no
				}
			}
			/* 			if pid, err := workspace.GetParentId(workspaceNo, workspaceIngressAction, currentAuth.Token.(string), currentAuth.LoginUrl); err == nil && pid > 0 {
			   				common.SmartIDELog.Ws_id = workspaceNo
			   				common.SmartIDELog.ParentId = pid
			   			} else {

			   			} */

			title := "??"
			if workspaceIngressAction == workspace.ActionEnum_Ingress_Enable {
				title = "创建Ingress"
			} else if workspaceIngressAction == workspace.ActionEnum_Ingress_Disable {
				title = "删除Ingress"
			}
			pid, err := workspace.CreateWsLog(workspaceNo, serverToken, currentAuth.LoginUrl, title, "", common.SmartIDELog.TekEventId)
			if err == nil {
				common.SmartIDELog.Ws_id = workspaceNo
				common.SmartIDELog.ParentId = pid
			}
		}
		common.SmartIDELog.Info("-----------------------")
		common.SmartIDELog.Info(i18nInstance.K8s.Info_start)

		//1. Get Workspace Info
		common.SmartIDELog.Info(i18nInstance.K8s.Info_log_get_workspace_start)
		workspaceInfo, err := getWorkspaceFromCmd(cmd, args)
		entryptionKey4Workspace(workspaceInfo) // 申明需要加密的文本
		common.CheckError(err)
		if workspaceInfo.IsNil() {
			workspaceIdStr := getWorkspaceIdFromFlagsOrArgs(cmd, args)
			common.SmartIDELog.Error(fmt.Sprintf("根据ID（%v）未找到工作区数据!", workspaceIdStr))
		}
		print := fmt.Sprintf(i18nInstance.Get.Info_workspace_detail_template,
			workspaceInfo.ID, workspaceInfo.Name, workspaceInfo.CliRunningEnv, workspaceInfo.Mode, workspaceInfo.ConfigFileRelativePath, workspaceInfo.WorkingDirectoryPath,
			workspaceInfo.GitCloneRepoUrl, workspaceInfo.GitRepoAuthType)
		common.SmartIDELog.Console(print)
		common.SmartIDELog.Info(i18nInstance.K8s.Info_log_get_workspace_success)

		baseDNSName := workspaceInfo.K8sInfo.IngressBaseDnsName
		namespace := workspaceInfo.K8sInfo.Namespace
		authType := workspaceInfo.K8sInfo.IngressAuthType
		username := workspaceInfo.K8sInfo.IngressLoginUserName
		password := workspaceInfo.K8sInfo.IngressLoginPassword

		//2. Save temp k8s config file
		k8sConfigDirPath := config.SmartIdeHome
		tempK8sConfigFileRelativePath := filepath.Join(k8sConfigDirPath, "tempconfig")
		err = os.WriteFile(tempK8sConfigFileRelativePath, []byte(workspaceInfo.K8sInfo.KubeConfigContent), 0777)
		if err != nil {
			common.SmartIDELog.Error(err)
		}
		k8sUtil, err := k8sLib.NewK8sUtilWithFile(tempK8sConfigFileRelativePath,
			workspaceInfo.K8sInfo.Context,
			namespace)
		common.CheckError(err)

		//3. Delete ingress if public-url flag is disable
		if publicUrl == "disable" {
			common.SmartIDELog.Info(i18nInstance.K8s.Info_log_disable_publicurl_start)
			ingressName := "ingress-" + namespace
			err = k8sUtil.ExecKubectlCommandRealtime(fmt.Sprintf("delete ingress %v -n %v", ingressName, namespace), "", false)
			common.CheckError(err)

			if authType == model.KubeAuthenticationTypeEnum_Basic {
				err = k8sUtil.ExecKubectlCommandRealtime(fmt.Sprintf("delete secret %v -n %v", "basic-auth", namespace), "", false)
				common.CheckError(err)
			}

			for index, _ := range workspaceInfo.Extend.Ports {
				workspaceInfo.Extend.Ports[index].IngressUrl = ""
				workspaceInfo.Extend.Ports[index].IsConnected = false
			}

			common.SmartIDELog.Info(i18nInstance.K8s.Info_log_disable_publicurl_success)
			common.SmartIDELog.Info("-----------------------")

			feedbackMap := make(map[string]interface{})
			feedbackMap["port"] = ""
			feedbackMap["url"] = ""

			err = server.FeeadbackExtend(currentAuth, workspaceInfo)
			if err != nil {
				common.SmartIDELog.Error(err)
			}
			return
		}

		//4. Get Resource by Resource ID
		resourceInfo, err := server.GetResourceByID(currentAuth, strconv.Itoa(workspaceInfo.ResourceID))
		common.CheckError(err)
		if resourceInfo == nil {
			common.SmartIDELog.Error(fmt.Sprintf("根据ID（%v）未找到资源数据！", workspaceInfo.ResourceID))
			return
		}

		//5. Construct Ingress Yaml Object
		var yamlData []byte

		//Use WorkspaceIngress construct when certificate policy is dynamic https, otherwise use WorkspaceIgnoreTLSIngress construct
		if resourceInfo.CertType == 3 {
			smartIdeIngress := &k8sLib.WorkspaceIngress{
				APIVersion: "networking.k8s.io/v1",
				Kind:       "Ingress",
				Metadata: struct {
					Name        string "yaml:\"name\""
					Namespace   string "yaml:\"namespace\""
					Annotations struct {
						NginxIngressKubernetesIoAuthType   string "yaml:\"nginx.ingress.kubernetes.io/auth-type\""
						NginxIngressKubernetesIoAuthSecret string "yaml:\"nginx.ingress.kubernetes.io/auth-secret\""
						NginxIngressKubernetesIoUseRegex   string "yaml:\"nginx.ingress.kubernetes.io/use-regex\""
						CertManagerIoClusterIssuer         string "yaml:\"cert-manager.io/cluster-issuer\""
					} "yaml:\"annotations\""
				}{
					Name:      "ingress-" + namespace,
					Namespace: namespace,
					Annotations: struct {
						NginxIngressKubernetesIoAuthType   string "yaml:\"nginx.ingress.kubernetes.io/auth-type\""
						NginxIngressKubernetesIoAuthSecret string "yaml:\"nginx.ingress.kubernetes.io/auth-secret\""
						NginxIngressKubernetesIoUseRegex   string "yaml:\"nginx.ingress.kubernetes.io/use-regex\""
						CertManagerIoClusterIssuer         string "yaml:\"cert-manager.io/cluster-issuer\""
					}{
						NginxIngressKubernetesIoUseRegex: "true",
						CertManagerIoClusterIssuer:       "letsencrypt",
					},
				},
				Spec: struct {
					IngressClassName string "yaml:\"ingressClassName\""
					TLS              []struct {
						Hosts      []string "yaml:\"hosts\""
						SecretName string   "yaml:\"secretName\""
					} "yaml:\"tls\""
					Rules []struct {
						Host string "yaml:\"host\""
						HTTP struct {
							Paths []struct {
								Path     string "yaml:\"path\""
								PathType string "yaml:\"pathType\""
								Backend  struct {
									Service struct {
										Name string "yaml:\"name\""
										Port struct {
											Number int "yaml:\"number\""
										} "yaml:\"port\""
									} "yaml:\"service\""
								} "yaml:\"backend\""
							} "yaml:\"paths\""
						} "yaml:\"http\""
					} "yaml:\"rules\""
				}{
					IngressClassName: "nginx",
					TLS: []struct {
						Hosts      []string "yaml:\"hosts\""
						SecretName string   "yaml:\"secretName\""
					}{},
					Rules: []struct {
						Host string "yaml:\"host\""
						HTTP struct {
							Paths []struct {
								Path     string "yaml:\"path\""
								PathType string "yaml:\"pathType\""
								Backend  struct {
									Service struct {
										Name string "yaml:\"name\""
										Port struct {
											Number int "yaml:\"number\""
										} "yaml:\"port\""
									} "yaml:\"service\""
								} "yaml:\"backend\""
							} "yaml:\"paths\""
						} "yaml:\"http\""
					}{},
				},
			}

			// Create Basic Secret
			if authType == model.KubeAuthenticationTypeEnum_Basic {
				// 运行htpasswd命令
				// e.g. htpasswd -b -c auth <USERNAME> <PASSWORD>
				common.SmartIDELog.Info(i18nInstance.K8s.Info_log_create_basic_secret_start)
				pwd, _ := os.Getwd()
				htpasswdCmd := exec.Command("htpasswd", "-b", "-c", "auth", username, password)
				htpasswdCmd.Stdout = os.Stdout
				htpasswdCmd.Stderr = os.Stderr
				if htpasswdCmdErr := htpasswdCmd.Run(); htpasswdCmdErr != nil {
					common.SmartIDELog.Error(htpasswdCmdErr)
				}
				// 运行kubectl create secret命令
				// e.g. kubectl create secret generic basic-auth --from-file=auth -n <NAMESPACE>
				err = k8sUtil.ExecKubectlCommandRealtime(fmt.Sprintf("create secret generic basic-auth --from-file=auth -n %v", namespace), pwd, false)
				common.CheckError(err)
				smartIdeIngress.Metadata.Annotations.NginxIngressKubernetesIoAuthType = "basic"
				smartIdeIngress.Metadata.Annotations.NginxIngressKubernetesIoAuthSecret = "basic-auth"
				common.SmartIDELog.Info(i18nInstance.K8s.Info_log_create_basic_secret_success)
			}

			// Genetrate AllInOne Ingress Yaml
			for index, portInfo := range workspaceInfo.Extend.Ports {
				if portInfo.HostPortDesc == "tools-ssh" {
					continue
				}
				host := fmt.Sprintf("%v-%v-p%v.%v", namespace, portInfo.ServiceName, portInfo.ClientPort, baseDNSName)
				//Append TLS
				smartIdeIngress.Spec.TLS = append(smartIdeIngress.Spec.TLS, struct {
					Hosts      []string "yaml:\"hosts\""
					SecretName string   "yaml:\"secretName\""
				}{
					Hosts:      []string{host},
					SecretName: fmt.Sprintf("%v-%v-%v", namespace, portInfo.ServiceName, portInfo.ClientPort),
				})
				//Append Rules
				smartIdeIngress.Spec.Rules = append(smartIdeIngress.Spec.Rules, struct {
					Host string "yaml:\"host\""
					HTTP struct {
						Paths []struct {
							Path     string "yaml:\"path\""
							PathType string "yaml:\"pathType\""
							Backend  struct {
								Service struct {
									Name string "yaml:\"name\""
									Port struct {
										Number int "yaml:\"number\""
									} "yaml:\"port\""
								} "yaml:\"service\""
							} "yaml:\"backend\""
						} "yaml:\"paths\""
					} "yaml:\"http\""
				}{
					Host: host,
					HTTP: struct {
						Paths []struct {
							Path     string "yaml:\"path\""
							PathType string "yaml:\"pathType\""
							Backend  struct {
								Service struct {
									Name string "yaml:\"name\""
									Port struct {
										Number int "yaml:\"number\""
									} "yaml:\"port\""
								} "yaml:\"service\""
							} "yaml:\"backend\""
						} "yaml:\"paths\""
					}{
						Paths: []struct {
							Path     string "yaml:\"path\""
							PathType string "yaml:\"pathType\""
							Backend  struct {
								Service struct {
									Name string "yaml:\"name\""
									Port struct {
										Number int "yaml:\"number\""
									} "yaml:\"port\""
								} "yaml:\"service\""
							} "yaml:\"backend\""
						}{
							{
								Path:     "/",
								PathType: "Prefix",
								Backend: struct {
									Service struct {
										Name string "yaml:\"name\""
										Port struct {
											Number int "yaml:\"number\""
										} "yaml:\"port\""
									} "yaml:\"service\""
								}{
									Service: struct {
										Name string "yaml:\"name\""
										Port struct {
											Number int "yaml:\"number\""
										} "yaml:\"port\""
									}{
										Name: portInfo.ServiceName,
										Port: struct {
											Number int "yaml:\"number\""
										}{
											Number: portInfo.ClientPort,
										},
									},
								},
							},
						},
					},
				})

				//Set Public URL
				workspaceInfo.Extend.Ports[index].IngressUrl = host
				workspaceInfo.Extend.Ports[index].IsConnected = host != ""
			}
			yamlData, err = yaml.Marshal(&smartIdeIngress)
			if err != nil {
				common.SmartIDELog.Error(err)
			}
		} else {
			smartIdeIngress := &k8sLib.WorkspaceIgnoreTLSIngress{
				APIVersion: "networking.k8s.io/v1",
				Kind:       "Ingress",
				Metadata: struct {
					Name        string "yaml:\"name\""
					Namespace   string "yaml:\"namespace\""
					Annotations struct {
						NginxIngressKubernetesIoAuthType         string "yaml:\"nginx.ingress.kubernetes.io/auth-type\""
						NginxIngressKubernetesIoAuthSecret       string "yaml:\"nginx.ingress.kubernetes.io/auth-secret\""
						NginxIngressKubernetesIoUseRegex         string "yaml:\"nginx.ingress.kubernetes.io/use-regex\""
						NginxIngressKubernetesIoForceSSLRedirect string "yaml:\"nginx.ingress.kubernetes.io/force-ssl-redirect\""
						NginxIngressKubernetesIoSSLPassThrough   string "yaml:\"nginx.ingress.kubernetes.io/ssl-passthrough\""
					} "yaml:\"annotations\""
				}{
					Name:      "ingress-" + namespace,
					Namespace: namespace,
					Annotations: struct {
						NginxIngressKubernetesIoAuthType         string "yaml:\"nginx.ingress.kubernetes.io/auth-type\""
						NginxIngressKubernetesIoAuthSecret       string "yaml:\"nginx.ingress.kubernetes.io/auth-secret\""
						NginxIngressKubernetesIoUseRegex         string "yaml:\"nginx.ingress.kubernetes.io/use-regex\""
						NginxIngressKubernetesIoForceSSLRedirect string "yaml:\"nginx.ingress.kubernetes.io/force-ssl-redirect\""
						NginxIngressKubernetesIoSSLPassThrough   string "yaml:\"nginx.ingress.kubernetes.io/ssl-passthrough\""
					}{
						NginxIngressKubernetesIoUseRegex: "true",
					},
				},
				Spec: struct {
					IngressClassName string "yaml:\"ingressClassName\""
					Rules            []struct {
						Host string "yaml:\"host\""
						HTTP struct {
							Paths []struct {
								Path     string "yaml:\"path\""
								PathType string "yaml:\"pathType\""
								Backend  struct {
									Service struct {
										Name string "yaml:\"name\""
										Port struct {
											Number int "yaml:\"number\""
										} "yaml:\"port\""
									} "yaml:\"service\""
								} "yaml:\"backend\""
							} "yaml:\"paths\""
						} "yaml:\"http\""
					} "yaml:\"rules\""
				}{
					IngressClassName: "nginx",
					Rules: []struct {
						Host string "yaml:\"host\""
						HTTP struct {
							Paths []struct {
								Path     string "yaml:\"path\""
								PathType string "yaml:\"pathType\""
								Backend  struct {
									Service struct {
										Name string "yaml:\"name\""
										Port struct {
											Number int "yaml:\"number\""
										} "yaml:\"port\""
									} "yaml:\"service\""
								} "yaml:\"backend\""
							} "yaml:\"paths\""
						} "yaml:\"http\""
					}{},
				},
			}

			// Auto Redirect http to https when certificate policy is static https
			if resourceInfo.CertType == 2 {
				smartIdeIngress.Metadata.Annotations.NginxIngressKubernetesIoForceSSLRedirect = "true"
				smartIdeIngress.Metadata.Annotations.NginxIngressKubernetesIoSSLPassThrough = "true"
			}

			// Create Basic Secret
			if authType == model.KubeAuthenticationTypeEnum_Basic {
				// 运行htpasswd命令
				// e.g. htpasswd -b -c auth <USERNAME> <PASSWORD>
				common.SmartIDELog.Info(i18nInstance.K8s.Info_log_create_basic_secret_start)
				pwd, _ := os.Getwd()
				htpasswdCmd := exec.Command("htpasswd", "-b", "-c", "auth", username, password)
				htpasswdCmd.Stdout = os.Stdout
				htpasswdCmd.Stderr = os.Stderr
				if htpasswdCmdErr := htpasswdCmd.Run(); htpasswdCmdErr != nil {
					common.SmartIDELog.Error(htpasswdCmdErr)
				}
				// 运行kubectl create secret命令
				// e.g. kubectl create secret generic basic-auth --from-file=auth -n <NAMESPACE>
				err = k8sUtil.ExecKubectlCommandRealtime(fmt.Sprintf("create secret generic basic-auth --from-file=auth -n %v", namespace), pwd, false)
				common.CheckError(err)
				smartIdeIngress.Metadata.Annotations.NginxIngressKubernetesIoAuthType = "basic"
				smartIdeIngress.Metadata.Annotations.NginxIngressKubernetesIoAuthSecret = "basic-auth"
				common.SmartIDELog.Info(i18nInstance.K8s.Info_log_create_basic_secret_success)
			}

			// Genetrate AllInOne Ingress Yaml
			for index, portInfo := range workspaceInfo.Extend.Ports {
				if portInfo.HostPortDesc == "tools-ssh" {
					continue
				}
				host := fmt.Sprintf("%v-%v-p%v.%v", namespace, portInfo.ServiceName, portInfo.ClientPort, baseDNSName)
				//Append Rules
				smartIdeIngress.Spec.Rules = append(smartIdeIngress.Spec.Rules, struct {
					Host string "yaml:\"host\""
					HTTP struct {
						Paths []struct {
							Path     string "yaml:\"path\""
							PathType string "yaml:\"pathType\""
							Backend  struct {
								Service struct {
									Name string "yaml:\"name\""
									Port struct {
										Number int "yaml:\"number\""
									} "yaml:\"port\""
								} "yaml:\"service\""
							} "yaml:\"backend\""
						} "yaml:\"paths\""
					} "yaml:\"http\""
				}{
					Host: host,
					HTTP: struct {
						Paths []struct {
							Path     string "yaml:\"path\""
							PathType string "yaml:\"pathType\""
							Backend  struct {
								Service struct {
									Name string "yaml:\"name\""
									Port struct {
										Number int "yaml:\"number\""
									} "yaml:\"port\""
								} "yaml:\"service\""
							} "yaml:\"backend\""
						} "yaml:\"paths\""
					}{
						Paths: []struct {
							Path     string "yaml:\"path\""
							PathType string "yaml:\"pathType\""
							Backend  struct {
								Service struct {
									Name string "yaml:\"name\""
									Port struct {
										Number int "yaml:\"number\""
									} "yaml:\"port\""
								} "yaml:\"service\""
							} "yaml:\"backend\""
						}{
							{
								Path:     "/",
								PathType: "Prefix",
								Backend: struct {
									Service struct {
										Name string "yaml:\"name\""
										Port struct {
											Number int "yaml:\"number\""
										} "yaml:\"port\""
									} "yaml:\"service\""
								}{
									Service: struct {
										Name string "yaml:\"name\""
										Port struct {
											Number int "yaml:\"number\""
										} "yaml:\"port\""
									}{
										Name: portInfo.ServiceName,
										Port: struct {
											Number int "yaml:\"number\""
										}{
											Number: portInfo.ClientPort,
										},
									},
								},
							},
						},
					},
				})

				//Set Public URL
				workspaceInfo.Extend.Ports[index].IngressUrl = host
				workspaceInfo.Extend.Ports[index].IsConnected = host != ""
			}
			yamlData, err = yaml.Marshal(&smartIdeIngress)
			if err != nil {
				common.SmartIDELog.Error(err)
			}
		}

		//7. Save AllInOne Ingress to Temp Yaml
		k8sDirPath := config.SmartIdeHome
		common.SmartIDELog.Info(i18nInstance.K8s.Info_log_save_temp_yaml_start)

		tempK8sYamlFileRelativePath := filepath.Join(k8sDirPath, "k8s_ingress_temp.yaml")
		err = os.WriteFile(tempK8sYamlFileRelativePath, []byte(yamlData), 0777)
		if err != nil {
			common.SmartIDELog.Error(err)
		}
		common.SmartIDELog.Info(i18nInstance.K8s.Info_log_save_temp_yaml_success)
		//8. Kubectl Apply
		common.SmartIDELog.Info(i18nInstance.K8s.Info_log_enable_publicurl_start)
		err = k8sUtil.ExecKubectlCommandRealtime(fmt.Sprintf("apply -f %v", tempK8sYamlFileRelativePath), "", false)
		if err != nil {
			common.SmartIDELog.Error(err)
		}
		common.SmartIDELog.Info(i18nInstance.K8s.Info_log_enable_publicurl_success)
		common.SmartIDELog.Info("-----------------------")

		//9. feedback
		err = server.Feedback_Finish(server.FeedbackCommandEnum_Ingress, cmd, true, nil, workspaceInfo, "", "")
		common.CheckError(err)
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
	k8sCmd.Flags().StringP(k8s_flag_workspaceid, "", "", i18nInstance.K8s.Info_help_flag_workspaceid)
	k8sCmd.Flags().StringP(k8s_flag_public_url, "", "", i18nInstance.K8s.Info_help_flag_publicurl)
	k8sCmd.Flags().StringP(k8s_flag_serverhost, "", "", i18nInstance.K8s.Info_help_flag_serverhost)
	k8sCmd.Flags().StringP(k8s_flag_mode, "", "", i18nInstance.K8s.Info_help_flag_mode)
	k8sCmd.Flags().StringP(k8s_flag_servertoken, "", "", i18nInstance.K8s.Info_help_flag_servertoken)
	k8sCmd.AddCommand(k8s.ApplySSHCmd)
	k8sCmd.AddCommand(k8s.K8sInitCmd)
}
