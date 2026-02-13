package types

// MemoryChunk is a piece of project context stored for RAG retrieval.
type MemoryChunk struct {
	ID        string   `yaml:"id" json:"id"`
	Project   string   `yaml:"project" json:"project"`
	Source    string   `yaml:"source" json:"source"`       // task, prd, session, user
	SourceID  string   `yaml:"source_id" json:"source_id"` // task ID, session ID, etc.
	Summary   string   `yaml:"summary" json:"summary"`
	Content   string   `yaml:"content" json:"content"`
	Tags      []string `yaml:"tags,omitempty" json:"tags,omitempty"`
	CreatedAt string   `yaml:"created_at" json:"created_at"`
}

// MemoryIndex stores all memory chunks for a project.
type MemoryIndex struct {
	Chunks []MemoryChunk `yaml:"chunks" json:"chunks"`
}

// SessionLog records a Claude Code session for a project.
type SessionLog struct {
	SessionID string         `yaml:"session_id" json:"session_id"`
	Project   string         `yaml:"project" json:"project"`
	Summary   string         `yaml:"summary" json:"summary"`
	Events    []SessionEvent `yaml:"events,omitempty" json:"events,omitempty"`
	StartedAt string         `yaml:"started_at" json:"started_at"`
	EndedAt   string         `yaml:"ended_at,omitempty" json:"ended_at,omitempty"`
}

// SessionEvent is a single event within a session log.
type SessionEvent struct {
	Type      string `yaml:"type" json:"type"` // tool_call, decision, output
	Summary   string `yaml:"summary" json:"summary"`
	Timestamp string `yaml:"timestamp" json:"timestamp"`
}

// SessionIndex stores all session logs for a project.
type SessionIndex struct {
	Sessions []SessionLog `yaml:"sessions" json:"sessions"`
}
