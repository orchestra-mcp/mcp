package engine

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"syscall"
	"time"
)

// ErrNotFound indicates the engine binary was not found.
var ErrNotFound = errors.New("orchestra-engine binary not found")

// Manager controls the lifecycle of the orchestra-engine subprocess.
type Manager struct {
	cmd     *exec.Cmd
	port    int
	running bool
	spawned bool // true if we started the process (vs. found existing)
}

// NewManager creates a new engine manager.
func NewManager() *Manager {
	return &Manager{port: Port()}
}

// Start finds the engine binary and starts it. Returns ErrNotFound if binary is missing.
func (m *Manager) Start(workspace string) error {
	// Check if something is already listening
	if portOpen(m.port) {
		m.running = true
		return nil
	}

	bin := Resolve()
	if bin == "" {
		return ErrNotFound
	}

	m.cmd = exec.Command(bin, "--workspace", workspace)
	m.cmd.Stdout = os.Stderr // engine stdout → MCP stderr (not JSON-RPC stdout)
	m.cmd.Stderr = os.Stderr

	// Process group for clean cleanup (unix only)
	if runtime.GOOS != "windows" {
		m.cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	}

	if err := m.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start engine: %w", err)
	}
	m.spawned = true

	// Wait for engine to become reachable
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if portOpen(m.port) {
			m.running = true
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Engine started but not listening — kill it
	m.kill()
	return errors.New("engine started but not reachable within 3s")
}

// Stop gracefully terminates the engine subprocess.
func (m *Manager) Stop() {
	if !m.spawned || m.cmd == nil || m.cmd.Process == nil {
		return
	}
	m.kill()
	m.running = false
}

// IsRunning returns true if the engine is available.
func (m *Manager) IsRunning() bool { return m.running }

// Addr returns the gRPC address.
func (m *Manager) Addr() string { return fmt.Sprintf("localhost:%d", m.port) }

func (m *Manager) kill() {
	if m.cmd == nil || m.cmd.Process == nil {
		return
	}
	if runtime.GOOS != "windows" {
		// Kill process group
		_ = syscall.Kill(-m.cmd.Process.Pid, syscall.SIGTERM)
	} else {
		_ = m.cmd.Process.Kill()
	}
	done := make(chan struct{})
	go func() { _ = m.cmd.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		_ = m.cmd.Process.Kill()
		<-done
	}
}

func portOpen(port int) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", port), 200*time.Millisecond)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
