/*
 * @Date: 2022-09-05 11:48:48
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-09-06 09:43:26
 * @FilePath: /cli/pkg/k8s/helper.go
 */

package k8s

import (
	"reflect"
	"regexp"

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
			relValue := filterSpecialCharacters4LabelValue(value)
			tmpLabels[key] = relValue
		}
		result.FieldByName("Spec").FieldByName("Template").FieldByName("Labels").
			Set(reflect.ValueOf(tmpLabels))
	}

	// 原类型中的labels赋值
	objMeta := origin.FieldByName("ObjectMeta").Interface().(v1.ObjectMeta)
	for key, value := range labels {
		relValue := filterSpecialCharacters4LabelValue(value)
		v1.SetMetaDataLabel(&objMeta, key, relValue)
	}
	result.FieldByName("ObjectMeta").Set(reflect.ValueOf(objMeta))

	return result.Interface()
}

func filterSpecialCharacters4LabelValue(content string) string {
	reg, _ := regexp.Compile(`[^-A-Za-z0-9_.]`)
	relValue := reg.ReplaceAllString(content, "-")

	regFirstAnEnd, _ := regexp.Compile(`^[^A-Za-z0-9]|[^A-Za-z0-9]$`)
	relValue = regFirstAnEnd.ReplaceAllString(relValue, "0")

	return relValue
}
