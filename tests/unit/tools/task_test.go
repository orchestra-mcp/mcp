package tools_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/orchestra-mcp/mcp/src/tools"
)

func setupStory(t *testing.T) (string, string, string) {
	t.Helper()
	ws, epicID := setupEpic(t)
	storyTools := tools.Story(ws)
	res, _ := storyTools[1].Handler(map[string]any{
		"project": "test-app", "epic_id": epicID,
		"title": "Login", "user_story": "As a user I want to login",
	})
	var data map[string]any
	json.Unmarshal([]byte(res.Content[0].Text), &data)
	return ws, epicID, data["id"].(string)
}

func TestCreateTask(t *testing.T) {
	ws, epicID, storyID := setupStory(t)
	taskTools := tools.Task(ws)

	res, err := taskTools[1].Handler(map[string]any{
		"project": "test-app", "epic_id": epicID, "story_id": storyID,
		"title": "Build API", "type": "task", "priority": "high",
	})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if res.IsError {
		t.Fatalf("returned error: %s", res.Content[0].Text)
	}
	var data map[string]any
	json.Unmarshal([]byte(res.Content[0].Text), &data)
	if data["type"] != "task" {
		t.Errorf("type = %v", data["type"])
	}
	if data["status"] != "backlog" {
		t.Errorf("status = %v", data["status"])
	}
}

func TestCreateBugType(t *testing.T) {
	ws, epicID, storyID := setupStory(t)
	taskTools := tools.Task(ws)
	res, _ := taskTools[1].Handler(map[string]any{
		"project": "test-app", "epic_id": epicID, "story_id": storyID,
		"title": "Fix crash", "type": "bug",
	})
	var data map[string]any
	json.Unmarshal([]byte(res.Content[0].Text), &data)
	if data["type"] != "bug" {
		t.Errorf("type = %v, want bug", data["type"])
	}
}

func TestListTasks(t *testing.T) {
	ws, epicID, storyID := setupStory(t)
	taskTools := tools.Task(ws)

	res, _ := taskTools[0].Handler(map[string]any{
		"project": "test-app", "epic_id": epicID, "story_id": storyID,
	})
	if res.Content[0].Text != "[]" && res.Content[0].Text != "null" {
		t.Errorf("expected empty list, got %s", res.Content[0].Text)
	}

	taskTools[1].Handler(map[string]any{
		"project": "test-app", "epic_id": epicID, "story_id": storyID,
		"title": "Task 1", "type": "task",
	})
	taskTools[1].Handler(map[string]any{
		"project": "test-app", "epic_id": epicID, "story_id": storyID,
		"title": "Task 2", "type": "task",
	})

	res, _ = taskTools[0].Handler(map[string]any{
		"project": "test-app", "epic_id": epicID, "story_id": storyID,
	})
	var tasks []map[string]any
	json.Unmarshal([]byte(res.Content[0].Text), &tasks)
	if len(tasks) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(tasks))
	}
}

func TestGetTask(t *testing.T) {
	ws, epicID, storyID := setupStory(t)
	taskTools := tools.Task(ws)
	createRes, _ := taskTools[1].Handler(map[string]any{
		"project": "test-app", "epic_id": epicID, "story_id": storyID,
		"title": "My Task", "type": "task",
	})
	var created map[string]any
	json.Unmarshal([]byte(createRes.Content[0].Text), &created)
	taskID := created["id"].(string)

	res, _ := taskTools[2].Handler(map[string]any{
		"project": "test-app", "epic_id": epicID, "story_id": storyID, "task_id": taskID,
	})
	var data map[string]any
	json.Unmarshal([]byte(res.Content[0].Text), &data)
	if data["id"] != taskID {
		t.Errorf("id = %v, want %s", data["id"], taskID)
	}
}

func TestUpdateTaskValidTransition(t *testing.T) {
	ws, epicID, storyID := setupStory(t)
	taskTools := tools.Task(ws)
	createRes, _ := taskTools[1].Handler(map[string]any{
		"project": "test-app", "epic_id": epicID, "story_id": storyID,
		"title": "Task", "type": "task",
	})
	var created map[string]any
	json.Unmarshal([]byte(createRes.Content[0].Text), &created)
	taskID := created["id"].(string)

	// Valid: backlog -> todo
	res, _ := taskTools[3].Handler(map[string]any{
		"project": "test-app", "epic_id": epicID, "story_id": storyID, "task_id": taskID,
		"status": "todo",
	})
	if res.IsError {
		t.Fatalf("valid transition error: %s", res.Content[0].Text)
	}
	var updated map[string]any
	json.Unmarshal([]byte(res.Content[0].Text), &updated)
	if updated["status"] != "todo" {
		t.Errorf("status = %v, want todo", updated["status"])
	}
}

func TestUpdateTaskInvalidTransition(t *testing.T) {
	ws, epicID, storyID := setupStory(t)
	taskTools := tools.Task(ws)
	createRes, _ := taskTools[1].Handler(map[string]any{
		"project": "test-app", "epic_id": epicID, "story_id": storyID,
		"title": "Task", "type": "task",
	})
	var created map[string]any
	json.Unmarshal([]byte(createRes.Content[0].Text), &created)
	taskID := created["id"].(string)

	// Invalid: backlog -> done
	res, _ := taskTools[3].Handler(map[string]any{
		"project": "test-app", "epic_id": epicID, "story_id": storyID, "task_id": taskID,
		"status": "done",
	})
	if !res.IsError {
		t.Error("expected error for invalid transition")
	}
	if !strings.Contains(res.Content[0].Text, "invalid transition") {
		t.Errorf("error = %s", res.Content[0].Text)
	}
}

func TestDeleteTask(t *testing.T) {
	ws, epicID, storyID := setupStory(t)
	taskTools := tools.Task(ws)
	createRes, _ := taskTools[1].Handler(map[string]any{
		"project": "test-app", "epic_id": epicID, "story_id": storyID,
		"title": "Delete Me", "type": "task",
	})
	var created map[string]any
	json.Unmarshal([]byte(createRes.Content[0].Text), &created)
	taskID := created["id"].(string)

	res, _ := taskTools[4].Handler(map[string]any{
		"project": "test-app", "epic_id": epicID, "story_id": storyID, "task_id": taskID,
	})
	if res.IsError {
		t.Fatalf("delete error: %s", res.Content[0].Text)
	}
	if !strings.Contains(res.Content[0].Text, "deleted") {
		t.Errorf("expected 'deleted', got: %s", res.Content[0].Text)
	}
}
