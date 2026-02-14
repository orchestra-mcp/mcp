package tools

import (
	"fmt"
	"os"

	h "github.com/orchestra-mcp/mcp/src/helpers"
	t "github.com/orchestra-mcp/mcp/src/types"
)

// Prd returns all PRD management tools.
func Prd(ws string) []t.Tool {
	return []t.Tool{
		{Definition: t.ToolDefinition{Name: "start_prd_session", Description: "Start guided PRD creation", InputSchema: sp()}, Handler: func(a map[string]any) (*t.ToolResult, error) {
			slug := h.GetString(a, "project")
			if !h.FileExists(h.ProjectDir(ws, slug)) {
				return h.ErrorResult("project not found"), nil
			}
			s := &t.PrdSession{Slug: slug, ProjectName: slug, Status: "in_progress"}
			if err := savePrd(ws, s); err != nil {
				return h.ErrorResult(err.Error()), nil
			}
			return h.JSONResult(nextQ(s)), nil
		}},
		{Definition: t.ToolDefinition{Name: "answer_prd_question", Description: "Answer current PRD question", InputSchema: t.InputSchema{Type: "object", Properties: map[string]any{"project": map[string]any{"type": "string"}, "answer": map[string]any{"type": "string"}}, Required: []string{"project", "answer"}}}, Handler: func(a map[string]any) (*t.ToolResult, error) {
			s, err := loadPrd(ws, h.GetString(a, "project"))
			if err != nil {
				return h.ErrorResult(err.Error()), nil
			}
			q := prdQuestions[s.CurrentIndex]
			s.Answers = append(s.Answers, t.PrdAnswer{Question: q.Key, Answer: h.GetString(a, "answer")})
			return advancePrd(ws, s), nil
		}},
		{Definition: t.ToolDefinition{Name: "get_prd_session", Description: "Get PRD session state", InputSchema: sp()}, Handler: func(a map[string]any) (*t.ToolResult, error) {
			s, err := loadPrd(ws, h.GetString(a, "project"))
			if err != nil {
				return h.ErrorResult(err.Error()), nil
			}
			return h.JSONResult(s), nil
		}},
		{Definition: t.ToolDefinition{Name: "abandon_prd_session", Description: "Abandon PRD session", InputSchema: sp()}, Handler: func(a map[string]any) (*t.ToolResult, error) {
			os.Remove(prdFile(ws, h.GetString(a, "project")))
			return h.TextResult("abandoned"), nil
		}},
		{Definition: t.ToolDefinition{Name: "skip_prd_question", Description: "Skip optional PRD question", InputSchema: sp()}, Handler: func(a map[string]any) (*t.ToolResult, error) {
			s, err := loadPrd(ws, h.GetString(a, "project"))
			if err != nil {
				return h.ErrorResult(err.Error()), nil
			}
			if prdQuestions[s.CurrentIndex].Required {
				return h.ErrorResult("cannot skip required question"), nil
			}
			return advancePrd(ws, s), nil
		}},
		{Definition: t.ToolDefinition{Name: "back_prd_question", Description: "Go back to previous PRD question", InputSchema: sp()}, Handler: func(a map[string]any) (*t.ToolResult, error) {
			s, err := loadPrd(ws, h.GetString(a, "project"))
			if err != nil {
				return h.ErrorResult(err.Error()), nil
			}
			if s.CurrentIndex == 0 {
				return h.ErrorResult("at first question"), nil
			}
			s.CurrentIndex--
			prev := prdQuestions[s.CurrentIndex].Key
			if n := len(s.Answers); n > 0 && s.Answers[n-1].Question == prev {
				s.Answers = s.Answers[:n-1]
			}
			if err := savePrd(ws, s); err != nil {
				return h.ErrorResult(err.Error()), nil
			}
			return h.JSONResult(nextQ(s)), nil
		}},
		{Definition: t.ToolDefinition{Name: "preview_prd", Description: "Preview PRD markdown", InputSchema: sp()}, Handler: func(a map[string]any) (*t.ToolResult, error) {
			s, err := loadPrd(ws, h.GetString(a, "project"))
			if err != nil {
				return h.ErrorResult(err.Error()), nil
			}
			return h.TextResult(generatePrdMarkdown(s)), nil
		}},
		splitPrd(ws),
		listPrdPhases(ws),
	}
}

