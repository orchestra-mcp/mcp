package tools_test

import (
	"encoding/json"
	"testing"

	"github.com/orchestra-mcp/mcp/src/tools"
)

func TestReportBug(t *testing.T) {
	ws, epicID, storyID := setupStory(t)
	bugTools := tools.Bugfix(ws)
	reportBug := bugTools[0]

	res, err := reportBug.Handler(map[string]any{
		"project": "test-app", "story_id": storyID,
		"title": "Login crashes", "severity": "high",
		"steps": "1. Click login", "expected": "Success", "actual": "Crash",
	})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if res.IsError {
		t.Fatalf("returned error: %s", res.Content[0].Text)
	}
	var data map[string]any
	json.Unmarshal([]byte(res.Content[0].Text), &data)
	if data["status"] != "created" {
		t.Errorf("status = %v", data["status"])
	}

	// Verify story has the bug as child
	storyTools := tools.Story(ws)
	storyRes, _ := storyTools[2].Handler(map[string]any{
		"project": "test-app", "epic_id": epicID, "story_id": storyID,
	})
	var story map[string]any
	json.Unmarshal([]byte(storyRes.Content[0].Text), &story)
	children := story["children"].([]any)
	found := false
	for _, c := range children {
		child := c.(map[string]any)
		if child["title"] == "Login crashes" {
			found = true
		}
	}
	if !found {
		t.Error("bug not found in story children")
	}
}

func TestReportBugStoryNotFound(t *testing.T) {
	ws := setupProject(t)
	bugTools := tools.Bugfix(ws)

	res, _ := bugTools[0].Handler(map[string]any{
		"project": "test-app", "story_id": "NONEXISTENT",
		"title": "Bug", "severity": "low",
	})
	if !res.IsError {
		t.Error("expected error for nonexistent story")
	}
}

func TestLogRequest(t *testing.T) {
	ws := setupProject(t)
	bugTools := tools.Bugfix(ws)
	logReq := bugTools[1]

	res, err := logReq.Handler(map[string]any{
		"project": "test-app", "type": "feature",
		"description": "Add dark mode",
	})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if res.IsError {
		t.Fatalf("returned error: %s", res.Content[0].Text)
	}
	var data map[string]any
	json.Unmarshal([]byte(res.Content[0].Text), &data)
	if data["status"] != "logged" {
		t.Errorf("status = %v", data["status"])
	}
	count, _ := data["count"].(float64)
	if count != 1 {
		t.Errorf("count = %v, want 1", data["count"])
	}
}
