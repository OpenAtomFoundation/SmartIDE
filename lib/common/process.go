package common

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func IsLaunchedByDebugger() bool {

	// gops executable must be in the path. See https://github.com/google/gops, go get github.com/google/gops
	gopsOut, err := exec.Command("gops", strconv.Itoa(os.Getppid())).Output()
	if err == nil && strings.Contains(string(gopsOut), "\\dlv.exe") {
		// our parent process is (probably) the Delve debugger
		return true
	}
	return false
}
