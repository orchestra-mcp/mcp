package tools

import (
	"fmt"
	"path/filepath"
	"strings"

	h "github.com/orchestra-mcp/mcp/src/helpers"
	"github.com/orchestra-mcp/mcp/src/toon"
	t "github.com/orchestra-mcp/mcp/src/types"
)

// Prompts returns all built-in MCP prompts.
func Prompts(ws string) []t.Prompt {
	return []t.Prompt{
		createPrdPrompt(),
		reviewTaskPrompt(ws),
		planSprintPrompt(ws),
	}
}

func createPrdPrompt() t.Prompt {
	return t.Prompt{
		Definition: t.PromptDefinition{
			Name:        "create_prd",
			Title:       "Create PRD",
			Description: "Guided product requirements document creation",
			Arguments: []t.PromptArgument{
				{Name: "project_name", Description: "Name of the project", Required: true},
				{Name: "description", Description: "Brief project description"},
			},
		},
		Handler: func(args map[string]string) (string, []t.PromptMessage, error) {
			name := args["project_name"]
			desc := args["description"]
			if desc == "" {
				desc = "No description provided"
			}
			return "Create a PRD for " + name, []t.PromptMessage{{
				Role: "user",
				Content: t.ContentBlock{Type: "text", Text: fmt.Sprintf(
					"Create a Product Requirements Document for %q.\nDescription: %s\n\n"+
						"Include: 1) Problem statement 2) Target users 3) Core features "+
						"4) Success metrics 5) Technical constraints 6) Timeline", name, desc)},
			}}, nil
		},
	}
}

func reviewTaskPrompt(ws string) t.Prompt {
	return t.Prompt{
		Definition: t.PromptDefinition{
			Name:        "review_task",
			Title:       "Review Task",
			Description: "Generate a code review prompt for a specific task",
			Arguments: []t.PromptArgument{
				{Name: "project", Description: "Project slug", Required: true},
				{Name: "epic_id", Description: "Epic ID", Required: true},
				{Name: "story_id", Description: "Story ID", Required: true},
				{Name: "task_id", Description: "Task ID to review", Required: true},
			},
		},
		Handler: func(args map[string]string) (string, []t.PromptMessage, error) {
			p := filepath.Join(h.ProjectDir(ws, args["project"]),
				"epics", args["epic_id"], "stories", args["story_id"],
				"tasks", args["task_id"]+".toon")
			var task t.IssueData
			if err := toon.ParseFile(p, &task); err != nil {
				return "", nil, fmt.Errorf("task not found: %w", err)
			}
			return "Review: " + task.Title, []t.PromptMessage{{
				Role: "user",
				Content: t.ContentBlock{Type: "text", Text: fmt.Sprintf(
					"Review the implementation for task %s: %s\n\nDescription: %s\n\n"+
						"Check: code quality, error handling, test coverage, security, patterns.",
					task.ID, task.Title, task.Description)},
			}}, nil
		},
	}
}

func planSprintPrompt(ws string) t.Prompt {
	return t.Prompt{
		Definition: t.PromptDefinition{
			Name:        "plan_sprint",
			Title:       "Plan Sprint",
			Description: "Generate a sprint planning prompt with current backlog",
			Arguments: []t.PromptArgument{
				{Name: "project", Description: "Project slug", Required: true},
			},
		},
		Handler: func(args map[string]string) (string, []t.PromptMessage, error) {
			statusPath := filepath.Join(h.ProjectDir(ws, args["project"]), "project-status.toon")
			var ps t.ProjectStatus
			if err := toon.ParseFile(statusPath, &ps); err != nil {
				return "", nil, fmt.Errorf("project not found: %w", err)
			}
			var backlog []string
			for _, tk := range ps.Tasks {
				if tk.Status == "backlog" || tk.Status == "todo" {
					backlog = append(backlog, fmt.Sprintf("- [%s] %s (%s)", tk.ID, tk.Title, tk.Status))
				}
			}
			return "Sprint Planning: " + ps.Project, []t.PromptMessage{{
				Role: "user",
				Content: t.ContentBlock{Type: "text", Text: fmt.Sprintf(
					"Plan the next sprint for %s.\n\nBacklog items:\n%s\n\n"+
						"Prioritize by impact and dependencies. Group into a focused sprint.",
					ps.Project, strings.Join(backlog, "\n"))},
			}}, nil
		},
	}
}
