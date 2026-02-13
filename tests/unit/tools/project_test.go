package tools_test

import (
	"encoding/json"
	"testing"

	"github.com/orchestra-mcp/mcp/src/tools"
)

func TestListProjectsEmptyDir(t *testing.T) {
	ws := t.TempDir()
	toolList := tools.Project(ws)
	list := toolList[0].Handler
	res, err := list(map[string]any{})
	if err != nil {
		t.Fatalf("list error: %v", err)
	}
	if res.IsError {
		t.Fatalf("list returned error: %s", res.Content[0].Text)
	}
	if res.Content[0].Text != "[]" {
		t.Errorf("expected empty JSON array, got %s", res.Content[0].Text)
	}
}

func TestCreateProject(t *testing.T) {
	ws := t.TempDir()
	toolList := tools.Project(ws)
	create := toolList[1].Handler
	res, err := create(map[string]any{"name": "Test App", "description": "A test"})
	if err != nil {
		t.Fatalf("create error: %v", err)
	}
	if res.IsError {
		t.Fatalf("create returned error: %s", res.Content[0].Text)
	}
	var data map[string]any
	if err := json.Unmarshal([]byte(res.Content[0].Text), &data); err != nil {
		t.Fatalf("json parse: %v", err)
	}
	if data["slug"] != "test-app" {
		t.Errorf("slug = %v, want test-app", data["slug"])
	}
	if data["status"] != "created" {
		t.Errorf("status = %v, want created", data["status"])
	}
}

func TestGetProjectStatus(t *testing.T) {
	ws := t.TempDir()
	toolList := tools.Project(ws)
	// Create first
	toolList[1].Handler(map[string]any{"name": "Demo", "description": "d"})
	// Get status
	res, err := toolList[2].Handler(map[string]any{"project": "demo"})
	if err != nil {
		t.Fatalf("get_project_status error: %v", err)
	}
	if res.IsError {
		t.Fatalf("get_project_status error: %s", res.Content[0].Text)
	}
	var ps map[string]any
	json.Unmarshal([]byte(res.Content[0].Text), &ps)
	if ps["slug"] != "demo" {
		t.Errorf("slug = %v, want demo", ps["slug"])
	}
}

func TestReadPrd(t *testing.T) {
	ws := t.TempDir()
	toolList := tools.Project(ws)
	toolList[1].Handler(map[string]any{"name": "Demo", "description": "My project"})
	res, err := toolList[3].Handler(map[string]any{"project": "demo"})
	if err != nil {
		t.Fatalf("read_prd error: %v", err)
	}
	if res.IsError {
		t.Fatalf("read_prd returned error: %s", res.Content[0].Text)
	}
	if res.Content[0].Text == "" {
		t.Error("expected non-empty PRD")
	}
}

func TestWritePrd(t *testing.T) {
	ws := t.TempDir()
	toolList := tools.Project(ws)
	toolList[1].Handler(map[string]any{"name": "Demo", "description": "d"})
	res, _ := toolList[4].Handler(map[string]any{
		"project": "demo", "content": "# Updated PRD\n\nNew content",
	})
	if res.IsError {
		t.Fatalf("write_prd error: %s", res.Content[0].Text)
	}
	if res.Content[0].Text != "PRD updated" {
		t.Errorf("expected 'PRD updated', got %s", res.Content[0].Text)
	}

	// Read back
	readRes, _ := toolList[3].Handler(map[string]any{"project": "demo"})
	if readRes.Content[0].Text != "# Updated PRD\n\nNew content" {
		t.Errorf("PRD content mismatch: %s", readRes.Content[0].Text)
	}
}

func TestCreateProjectDuplicate(t *testing.T) {
	ws := t.TempDir()
	toolList := tools.Project(ws)
	toolList[1].Handler(map[string]any{"name": "Demo", "description": "d"})
	res, _ := toolList[1].Handler(map[string]any{"name": "Demo", "description": "d"})
	if !res.IsError {
		t.Error("expected error for duplicate project")
	}
}
