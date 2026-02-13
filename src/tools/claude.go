package tools

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/orchestra-mcp/mcp/src/bootstrap"
	h "github.com/orchestra-mcp/mcp/src/helpers"
	"github.com/orchestra-mcp/mcp/src/toon"
	t "github.com/orchestra-mcp/mcp/src/types"
)

const maxHookEvents = 100

// Claude returns Claude Code awareness tools.
func Claude(ws string) []t.Tool {
	return []t.Tool{
		listSkills(ws), listAgents(ws),
		receiveHookEvent(ws), getHookEvents(ws),
		installSkills(ws), installAgents(ws), installDocsTool(ws),
	}
}

func listSkills(ws string) t.Tool {
	return t.Tool{
		Definition: t.ToolDefinition{
			Name: "list_skills", Description: "List available skills in the project",
			InputSchema: t.InputSchema{Type: "object", Properties: map[string]any{}, Required: []string{}},
		},
		Handler: func(args map[string]any) (*t.ToolResult, error) {
			skillsDir := filepath.Join(ws, ".claude", "skills")
			entries, err := os.ReadDir(skillsDir)
			if err != nil {
				if os.IsNotExist(err) {
					return h.JSONResult([]any{}), nil
				}
				return h.ErrorResult(err.Error()), nil
			}
			type skillInfo struct {
				Name        string `json:"name"`
				Description string `json:"description"`
			}
			var skills []skillInfo
			for _, e := range entries {
				if !e.IsDir() {
					continue
				}
				md := filepath.Join(skillsDir, e.Name(), "SKILL.md")
				desc := readFirstContentLine(md)
				skills = append(skills, skillInfo{Name: e.Name(), Description: desc})
			}
			return h.JSONResult(skills), nil
		},
	}
}

func listAgents(ws string) t.Tool {
	return t.Tool{
		Definition: t.ToolDefinition{
			Name: "list_agents", Description: "List available agents in the project",
			InputSchema: t.InputSchema{Type: "object", Properties: map[string]any{}, Required: []string{}},
		},
		Handler: func(args map[string]any) (*t.ToolResult, error) {
			agentsDir := filepath.Join(ws, ".claude", "agents")
			entries, err := os.ReadDir(agentsDir)
			if err != nil {
				if os.IsNotExist(err) {
					return h.JSONResult([]any{}), nil
				}
				return h.ErrorResult(err.Error()), nil
			}
			type agentInfo struct {
				Name        string `json:"name"`
				Description string `json:"description"`
			}
			var agents []agentInfo
			for _, e := range entries {
				if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
					continue
				}
				name := strings.TrimSuffix(e.Name(), ".md")
				desc := readFirstContentLine(filepath.Join(agentsDir, e.Name()))
				agents = append(agents, agentInfo{Name: name, Description: desc})
			}
			return h.JSONResult(agents), nil
		},
	}
}

func receiveHookEvent(ws string) t.Tool {
	return t.Tool{
		Definition: t.ToolDefinition{
			Name: "receive_hook_event", Description: "Receive a Claude Code hook event",
			InputSchema: t.InputSchema{Type: "object", Properties: map[string]any{
				"event_type": map[string]any{"type": "string"},
				"session_id": map[string]any{"type": "string"},
				"tool_name":  map[string]any{"type": "string"},
				"agent_type": map[string]any{"type": "string"},
				"data":       map[string]any{"type": "object"},
			}, Required: []string{"event_type"}},
		},
		Handler: func(args map[string]any) (*t.ToolResult, error) {
			event := t.HookEvent{
				EventType: h.GetString(args, "event_type"),
				SessionID: h.GetString(args, "session_id"),
				ToolName:  h.GetString(args, "tool_name"),
				AgentType: h.GetString(args, "agent_type"),
				Timestamp: h.Now(),
			}
			if d, ok := args["data"].(map[string]any); ok {
				event.Data = d
			}
			eventsDir := filepath.Join(ws, ".projects", ".events")
			_ = os.MkdirAll(eventsDir, 0o755)
			logPath := filepath.Join(eventsDir, "hook-events.toon")
			var log t.HookEventLog
			_ = toon.ParseFile(logPath, &log) // ignore if not exists
			log.Events = append(log.Events, event)
			if len(log.Events) > maxHookEvents {
				log.Events = log.Events[len(log.Events)-maxHookEvents:]
			}
			_ = toon.WriteFile(logPath, &log)
			return h.JSONResult(map[string]any{"stored": true, "event_type": event.EventType}), nil
		},
	}
}

