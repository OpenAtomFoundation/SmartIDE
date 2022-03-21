/*
 * @Author: Jason Chen
 * @Date: 2022-03-15 15:00:10
 * @LastEditTime: 2022-03-15 15:04:47
 * @LastEditors: Jason Chen
 * @Description:
 * @FilePath: /smartide-cli/pkg/common/url.go
 */
package common

import (
	"fmt"
	"net/url"
	"path"
)

//
func UrlJoin(basePath string, paths ...string) (*url.URL, error) {

	u, err := url.Parse(basePath)

	if err != nil {
		return nil, fmt.Errorf("invalid url")
	}

	p2 := append([]string{u.Path}, paths...)

	result := path.Join(p2...)

	u.Path = result

	return u, nil
}
