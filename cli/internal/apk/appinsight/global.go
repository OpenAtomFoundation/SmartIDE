package appinsight

type GlobalInfo struct {
	Version          string
	Mode             string
	Serverhost       string
	Cloud_RoleName   string
	CmdType          string
	ServerUserName string
	ServerUserGuid string
	ServerWorkSpaceId string
	IsInsightEnabled bool
}

var Global = new(GlobalInfo)
