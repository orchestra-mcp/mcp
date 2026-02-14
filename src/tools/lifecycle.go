package tools

import (
	"fmt"
	"os"
	"path/filepath"

	h "github.com/orchestra-mcp/mcp/src/helpers"
	"github.com/orchestra-mcp/mcp/src/toon"
	t "github.com/orchestra-mcp/mcp/src/types"
	"github.com/orchestra-mcp/mcp/src/workflow"
)

// Lifecycle returns advance/reject lifecycle tools.
func Lifecycle(ws string) []t.Tool {
	return []t.Tool{advanceTask(ws), rejectTask(ws)}
}

// gateRequirements maps transitions that require evidence before advancing.
var gateRequirements = map[string]string{
	"in-progress": "tests_passed: Run tests and confirm all pass. Include test command and results summary.",
	"in-testing":  "tests_verified: Confirm coverage is acceptable and edge cases are covered.",
	"in-docs":     "docs_written: Describe what documentation was added or updated.",
	"in-review":   "review_approved: Describe what was reviewed and confirm code quality.",
}

// gateHints tells the agent what to do at each stage before the next advance.
var gateHints = map[string]string{
	"ready-for-testing": "ACTION REQUIRED: Run tests (use qa-go/qa-rust/qa-node agent). Provide test results as evidence when advancing.",
	"in-testing":        "ACTION REQUIRED: Verify test coverage and edge cases. Provide verification summary as evidence.",
	"ready-for-docs":    "ACTION REQUIRED: Write or update documentation for this task. Provide docs summary as evidence.",
	"in-docs":           "ACTION REQUIRED: Finalize documentation. Provide summary as evidence when advancing.",
	"documented":        "ACTION REQUIRED: Review code quality, patterns, security, error handling. Provide review findings as evidence.",
	"in-review":         "ACTION REQUIRED: Final review and approval. Provide approval summary as evidence.",
}

func advanceTask(ws string) t.Tool {
	return t.Tool{
		Definition: t.ToolDefinition{
			Name:        "advance_task",
			Description: "Advance task to next lifecycle stage. Gated transitions (from in-progress, in-testing, in-docs, in-review) require 'evidence' describing work done.",
			InputSchema: t.InputSchema{Type: "object", Properties: map[string]any{
				"project":  map[string]any{"type": "string"},
				"epic_id":  map[string]any{"type": "string"},
				"story_id": map[string]any{"type": "string"},
				"task_id":  map[string]any{"type": "string"},
				"evidence": map[string]any{"type": "string", "description": "Required for gated transitions. Describe tests run, docs written, or review findings."},
			}, Required: []string{"project", "epic_id", "story_id", "task_id"}},
		},
		Handler: func(args map[string]any) (*t.ToolResult, error) {
			slug := h.GetString(args, "project")
			epicID := h.GetString(args, "epic_id")
			storyID := h.GetString(args, "story_id")
			taskID := h.GetString(args, "task_id")
			evidence := h.GetString(args, "evidence")
			projDir := h.ProjectDir(ws, slug)
			taskPath := filepath.Join(projDir, "epics", epicID, "stories", storyID, "tasks", taskID+".toon")
			var task t.IssueData
			if err := toon.ParseFile(taskPath, &task); err != nil {
				return h.ErrorResult(err.Error()), nil
			}
			next, ok := workflow.AdvanceMap[task.Status]
			if !ok {
				return h.ErrorResult(fmt.Sprintf("cannot advance %s from %s", taskID, task.Status)), nil
			}
			// Enforce evidence gates for critical transitions.
			if gate, gated := gateRequirements[task.Status]; gated && evidence == "" {
				return h.ErrorResult(fmt.Sprintf(
					"GATE BLOCKED: Cannot advance %s from '%s' without evidence.\nRequired: %s\nProvide 'evidence' parameter describing work done at this stage.",
					taskID, task.Status, gate,
				)), nil
			}
			from := task.Status
			task.Status = next
			task.UpdatedAt = h.Now()
			_ = toon.WriteFile(taskPath, &task)
			workflow.Emit(workflow.TransitionEvent{
				Project: slug, EpicID: epicID, StoryID: storyID, TaskID: taskID,
				Type: task.Type, From: from, To: next, Time: task.UpdatedAt,
			})
			cascadeParents(projDir, epicID, storyID, taskID, task)
			result := map[string]any{"task": task, "from": from, "to": next}
			if evidence != "" {
				result["evidence"] = evidence
			}
			if hint, ok := gateHints[next]; ok {
				result["gate"] = hint
			}
			return h.JSONResult(result), nil
		},
	}
}

