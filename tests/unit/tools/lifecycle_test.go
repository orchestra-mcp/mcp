package tools_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/orchestra-mcp/mcp/src/tools"
)

// setupTaskInProgress creates project/epic/story/task and moves task to in-progress.
// Returns (workspace, epicID, storyID, taskID).
func setupTaskInProgress(t *testing.T) (string, string, string, string) {
	t.Helper()
	ws, epicID, storyID := setupStory(t)
	taskTools := tools.Task(ws)

	// Create task
	createRes, _ := taskTools[1].Handler(map[string]any{
		"project": "test-app", "epic_id": epicID, "story_id": storyID,
		"title": "API Handler", "type": "task",
	})
	var created map[string]any
	json.Unmarshal([]byte(createRes.Content[0].Text), &created)
	taskID := created["id"].(string)

	// Move to todo, then in-progress
	taskTools[3].Handler(map[string]any{
		"project": "test-app", "epic_id": epicID, "story_id": storyID, "task_id": taskID,
		"status": "todo",
	})
	taskTools[3].Handler(map[string]any{
		"project": "test-app", "epic_id": epicID, "story_id": storyID, "task_id": taskID,
		"status": "in-progress",
	})
	return ws, epicID, storyID, taskID
}

func TestAdvanceTask(t *testing.T) {
	ws, epicID, storyID, taskID := setupTaskInProgress(t)
	lcTools := tools.Lifecycle(ws)
	advance := lcTools[0]

	// Advance from in-progress -> ready-for-testing (gated, needs evidence)
	res, err := advance.Handler(map[string]any{
		"project": "test-app", "epic_id": epicID, "story_id": storyID, "task_id": taskID,
		"evidence": "go test ./... — 5/5 passed",
	})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if res.IsError {
		t.Fatalf("returned error: %s", res.Content[0].Text)
	}
	var data map[string]any
	json.Unmarshal([]byte(res.Content[0].Text), &data)
	if data["from"] != "in-progress" {
		t.Errorf("from = %v", data["from"])
	}
	if data["to"] != "ready-for-testing" {
		t.Errorf("to = %v", data["to"])
	}
	if data["evidence"] != "go test ./... — 5/5 passed" {
		t.Errorf("evidence not stored in response")
	}
}

func TestAdvanceGateBlocked(t *testing.T) {
	ws, epicID, storyID, taskID := setupTaskInProgress(t)
	lcTools := tools.Lifecycle(ws)
	advance := lcTools[0]

	// Try advancing from in-progress WITHOUT evidence — should be blocked
	res, _ := advance.Handler(map[string]any{
		"project": "test-app", "epic_id": epicID, "story_id": storyID, "task_id": taskID,
	})
	if !res.IsError {
		t.Error("expected gate block when advancing without evidence")
	}
	if !strings.Contains(res.Content[0].Text, "GATE BLOCKED") {
		t.Errorf("error = %s, expected GATE BLOCKED", res.Content[0].Text)
	}
}

func TestAdvanceFullChain(t *testing.T) {
	ws, epicID, storyID, taskID := setupTaskInProgress(t)
	lcTools := tools.Lifecycle(ws)
	advance := lcTools[0]

	// Gated transitions need evidence; non-gated don't.
	type step struct {
		to       string
		evidence string
	}
	steps := []step{
		{"ready-for-testing", "go test — 5/5 passed"}, // GATE 1: in-progress
		{"in-testing", ""}, // no gate
		{"ready-for-docs", "coverage 90%, edge cases OK"}, // GATE 2: in-testing
		{"in-docs", ""}, // no gate
		{"documented", "added godoc to all exported funcs"}, // GATE 3: in-docs
		{"in-review", ""}, // no gate
		{"done", "reviewed: code quality OK, no issues"}, // GATE 4: in-review
	}
	for _, s := range steps {
		args := map[string]any{
			"project": "test-app", "epic_id": epicID, "story_id": storyID, "task_id": taskID,
		}
		if s.evidence != "" {
			args["evidence"] = s.evidence
		}
		res, _ := advance.Handler(args)
		if res.IsError {
			t.Fatalf("advance error at %s: %s", s.to, res.Content[0].Text)
		}
		var data map[string]any
		json.Unmarshal([]byte(res.Content[0].Text), &data)
		if data["to"] != s.to {
			t.Errorf("advance to = %v, want %s", data["to"], s.to)
		}
	}
}

