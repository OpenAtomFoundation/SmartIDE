package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/leansoftX/smartide-cli/internal/biz/workspace"
	"github.com/leansoftX/smartide-cli/internal/model"
	"github.com/leansoftX/smartide-cli/pkg/common"
	"github.com/spf13/cobra"
	"github.com/thedevsaddam/gojsonq"
)

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
func Trigger_Action(action string, serverWorkspaceNo string, auth model.Auth, datas map[string]interface{}) error {

	if action != "stop" && action != "remove" {
		return errors.New("当前方法仅支持stop 或 remove")
	}
	datas["no"] = serverWorkspaceNo

	headers := map[string]string{
		"Content-Type": "application/json",
		"x-token":      auth.Token.(string),
	}

	url, err := common.UrlJoin(auth.LoginUrl, "/api/smartide/workspace/", action)
	if err != nil {
		return err
	}

	httpClient := common.CreateHttpClient(6, 30*time.Second)
	response, err := httpClient.Put(url.String(), datas, headers) // post 请求
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
	isSuccess bool, webidePort *int, workspaceInfo workspace.WorkspaceInfo, message string, containerId string) error {
	if workspaceInfo.CliRunningEnv != workspace.CliRunningEvnEnum_Server {
		return errors.New("当前仅支持在 mode=server 的模式下运行！")
	}

	// 验证参数是否有值
	//Check(cmd)

	//1. 从参数中获取相应值
	serverModeInfo, err := GetServerModeInfo(cmd)
	//1.1. validate
	if err != nil {
		return err
	}
	/* 	if serverModeInfo.ServerUsername == "" {
		return errors.New("ServerUserName is nil")
	} */
	if serverModeInfo.ServerHost == "" {
		return errors.New("ServerHost is nil")
	}
	//1.2.
	serverFeedbackUrl, _ := common.UrlJoin(serverModeInfo.ServerHost, "/api/smartide/workspace/finish")
	configFileContent, _ := workspaceInfo.ConfigYaml.ToYaml()
	tempYamlContent, _ := workspaceInfo.TempDockerCompose.ToYaml()
	linkYmalContent, _ := workspaceInfo.LinkDockerCompose.ToYaml()
	if workspaceInfo.Mode == workspace.WorkingMode_K8s {
		tempYamlContent, _ = workspaceInfo.K8sInfo.TempK8sConfig.ConvertToK8sYaml()
		linkYmalContent, _ = workspaceInfo.K8sInfo.OriginK8sYaml.ConvertToK8sYaml()
	}
	//1.3. workspace id
	worksspaceId := strconv.Itoa(int(workspaceInfo.ServerWorkSpace.ID))
	if worksspaceId == "" {
		worksspaceId = serverModeInfo.ServerWorkspaceid
	}
	if worksspaceId == "" {
		return errors.New("workspace id is nil")
	} else if strings.Contains(strings.ToLower(worksspaceId), "WS") {
		flysnowRegexp := regexp.MustCompile(`[1-9]{1}[0-9].*`)
		params := flysnowRegexp.FindStringSubmatch(worksspaceId)
		worksspaceId = params[0]
	}
	//1.4. extend
	extend := workspaceInfo.Extend.ToJson()

	//2.
	datas := map[string]interface{}{
		"command":           string(feedbackCommand),
		"serverWorkspaceid": worksspaceId,
		"serverUserName":    serverModeInfo.ServerUsername,
		"isSuccess":         isSuccess,
		"message":           message,
		"containerId":       containerId,
	}
	if feedbackCommand == FeedbackCommandEnum_Start { // 只有start的时候，才需要传递文件内容
		datas["configFileContent"] = configFileContent
		datas["tempDockerComposeContent"] = tempYamlContent
		datas["linkDockerCompose"] = linkYmalContent
		datas["extend"] = extend
	} else if feedbackCommand == FeedbackCommandEnum_Ingress || feedbackCommand == FeedbackCommandEnum_ApplySSH {
		datas["extend"] = extend
	}
	headers := map[string]string{"Content-Type": "application/json", "x-token": serverModeInfo.ServerToken}

	httpClient := common.CreateHttpClient(6, 30*time.Second)
	_, err = httpClient.PostJson(serverFeedbackUrl.String(), datas, headers) // post 请求

	if err != nil {
		return err
	}
	//common.SmartIDELog.Info(response)

	return nil
}

// 反馈server工作区的创建情况
func Send_WorkspaceInfo(callbackAPI string, feedbackCommand FeedbackCommandEnum, cmd *cobra.Command,
	isSuccess bool, webidePort *int, workspaceInfo workspace.WorkspaceInfo) error {

	serverFeedbackUrl := callbackAPI
	configFileContent, _ := workspaceInfo.ConfigYaml.ToYaml()
	tempYamlContent, _ := workspaceInfo.TempDockerCompose.ToYaml()
	linkYmalContent, _ := workspaceInfo.LinkDockerCompose.ToYaml()
	if workspaceInfo.Mode == workspace.WorkingMode_K8s {
		tempYamlContent, _ = workspaceInfo.K8sInfo.TempK8sConfig.ConvertToK8sYaml()
		linkYmalContent, _ = workspaceInfo.K8sInfo.OriginK8sYaml.ConvertToK8sYaml()
	}

	extend := workspaceInfo.Extend.ToJson()

	//2.
	datas := map[string]interface{}{
		"command":   string(feedbackCommand),
		"isSuccess": isSuccess,
	}
	if feedbackCommand == FeedbackCommandEnum_Start { // 只有start的时候，才需要传递文件内容
		datas["configFileContent"] = configFileContent
		datas["tempDockerComposeContent"] = tempYamlContent
		datas["linkDockerCompose"] = linkYmalContent
		datas["extend"] = extend
	} else if feedbackCommand == FeedbackCommandEnum_Ingress || feedbackCommand == FeedbackCommandEnum_ApplySSH {
		datas["extend"] = extend
	}
	headers := map[string]string{"Content-Type": "application/json"}

	httpClient := common.CreateHttpClient(6, 30*time.Second)
	_, err := httpClient.PostJson(serverFeedbackUrl, datas, headers) // post 请求

	if err != nil {
		return err
	}

	common.SmartIDELog.InfoF(i18nInstance.VmStart.Info_callback_msg, callbackAPI)

	return nil
}
