/*
 * @Date: 2022-09-03 16:14:15
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-09-22 09:37:14
 * @FilePath: /cli/internal/apk/server/version.go
 */
package server

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/leansoftX/smartide-cli/internal/model"
	"github.com/leansoftX/smartide-cli/pkg/common"
)

// 获取工作区详情
func GetVersionByApi(auth model.Auth) (version string, err error) {
	if (auth == model.Auth{}) {
		err = errors.New("用户未登录！")
		return
	}

	url := fmt.Sprint(auth.LoginUrl, "/api/system/getVersion")
	queryparams := map[string]string{}

	httpClient := common.CreateHttpClientEnableRetry()
	response, err := httpClient.Get(url,
		queryparams, map[string]string{
			"Content-Type": "application/json",
			"x-token":      auth.Token.(string),
		})

	if err != nil {
		return
	}
	if response == "" {
		return "", errors.New("服务器访问空数据！")
	}

	var objResponse struct {
		Code int `json:"code"`
		Data struct {
			Version string `json:"version"`
		} `json:"data"`
	}
	err = json.Unmarshal([]byte(response), &objResponse)
	if err != nil {
		return "", err
	}

	if objResponse.Code == 0 {
		version = objResponse.Data.Version
	}
	return version, err
}
