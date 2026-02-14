package tools

import (
	"fmt"
	"path/filepath"

	h "github.com/orchestra-mcp/mcp/src/helpers"
	"github.com/orchestra-mcp/mcp/src/toon"
	t "github.com/orchestra-mcp/mcp/src/types"
	"github.com/orchestra-mcp/mcp/src/workflow"
)

func setCurrentTask(ws string) t.Tool {
	return t.Tool{
		Definition: t.ToolDefinition{
			Name: "set_current_task", Description: "Set task to in-progress, cascade parents",
			InputSchema: t.InputSchema{Type: "object", Properties: map[string]any{
				"project": map[string]any{"type": "string"}, "epic_id": map[string]any{"type": "string"},
				"story_id": map[string]any{"type": "string"}, "task_id": map[string]any{"type": "string"},
			}, Required: []string{"project", "epic_id", "story_id", "task_id"}},
		},
		Handler: func(args map[string]any) (*t.ToolResult, error) {
			slug := h.GetString(args, "project")
			epicID := h.GetString(args, "epic_id")
			storyID := h.GetString(args, "story_id")
			taskID := h.GetString(args, "task_id")
			projDir := h.ProjectDir(ws, slug)
			taskPath := filepath.Join(projDir, "epics", epicID, "stories", storyID, "tasks", taskID+".toon")
			var task t.IssueData
			if err := toon.ParseFile(taskPath, &task); err != nil {
				return h.ErrorResult(err.Error()), nil
			}
			if !workflow.IsValid(task.Status, statusInProgress) {
				return h.ErrorResult(fmt.Sprintf("cannot transition %s -> in-progress from %s", taskID, task.Status)), nil
			}
			from := task.Status
			task.Status = statusInProgress
			task.UpdatedAt = h.Now()
			_ = toon.WriteFile(taskPath, &task)
			workflow.Emit(workflow.TransitionEvent{
				Project: slug, EpicID: epicID, StoryID: storyID, TaskID: taskID,
				Type: task.Type, From: from, To: statusInProgress, Time: task.UpdatedAt,
			})
			// Cascade to story
			storyPath := filepath.Join(projDir, "epics", epicID, "stories", storyID, "story.toon")
			var story t.IssueData
			if toon.ParseFile(storyPath, &story) == nil {
				if story.Status == statusBacklog || story.Status == statusTodo {
					story.Status = statusInProgress
					story.UpdatedAt = h.Now()
				}
				_ = h.UpdateParentChildren(storyPath, "update", t.IssueChild{ID: taskID, Title: task.Title, Status: task.Status})
			}
			// Cascade to epic
			epicPath := filepath.Join(projDir, "epics", epicID, "epic.toon")
			var epic t.IssueData
			if toon.ParseFile(epicPath, &epic) == nil {
				if epic.Status == statusBacklog || epic.Status == statusTodo {
					epic.Status = statusInProgress
					epic.UpdatedAt = h.Now()
					_ = toon.WriteFile(epicPath, &epic)
				}
				_ = h.UpdateParentChildren(epicPath, "update", t.IssueChild{ID: storyID, Title: story.Title, Status: story.Status})
			}
			// Update project status
			statusPath := filepath.Join(projDir, "project-status.toon")
			var ps t.ProjectStatus
			if toon.ParseFile(statusPath, &ps) == nil {
				h.UpdateProjectStatus(&ps, task)
				h.UpdateProjectStatus(&ps, story)
				h.UpdateProjectStatus(&ps, epic)
				ps.UpdatedAt = h.Now()
				_ = toon.WriteFile(statusPath, &ps)
			}
			return h.JSONResult(task), nil
		},
	}
}

func completeTask(ws string) t.Tool {
	return t.Tool{
		Definition: t.ToolDefinition{
			Name: "complete_task", Description: "Complete task, cascade done if all siblings done",
			InputSchema: t.InputSchema{Type: "object", Properties: map[string]any{
				"project": map[string]any{"type": "string"}, "epic_id": map[string]any{"type": "string"},
				"story_id": map[string]any{"type": "string"}, "task_id": map[string]any{"type": "string"},
			}, Required: []string{"project", "epic_id", "story_id", "task_id"}},
		},
		Handler: func(args map[string]any) (*t.ToolResult, error) {
			slug := h.GetString(args, "project")
			epicID := h.GetString(args, "epic_id")
			storyID := h.GetString(args, "story_id")
			taskID := h.GetString(args, "task_id")
			projDir := h.ProjectDir(ws, slug)
			taskPath := filepath.Join(projDir, "epics", epicID, "stories", storyID, "tasks", taskID+".toon")
			var task t.IssueData
			if err := toon.ParseFile(taskPath, &task); err != nil {
				return h.ErrorResult(err.Error()), nil
			}
			if !workflow.IsValid(task.Status, statusReadyForTesting) {
				return h.ErrorResult(fmt.Sprintf("cannot complete %s from %s (needs in-progress state)", taskID, task.Status)), nil
			}
			from := task.Status
			task.Status = statusReadyForTesting
			task.UpdatedAt = h.Now()
			_ = toon.WriteFile(taskPath, &task)
			workflow.Emit(workflow.TransitionEvent{
				Project: slug, EpicID: epicID, StoryID: storyID, TaskID: taskID,
				Type: task.Type, From: from, To: statusReadyForTesting, Time: task.UpdatedAt,
			})
			// Update story children
			storyPath := filepath.Join(projDir, "epics", epicID, "stories", storyID, "story.toon")
			var story t.IssueData
			if toon.ParseFile(storyPath, &story) == nil {
				_ = h.UpdateParentChildren(storyPath, "update", t.IssueChild{ID: taskID, Title: task.Title, Status: task.Status})
				_ = toon.ParseFile(storyPath, &story)
				if allChildrenDone(story.Children) {
					story.Status = statusDone
					story.UpdatedAt = h.Now()
					_ = toon.WriteFile(storyPath, &story)
				}
			}
			// Update epic children
			epicPath := filepath.Join(projDir, "epics", epicID, "epic.toon")
			var epic t.IssueData
			if toon.ParseFile(epicPath, &epic) == nil {
				_ = h.UpdateParentChildren(epicPath, "update", t.IssueChild{ID: storyID, Title: story.Title, Status: story.Status})
				_ = toon.ParseFile(epicPath, &epic)
				if allChildrenDone(epic.Children) {
					epic.Status = statusDone
					epic.UpdatedAt = h.Now()
					_ = toon.WriteFile(epicPath, &epic)
				}
			}
			statusPath := filepath.Join(projDir, "project-status.toon")
			var ps t.ProjectStatus
			if toon.ParseFile(statusPath, &ps) == nil {
				h.UpdateProjectStatus(&ps, task)
				h.UpdateProjectStatus(&ps, story)
				h.UpdateProjectStatus(&ps, epic)
				ps.UpdatedAt = h.Now()
				_ = toon.WriteFile(statusPath, &ps)
			}
			return h.JSONResult(task), nil
		},
	}
}

func allChildrenDone(children []t.IssueChild) bool {
	if len(children) == 0 {
		return false
	}
	for _, c := range children {
		if !workflow.CompletedStatuses[c.Status] {
			return false
		}
	}
	return true
}
