package engine_test

import (
	"os"
	"testing"

	"github.com/orchestra-mcp/mcp/src/engine"
)

func TestDefaultPort(t *testing.T) {
	if engine.DefaultPort != 50051 {
		t.Errorf("DefaultPort = %d, want 50051", engine.DefaultPort)
	}
}

func TestPortDefault(t *testing.T) {
	// Unset env var to test default
	os.Unsetenv(engine.PortEnvVar)
	if p := engine.Port(); p != engine.DefaultPort {
		t.Errorf("Port() = %d, want %d", p, engine.DefaultPort)
	}
}

func TestPortFromEnv(t *testing.T) {
	os.Setenv(engine.PortEnvVar, "9999")
	defer os.Unsetenv(engine.PortEnvVar)
	if p := engine.Port(); p != 9999 {
		t.Errorf("Port() = %d, want 9999", p)
	}
}

func TestPortFromEnvInvalid(t *testing.T) {
	os.Setenv(engine.PortEnvVar, "not-a-number")
	defer os.Unsetenv(engine.PortEnvVar)
	if p := engine.Port(); p != engine.DefaultPort {
		t.Errorf("Port() = %d, want default %d for invalid env", p, engine.DefaultPort)
	}
}

func TestBridgeNilClient(t *testing.T) {
	bridge := engine.NewBridge(nil, "/tmp/test")
	if bridge.UsingEngine() {
		t.Error("UsingEngine() should be false with nil client")
	}
}

func TestNewManager(t *testing.T) {
	os.Unsetenv(engine.PortEnvVar)
	mgr := engine.NewManager()
	if mgr.IsRunning() {
		t.Error("new manager should not be running")
	}
	if mgr.Addr() != "localhost:50051" {
		t.Errorf("Addr() = %q, want localhost:50051", mgr.Addr())
	}
}

func TestManagerStartNotFound(t *testing.T) {
	// Set PATH to empty to ensure binary is not found
	oldPath := os.Getenv("PATH")
	os.Setenv("PATH", "")
	defer os.Setenv("PATH", oldPath)

	mgr := engine.NewManager()
	err := mgr.Start(t.TempDir())
	if err == nil {
		mgr.Stop()
		t.Fatal("expected error when binary not found")
	}
	if err != engine.ErrNotFound {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestManagerStopNoOp(t *testing.T) {
	mgr := engine.NewManager()
	// Stop on non-started manager should not panic
	mgr.Stop()
}

func TestBinaryName(t *testing.T) {
	if engine.BinaryName != "orchestra-engine" {
		t.Errorf("BinaryName = %q", engine.BinaryName)
	}
}
