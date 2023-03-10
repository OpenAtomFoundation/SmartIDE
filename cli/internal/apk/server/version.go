/*
SmartIDE - CLI
Copyright (C) 2023 leansoftX.com

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
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
