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
