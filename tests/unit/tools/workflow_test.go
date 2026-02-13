package tools_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/orchestra-mcp/mcp/src/tools"
)

func TestGetNextTask(t *testing.T) {
	ws, epicID, storyID := setupStory(t)
	taskTools := tools.Task(ws)
	wfTools := tools.Workflow(ws)

	// Create two tasks
	taskTools[1].Handler(map[string]any{
		"project": "test-app", "epic_id": epicID, "story_id": storyID,
		"title": "Task A", "type": "task",
	})
	taskTools[1].Handler(map[string]any{
		"project": "test-app", "epic_id": epicID, "story_id": storyID,
		"title": "Bug B", "type": "bug",
	})

	res, err := wfTools[0].Handler(map[string]any{"project": "test-app"})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if res.IsError {
		t.Fatalf("returned error: %s", res.Content[0].Text)
	}
	// Bug should come first (higher priority)
	var data map[string]any
	json.Unmarshal([]byte(res.Content[0].Text), &data)
	if data["type"] != "bug" {
		t.Errorf("expected bug first, got type=%v", data["type"])
	}
}

func TestGetNextTaskEmpty(t *testing.T) {
	ws := setupProject(t)
	wfTools := tools.Workflow(ws)

	res, _ := wfTools[0].Handler(map[string]any{"project": "test-app"})
	if !strings.Contains(res.Content[0].Text, "no actionable") {
		t.Errorf("expected 'no actionable', got: %s", res.Content[0].Text)
	}
}

func TestSetCurrentTask(t *testing.T) {
	ws, epicID, storyID := setupStory(t)
	taskTools := tools.Task(ws)
	wfTools := tools.Workflow(ws)

	createRes, _ := taskTools[1].Handler(map[string]any{
		"project": "test-app", "epic_id": epicID, "story_id": storyID,
		"title": "Task", "type": "task",
	})
	var created map[string]any
	json.Unmarshal([]byte(createRes.Content[0].Text), &created)
	taskID := created["id"].(string)

	// Move to todo first (backlog -> todo)
	taskTools[3].Handler(map[string]any{
		"project": "test-app", "epic_id": epicID, "story_id": storyID, "task_id": taskID,
		"status": "todo",
	})

	// set_current_task (todo -> in-progress)
	res, _ := wfTools[1].Handler(map[string]any{
		"project": "test-app", "epic_id": epicID, "story_id": storyID, "task_id": taskID,
	})
	if res.IsError {
		t.Fatalf("set_current error: %s", res.Content[0].Text)
	}
	var data map[string]any
	json.Unmarshal([]byte(res.Content[0].Text), &data)
	if data["status"] != "in-progress" {
		t.Errorf("status = %v, want in-progress", data["status"])
	}

	// Verify epic cascaded to in-progress
	epicTools := tools.Epic(ws)
	epicRes, _ := epicTools[2].Handler(map[string]any{"project": "test-app", "epic_id": epicID})
	var epic map[string]any
	json.Unmarshal([]byte(epicRes.Content[0].Text), &epic)
	if epic["status"] != "in-progress" {
		t.Errorf("epic status = %v, want in-progress", epic["status"])
	}
}

func TestCompleteTask(t *testing.T) {
	ws, epicID, storyID, taskID := setupTaskInProgress(t)
	wfTools := tools.Workflow(ws)

	// complete_task (in-progress -> ready-for-testing)
	res, _ := wfTools[2].Handler(map[string]any{
		"project": "test-app", "epic_id": epicID, "story_id": storyID, "task_id": taskID,
	})
	if res.IsError {
		t.Fatalf("complete_task error: %s", res.Content[0].Text)
	}
	var data map[string]any
	json.Unmarshal([]byte(res.Content[0].Text), &data)
	if data["status"] != "ready-for-testing" {
		t.Errorf("status = %v, want ready-for-testing", data["status"])
	}
}

func TestCompleteTaskFromInvalidState(t *testing.T) {
	ws, epicID, storyID := setupStory(t)
	taskTools := tools.Task(ws)
	wfTools := tools.Workflow(ws)

	createRes, _ := taskTools[1].Handler(map[string]any{
		"project": "test-app", "epic_id": epicID, "story_id": storyID,
		"title": "Task", "type": "task",
	})
	var created map[string]any
	json.Unmarshal([]byte(createRes.Content[0].Text), &created)
	taskID := created["id"].(string)

	// Try complete from backlog (should fail)
	res, _ := wfTools[2].Handler(map[string]any{
		"project": "test-app", "epic_id": epicID, "story_id": storyID, "task_id": taskID,
	})
	if !res.IsError {
		t.Error("expected error completing from backlog")
	}
}

func TestSearchIssues(t *testing.T) {
	ws, epicID, storyID := setupStory(t)
	taskTools := tools.Task(ws)
	wfTools := tools.Workflow(ws)

	taskTools[1].Handler(map[string]any{
		"project": "test-app", "epic_id": epicID, "story_id": storyID,
		"title": "Login API endpoint", "type": "task",
	})
	taskTools[1].Handler(map[string]any{
		"project": "test-app", "epic_id": epicID, "story_id": storyID,
		"title": "Dashboard UI", "type": "task",
	})

	// Search for "login" — should match the epic (Auth), story (Login), and task
	res, _ := wfTools[3].Handler(map[string]any{"project": "test-app", "query": "login"})
	var matches []map[string]any
	json.Unmarshal([]byte(res.Content[0].Text), &matches)
	if len(matches) == 0 {
		t.Error("expected matches for 'login'")
	}

	// Search with type filter
	res, _ = wfTools[3].Handler(map[string]any{"project": "test-app", "query": "login", "type": "task"})
	json.Unmarshal([]byte(res.Content[0].Text), &matches)
	for _, m := range matches {
		if m["type"] != "task" {
			t.Errorf("type filter failed, got type=%v", m["type"])
		}
	}
}

func TestGetWorkflowStatus(t *testing.T) {
	ws, epicID, storyID := setupStory(t)
	taskTools := tools.Task(ws)
	wfTools := tools.Workflow(ws)

	taskTools[1].Handler(map[string]any{
		"project": "test-app", "epic_id": epicID, "story_id": storyID,
		"title": "T1", "type": "task",
	})
	taskTools[1].Handler(map[string]any{
		"project": "test-app", "epic_id": epicID, "story_id": storyID,
		"title": "T2", "type": "bug",
	})

	res, _ := wfTools[4].Handler(map[string]any{"project": "test-app"})
	if res.IsError {
		t.Fatalf("error: %s", res.Content[0].Text)
	}
	var data map[string]any
	json.Unmarshal([]byte(res.Content[0].Text), &data)
	total, _ := data["total"].(float64)
	if total != 2 {
		t.Errorf("total = %v, want 2", data["total"])
	}
}
