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
	SmartIDELog.Debug(fmt.Sprintf("local realtime (%v) exec %v -> %v ",
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
	SmartIDELog.Debug(fmt.Sprintf("local combine (%v) exec -> %v >>\n%v", runtime.GOOS, command, output))

	return output, err
}
