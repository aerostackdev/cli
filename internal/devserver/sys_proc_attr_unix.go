//go:build !windows
// +build !windows

package devserver

import (
	"os"
	"syscall"
)

func setProcessGroup(attr *syscall.SysProcAttr) {
	if attr != nil {
		attr.Setpgid = true
	}
}

func KillProcessGroup(p *os.Process) {
	if p != nil {
		// Kill the whole process group (negative PID)
		_ = syscall.Kill(-p.Pid, syscall.SIGKILL)
	}
}
