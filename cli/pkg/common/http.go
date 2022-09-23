/*
 * @Author: kenan
 * @Date: 2022-02-10 18:11:42
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-09-23 13:38:33
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
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type ResponseBodyTypeEnum string

const (
	ResponseBodyTypeEnum_JSON ResponseBodyTypeEnum = "json"
	ResponseBodyTypeEnum_HTML ResponseBodyTypeEnum = "html"
)

type HttpClient struct {
	RetryMax         uint
	Sleep            time.Duration
	TimeOut          time.Duration
	ResponseBodyType ResponseBodyTypeEnum
}

func CreateHttpClient(retryMax uint,
	timeOut time.Duration, sleep time.Duration,
	responseBodyType ResponseBodyTypeEnum) HttpClient {
	return HttpClient{
		RetryMax:         retryMax,
		TimeOut:          timeOut,
		Sleep:            sleep,
		ResponseBodyType: responseBodyType,
	}
}

func CreateHttpClientDisableRetry() HttpClient {
	return CreateHttpClient(0, defaultTimeout, 0, "application/json")
}

func CreateHttpClientEnableRetry() HttpClient {
	return CreateHttpClient(3, defaultTimeout, time.Second*3, "application/json")
}

func (target HttpClient) Get(reqUrl string,
	reqParams map[string]string, headers map[string]string) (string, error) {
	result := ""
	var err error = nil

	Retry(target.RetryMax, target.Sleep, func() error {
		result, err = target.get(reqUrl, reqParams, headers)
		if err == nil {
			if result == "" {
				err = errors.New("response body is empty")
			} else if !IsJSON(result) {
				err = errors.New("response body is not json")
			}
		}
		return err
	})

	return result, err
}

func (target HttpClient) Put(reqUrl string,
	reqParams map[string]interface{}, headers map[string]string) (string, error) {

	result := ""
	var err error = nil

	Retry(target.RetryMax, target.Sleep, func() error {
		result, err = target.put(reqUrl, reqParams, headers)
		if err == nil {
			if result == "" {
				err = errors.New("response body is empty")
			} else if !IsJSON(result) {
				err = errors.New("response body is not json")
			}
		}
		return err
	})

	return result, err
}

func (target HttpClient) PostJson(reqUrl string,
	reqParams map[string]interface{}, headers map[string]string) (string, error) {
	return postRetry(target, reqUrl, reqParams, "application/json", nil, headers)
}

func (target HttpClient) PostForm(reqUrl string,
	reqParams map[string]interface{}, headers map[string]string) (string, error) {
	return postRetry(target, reqUrl, reqParams, "application/x-www-form-urlencoded", nil, headers)
}

func (target HttpClient) PostFile(reqUrl string,
	reqParams map[string]interface{}, files []UploadFile, headers map[string]string) (string, error) {
	return postRetry(target, reqUrl, reqParams, "multipart/form-data", files, headers)
}

func postRetry(target HttpClient, reqUrl string,
	reqParams map[string]interface{}, contentType string, files []UploadFile, headers map[string]string) (string, error) {
	result := ""
	var err error = nil

	if target.Sleep > 0 {
		defaultTimeout = target.Sleep
	}

	Retry(target.RetryMax, target.Sleep, func() error {
		result, err = target.post(reqUrl, reqParams, contentType, files, headers)
		if err == nil {
			if result == "" {
				err = errors.New("response body is empty")
			} else if !IsJSON(result) && contentType == "application/json" {
				err = errors.New("response body is not json")
			}
		}
		return err
	})

	return result, err
}

type UploadFile struct {
	// 表单名称
	Name string
	// 文件全路径
	Filepath string
}

var defaultTimeout time.Duration = 10 * time.Second

/* func SetTimeOut(timeDuration time.Duration) {
	if timeDuration != 0 {
		defaultTimeout = timeDuration
	}
} */

func (target HttpClient) get(reqUrl string, reqParams map[string]string, headers map[string]string) (string, error) {
	urlParams := url.Values{}
	Url, _ := url.Parse(reqUrl)
	for key, val := range reqParams {
		urlParams.Set(key, val)
	}

	// client
	httpClient := getClient(reqUrl, target.TimeOut)

	//如果参数中有中文参数,这个方法会进行URLEncode
	Url.RawQuery = urlParams.Encode()
	// 得到完整的url，http://xx?query
	urlPath := Url.String()

	httpRequest, err := http.NewRequest("GET", urlPath, nil)
	if err != nil {
		return "", err
	}
	// 添加请求头
	for k, v := range headers {
		httpRequest.Header.Add(k, v)
	}

	// 发送请求
	SmartIDELog.Debug(formatRequest(httpRequest, nil))
	httpResponse, err := httpClient.Do(httpRequest)
	if err != nil {
		return "", err
	}
	responseBody, err := parsingResponse(httpResponse)
	if err != nil {
		return "", err
	}
	defer httpResponse.Body.Close()
	if httpResponse.StatusCode != 200 {
		return "", errors.New(httpResponse.Status)
	}

	if len(string(responseBody)) == 0 {
		return "", errors.New("reponse body is empty")
	}
	return string(responseBody), nil
}

