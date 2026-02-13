package tools_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/orchestra-mcp/mcp/src/tools"
)

func TestReceiveHookEvent(t *testing.T) {
	ws := t.TempDir()
	os.MkdirAll(filepath.Join(ws, ".projects"), 0o755)
	claudeTools := tools.Claude(ws)
	receive := claudeTools[2] // receiveHookEvent

	res, err := receive.Handler(map[string]any{
		"event_type": "PostToolUse", "session_id": "s1", "tool_name": "Read",
	})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if res.IsError {
		t.Fatalf("returned error: %s", res.Content[0].Text)
	}
	var data map[string]any
	json.Unmarshal([]byte(res.Content[0].Text), &data)
	if data["stored"] != true {
		t.Errorf("stored = %v", data["stored"])
	}
}

func TestGetHookEvents(t *testing.T) {
	ws := t.TempDir()
	os.MkdirAll(filepath.Join(ws, ".projects"), 0o755)
	claudeTools := tools.Claude(ws)
	receive := claudeTools[2]
	getEvents := claudeTools[3]

	// Store two events
	receive.Handler(map[string]any{"event_type": "PostToolUse", "tool_name": "Read"})
	receive.Handler(map[string]any{"event_type": "Notification", "tool_name": ""})

	// Get all
	res, _ := getEvents.Handler(map[string]any{})
	var events []map[string]any
	json.Unmarshal([]byte(res.Content[0].Text), &events)
	if len(events) != 2 {
		t.Errorf("expected 2 events, got %d", len(events))
	}

	// Filter by type
	res, _ = getEvents.Handler(map[string]any{"event_type": "PostToolUse"})
	json.Unmarshal([]byte(res.Content[0].Text), &events)
	if len(events) != 1 {
		t.Errorf("expected 1 filtered event, got %d", len(events))
	}
}

func TestGetHookEventsLimit(t *testing.T) {
	ws := t.TempDir()
	os.MkdirAll(filepath.Join(ws, ".projects"), 0o755)
	claudeTools := tools.Claude(ws)
	receive := claudeTools[2]
	getEvents := claudeTools[3]

	for i := 0; i < 5; i++ {
		receive.Handler(map[string]any{"event_type": "PostToolUse"})
	}

	res, _ := getEvents.Handler(map[string]any{"limit": 3})
	var events []map[string]any
	json.Unmarshal([]byte(res.Content[0].Text), &events)
	if len(events) != 3 {
		t.Errorf("expected 3 (limited), got %d", len(events))
	}
}

func TestGetHookEventsEmpty(t *testing.T) {
	ws := t.TempDir()
	claudeTools := tools.Claude(ws)
	getEvents := claudeTools[3]

	res, _ := getEvents.Handler(map[string]any{})
	if res.Content[0].Text != "[]" {
		t.Errorf("expected empty, got %s", res.Content[0].Text)
	}
}

func TestListSkillsEmpty(t *testing.T) {
	ws := t.TempDir()
	claudeTools := tools.Claude(ws)
	res, _ := claudeTools[0].Handler(map[string]any{})
	if res.Content[0].Text != "[]" {
		t.Errorf("expected empty, got %s", res.Content[0].Text)
	}
}

func TestListAgentsEmpty(t *testing.T) {
	ws := t.TempDir()
	claudeTools := tools.Claude(ws)
	res, _ := claudeTools[1].Handler(map[string]any{})
	if res.Content[0].Text != "[]" {
		t.Errorf("expected empty, got %s", res.Content[0].Text)
	}
}

func TestListSkillsWithDir(t *testing.T) {
	ws := t.TempDir()
	skillDir := filepath.Join(ws, ".claude", "skills", "go-backend")
	os.MkdirAll(skillDir, 0o755)
	os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# Go Backend\n\nGo API patterns"), 0o644)

	claudeTools := tools.Claude(ws)
	res, _ := claudeTools[0].Handler(map[string]any{})
	var skills []map[string]any
	json.Unmarshal([]byte(res.Content[0].Text), &skills)
	if len(skills) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(skills))
	}
	if skills[0]["name"] != "go-backend" {
		t.Errorf("name = %v", skills[0]["name"])
	}
}

func TestListAgentsWithDir(t *testing.T) {
	ws := t.TempDir()
	agentsDir := filepath.Join(ws, ".claude", "agents")
	os.MkdirAll(agentsDir, 0o755)
	os.WriteFile(filepath.Join(agentsDir, "go-architect.md"), []byte("# Go Architect\n\nGo backend design"), 0o644)

	claudeTools := tools.Claude(ws)
	res, _ := claudeTools[1].Handler(map[string]any{})
	var agents []map[string]any
	json.Unmarshal([]byte(res.Content[0].Text), &agents)
	if len(agents) != 1 {
		t.Fatalf("expected 1 agent, got %d", len(agents))
	}
	if agents[0]["name"] != "go-architect" {
		t.Errorf("name = %v", agents[0]["name"])
	}
}
