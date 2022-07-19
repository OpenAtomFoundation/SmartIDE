/*
 * @Author: kenan
 * @Date: 2022-02-10 18:11:42
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-07-19 11:18:24
 * @FilePath: /cli/pkg/common/http.go
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
	"fmt"
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

type HttpClient struct {
	RetryMax uint
	TimeOut  time.Duration
}

func CreateHttpClient(retryMax uint, timeOut time.Duration) HttpClient {
	return HttpClient{
		RetryMax: retryMax,
		TimeOut:  timeOut,
	}
}

func (target HttpClient) PostJson(reqUrl string, reqParams map[string]interface{}, headers map[string]string) (string, error) {

	result := ""
	var err error = nil

	var retryCount uint = 0
	for {
		result, err = post(reqUrl, reqParams, "application/json", nil, headers)
		if err != nil && target.RetryMax > retryCount {
			retryCount++
		}
	}

	return result, err
}

type UploadFile struct {
	// 表单名称
	Name string
	// 文件全路径
	Filepath string
}

var timeout time.Duration = 10 * time.Second

func SetTimeOut(timeDuration time.Duration) {
	if timeDuration != 0 {
		timeout = timeDuration
	}
}

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

	// debug
	SmartIDELog.Debug(formatRequest(httpRequest, nil))

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
	SmartIDELog.Debug("response: " + string(response))
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

	// debug
	SmartIDELog.Debug(formatRequest(httpRequest, reqParams))

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
	SmartIDELog.Debug("response: " + string(response))
	return string(response), err
}

//
func getClient(url string) *http.Client {
	_httpClient := &http.Client{
		Timeout: timeout, // 请求超时时间
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

	// debug
	SmartIDELog.Debug(formatRequest(httpRequest, reqParams))

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
	SmartIDELog.Debug("response: " + string(response))
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
				SmartIDELog.ImportanceWithError(err)
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

// formatRequest generates ascii representation of a request
func formatRequest(r *http.Request, reqParams map[string]interface{}) string {
	// Create return string
	var request []string
	// Add the request string
	url := fmt.Sprintf("%v %v %v", r.Method, r.URL, r.Proto)
	request = append(request, url)
	// Add the host
	request = append(request, fmt.Sprintf("Host: %v", r.Host))
	// Loop through headers
	for name, headers := range r.Header {
		name = strings.ToLower(name)
		for _, h := range headers {
			request = append(request, fmt.Sprintf("%v: %v", name, h))
		}
	}

	// If this is a POST, add post data
	if r.Method == "POST" {
		r.ParseForm()
		request = append(request, "\n")
		request = append(request, r.Form.Encode())
	}

	if reqParams != nil {
		j, _ := json.Marshal(reqParams)
		request = append(request, string(j))
	}

	// Return the request as a string
	return "request: " + strings.Join(request, "\n")
}
