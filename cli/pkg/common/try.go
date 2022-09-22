/*
 * @Author: jason chen (jasonchen@leansoftx.com, http://smallidea.cnblogs.com)
 * @Description:
 * @Date: 2021-11
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-09-22 14:38:48
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
		SmartIDELog.InfoF("This is attempt number %v / %v", i+1, attempts)
		// calling the important function
		err = f()
		if err != nil {
			SmartIDELog.Importance(fmt.Sprintf("error occured after attempt number %d/%d: %s", i+1, attempts, err.Error()))

			if i+1 == int(attempts) {
				continue
			}

			SmartIDELog.InfoF("sleeping for: %v", sleep.String())
			time.Sleep(sleep)
			sleep *= 2
			continue
		}
		break
	}
	return err
}