func TestAdvanceFromNonAdvanceable(t *testing.T) {
	ws := setupProject(t)
	epicTools := tools.Epic(ws)
	storyTools := tools.Story(ws)
	taskTools := tools.Task(ws)
	lcTools := tools.Lifecycle(ws)

	epicRes, _ := epicTools[1].Handler(map[string]any{"project": "test-app", "title": "E"})
	var epic map[string]any
	json.Unmarshal([]byte(epicRes.Content[0].Text), &epic)
	epicID := epic["id"].(string)

	storyRes, _ := storyTools[1].Handler(map[string]any{
		"project": "test-app", "epic_id": epicID, "title": "S", "user_story": "s",
	})
	var story map[string]any
	json.Unmarshal([]byte(storyRes.Content[0].Text), &story)
	storyID := story["id"].(string)

	taskRes, _ := taskTools[1].Handler(map[string]any{
		"project": "test-app", "epic_id": epicID, "story_id": storyID,
		"title": "T", "type": "task",
	})
	var task map[string]any
	json.Unmarshal([]byte(taskRes.Content[0].Text), &task)
	taskID := task["id"].(string)

	// Task is in backlog — cannot advance
	res, _ := lcTools[0].Handler(map[string]any{
		"project": "test-app", "epic_id": epicID, "story_id": storyID, "task_id": taskID,
	})
	if !res.IsError {
		t.Error("expected error when advancing from backlog")
	}
	if !strings.Contains(res.Content[0].Text, "cannot advance") {
		t.Errorf("error = %s", res.Content[0].Text)
	}
}

func TestRejectTask(t *testing.T) {
	ws, epicID, storyID, taskID := setupTaskInProgress(t)
	lcTools := tools.Lifecycle(ws)
	advance := lcTools[0]
	reject := lcTools[1]

	// Advance to in-review with evidence at gated transitions
	gatedAdvances := []map[string]any{
		{"project": "test-app", "epic_id": epicID, "story_id": storyID, "task_id": taskID, "evidence": "tests passed"}, // in-progress → ready-for-testing
		{"project": "test-app", "epic_id": epicID, "story_id": storyID, "task_id": taskID},                             // ready-for-testing → in-testing
		{"project": "test-app", "epic_id": epicID, "story_id": storyID, "task_id": taskID, "evidence": "coverage OK"},  // in-testing → ready-for-docs
		{"project": "test-app", "epic_id": epicID, "story_id": storyID, "task_id": taskID},                             // ready-for-docs → in-docs
		{"project": "test-app", "epic_id": epicID, "story_id": storyID, "task_id": taskID, "evidence": "docs written"}, // in-docs → documented
		{"project": "test-app", "epic_id": epicID, "story_id": storyID, "task_id": taskID},                             // documented → in-review
	}
	for _, args := range gatedAdvances {
		advance.Handler(args)
	}

	// Now reject from in-review
	res, err := reject.Handler(map[string]any{
		"project": "test-app", "epic_id": epicID, "story_id": storyID, "task_id": taskID,
		"reason": "Needs more tests",
	})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if res.IsError {
		t.Fatalf("returned error: %s", res.Content[0].Text)
	}
	var data map[string]any
	json.Unmarshal([]byte(res.Content[0].Text), &data)
	if data["rejected"] == nil {
		t.Error("expected rejected task in response")
	}
	if data["bug_created"] == nil {
		t.Error("expected bug_created in response")
	}
}

func TestRejectFromNonReview(t *testing.T) {
	ws, epicID, storyID, taskID := setupTaskInProgress(t)
	lcTools := tools.Lifecycle(ws)
	reject := lcTools[1]

	// Try to reject from in-progress (should fail)
	res, _ := reject.Handler(map[string]any{
		"project": "test-app", "epic_id": epicID, "story_id": storyID, "task_id": taskID,
	})
	if !res.IsError {
		t.Error("expected error when rejecting from in-progress")
	}
	if !strings.Contains(res.Content[0].Text, "cannot reject") {
		t.Errorf("error = %s", res.Content[0].Text)
	}
}
