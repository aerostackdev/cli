//go:build !windows
// +build !windows

package devserver

import "syscall"

func setProcessGroup(attr *syscall.SysProcAttr) {
	if attr != nil {
		attr.Setpgid = true
	}
}
