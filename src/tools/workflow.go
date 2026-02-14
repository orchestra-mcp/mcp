package tools

import (
	"fmt"
	"sort"

	h "github.com/orchestra-mcp/mcp/src/helpers"
	t "github.com/orchestra-mcp/mcp/src/types"
	"github.com/orchestra-mcp/mcp/src/workflow"
)

const (
	statusBacklog         = "backlog"
	statusTodo            = "todo"
	statusInProgress      = "in-progress"
	statusBlocked         = "blocked"
	statusReadyForTesting = "ready-for-testing"
	statusInTesting       = "in-testing"
	statusReadyForDocs    = "ready-for-docs"
	statusInDocs          = "in-docs"
	statusDocumented      = "documented"
	statusInReview        = "in-review"
	statusDone            = "done"
	statusRejected        = "rejected"
)

var (
	typePriority   = map[string]int{"hotfix": 0, "bug": 1, "task": 2}
	statusPriority = map[string]int{statusInProgress: 0, statusTodo: 1, statusBacklog: 2}
)

// Workflow returns all workflow management tools.
func Workflow(ws string) []t.Tool {
	return []t.Tool{
		getNextTask(ws), setCurrentTask(ws), completeTask(ws),
		searchIssues(ws), getWorkflowStatus(ws),
	}
}

func getNextTask(ws string) t.Tool {
	return t.Tool{
		Definition: t.ToolDefinition{
			Name: "get_next_task", Description: "Get highest priority actionable task",
			InputSchema: t.InputSchema{Type: "object", Properties: map[string]any{
				"project": map[string]any{"type": "string"},
			}, Required: []string{"project"}},
		},
		Handler: func(args map[string]any) (*t.ToolResult, error) {
			tasks := h.ScanAllTasks(ws, h.GetString(args, "project"))
			var actionable []h.ScannedTask
			for _, tk := range tasks {
				s := tk.Data.Status
				if s == statusInProgress || s == statusTodo || s == statusBacklog ||
					s == statusReadyForTesting || s == statusReadyForDocs ||
					s == statusDocumented {
					actionable = append(actionable, tk)
				}
			}
			if len(actionable) == 0 {
				return h.TextResult("no actionable tasks"), nil
			}
			sort.Slice(actionable, func(i, j int) bool {
				ti, tj := typePriority[actionable[i].Data.Type], typePriority[actionable[j].Data.Type]
				if ti != tj {
					return ti < tj
				}
				return statusPriority[actionable[i].Data.Status] < statusPriority[actionable[j].Data.Status]
			})
			return h.JSONResult(actionable[0].Data), nil
		},
	}
}

func searchIssues(ws string) t.Tool {
	return t.Tool{
		Definition: t.ToolDefinition{
			Name: "search", Description: "Search issues by text, optional type filter",
			InputSchema: t.InputSchema{Type: "object", Properties: map[string]any{
				"project": map[string]any{"type": "string"},
				"query":   map[string]any{"type": "string"},
				"type":    map[string]any{"type": "string", "enum": []string{"epic", "story", "task", "bug", "hotfix"}},
			}, Required: []string{"project", "query"}},
		},
		Handler: func(args map[string]any) (*t.ToolResult, error) {
			slug := h.GetString(args, "project")
			query := h.GetString(args, "query")
			typeFilter := h.GetString(args, "type")
			issues := h.ScanAllIssues(ws, slug)
			var matches []t.IssueData
			for _, iss := range issues {
				if typeFilter != "" && iss.Data.Type != typeFilter {
					continue
				}
				text := iss.Data.Title + " " + iss.Data.Description
				if containsCI(text, query) {
					matches = append(matches, iss.Data)
				}
			}
			return h.JSONResult(matches), nil
		},
	}
}

func getWorkflowStatus(ws string) t.Tool {
	return t.Tool{
		Definition: t.ToolDefinition{
			Name: "get_workflow_status", Description: "Get workflow stats: counts, blocked, completion %",
			InputSchema: t.InputSchema{Type: "object", Properties: map[string]any{
				"project": map[string]any{"type": "string"},
			}, Required: []string{"project"}},
		},
		Handler: func(args map[string]any) (*t.ToolResult, error) {
			tasks := h.ScanAllTasks(ws, h.GetString(args, "project"))
			byStatus := map[string]int{}
			byType := map[string]int{}
			var blocked, inProgress, ready []string
			var testing, documenting, reviewing []string
			total, done := 0, 0
			for _, tk := range tasks {
				total++
				byStatus[tk.Data.Status]++
				byType[tk.Data.Type]++
				if workflow.CompletedStatuses[tk.Data.Status] {
					done++
				}
				switch tk.Data.Status {
				case statusBlocked:
					blocked = append(blocked, tk.Data.ID)
				case statusInProgress:
					inProgress = append(inProgress, tk.Data.ID)
				case statusTodo:
					ready = append(ready, tk.Data.ID)
				case statusReadyForTesting, statusInTesting:
					testing = append(testing, tk.Data.ID)
				case statusReadyForDocs, statusInDocs, statusDocumented:
					documenting = append(documenting, tk.Data.ID)
				case statusInReview:
					reviewing = append(reviewing, tk.Data.ID)
				}
			}
			pct := 0.0
			if total > 0 {
				pct = float64(done) / float64(total) * 100
			}
			return h.JSONResult(map[string]any{
				"total": total, "done": done, "completion_pct": fmt.Sprintf("%.1f", pct),
				"by_status": byStatus, "by_type": byType,
				"blocked": blocked, "in_progress": inProgress, "ready": ready,
				"testing": testing, "documenting": documenting, "reviewing": reviewing,
			}), nil
		},
	}
}
