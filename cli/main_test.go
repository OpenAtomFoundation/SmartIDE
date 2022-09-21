/*
 * @Date: 2022-07-15 09:23:34
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-09-20 17:07:32
 * @FilePath: /cli/main_test.go
 */
package main

import (
	"fmt"
	"testing"
)

// e.g. https://github.com/spf13/cobra/blob/master/command_test.go
func TestMain(t *testing.T) {
	bs := []byte{}
	fmt.Print("response:" + string(bs))

	// 基本的测试，查看命令是否可以正常跑起来
	main()
}
