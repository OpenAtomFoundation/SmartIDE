/*
 * @Author: jason chen (jasonchen@leansoftx.com, http://smallidea.cnblogs.com)
 * @Description:
 * @Date: 2021-11
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-07-20 11:12:48
 */
package common

import "encoding/json"

//
func ConvertToJson(obj interface{}) string {
	b, err := json.Marshal(obj)
	if err != nil {
		CheckError(err)
	}
	return string(b)
}

func IsJSON(s string) bool {
	var js interface{}
	return json.Unmarshal([]byte(s), &js) == nil
}
