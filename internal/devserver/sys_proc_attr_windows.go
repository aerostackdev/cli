//go:build windows
// +build windows

package devserver

import "syscall"

func setProcessGroup(attr *syscall.SysProcAttr) {
	// Setpgid is not available on Windows
}
