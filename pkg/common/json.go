/*
 * @Author: jason chen (jasonchen@leansoftx.com, http://smallidea.cnblogs.com)
 * @Description:
 * @Date: 2021-11
 * @LastEditors:
 * @LastEditTime:
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
