package types

// HookEvent represents a Claude Code hook event received by the MCP server.
type HookEvent struct {
	EventType string         `yaml:"event_type" json:"event_type"`
	SessionID string         `yaml:"session_id" json:"session_id"`
	ToolName  string         `yaml:"tool_name,omitempty" json:"tool_name,omitempty"`
	AgentType string         `yaml:"agent_type,omitempty" json:"agent_type,omitempty"`
	Data      map[string]any `yaml:"data,omitempty" json:"data,omitempty"`
	Timestamp string         `yaml:"timestamp" json:"timestamp"`
}

// HookEventLog stores a rolling list of hook events.
type HookEventLog struct {
	Events []HookEvent `yaml:"events" json:"events"`
}