func (target HttpClient) put(reqUrl string, reqParams map[string]interface{}, headers map[string]string) (string, error) {
	jsonBytes, err := json.Marshal(reqParams)
	if err != nil {
		return "", err
	}
	httpRequest, err := http.NewRequest("PUT", reqUrl, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return "", err
	}

	// client
	httpClient := getClient(reqUrl, target.TimeOut)

	// 添加请求头
	headers["Content-Type"] = "application/json" // 后续可根据需要增加其他的类型
	for k, v := range headers {
		httpRequest.Header.Add(k, v)
	}

	// 发送请求
	SmartIDELog.Debug(formatRequest(httpRequest, reqParams))
	httpResponse, err := httpClient.Do(httpRequest)
	if err != nil {
		return "", err
	}
	responseBody, err := parsingResponse(httpResponse)
	if err != nil {
		return "", err
	}
	defer httpResponse.Body.Close()
	if httpResponse.StatusCode != 200 {
		return "", errors.New(httpResponse.Status)
	}
	if len(string(responseBody)) == 0 {
		return "", errors.New("reponse body is empty")
	}
	return string(responseBody), err

}

func (target HttpClient) post(reqUrl string,
	reqParams map[string]interface{}, contentType string, files []UploadFile, headers map[string]string) (
	string, error) {
	requestBody, realContentType := getReader(reqParams, contentType, files)
	httpRequest, err := http.NewRequest("POST", reqUrl, requestBody)
	if err != nil {
		return "", err
	}

	// client
	httpClient := getClient(reqUrl, target.TimeOut)

	// 添加请求头
	httpRequest.Header.Add("Content-Type", realContentType)
	for k, v := range headers {
		httpRequest.Header.Add(k, v)
	}

	// 发送请求
	SmartIDELog.Debug(formatRequest(httpRequest, reqParams))
	httpResponse, err := httpClient.Do(httpRequest)
	if err != nil {
		return "", err
	}
	responseBody, err := parsingResponse(httpResponse)
	if err != nil {
		return "", err
	}
	defer httpResponse.Body.Close()
	if httpResponse.StatusCode != 200 {
		return "", errors.New(httpResponse.Status)
	}
	//responseBody, err := io.ReadAll(httpResponse.Body)
	if len(string(responseBody)) == 0 {
		return "", errors.New("reponse body is empty")
	}
	return string(responseBody), err

}

func getClient(url string, timeout time.Duration) *http.Client {
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
	var messages []string
	// Add the request string
	url := fmt.Sprintf("%v %v %v", r.Method, r.URL, r.Proto)
	messages = append(messages, url)
	// Add the host
	messages = append(messages, fmt.Sprintf("Host: %v", r.Host))
	// Loop through headers
	for name, headers := range r.Header {
		name = strings.ToLower(name)
		for _, h := range headers {
			messages = append(messages, fmt.Sprintf("%v: %v", name, h))
		}
	}

	// If this is a POST, add post data
	if r.Method == "POST" {
		r.ParseForm()
		messages = append(messages, "\n")
		messages = append(messages, r.Form.Encode())
	}

	if reqParams != nil {
		j, _ := json.Marshal(reqParams)
		messages = append(messages, string(j))
	}

	// Return the request as a string
	return "REQUEST \n\t" + strings.Join(messages, "\n\t")
}

// response format
func parsingResponse(resp *http.Response) (string, error) {
	if resp == nil {
		return "", errors.New("response is nil")
	}

	responseBody := []byte{}
	if resp.Body != nil {
		responseBody, _ = io.ReadAll(resp.Body)
		if len(responseBody) == 0 {
			return "", nil
		}
	}

	result := string(responseBody)

	// 压缩json数据，减少打印的信息
	simpleResult := result
	regex := regexp.MustCompile(`:".*?[^\\]"`)
	simpleResult = regex.ReplaceAllStringFunc(simpleResult, func(old string) string {
		if len(old) > 26 {
			return old[:23] + "...\""
		}
		return old
	})

	printRespStr := fmt.Sprintf("RESPONSE \n\tcode: %v \n\thead: %v \n\tbody: %s",
		resp.StatusCode, resp.Header, simpleResult)
	SmartIDELog.Debug(printRespStr)

	return result, nil
}
