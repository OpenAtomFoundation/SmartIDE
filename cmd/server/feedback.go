package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/leansoftX/smartide-cli/internal/biz/workspace"
	"github.com/leansoftX/smartide-cli/internal/model"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/spf13/cobra"
	"github.com/thedevsaddam/gojsonq"
)

//
type feedbackRequest struct {
	Command string `json:"command"`

	ServerWorkspaceId string `json:"serverWorkspaceId"`
	ServerUserName    string `json:"serverUserName"`
	ServerUserGuid    string `json:"serverUserGuid"`
	IsSuccess         bool   `json:"isSuccess"`
	WebidePort        int    `json:"webidePort"`
	Message           string `json:"message"`

	ConfigFileContent        string `json:"configFileContent"`
	TempDockerComposeContent string `json:"tempDockerComposeContent"`
	LinkDockerCompose        string `json:"linkDockerComposeContent"`
	Extend                   string `json:"extend"`
}

func (feedbackReq feedbackRequest) Check() error {
	if feedbackReq.ServerUserName == "" {
		return errors.New("ServerUserName is nil")
	}

	if feedbackReq.ServerUserGuid == "" {
		return errors.New("ServerUserGuid is nil")
	}

	if feedbackReq.ServerWorkspaceId == "" {
		return errors.New("ServerWorkspaceId is nil")
	}

	return nil
}

func FeeadbackExtend(auth model.Auth, workspaceInfo workspace.WorkspaceInfo) error {
	var _feedbackRequest struct {
		ID     uint
		Extend string
	}
	_feedbackRequest.ID = workspaceInfo.ServerWorkSpace.ID
	_feedbackRequest.Extend = workspaceInfo.Extend.ToJson()

	// 请求体
	jsonBytes, err := json.Marshal(_feedbackRequest)
	if err != nil {
		return err
	}
	url := fmt.Sprint(auth.LoginUrl, "/api/smartide/workspace/update")
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-token", auth.Token.(string))

	//
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Do(req)

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// request
	reqBody, _ := ioutil.ReadAll(req.Body)
	printReqStr := fmt.Sprintf("request head: %v, body: %s", req.Header, reqBody)
	common.SmartIDELog.Debug(printReqStr)

	// response
	respBody, _ := ioutil.ReadAll(resp.Body)
	printRespStr := fmt.Sprintf("response status code: %v, head: %v, body: %s", resp.StatusCode, resp.Header, string(respBody))
	common.SmartIDELog.Debug(printRespStr)

	return nil
}

// 触发 remove
func Trigger_Action(action string, serverWorkspaceNo string, host string, auth model.Auth, datas map[string]interface{}) error {

	if action != "stop" && action != "remove" {
		return errors.New("当前方法仅支持stop 或 remove")
	}

	url, err := common.UrlJoin(host, "/api/smartide/workspace/", action)
	if err != nil {
		return err
	}
	datas["no"] = serverWorkspaceNo

	header := map[string]string{
		"Content-Type": "application/json",
		"x-token":      auth.Token.(string),
	}

	response, err := common.Put(url.String(), datas, header)
	if err != nil {
		return err
	}

	code := gojsonq.New().JSONString(response).Find("code").(float64)
	if code != 0 {
		msg := gojsonq.New().JSONString(response).Find("msg")
		return fmt.Errorf("stop fail: %q", msg)
	}

	return nil
}

// 反馈server工作区的创建情况
func Feedback_Finish(feedbackCommand FeedbackCommandEnum, cmd *cobra.Command,
	isSuccess bool, webidePort int, workspaceInfo workspace.WorkspaceInfo, message string) error {

	fflags := cmd.Flags()

	mode, _ := cmd.Flags().GetString(Flags_Mode)
	if strings.ToLower(mode) != "server" {
		return errors.New("当前仅支持在 mode=server 的模式下运行！")
	}

	// 验证参数是否有值
	Check(cmd)

	// 从参数中获取相应值
	serverWorkspaceid, _ := fflags.GetString(Flags_ServerWorkspaceid)
	serverFeedbackUrl, _ := fflags.GetString(Flags_ServerFeedback)
	serverToken, _ := fflags.GetString(Flags_ServerToken)
	serverUserName, _ := fflags.GetString(Flags_ServerUsername)
	serverUserGuid, _ := fflags.GetString(Flags_ServerUserGUID)
	configFileContent, _ := workspaceInfo.ConfigYaml.ToYaml()
	tempDockerComposeContent, _ := workspaceInfo.TempDockerCompose.ToYaml()
	linkDockerCompose, _ := workspaceInfo.LinkDockerCompose.ToYaml()
	extend := workspaceInfo.Extend.ToJson()
	/* 	_feedbackRequest := feedbackRequest{
	   		Command: string(feedbackCommand),

	   		ServerWorkspaceId:        serverWorkspaceid,
	   		ServerUserName:           serverUserName,
	   		ServerUserGuid:           serverUserGuid,
	   		IsSuccess:                isSuccess,
	   		WebidePort:               webidePort,
	   		Message:                  message,
	   		ConfigFileContent:        configFileContent,
	   		TempDockerComposeContent: tempDockerComposeContent,
	   		LinkDockerCompose:        linkDockerCompose,
	   		Extend:                   extend,
	   	}
	   	err := _feedbackRequest.Check()

	   	if err != nil {
	   		return err
	   	}
	*/

	if serverUserName == "" {
		return errors.New("ServerUserName is nil")
	}
	if serverUserGuid == "" {
		return errors.New("ServerUserGuid is nil")
	}
	if serverWorkspaceid == "" {
		return errors.New("ServerWorkspaceId is nil")
	}
	if serverFeedbackUrl == "" {
		return errors.New("serverFeedbackUrl is nil")
	}

	datas := map[string]interface{}{
		"serverWorkspaceid": serverWorkspaceid,
		"serverUserName":    serverUserName,
		"serverUserGuid":    serverUserGuid,
		"isSuccess":         isSuccess,
		"webidePort":        webidePort,
		"message":           message,
	}
	if feedbackCommand == FeedbackCommandEnum_Start { // 只有start的时候，才需要传递文件内容
		datas["configFileContent"] = configFileContent
		datas["tempDockerComposeContent"] = tempDockerComposeContent
		datas["linkDockerCompose"] = linkDockerCompose
		datas["extend"] = extend
	}
	headers := map[string]string{"Content-Type": "application/json", "x-token": serverToken}
	response, err := common.PostJson(serverFeedbackUrl, datas, headers)

	if err != nil {
		return err
	}
	common.SmartIDELog.Info(response)

	/* // 请求体
	jsonBytes, err := json.Marshal(_feedbackRequest)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", serverFeedbackUrl, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-token", serverToken)

	//
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Do(req)

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// request
	reqBody, _ := ioutil.ReadAll(req.Body)
	printReqStr := fmt.Sprintf("request head: %v, body: %s", req.Header, reqBody)
	common.SmartIDELog.Debug(printReqStr)

	// response
	respBody, _ := ioutil.ReadAll(resp.Body)
	printRespStr := fmt.Sprintf("response status code: %v, head: %v, body: %s", resp.StatusCode, resp.Header, string(respBody))
	common.SmartIDELog.Debug(printRespStr)
	common.SmartIDELog.Info(string(respBody)) */

	return nil
}
