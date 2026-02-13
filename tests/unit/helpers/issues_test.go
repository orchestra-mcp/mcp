package helpers_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/orchestra-mcp/mcp/src/helpers"
	"github.com/orchestra-mcp/mcp/src/toon"
	"github.com/orchestra-mcp/mcp/src/types"
)

func TestUpdateProjectStatusAdd(t *testing.T) {
	ps := &types.ProjectStatus{Project: "test"}
	issue := types.IssueData{ID: "TA-1", Title: "Epic", Type: "epic", Status: "backlog"}
	helpers.UpdateProjectStatus(ps, issue)
	if len(ps.Tasks) != 1 {
		t.Errorf("expected 1 task entry, got %d", len(ps.Tasks))
	}
	if ps.Tasks[0].ID != "TA-1" {
		t.Errorf("id = %s", ps.Tasks[0].ID)
	}
}

func TestUpdateProjectStatusUpdate(t *testing.T) {
	ps := &types.ProjectStatus{
		Project: "test",
		Tasks:   []types.IssueEntry{{ID: "TA-1", Title: "Old", Status: "backlog"}},
	}
	issue := types.IssueData{ID: "TA-1", Title: "New", Type: "task", Status: "todo"}
	helpers.UpdateProjectStatus(ps, issue)
	if len(ps.Tasks) != 1 {
		t.Errorf("expected 1 entry, got %d", len(ps.Tasks))
	}
	if ps.Tasks[0].Title != "New" {
		t.Errorf("title = %s, want New", ps.Tasks[0].Title)
	}
	if ps.Tasks[0].Status != "todo" {
		t.Errorf("status = %s, want todo", ps.Tasks[0].Status)
	}
}

func TestRemoveEntry(t *testing.T) {
	entries := []types.IssueEntry{
		{ID: "A", Title: "a"}, {ID: "B", Title: "b"}, {ID: "C", Title: "c"},
	}
	result := helpers.RemoveEntry(entries, "B")
	if len(result) != 2 {
		t.Fatalf("expected 2, got %d", len(result))
	}
	for _, e := range result {
		if e.ID == "B" {
			t.Error("B should be removed")
		}
	}
}

func TestRemoveEntryNotFound(t *testing.T) {
	entries := []types.IssueEntry{{ID: "A"}}
	result := helpers.RemoveEntry(entries, "Z")
	if len(result) != 1 {
		t.Errorf("expected 1, got %d", len(result))
	}
}

func TestUpdateParentChildrenAdd(t *testing.T) {
	dir := t.TempDir()
	parentPath := filepath.Join(dir, "parent.toon")
	parent := types.IssueData{ID: "E-1", Type: "epic", Title: "Epic"}
	toon.WriteFile(parentPath, &parent)

	err := helpers.UpdateParentChildren(parentPath, "add", types.IssueChild{ID: "S-1", Title: "Story", Status: "backlog"})
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	var updated types.IssueData
	toon.ParseFile(parentPath, &updated)
	if len(updated.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(updated.Children))
	}
	if updated.Children[0].ID != "S-1" {
		t.Errorf("child id = %s", updated.Children[0].ID)
	}
}

func TestUpdateParentChildrenUpdate(t *testing.T) {
	dir := t.TempDir()
	parentPath := filepath.Join(dir, "parent.toon")
	parent := types.IssueData{
		ID: "E-1", Type: "epic", Title: "Epic",
		Children: []types.IssueChild{{ID: "S-1", Title: "Old", Status: "backlog"}},
	}
	toon.WriteFile(parentPath, &parent)

	helpers.UpdateParentChildren(parentPath, "update", types.IssueChild{ID: "S-1", Title: "New", Status: "in-progress"})

	var updated types.IssueData
	toon.ParseFile(parentPath, &updated)
	if updated.Children[0].Title != "New" {
		t.Errorf("title = %s", updated.Children[0].Title)
	}
	if updated.Children[0].Status != "in-progress" {
		t.Errorf("status = %s", updated.Children[0].Status)
	}
}

func TestUpdateParentChildrenRemove(t *testing.T) {
	dir := t.TempDir()
	parentPath := filepath.Join(dir, "parent.toon")
	parent := types.IssueData{
		ID: "E-1", Type: "epic",
		Children: []types.IssueChild{
			{ID: "S-1", Title: "A"}, {ID: "S-2", Title: "B"},
		},
	}
	toon.WriteFile(parentPath, &parent)

	helpers.UpdateParentChildren(parentPath, "remove", types.IssueChild{ID: "S-1"})

	var updated types.IssueData
	toon.ParseFile(parentPath, &updated)
	if len(updated.Children) != 1 {
		t.Fatalf("expected 1, got %d", len(updated.Children))
	}
	if updated.Children[0].ID != "S-2" {
		t.Errorf("remaining = %s", updated.Children[0].ID)
	}
}

func TestScanAllTasks(t *testing.T) {
	ws := t.TempDir()
	projDir := filepath.Join(ws, ".projects", "my-app")
	tasksDir := filepath.Join(projDir, "epics", "E-1", "stories", "S-1", "tasks")
	os.MkdirAll(tasksDir, 0o755)

	task := types.IssueData{ID: "T-1", Type: "task", Title: "Test task", Status: "todo"}
	toon.WriteFile(filepath.Join(tasksDir, "T-1.toon"), &task)

	tasks := helpers.ScanAllTasks(ws, "my-app")
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	if tasks[0].Data.ID != "T-1" {
		t.Errorf("id = %s", tasks[0].Data.ID)
	}
	if tasks[0].EpicID != "E-1" {
		t.Errorf("epicID = %s", tasks[0].EpicID)
	}
	if tasks[0].StoryID != "S-1" {
		t.Errorf("storyID = %s", tasks[0].StoryID)
	}
}

func TestScanAllIssues(t *testing.T) {
	ws := t.TempDir()
	projDir := filepath.Join(ws, ".projects", "my-app")

	// Create epic
	epicDir := filepath.Join(projDir, "epics", "E-1")
	os.MkdirAll(filepath.Join(epicDir, "stories"), 0o755)
	toon.WriteFile(filepath.Join(epicDir, "epic.toon"), &types.IssueData{
		ID: "E-1", Type: "epic", Title: "Epic",
	})

	// Create story
	storyDir := filepath.Join(epicDir, "stories", "S-1")
	os.MkdirAll(filepath.Join(storyDir, "tasks"), 0o755)
	toon.WriteFile(filepath.Join(storyDir, "story.toon"), &types.IssueData{
		ID: "S-1", Type: "story", Title: "Story",
	})

	// Create task
	toon.WriteFile(filepath.Join(storyDir, "tasks", "T-1.toon"), &types.IssueData{
		ID: "T-1", Type: "task", Title: "Task",
	})

	issues := helpers.ScanAllIssues(ws, "my-app")
	if len(issues) != 3 {
		t.Fatalf("expected 3 issues (epic+story+task), got %d", len(issues))
	}

	types := map[string]bool{}
	for _, i := range issues {
		types[i.Type] = true
	}
	if !types["epic"] || !types["story"] || !types["task"] {
		t.Errorf("missing types: %v", types)
	}
}

func TestScanAllTasksEmpty(t *testing.T) {
	ws := t.TempDir()
	tasks := helpers.ScanAllTasks(ws, "nonexistent")
	if len(tasks) != 0 {
		t.Errorf("expected 0 tasks, got %d", len(tasks))
	}
}
