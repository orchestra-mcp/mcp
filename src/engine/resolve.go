package engine

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

const (
	BinaryName = "orchestra-engine"
	DefaultPort = 50051
	PortEnvVar  = "ORCHESTRA_ENGINE_PORT"
)

// Resolve returns the absolute path to the orchestra-engine binary, or "" if not found.
func Resolve() string {
	// 1. Same directory as the running binary
	if self, err := os.Executable(); err == nil {
		candidate := filepath.Join(filepath.Dir(self), BinaryName)
		if fileIsExecutable(candidate) {
			return candidate
		}
	}
	// 2. PATH lookup
	if p, err := exec.LookPath(BinaryName); err == nil {
		return p
	}
	return ""
}

// Port returns the engine port from env or default.
func Port() int {
	if v := os.Getenv(PortEnvVar); v != "" {
		if p, err := strconv.Atoi(v); err == nil && p > 0 {
			return p
		}
	}
	return DefaultPort
}

func fileIsExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir() && info.Mode()&0o111 != 0
}
