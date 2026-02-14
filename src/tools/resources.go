package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	h "github.com/orchestra-mcp/mcp/src/helpers"
	"github.com/orchestra-mcp/mcp/src/toon"
	t "github.com/orchestra-mcp/mcp/src/types"
)

// Resources returns all built-in MCP resources.
func Resources(ws string) []t.Resource {
	return []t.Resource{
		projectPrdResource(ws),
		projectStatusResource(ws),
		taskDetailResource(ws),
	}
}

func projectPrdResource(ws string) t.Resource {
	return t.Resource{
		Definition: t.ResourceDefinition{
			URI:         "toon://project/{slug}/prd",
			Name:        "project_prd",
			Title:       "Project PRD Document",
			Description: "The Product Requirements Document for a project",
			MimeType:    "text/markdown",
		},
		Handler: func(uri string) ([]t.ResourceContent, error) {
			slug := extractParam("toon://project/{slug}/prd", uri, "slug")
			data, err := os.ReadFile(filepath.Join(h.ProjectDir(ws, slug), "prd.md"))
			if err != nil {
				return nil, fmt.Errorf("prd not found for %s: %w", slug, err)
			}
			return []t.ResourceContent{{URI: uri, MimeType: "text/markdown", Text: string(data)}}, nil
		},
	}
}

func projectStatusResource(ws string) t.Resource {
	return t.Resource{
		Definition: t.ResourceDefinition{
			URI:         "toon://project/{slug}/status",
			Name:        "project_status",
			Title:       "Project Status",
			Description: "Current project status with epic/story/task summaries",
			MimeType:    "application/json",
		},
		Handler: func(uri string) ([]t.ResourceContent, error) {
			slug := extractParam("toon://project/{slug}/status", uri, "slug")
			var ps t.ProjectStatus
			p := filepath.Join(h.ProjectDir(ws, slug), "project-status.toon")
			if err := toon.ParseFile(p, &ps); err != nil {
				return nil, err
			}
			data, _ := json.MarshalIndent(ps, "", "  ")
			return []t.ResourceContent{{URI: uri, MimeType: "application/json", Text: string(data)}}, nil
		},
	}
}

func taskDetailResource(ws string) t.Resource {
	return t.Resource{
		Definition: t.ResourceDefinition{
			URI:         "toon://project/{slug}/task/{epicId}/{storyId}/{taskId}",
			Name:        "task_detail",
			Title:       "Task Detail",
			Description: "Full detail of a specific task",
			MimeType:    "application/json",
		},
		Handler: func(uri string) ([]t.ResourceContent, error) {
			pattern := "toon://project/{slug}/task/{epicId}/{storyId}/{taskId}"
			slug := extractParam(pattern, uri, "slug")
			epicID := extractParam(pattern, uri, "epicId")
			storyID := extractParam(pattern, uri, "storyId")
			taskID := extractParam(pattern, uri, "taskId")
			p := filepath.Join(h.ProjectDir(ws, slug),
				"epics", epicID, "stories", storyID, "tasks", taskID+".toon")
			var task t.IssueData
			if err := toon.ParseFile(p, &task); err != nil {
				return nil, err
			}
			data, _ := json.MarshalIndent(task, "", "  ")
			return []t.ResourceContent{{URI: uri, MimeType: "application/json", Text: string(data)}}, nil
		},
	}
}

// extractParam extracts a named {param} from a URI given the pattern.
func extractParam(pattern, uri, name string) string {
	pp := strings.Split(pattern, "/")
	up := strings.Split(uri, "/")
	for i, seg := range pp {
		if seg == "{"+name+"}" && i < len(up) {
			return up[i]
		}
	}
	return ""
}
