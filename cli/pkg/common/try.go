/*
SmartIDE - Dev Containers
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

package common

import (
	"fmt"
	"time"
)

/*
example:

	Block{
	        Try: func() {
	            fmt.Println("I tried")
	            Throw("Oh,...sh...")
	        },
	        Catch: func(e Exception) {
	            fmt.Printf("Caught %v\n", e)
	        },
	        Finally: func() {
	            fmt.Println("Finally...")
	        },
	    }.Do()
*/
type Block struct {
	Try     func()
	Catch   func(Exception)
	Finally func()
}

type Exception interface{}

func Throw(up Exception) {
	panic(up)
}

func (tcf Block) Do() {
	if tcf.Finally != nil {

		defer tcf.Finally()
	}
	if tcf.Catch != nil {
		defer func() {
			if r := recover(); r != nil {
				tcf.Catch(r)
			}
		}()
	}
	tcf.Try()
}

// 重试
func Retry(attempts uint, sleep time.Duration, f func() error) (err error) {
	err = f()
	if err == nil || attempts == 0 {
		return err
	}

	for i := 0; i < int(attempts); i++ {
		SmartIDELog.DebugF("This is attempt number %v / %v", i+1, attempts)
		// calling the important function
		err = f()
		if err != nil {
			warning := fmt.Sprintf("error occured after attempt number %d/%d: %s", i+1, attempts, err.Error())
			SmartIDELog.WarningF(warning)
			if i+1 == int(attempts) {
				continue
			}

			SmartIDELog.DebugF("sleeping for: %v", sleep.String())
			time.Sleep(sleep)
			sleep *= 2
			continue
		}
		break
	}
	return err
}
