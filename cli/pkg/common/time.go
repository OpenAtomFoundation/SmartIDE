/*
 * @Date: 2022-06-02 23:40:22
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-06-02 23:42:56
 * @FilePath: /smartide-cli/pkg/common/time.go
 */

package common

import "time"

func LocalTimeStr(dt time.Time) string {
	local, _ := time.LoadLocation("Local")            // 北京时区
	str := dt.In(local).Format("2006-01-02 15:04:05") // 格式化输出
	return str
}
