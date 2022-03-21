/*
 * @Author: kenan
 * @Date: 2022-02-10 18:11:42
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-03-17 09:17:38
 * @FilePath: /smartide-cli/pkg/common/http.go
 * @Description:
 *
 * Copyright (c) 2022 by kenanlu@leansoftx.com, All Rights Reserved.
 */

package common

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type UploadFile struct {
	// 表单名称
	Name string
	// 文件全路径
	Filepath string
}

/* // 请求客户端
var httpClient = &http.Client{
	Timeout: 10 * time.Second,
} */

func Get(reqUrl string, reqParams map[string]string, headers map[string]string) (string, error) {
	urlParams := url.Values{}
	Url, _ := url.Parse(reqUrl)
	for key, val := range reqParams {
		urlParams.Set(key, val)
	}

	// client
	httpClient := getClient(reqUrl)

	//如果参数中有中文参数,这个方法会进行URLEncode
	Url.RawQuery = urlParams.Encode()
	// 得到完整的url，http://xx?query
	urlPath := Url.String()

	httpRequest, _ := http.NewRequest("GET", urlPath, nil)
	// 添加请求头
	for k, v := range headers {
		httpRequest.Header.Add(k, v)
	}
	// 发送请求
	resp, err := httpClient.Do(httpRequest)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", errors.New(resp.Status)
	}

	response, _ := ioutil.ReadAll(resp.Body)
	return string(response), nil
}

func PostForm(reqUrl string, reqParams map[string]interface{}, headers map[string]string) (string, error) {
	return post(reqUrl, reqParams, "application/x-www-form-urlencoded", nil, headers)
}

func PostJson(reqUrl string, reqParams map[string]interface{}, headers map[string]string) (string, error) {
	return post(reqUrl, reqParams, "application/json", nil, headers)
}

func PostFile(reqUrl string, reqParams map[string]interface{}, files []UploadFile, headers map[string]string) (string, error) {
	return post(reqUrl, reqParams, "multipart/form-data", files, headers)
}

func Put(reqUrl string, reqParams map[string]interface{}, headers map[string]string) (string, error) {
	requestBody, realContentType := getReader(reqParams, "application/json", nil)
	httpRequest, _ := http.NewRequest("PUT", reqUrl, requestBody)

	// client
	httpClient := getClient(reqUrl)

	// 添加请求头
	httpRequest.Header.Add("Content-Type", realContentType)
	for k, v := range headers {
		httpRequest.Header.Add(k, v)
	}
	// 发送请求
	resp, err := httpClient.Do(httpRequest)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", errors.New(resp.Status)
	}
	response, err := ioutil.ReadAll(resp.Body)
	SmartIDELog.Debug(string(response))
	return string(response), err
}

var timeout time.Duration = 10 * time.Second

//
func getClient(url string) *http.Client {
	_httpClient := &http.Client{
		Timeout: timeout,
	}
	if strings.HasPrefix(url, "https") {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		_httpClient.Transport = tr
	}
	return _httpClient
}

func post(reqUrl string, reqParams map[string]interface{}, contentType string, files []UploadFile, headers map[string]string) (string, error) {
	requestBody, realContentType := getReader(reqParams, contentType, files)
	httpRequest, _ := http.NewRequest("POST", reqUrl, requestBody)

	// client
	httpClient := getClient(reqUrl)

	// 添加请求头
	httpRequest.Header.Add("Content-Type", realContentType)
	for k, v := range headers {
		httpRequest.Header.Add(k, v)
	}
	// 发送请求
	resp, err := httpClient.Do(httpRequest)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", errors.New(resp.Status)
	}
	response, err := ioutil.ReadAll(resp.Body)
	SmartIDELog.Debug(string(response))
	return string(response), err
}

func getReader(reqParams map[string]interface{}, contentType string, files []UploadFile) (io.Reader, string) {
	if strings.Contains(contentType, "json") {
		bytesData, _ := json.Marshal(reqParams)
		return bytes.NewReader(bytesData), contentType
	} else if files != nil {
		body := &bytes.Buffer{}
		// 文件写入 body
		writer := multipart.NewWriter(body)
		for _, uploadFile := range files {
			file, err := os.Open(uploadFile.Filepath)
			if err != nil {
				panic(err)
			}
			part, err := writer.CreateFormFile(uploadFile.Name, filepath.Base(uploadFile.Filepath))
			if err != nil {
				panic(err)
			}
			_, err = io.Copy(part, file)
			if err != nil {
				SmartIDELog.Importance(err.Error())
			}
			file.Close()
		}
		// 其他参数列表写入 body
		for k, v := range reqParams {
			if err := writer.WriteField(k, v.(string)); err != nil {
				panic(err)
			}
		}
		if err := writer.Close(); err != nil {
			panic(err)
		}
		// 上传文件需要自己专用的contentType
		return body, writer.FormDataContentType()
	} else {
		urlValues := url.Values{}
		for key, val := range reqParams {
			urlValues.Set(key, val.(string))
		}
		reqBody := urlValues.Encode()
		return strings.NewReader(reqBody), contentType
	}
}
