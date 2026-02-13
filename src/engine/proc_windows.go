//go:build windows

package engine

import "os/exec"

func setProcAttr(cmd *exec.Cmd) {
	// No process group on Windows
}

func killProcess(cmd *exec.Cmd) {
	_ = cmd.Process.Kill()
}
