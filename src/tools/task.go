package tools

import (
	"fmt"
	"os"
	"path/filepath"

	h "github.com/orchestra-mcp/mcp/src/helpers"
	"github.com/orchestra-mcp/mcp/src/toon"
	t "github.com/orchestra-mcp/mcp/src/types"
)

// Task returns all task management tools.
func Task(ws string) []t.Tool {
	return []t.Tool{listTasks(ws), createTask(ws), getTask(ws), updateTask(ws), deleteTask(ws)}
}

func listTasks(ws string) t.Tool {
	return t.Tool{
		Definition: t.ToolDefinition{
			Name: "list_tasks", Description: "List tasks in a story",
			InputSchema: t.InputSchema{Type: "object", Properties: map[string]any{
				"project": map[string]any{"type": "string"}, "epic_id": map[string]any{"type": "string"},
				"story_id": map[string]any{"type": "string"},
			}, Required: []string{"project", "epic_id", "story_id"}},
		},
		Handler: func(args map[string]any) (*t.ToolResult, error) {
			dir := filepath.Join(h.ProjectDir(ws, h.GetString(args, "project")),
				"epics", h.GetString(args, "epic_id"), "stories", h.GetString(args, "story_id"), "tasks")
			entries, err := os.ReadDir(dir)
			if err != nil {
				if os.IsNotExist(err) {
					return h.JSONResult([]any{}), nil
				}
				return h.ErrorResult(err.Error()), nil
			}
			var tasks []t.IssueData
			for _, e := range entries {
				if e.IsDir() || filepath.Ext(e.Name()) != ".toon" {
					continue
				}
				var task t.IssueData
				if toon.ParseFile(filepath.Join(dir, e.Name()), &task) == nil {
					tasks = append(tasks, task)
				}
			}
			return h.JSONResult(tasks), nil
		},
	}
}

func createTask(ws string) t.Tool {
	return t.Tool{
		Definition: t.ToolDefinition{
			Name: "create_task", Description: "Create a task/bug/hotfix under a story",
			InputSchema: t.InputSchema{Type: "object", Properties: map[string]any{
				"project": map[string]any{"type": "string"}, "epic_id": map[string]any{"type": "string"},
				"story_id": map[string]any{"type": "string"}, "title": map[string]any{"type": "string"},
				"type":        map[string]any{"type": "string", "enum": []string{"task", "bug", "hotfix"}},
				"description": map[string]any{"type": "string"},
				"priority":    map[string]any{"type": "string"},
			}, Required: []string{"project", "epic_id", "story_id", "title", "type"}},
		},
		Handler: func(args map[string]any) (*t.ToolResult, error) {
			slug := h.GetString(args, "project")
			epicID := h.GetString(args, "epic_id")
			storyID := h.GetString(args, "story_id")
			projDir := h.ProjectDir(ws, slug)
			statusPath := filepath.Join(projDir, "project-status.toon")
			var ps t.ProjectStatus
			if err := toon.ParseFile(statusPath, &ps); err != nil {
				return h.ErrorResult(err.Error()), nil
			}
			key := h.DeriveKey(ps.Project)
			id := fmt.Sprintf("%s-%d", key, len(ps.Epics)+len(ps.Stories)+len(ps.Tasks)+1)
			tasksDir := filepath.Join(projDir, "epics", epicID, "stories", storyID, "tasks")
			if err := os.MkdirAll(tasksDir, 0o755); err != nil {
				return h.ErrorResult(err.Error()), nil
			}
			task := t.IssueData{
				ID: id, Type: h.GetString(args, "type"), Title: h.GetString(args, "title"),
				Status: "backlog", Description: h.GetString(args, "description"),
				Priority: h.GetString(args, "priority"), CreatedAt: h.Now(),
			}
			if err := toon.WriteFile(filepath.Join(tasksDir, id+".toon"), &task); err != nil {
				return h.ErrorResult(err.Error()), nil
			}
			storyPath := filepath.Join(projDir, "epics", epicID, "stories", storyID, "story.toon")
			_ = h.UpdateParentChildren(storyPath, "add", t.IssueChild{ID: id, Title: task.Title, Status: task.Status})
			h.UpdateProjectStatus(&ps, task)
			ps.UpdatedAt = h.Now()
			_ = toon.WriteFile(statusPath, &ps)
			return h.JSONResult(task), nil
		},
	}
}

func getTask(ws string) t.Tool {
	return t.Tool{
		Definition: t.ToolDefinition{
			Name: "get_task", Description: "Get task details",
			InputSchema: t.InputSchema{Type: "object", Properties: map[string]any{
				"project": map[string]any{"type": "string"}, "epic_id": map[string]any{"type": "string"},
				"story_id": map[string]any{"type": "string"}, "task_id": map[string]any{"type": "string"},
			}, Required: []string{"project", "epic_id", "story_id", "task_id"}},
		},
		Handler: func(args map[string]any) (*t.ToolResult, error) {
			p := filepath.Join(h.ProjectDir(ws, h.GetString(args, "project")),
				"epics", h.GetString(args, "epic_id"), "stories", h.GetString(args, "story_id"),
				"tasks", h.GetString(args, "task_id")+".toon")
			var task t.IssueData
			if err := toon.ParseFile(p, &task); err != nil {
				return h.ErrorResult(err.Error()), nil
			}
			return h.JSONResult(task), nil
		},
	}
}
