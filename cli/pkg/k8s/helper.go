/*
SmartIDE - CLI
Copyright (C) 2023 leansoftX.com

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package k8s

import (
	"reflect"
	"regexp"

	"github.com/jinzhu/copier"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// 添加label
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
		originLabels := origin.FieldByName("Spec").FieldByName("Template").FieldByName("Labels").
			Interface().(map[string]string)
		currentLabels := make(map[string]string)
		if originLabels != nil {
			copier.Copy(&currentLabels, originLabels)
		}
		for key, value := range labels {
			relValue := filterSpecialCharacters4LabelValue(value)
			currentLabels[key] = relValue
		}
		result.FieldByName("Spec").FieldByName("Template").FieldByName("Labels").
			Set(reflect.ValueOf(currentLabels))
	}

	// 原类型中的labels赋值
	objMeta := origin.FieldByName("ObjectMeta").Interface().(v1.ObjectMeta)
	if objMeta.Labels != nil {
		currentLabels := make(map[string]string)
		copier.Copy(&currentLabels, objMeta.Labels)
		objMeta.Labels = currentLabels
	}
	for key, value := range labels {
		relValue := filterSpecialCharacters4LabelValue(value)
		v1.SetMetaDataLabel(&objMeta, key, relValue)
	}
	result.FieldByName("ObjectMeta").Set(reflect.ValueOf(objMeta))

	return result.Interface()
}

func filterSpecialCharacters4LabelValue(content string) string {
	if len(content) >= 63 {
		content = content[:62]
	}

	reg, _ := regexp.Compile(`[^-A-Za-z0-9_.]`)
	relValue := reg.ReplaceAllString(content, "-")

	regFirstAnEnd, _ := regexp.Compile(`^[^A-Za-z0-9]|[^A-Za-z0-9]$`)
	relValue = regFirstAnEnd.ReplaceAllString(relValue, "0")

	return relValue
}
