/*
 * @Date: 2022-03-29 14:16:33
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-04-29 14:14:40
 * @FilePath: /smartide-cli/pkg/common/exec.go
 */

package common

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

type execOperation struct{}

var EXEC execOperation

func init() {
	EXEC = execOperation{}

}

// 实时运行
func (eo *execOperation) Realtime(command string, rootDir string) error {

	var execCommand *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		execCommand = exec.Command("powershell", "/c", command)
	case "darwin":
		execCommand = exec.Command("bash", "-c", command)
	case "linux":
		execCommand = exec.Command("bash", "-c", command)
	}
	if rootDir != "" {
		execCommand.Dir = rootDir
	}

	SmartIDELog.Debug(fmt.Sprintf("local (%v) exec -> %v ", runtime.GOOS, command))

	execCommand.Stdout = os.Stdout
	execCommand.Stderr = os.Stderr
	err := execCommand.Run()
	if err != nil {
		return err
	}

	return nil
}

// 一次性返回结果
func (eo *execOperation) CombinedOutput(command string, rootDir string) (string, error) {

	var execCommand *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		execCommand = exec.Command("powershell", "/c", command)
	case "darwin":
		execCommand = exec.Command("bash", "-c", command)
	case "linux":
		execCommand = exec.Command("bash", "-c", command)
	}
	if rootDir != "" {
		execCommand.Dir = rootDir
	}

	bytes, err := execCommand.CombinedOutput()
	return string(bytes), err
}