func splitPrd(ws string) t.Tool {
	return t.Tool{
		Definition: t.ToolDefinition{
			Name: "split_prd", Description: "Split completed PRD into numbered phases",
			InputSchema: t.InputSchema{Type: "object", Properties: map[string]any{
				"project": map[string]any{"type": "string", "description": "Project slug"},
				"phases":  map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Phase names in order"},
			}, Required: []string{"project", "phases"}},
		},
		Handler: func(a map[string]any) (*t.ToolResult, error) {
			slug := h.GetString(a, "project")
			parent, err := loadPrd(ws, slug)
			if err != nil {
				return h.ErrorResult(err.Error()), nil
			}
			if parent.Status != "complete" {
				return h.ErrorResult("PRD must be complete before splitting"), nil
			}
			rawPhases, ok := a["phases"].([]any)
			if !ok || len(rawPhases) < 2 {
				return h.ErrorResult("provide at least 2 phase names"), nil
			}
			var phases []string
			for _, p := range rawPhases {
				if s, ok := p.(string); ok && s != "" {
					phases = append(phases, s)
				}
			}
			if len(phases) < 2 {
				return h.ErrorResult("provide at least 2 phase names"), nil
			}
			parent.Phases = make([]string, len(phases))
			for i, name := range phases {
				phaseSlug := fmt.Sprintf("%s-phase-%d", slug, i+1)
				parent.Phases[i] = phaseSlug
				child := &t.PrdSession{
					Slug:        phaseSlug,
					ProjectName: fmt.Sprintf("%s — Phase %d: %s", parent.ProjectName, i+1, name),
					Status:      "pending",
					ParentSlug:  slug,
					Phase:       i + 1,
				}
				dir := h.ProjectDir(ws, phaseSlug)
				if err := os.MkdirAll(dir, 0o755); err != nil {
					return h.ErrorResult(err.Error()), nil
				}
				if err := savePrd(ws, child); err != nil {
					return h.ErrorResult(err.Error()), nil
				}
			}
			if err := savePrd(ws, parent); err != nil {
				return h.ErrorResult(err.Error()), nil
			}
			return h.JSONResult(map[string]any{"phases": parent.Phases, "count": len(phases)}), nil
		},
	}
}

func listPrdPhases(ws string) t.Tool {
	return t.Tool{
		Definition: t.ToolDefinition{
			Name: "list_prd_phases", Description: "List PRD phases for a project",
			InputSchema: sp(),
		},
		Handler: func(a map[string]any) (*t.ToolResult, error) {
			slug := h.GetString(a, "project")
			s, err := loadPrd(ws, slug)
			if err != nil {
				return h.ErrorResult(err.Error()), nil
			}
			if len(s.Phases) == 0 {
				return h.JSONResult(map[string]any{"phases": []any{}, "message": "no phases — use split_prd first"}), nil
			}
			type phaseInfo struct {
				Slug   string `json:"slug"`
				Name   string `json:"name"`
				Phase  int    `json:"phase"`
				Status string `json:"status"`
			}
			var phases []phaseInfo
			for _, ps := range s.Phases {
				child, err := loadPrd(ws, ps)
				if err != nil {
					phases = append(phases, phaseInfo{Slug: ps, Status: "missing"})
					continue
				}
				phases = append(phases, phaseInfo{Slug: ps, Name: child.ProjectName, Phase: child.Phase, Status: child.Status})
			}
			return h.JSONResult(phases), nil
		},
	}
}