func getHookEvents(ws string) t.Tool {
	return t.Tool{
		Definition: t.ToolDefinition{
			Name: "get_hook_events", Description: "Get recent Claude Code hook events",
			InputSchema: t.InputSchema{Type: "object", Properties: map[string]any{
				"event_type": map[string]any{"type": "string", "description": "Filter by event type"},
				"limit":      map[string]any{"type": "number", "description": "Max events to return"},
			}, Required: []string{}},
		},
		Handler: func(args map[string]any) (*t.ToolResult, error) {
			logPath := filepath.Join(ws, ".projects", ".events", "hook-events.toon")
			var log t.HookEventLog
			if toon.ParseFile(logPath, &log) != nil {
				return h.JSONResult([]any{}), nil
			}
			events := log.Events
			if typeFilter := h.GetString(args, "event_type"); typeFilter != "" {
				var filtered []t.HookEvent
				for _, e := range events {
					if e.EventType == typeFilter {
						filtered = append(filtered, e)
					}
				}
				events = filtered
			}
			limit := h.GetInt(args, "limit")
			if limit > 0 && len(events) > limit {
				events = events[len(events)-limit:]
			}
			return h.JSONResult(events), nil
		},
	}
}

func installSkills(ws string) t.Tool {
	return t.Tool{
		Definition: t.ToolDefinition{
			Name: "install_skills", Description: "Install bundled skills to project",
			InputSchema: t.InputSchema{Type: "object", Properties: map[string]any{
				"names": map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Skill names to install (empty = all)"},
			}, Required: []string{}},
		},
		Handler: func(args map[string]any) (*t.ToolResult, error) {
			target := filepath.Join(ws, ".claude", "skills")
			count, err := bootstrap.InstallSkills(target)
			if err != nil {
				return h.ErrorResult(err.Error()), nil
			}
			return h.JSONResult(map[string]any{
				"installed": count, "available": bootstrap.ListBundledSkills(),
			}), nil
		},
	}
}

func installAgents(ws string) t.Tool {
	return t.Tool{
		Definition: t.ToolDefinition{
			Name: "install_agents", Description: "Install bundled agents to project",
			InputSchema: t.InputSchema{Type: "object", Properties: map[string]any{
				"names": map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Agent names to install (empty = all)"},
			}, Required: []string{}},
		},
		Handler: func(args map[string]any) (*t.ToolResult, error) {
			target := filepath.Join(ws, ".claude", "agents")
			count, err := bootstrap.InstallAgents(target)
			if err != nil {
				return h.ErrorResult(err.Error()), nil
			}
			return h.JSONResult(map[string]any{
				"installed": count, "available": bootstrap.ListBundledAgents(),
			}), nil
		},
	}
}

func installDocsTool(ws string) t.Tool {
	return t.Tool{
		Definition: t.ToolDefinition{
			Name: "install_docs", Description: "Install CLAUDE.md, AGENTS.md, CONTEXT.md to project root",
			InputSchema: t.InputSchema{Type: "object", Properties: map[string]any{}, Required: []string{}},
		},
		Handler: func(args map[string]any) (*t.ToolResult, error) {
			count := bootstrap.InstallDocs(ws)
			return h.JSONResult(map[string]any{
				"installed": count,
				"files":     []string{"CLAUDE.md", "AGENTS.md", "CONTEXT.md"},
			}), nil
		},
	}
}

// readFirstContentLine reads the first non-empty, non-heading line from a markdown file.
func readFirstContentLine(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		return line
	}
	return ""
}
