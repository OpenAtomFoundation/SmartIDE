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

package i18n

import (

	// 这里不要忘记引入驱动,如引入默认的json驱动

	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/jeandeaual/go-locale"
	_ "github.com/leansoftX/i18n/parser_json"
)

// 测试双语配置是否正常
func TestI18n(t *testing.T) {

	// locale
	currentLang, _ := locale.GetLocale()
	if strings.Index(currentLang, "zh-") == 0 { // 如果不是简体中文，就是英文
		currentLang = "zh_cn"
	} else {
		currentLang = "en_us"
	}

	err := testI18n("zh_cn")
	if err != nil {
		t.Error(err)
	}

	err = testI18n("en_us")
	if err != nil {
		t.Error(err)
	}

}

func testI18n(currentLang string) error {
	// loading and parse json
	data, _ := f.ReadFile("language/" + currentLang + "/info.json")
	json.Unmarshal(data, &instance)

	return sanitize(instance, currentLang)
}

func sanitize(s interface{}, parent string) error {
	if s == nil {
		return nil
	}

	val := reflect.ValueOf(s)

	// If it's an interface or a pointer, unwrap it.
	if val.Kind() == reflect.Ptr && val.Elem().Kind() == reflect.Struct {
		val = val.Elem()
	} else {
		return errors.New("s must be a struct")
	}

	valNumFields := val.NumField()

	for i := 0; i < valNumFields; i++ {
		field := val.Field(i)
		fieldKind := field.Kind()

		fieldName := val.Type().Field(i).Name
		if parent != "" {
			fieldName = fmt.Sprintf("%v.%v", parent, fieldName)
		}

		// Check if it's a pointer to a struct.
		if fieldKind == reflect.Ptr && field.Elem().Kind() == reflect.Struct {
			if field.CanInterface() {
				// Recurse using an interface of the field.
				err := sanitize(field.Interface(), fieldName)
				if err != nil {
					return err
				}
			}

			// Move onto the next field.
			continue
		}

		// Check if it's a struct value.
		if fieldKind == reflect.Struct {
			if field.CanAddr() && field.Addr().CanInterface() {
				// Recurse using an interface of the pointer value of the field.
				err := sanitize(field.Addr().Interface(), fieldName)
				if err != nil {
					return err
				}
			}

			// Move onto the next field.
			continue
		}

		fieldValue := val.Field(i).Interface()
		if fieldValue == "" {
			//fmt.Printf("name: %v, value: %v \n", fieldName, val.Field(i).Interface())
			msg := fmt.Sprintf("%v is null", fieldName)
			return errors.New(msg)
		}

		/* // Check if it's a string or a pointer to a string.
		if fieldKind == reflect.String || (fieldKind == reflect.Ptr && field.Elem().Kind() == reflect.String) {
			typeField := val.Type().Field(i)

			if typeField.Tag == "" {
				fmt.Printf("name: %v, value: %v", typeField.Name, typeField.Tag)
			}

			continue
		} */
	}

	return nil
}
