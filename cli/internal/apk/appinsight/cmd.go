package appinsight

type AppinsightCmd string

const (
	Cli_Reset  string = "cli-reset"
	Cli_Update string = "cli-update"

	Cli_Add_Host    string = "cli-add-host"
	Cli_Remove_Host string = "cli-remove-host"

	Cli_Local_Start  string = "cli-local-start"
	Cli_Local_Remove string = "cli-local-remove"
	Cli_Local_Stop   string = "cli-local-stop"
	Cli_Local_New    string = "cli-local-new"
	Cli_Local_Init   string = "cli-local-init"

	Cli_Host_Start  string = "cli-host-start"
	Cli_Host_Remove string = "cli-host-remove"
	Cli_Host_Stop   string = "cli-host-stop"
	Cli_Host_New    string = "cli-host-new"
	Cli_Host_Init   string = "cli-host-init"

	Cli_K8s_Start         string = "cli-k8s-start"
	Cli_K8s_Remove        string = "cli-k8s-remove"
	Cli_K8s_New           string = "cli-k8s-new"
	Cli_K8s_Ingress_Apply string = "cli-k8s-ingress-apply"
	Cli_K8s_Ssh_Apply     string = "cli-k8s-ssh-apply"

	Cli_Server_Connect string = "cli-server-connect"
	Cli_Server_Login   string = "cli-server-login"
	Cli_Server_Logout  string = "cli-server-logout"
)
