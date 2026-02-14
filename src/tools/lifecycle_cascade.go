package tools

import (
	"fmt"
	"os"
	"path/filepath"

	h "github.com/orchestra-mcp/mcp/src/helpers"
	"github.com/orchestra-mcp/mcp/src/toon"
	t "github.com/orchestra-mcp/mcp/src/types"
)

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
