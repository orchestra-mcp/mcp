package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	h "github.com/orchestra-mcp/mcp/src/helpers"
	"github.com/orchestra-mcp/mcp/src/toon"
	t "github.com/orchestra-mcp/mcp/src/types"
)

var prdQuestions = []t.PrdQuestion{
	{Index: 0, Key: "project_name", Section: "overview", Question: "What is the project name?", Required: true},
	{Index: 1, Key: "project_description", Section: "overview", Question: "Describe the project.", Required: true},
	{Index: 2, Key: "target_audience", Section: "overview", Question: "Who is the target audience?", Required: true, Options: []string{"Developers", "End users", "Enterprise teams", "Internal team"}},
	{Index: 3, Key: "primary_goals", Section: "goals", Question: "What are the primary goals?", Required: true},
	{Index: 4, Key: "success_metrics", Section: "goals", Question: "How will success be measured?", Required: false, Options: []string{"User adoption rate", "Performance benchmarks", "Revenue targets", "User satisfaction score"}},
	{Index: 5, Key: "functional_requirements", Section: "requirements", Question: "Functional requirements?", Required: true},
	{Index: 6, Key: "non_functional_requirements", Section: "requirements", Question: "Non-functional requirements?", Required: false, Options: []string{"High availability (99.9%)", "Sub-100ms latency", "GDPR compliance", "Offline support"}},
	{Index: 7, Key: "constraints", Section: "requirements", Question: "Constraints or limitations?", Required: false},
	{Index: 8, Key: "tech_stack", Section: "technical", Question: "Tech stack?", Required: false, Options: []string{"Go + React", "Python + React", "Node.js + React", "Rust + React"}},
	{Index: 9, Key: "integrations", Section: "technical", Question: "Third-party integrations?", Required: false},
	{Index: 10, Key: "milestones", Section: "timeline", Question: "Key milestones?", Required: false},
	{Index: 11, Key: "deadline", Section: "timeline", Question: "Target deadline?", Required: false, Options: []string{"1 month", "3 months", "6 months", "1 year"}},
}

var sectionTitles = map[string]string{
	"overview": "Overview", "goals": "Goals", "requirements": "Requirements",
	"technical": "Technical", "timeline": "Timeline",
}

var questionLabels = map[string]string{
	"project_name": "Project Name", "project_description": "Description",
	"target_audience": "Target Audience", "primary_goals": "Primary Goals",
	"success_metrics": "Success Metrics", "functional_requirements": "Functional Requirements",
	"non_functional_requirements": "Non-Functional Requirements", "constraints": "Constraints",
	"tech_stack": "Tech Stack", "integrations": "Integrations",
	"milestones": "Milestones", "deadline": "Deadline",
}

func prdFile(ws, slug string) string {
	return filepath.Join(h.ProjectDir(ws, slug), "prd-session.toon")
}

func loadPrd(ws, slug string) (*t.PrdSession, error) {
	var s t.PrdSession
	return &s, toon.ParseFile(prdFile(ws, slug), &s)
}

func savePrd(ws string, s *t.PrdSession) error { return toon.WriteFile(prdFile(ws, s.Slug), s) }

func generatePrdMarkdown(s *t.PrdSession) string {
	ans := map[string]string{}
	for _, a := range s.Answers {
		ans[a.Question] = a.Answer
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("# %s\n\n", s.ProjectName))
	sec := ""
	for _, q := range prdQuestions {
		v := ans[q.Key]
		if v == "" {
			continue
		}
		if q.Section != sec {
			sec = q.Section
			b.WriteString(fmt.Sprintf("## %s\n\n", sectionTitles[sec]))
		}
		b.WriteString(fmt.Sprintf("### %s\n\n%s\n\n", questionLabels[q.Key], v))
	}
	return b.String()
}

func nextQ(s *t.PrdSession) map[string]any {
	if s.CurrentIndex >= len(prdQuestions) {
		return map[string]any{"status": "complete"}
	}
	q := prdQuestions[s.CurrentIndex]
	r := map[string]any{"status": "in_progress", "question": q.Question, "key": q.Key, "index": s.CurrentIndex, "required": q.Required}
	if len(q.Options) > 0 {
		r["options"] = q.Options
	}
	return r
}

func finishPrd(ws string, s *t.PrdSession) *t.ToolResult {
	s.Status = "complete"
	if err := os.WriteFile(filepath.Join(h.ProjectDir(ws, s.Slug), "prd.md"), []byte(generatePrdMarkdown(s)), 0o644); err != nil {
		return h.ErrorResult(err.Error())
	}
	if err := savePrd(ws, s); err != nil {
		return h.ErrorResult(err.Error())
	}
	return h.JSONResult(map[string]any{"status": "complete", "file": "prd.md"})
}

func advancePrd(ws string, s *t.PrdSession) *t.ToolResult {
	s.CurrentIndex++
	if s.CurrentIndex >= len(prdQuestions) {
		return finishPrd(ws, s)
	}
	if err := savePrd(ws, s); err != nil {
		return h.ErrorResult(err.Error())
	}
	return h.JSONResult(nextQ(s))
}

func sp() t.InputSchema {
	return t.InputSchema{Type: "object", Properties: map[string]any{
		"project": map[string]any{"type": "string", "description": "Project slug"},
	}, Required: []string{"project"}}
}
