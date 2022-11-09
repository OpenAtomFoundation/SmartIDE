/*
 * @Date: 2022-03-29 14:16:33
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-11-09 13:39:16
 * @FilePath: /cli/pkg/common/exec.go
 */

package common

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
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

	currentWorkingDir := ""
	if execCommand.Dir != "" {
		currentWorkingDir = fmt.Sprintf("> %v", execCommand.Dir)
	}
	SmartIDELog.Debug(fmt.Sprintf("local (%v) exec %v -> %v ",
		runtime.GOOS, currentWorkingDir, command))

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
	} else {
		homeDir, _ := os.UserHomeDir()
		if homeDir != "" {
			execCommand.Dir = homeDir
		}
	}

	bytes, err := execCommand.CombinedOutput()

	output := strings.TrimSpace(string(bytes))
	SmartIDELog.Debug(fmt.Sprintf("local (%v) exec -> %v >>\n%v", runtime.GOOS, command, output))

	return output, err
}