func rejectTask(ws string) t.Tool {
	return t.Tool{
		Definition: t.ToolDefinition{
			Name: "reject_task", Description: "Reject task from review, auto-creates bug",
			InputSchema: t.InputSchema{Type: "object", Properties: map[string]any{
				"project": map[string]any{"type": "string"}, "epic_id": map[string]any{"type": "string"},
				"story_id": map[string]any{"type": "string"}, "task_id": map[string]any{"type": "string"},
				"reason": map[string]any{"type": "string", "description": "Rejection reason"},
			}, Required: []string{"project", "epic_id", "story_id", "task_id"}},
		},
		Handler: func(args map[string]any) (*t.ToolResult, error) {
			slug := h.GetString(args, "project")
			epicID := h.GetString(args, "epic_id")
			storyID := h.GetString(args, "story_id")
			taskID := h.GetString(args, "task_id")
			reason := h.GetString(args, "reason")
			projDir := h.ProjectDir(ws, slug)
			taskPath := filepath.Join(projDir, "epics", epicID, "stories", storyID, "tasks", taskID+".toon")
			var task t.IssueData
			if err := toon.ParseFile(taskPath, &task); err != nil {
				return h.ErrorResult(err.Error()), nil
			}
			if !workflow.IsValid(task.Status, statusRejected) {
				return h.ErrorResult(fmt.Sprintf("cannot reject %s from %s (must be in-review)", taskID, task.Status)), nil
			}
			task.Status = statusRejected
			task.UpdatedAt = h.Now()
			_ = toon.WriteFile(taskPath, &task)
			workflow.Emit(workflow.TransitionEvent{
				Project: slug, EpicID: epicID, StoryID: storyID, TaskID: taskID,
				Type: task.Type, From: "in-review", To: statusRejected, Time: task.UpdatedAt,
			})
			// Auto-create bug under same story
			bug, err := createRejectionBug(projDir, epicID, storyID, task, reason)
			if err != nil {
				return h.ErrorResult(err.Error()), nil
			}
			cascadeParents(projDir, epicID, storyID, taskID, task)
			return h.JSONResult(map[string]any{
				"rejected": task, "bug_created": bug,
			}), nil
		},
	}
}

func createRejectionBug(projDir, epicID, storyID string, task t.IssueData, reason string) (t.IssueData, error) {
	statusPath := filepath.Join(projDir, "project-status.toon")
	var ps t.ProjectStatus
	if err := toon.ParseFile(statusPath, &ps); err != nil {
		return t.IssueData{}, err
	}
	key := h.DeriveKey(ps.Project)
	id := fmt.Sprintf("%s-%d", key, len(ps.Epics)+len(ps.Stories)+len(ps.Tasks)+1)
	desc := fmt.Sprintf("Rejected from %s: %s", task.ID, task.Title)
	if reason != "" {
		desc += "\n\nReason: " + reason
	}
	bug := t.IssueData{
		ID: id, Type: "bug", Title: "Fix: " + task.Title,
		Status: statusBacklog, Description: desc,
		Priority: "high", CreatedAt: h.Now(),
	}
	tasksDir := filepath.Join(projDir, "epics", epicID, "stories", storyID, "tasks")
	_ = os.MkdirAll(tasksDir, 0o755)
	if err := toon.WriteFile(filepath.Join(tasksDir, id+".toon"), &bug); err != nil {
		return t.IssueData{}, err
	}
	storyPath := filepath.Join(projDir, "epics", epicID, "stories", storyID, "story.toon")
	_ = h.UpdateParentChildren(storyPath, "add", t.IssueChild{ID: id, Title: bug.Title, Status: bug.Status})
	h.UpdateProjectStatus(&ps, bug)
	ps.UpdatedAt = h.Now()
	_ = toon.WriteFile(statusPath, &ps)
	return bug, nil
}

func cascadeParents(projDir, epicID, storyID, taskID string, task t.IssueData) {
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
}
