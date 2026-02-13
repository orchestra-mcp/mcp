package tools_test

import (
	"encoding/json"
	"testing"

	"github.com/orchestra-mcp/mcp/src/engine"
	"github.com/orchestra-mcp/mcp/src/tools"
)

func memoryBridge(ws string) *engine.Bridge {
	return engine.NewBridge(nil, ws) // TOON fallback
}

func TestSaveMemory(t *testing.T) {
	ws := setupProject(t)
	bridge := memoryBridge(ws)
	memTools := tools.Memory(ws, bridge)

	res, err := memTools[0].Handler(map[string]any{
		"project": "test-app", "content": "Use JWT for auth",
		"summary": "Auth decision", "source": "task", "source_id": "T-1",
		"tags": []any{"auth", "jwt"},
	})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if res.IsError {
		t.Fatalf("returned error: %s", res.Content[0].Text)
	}
	var chunk map[string]any
	json.Unmarshal([]byte(res.Content[0].Text), &chunk)
	if chunk["id"] != "mem-1" {
		t.Errorf("id = %v, want mem-1", chunk["id"])
	}
	if chunk["summary"] != "Auth decision" {
		t.Errorf("summary = %v", chunk["summary"])
	}
}

func TestSearchMemory(t *testing.T) {
	ws := setupProject(t)
	bridge := memoryBridge(ws)
	memTools := tools.Memory(ws, bridge)

	// Save two chunks
	memTools[0].Handler(map[string]any{
		"project": "test-app", "content": "Use JWT for authentication",
		"summary": "Auth decision", "tags": []any{"auth", "jwt"},
	})
	memTools[0].Handler(map[string]any{
		"project": "test-app", "content": "Use PostgreSQL for database",
		"summary": "DB decision", "tags": []any{"database", "postgresql"},
	})

	// Search for "JWT" - should match first chunk
	res, _ := memTools[1].Handler(map[string]any{
		"project": "test-app", "query": "jwt authentication",
	})
	var results []map[string]any
	json.Unmarshal([]byte(res.Content[0].Text), &results)
	if len(results) == 0 {
		t.Fatal("expected search results for 'jwt authentication'")
	}
	chunk := results[0]["chunk"].(map[string]any)
	if chunk["id"] != "mem-1" {
		t.Errorf("expected mem-1 first, got %v", chunk["id"])
	}
}

func TestSearchMemoryNoResults(t *testing.T) {
	ws := setupProject(t)
	bridge := memoryBridge(ws)
	memTools := tools.Memory(ws, bridge)

	res, _ := memTools[1].Handler(map[string]any{
		"project": "test-app", "query": "nonexistent",
	})
	if res.Content[0].Text != "[]" && res.Content[0].Text != "null" {
		t.Errorf("expected empty, got %s", res.Content[0].Text)
	}
}

func TestGetContext(t *testing.T) {
	ws := setupProject(t)
	bridge := memoryBridge(ws)
	memTools := tools.Memory(ws, bridge)

	memTools[0].Handler(map[string]any{
		"project": "test-app", "content": "JWT auth tokens",
		"summary": "Auth approach", "tags": []any{"auth"},
	})

	res, _ := memTools[2].Handler(map[string]any{
		"project": "test-app", "query": "auth approach",
	})
	var items []map[string]any
	json.Unmarshal([]byte(res.Content[0].Text), &items)
	if len(items) == 0 {
		t.Error("expected context results")
	}
}

func TestSaveSession(t *testing.T) {
	ws := setupProject(t)
	bridge := memoryBridge(ws)
	memTools := tools.Memory(ws, bridge)

	res, _ := memTools[3].Handler(map[string]any{
		"project": "test-app", "session_id": "sess-001",
		"summary": "Implemented auth", "events": []any{
			map[string]any{"type": "task_completed", "summary": "Auth done"},
		},
	})
	if res.IsError {
		t.Fatalf("error: %s", res.Content[0].Text)
	}
	var session map[string]any
	json.Unmarshal([]byte(res.Content[0].Text), &session)
	if session["session_id"] != "sess-001" {
		t.Errorf("session_id = %v", session["session_id"])
	}
}

func TestListSessions(t *testing.T) {
	ws := setupProject(t)
	bridge := memoryBridge(ws)
	memTools := tools.Memory(ws, bridge)

	// Empty at first
	res, _ := memTools[4].Handler(map[string]any{"project": "test-app"})
	if res.Content[0].Text != "[]" && res.Content[0].Text != "null" {
		t.Errorf("expected empty, got %s", res.Content[0].Text)
	}

	// Save a session
	memTools[3].Handler(map[string]any{
		"project": "test-app", "session_id": "sess-001", "summary": "Session 1",
	})

	res, _ = memTools[4].Handler(map[string]any{"project": "test-app"})
	var sessions []map[string]any
	json.Unmarshal([]byte(res.Content[0].Text), &sessions)
	if len(sessions) != 1 {
		t.Errorf("expected 1 session, got %d", len(sessions))
	}
}

func TestGetSession(t *testing.T) {
	ws := setupProject(t)
	bridge := memoryBridge(ws)
	memTools := tools.Memory(ws, bridge)

	memTools[3].Handler(map[string]any{
		"project": "test-app", "session_id": "sess-001", "summary": "Test session",
	})

	res, _ := memTools[5].Handler(map[string]any{
		"project": "test-app", "session_id": "sess-001",
	})
	if res.IsError {
		t.Fatalf("error: %s", res.Content[0].Text)
	}
	var session map[string]any
	json.Unmarshal([]byte(res.Content[0].Text), &session)
	if session["session_id"] != "sess-001" {
		t.Errorf("session_id = %v", session["session_id"])
	}
}
