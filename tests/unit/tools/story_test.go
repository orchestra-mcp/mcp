package tools_test

import (
	"encoding/json"
	"testing"

	"github.com/orchestra-mcp/mcp/src/tools"
)

func setupEpic(t *testing.T) (string, string) {
	t.Helper()
	ws := setupProject(t)
	epicTools := tools.Epic(ws)
	res, _ := epicTools[1].Handler(map[string]any{"project": "test-app", "title": "Auth"})
	var data map[string]any
	json.Unmarshal([]byte(res.Content[0].Text), &data)
	return ws, data["id"].(string)
}

func TestCreateStory(t *testing.T) {
	ws, epicID := setupEpic(t)
	storyTools := tools.Story(ws)

	res, err := storyTools[1].Handler(map[string]any{
		"project": "test-app", "epic_id": epicID,
		"title": "Login Flow", "user_story": "As a user I want to login",
	})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if res.IsError {
		t.Fatalf("returned error: %s", res.Content[0].Text)
	}
	var data map[string]any
	json.Unmarshal([]byte(res.Content[0].Text), &data)
	if data["type"] != "story" {
		t.Errorf("type = %v, want story", data["type"])
	}
	if data["status"] != "backlog" {
		t.Errorf("status = %v, want backlog", data["status"])
	}
}

func TestListStoriesEmpty(t *testing.T) {
	ws, epicID := setupEpic(t)
	storyTools := tools.Story(ws)
	res, _ := storyTools[0].Handler(map[string]any{"project": "test-app", "epic_id": epicID})
	if res.Content[0].Text != "[]" && res.Content[0].Text != "null" {
		t.Errorf("expected empty list, got %s", res.Content[0].Text)
	}
}

func TestListStoriesAfterCreate(t *testing.T) {
	ws, epicID := setupEpic(t)
	storyTools := tools.Story(ws)
	storyTools[1].Handler(map[string]any{
		"project": "test-app", "epic_id": epicID,
		"title": "S1", "user_story": "story",
	})
	res, _ := storyTools[0].Handler(map[string]any{"project": "test-app", "epic_id": epicID})
	var stories []map[string]any
	json.Unmarshal([]byte(res.Content[0].Text), &stories)
	if len(stories) != 1 {
		t.Errorf("expected 1 story, got %d", len(stories))
	}
}

func TestGetStory(t *testing.T) {
	ws, epicID := setupEpic(t)
	storyTools := tools.Story(ws)
	createRes, _ := storyTools[1].Handler(map[string]any{
		"project": "test-app", "epic_id": epicID,
		"title": "Login", "user_story": "As a user",
	})
	var created map[string]any
	json.Unmarshal([]byte(createRes.Content[0].Text), &created)
	storyID := created["id"].(string)

	res, _ := storyTools[2].Handler(map[string]any{
		"project": "test-app", "epic_id": epicID, "story_id": storyID,
	})
	var data map[string]any
	json.Unmarshal([]byte(res.Content[0].Text), &data)
	if data["id"] != storyID {
		t.Errorf("id = %v, want %s", data["id"], storyID)
	}
}

func TestUpdateStory(t *testing.T) {
	ws, epicID := setupEpic(t)
	storyTools := tools.Story(ws)
	createRes, _ := storyTools[1].Handler(map[string]any{
		"project": "test-app", "epic_id": epicID,
		"title": "Old", "user_story": "story",
	})
	var created map[string]any
	json.Unmarshal([]byte(createRes.Content[0].Text), &created)
	storyID := created["id"].(string)

	res, _ := storyTools[3].Handler(map[string]any{
		"project": "test-app", "epic_id": epicID, "story_id": storyID,
		"title": "New Title",
	})
	var updated map[string]any
	json.Unmarshal([]byte(res.Content[0].Text), &updated)
	if updated["title"] != "New Title" {
		t.Errorf("title = %v", updated["title"])
	}
}

func TestDeleteStory(t *testing.T) {
	ws, epicID := setupEpic(t)
	storyTools := tools.Story(ws)
	createRes, _ := storyTools[1].Handler(map[string]any{
		"project": "test-app", "epic_id": epicID,
		"title": "Delete Me", "user_story": "story",
	})
	var created map[string]any
	json.Unmarshal([]byte(createRes.Content[0].Text), &created)
	storyID := created["id"].(string)

	res, _ := storyTools[4].Handler(map[string]any{
		"project": "test-app", "epic_id": epicID, "story_id": storyID,
	})
	if res.IsError {
		t.Fatalf("delete error: %s", res.Content[0].Text)
	}
	listRes, _ := storyTools[0].Handler(map[string]any{"project": "test-app", "epic_id": epicID})
	if listRes.Content[0].Text != "[]" && listRes.Content[0].Text != "null" {
		t.Errorf("expected empty after delete, got %s", listRes.Content[0].Text)
	}
}

func TestStoryParentChildUpdate(t *testing.T) {
	ws, epicID := setupEpic(t)
	storyTools := tools.Story(ws)
	epicTools := tools.Epic(ws)

	storyTools[1].Handler(map[string]any{
		"project": "test-app", "epic_id": epicID,
		"title": "Login", "user_story": "story",
	})

	// Epic should now have a child
	epicRes, _ := epicTools[2].Handler(map[string]any{"project": "test-app", "epic_id": epicID})
	var epic map[string]any
	json.Unmarshal([]byte(epicRes.Content[0].Text), &epic)
	children, ok := epic["children"].([]any)
	if !ok || len(children) != 1 {
		t.Errorf("expected 1 child on epic, got %v", epic["children"])
	}
}
