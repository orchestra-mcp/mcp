package tools_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/orchestra-mcp/mcp/src/tools"
)

// setupProject creates a project in a temp workspace and returns the workspace path.
func setupProject(t *testing.T) string {
	t.Helper()
	ws := t.TempDir()
	projTools := tools.Project(ws)
	res, err := projTools[1].Handler(map[string]any{"name": "Test App", "description": "A test project"})
	if err != nil || res.IsError {
		t.Fatalf("failed to create project: %v", err)
	}
	return ws
}

func TestCreateEpic(t *testing.T) {
	ws := setupProject(t)
	epicTools := tools.Epic(ws)

	res, err := epicTools[1].Handler(map[string]any{
		"project": "test-app", "title": "Auth System", "priority": "high",
	})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if res.IsError {
		t.Fatalf("returned error: %s", res.Content[0].Text)
	}
	var data map[string]any
	json.Unmarshal([]byte(res.Content[0].Text), &data)
	if data["type"] != "epic" {
		t.Errorf("type = %v, want epic", data["type"])
	}
	if data["status"] != "backlog" {
		t.Errorf("status = %v, want backlog", data["status"])
	}
	if data["title"] != "Auth System" {
		t.Errorf("title = %v", data["title"])
	}
}

func TestListEpicsEmpty(t *testing.T) {
	ws := setupProject(t)
	epicTools := tools.Epic(ws)

	res, _ := epicTools[0].Handler(map[string]any{"project": "test-app"})
	if res.Content[0].Text != "[]" && res.Content[0].Text != "null" {
		t.Errorf("expected empty list, got %s", res.Content[0].Text)
	}
}

func TestListEpicsAfterCreate(t *testing.T) {
	ws := setupProject(t)
	epicTools := tools.Epic(ws)
	epicTools[1].Handler(map[string]any{"project": "test-app", "title": "Epic One"})
	epicTools[1].Handler(map[string]any{"project": "test-app", "title": "Epic Two"})

	res, _ := epicTools[0].Handler(map[string]any{"project": "test-app"})
	var epics []map[string]any
	json.Unmarshal([]byte(res.Content[0].Text), &epics)
	if len(epics) != 2 {
		t.Errorf("expected 2 epics, got %d", len(epics))
	}
}

func TestGetEpic(t *testing.T) {
	ws := setupProject(t)
	epicTools := tools.Epic(ws)
	createRes, _ := epicTools[1].Handler(map[string]any{"project": "test-app", "title": "Auth"})
	var created map[string]any
	json.Unmarshal([]byte(createRes.Content[0].Text), &created)
	epicID := created["id"].(string)

	res, _ := epicTools[2].Handler(map[string]any{"project": "test-app", "epic_id": epicID})
	var data map[string]any
	json.Unmarshal([]byte(res.Content[0].Text), &data)
	if data["id"] != epicID {
		t.Errorf("id = %v, want %s", data["id"], epicID)
	}
}

func TestUpdateEpic(t *testing.T) {
	ws := setupProject(t)
	epicTools := tools.Epic(ws)
	createRes, _ := epicTools[1].Handler(map[string]any{"project": "test-app", "title": "Old"})
	var created map[string]any
	json.Unmarshal([]byte(createRes.Content[0].Text), &created)
	epicID := created["id"].(string)

	res, _ := epicTools[3].Handler(map[string]any{
		"project": "test-app", "epic_id": epicID,
		"title": "New Title", "priority": "critical",
	})
	var updated map[string]any
	json.Unmarshal([]byte(res.Content[0].Text), &updated)
	if updated["title"] != "New Title" {
		t.Errorf("title = %v", updated["title"])
	}
	if updated["priority"] != "critical" {
		t.Errorf("priority = %v", updated["priority"])
	}
}

func TestDeleteEpic(t *testing.T) {
	ws := setupProject(t)
	epicTools := tools.Epic(ws)
	createRes, _ := epicTools[1].Handler(map[string]any{"project": "test-app", "title": "Delete Me"})
	var created map[string]any
	json.Unmarshal([]byte(createRes.Content[0].Text), &created)
	epicID := created["id"].(string)

	res, _ := epicTools[4].Handler(map[string]any{"project": "test-app", "epic_id": epicID})
	if res.IsError {
		t.Fatalf("delete error: %s", res.Content[0].Text)
	}

	epicDir := filepath.Join(ws, ".projects", "test-app", "epics", epicID)
	if _, err := os.Stat(epicDir); !os.IsNotExist(err) {
		t.Error("epic directory should be deleted")
	}
}
