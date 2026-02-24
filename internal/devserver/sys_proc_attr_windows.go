//go:build windows
// +build windows

package devserver

import (
	"os"
	"syscall"
)

func setProcessGroup(attr *syscall.SysProcAttr) {
	// Setpgid is not available on Windows
}

func KillProcessGroup(p *os.Process) {
	if p != nil {
		_ = p.Kill()
	}
}
