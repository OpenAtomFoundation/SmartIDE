/*
 * @Date: 2022-09-05 11:48:48
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-09-05 11:50:56
 * @FilePath: /cli/pkg/k8s/helper.go
 */

package k8s

import (
	"fmt"
	"reflect"
	"strings"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func AddLabels(kind interface{}, labels map[string]string) interface{} {
	origin := reflect.ValueOf(kind)
	if origin.Kind() == reflect.Ptr {
		origin = origin.Elem()
	}

	// 返回的对象
	result := reflect.New(origin.Type()).Elem() // 实例
	result.Set(origin)                          // 赋值

	// deployment的时候需要给templte 中的labels 进行赋值
	if origin.FieldByName("Kind").String() == "Deployment" {
		tmpLabels := origin.FieldByName("Spec").FieldByName("Template").FieldByName("Labels").
			Interface().(map[string]string)
		if tmpLabels == nil {
			tmpLabels = make(map[string]string)
		}
		for key, value := range labels {
			tmpLabels[key] = value
		}
		result.FieldByName("Spec").FieldByName("Template").FieldByName("Labels").
			Set(reflect.ValueOf(tmpLabels))
	}

	// 原类型中的labels赋值
	objMeta := origin.FieldByName("ObjectMeta").Interface().(v1.ObjectMeta)
	for key, value := range labels {
		relValue := fmt.Sprintf("\\\"%v\\\"", value)
		relValue = strings.ReplaceAll(relValue, "/", "_")
		relValue = strings.ReplaceAll(relValue, ":", "_") //TODO 需要使用正则表达式进行replace
		v1.SetMetaDataLabel(&objMeta, key, relValue)
	}
	result.FieldByName("ObjectMeta").Set(reflect.ValueOf(objMeta))

	return result.Interface()
}
