package transport

import (
	"context"
	"sync"

	"github.com/google/uuid"
)

// SSESession represents a single SSE client connection.
type SSESession struct {
	ID       string
	Messages chan []byte // outbound SSE messages
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewSSESession creates a new session with a buffered message channel.
func NewSSESession() *SSESession {
	ctx, cancel := context.WithCancel(context.Background())
	return &SSESession{
		ID:       uuid.New().String(),
		Messages: make(chan []byte, 32),
		ctx:      ctx,
		cancel:   cancel,
	}
}

// Close terminates the session.
func (s *SSESession) Close() { s.cancel() }

// Context returns the session context (cancelled when closed).
func (s *SSESession) Context() context.Context { return s.ctx }

// SSESessionManager tracks all active SSE sessions.
type SSESessionManager struct {
	mu       sync.RWMutex
	sessions map[string]*SSESession
}

// NewSSESessionManager creates a session manager.
func NewSSESessionManager() *SSESessionManager {
	return &SSESessionManager{sessions: make(map[string]*SSESession)}
}

// Create creates and registers a new session.
func (m *SSESessionManager) Create() *SSESession {
	sess := NewSSESession()
	m.mu.Lock()
	m.sessions[sess.ID] = sess
	m.mu.Unlock()
	return sess
}

// Get retrieves a session by ID.
func (m *SSESessionManager) Get(id string) (*SSESession, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	sess, ok := m.sessions[id]
	return sess, ok
}

// Remove unregisters and closes a session.
func (m *SSESessionManager) Remove(id string) {
	m.mu.Lock()
	sess, ok := m.sessions[id]
	if ok {
		delete(m.sessions, id)
	}
	m.mu.Unlock()
	if ok {
		sess.Close()
	}
}

// Count returns the number of active sessions.
func (m *SSESessionManager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.sessions)
}
