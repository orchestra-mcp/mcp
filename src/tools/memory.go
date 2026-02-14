package tools

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/orchestra-mcp/mcp/src/engine"
	pb "github.com/orchestra-mcp/mcp/src/gen/memoryv1"
	h "github.com/orchestra-mcp/mcp/src/helpers"
	t "github.com/orchestra-mcp/mcp/src/types"
)

// Memory returns all project memory/RAG tools.
func Memory(ws string, bridge *engine.Bridge) []t.Tool {
	return []t.Tool{
		saveMemory(ws, bridge), searchMemory(ws, bridge), getContext(ws, bridge),
		saveSession(ws, bridge), listSessions(ws, bridge), getSession(ws, bridge),
	}
}

func memoryDir(ws, slug string) string {
	return filepath.Join(h.ProjectDir(ws, slug), ".memory")
}

func chunksPath(ws, slug string) string {
	return filepath.Join(memoryDir(ws, slug), "chunks.toon")
}

func sessionsDir(ws, slug string) string {
	return filepath.Join(memoryDir(ws, slug), "sessions")
}

func saveMemory(ws string, bridge *engine.Bridge) t.Tool {
	return t.Tool{
		Definition: t.ToolDefinition{
			Name: "save_memory", Description: "Save a context chunk to project memory",
			InputSchema: t.InputSchema{Type: "object", Properties: map[string]any{
				"project":   map[string]any{"type": "string"},
				"content":   map[string]any{"type": "string", "description": "Content to remember"},
				"summary":   map[string]any{"type": "string", "description": "Short summary"},
				"source":    map[string]any{"type": "string", "description": "Source type: task, prd, session, user"},
				"source_id": map[string]any{"type": "string", "description": "Source ID (task ID, session ID)"},
				"tags":      map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			}, Required: []string{"project", "content", "summary"}},
		},
		Handler: func(args map[string]any) (*t.ToolResult, error) {
			slug := h.GetString(args, "project")
			tags := extractTags(args)

			if bridge.UsingEngine() {
				resp, err := bridge.Client.StoreChunk(slug, h.GetString(args, "source"),
					h.GetString(args, "source_id"), h.GetString(args, "summary"),
					h.GetString(args, "content"), tags)
				if err == nil {
					return h.JSONResult(resp.Chunk), nil
				}
				logFallback("save_memory", err)
			}
			return toonSaveMemory(ws, slug, args, tags)
		},
	}
}

func searchMemory(ws string, bridge *engine.Bridge) t.Tool {
	return t.Tool{
		Definition: t.ToolDefinition{
			Name: "search_memory", Description: "Search project memory by keyword",
			InputSchema: t.InputSchema{Type: "object", Properties: map[string]any{
				"project": map[string]any{"type": "string"},
				"query":   map[string]any{"type": "string"},
				"limit":   map[string]any{"type": "number"},
			}, Required: []string{"project", "query"}},
		},
		Handler: func(args map[string]any) (*t.ToolResult, error) {
			slug := h.GetString(args, "project")
			query := h.GetString(args, "query")
			limit := h.GetInt(args, "limit")
			if limit <= 0 {
				limit = 10
			}

			if bridge.UsingEngine() {
				resp, err := bridge.Client.SearchMemory(slug, query, int32(limit))
				if err == nil {
					return h.JSONResult(resp.Results), nil
				}
				logFallback("search_memory", err)
			}
			return toonSearchMemory(ws, slug, query, limit)
		},
	}
}

func getContext(ws string, bridge *engine.Bridge) t.Tool {
	return t.Tool{
		Definition: t.ToolDefinition{
			Name: "get_context", Description: "Get relevant context for current work",
			InputSchema: t.InputSchema{Type: "object", Properties: map[string]any{
				"project": map[string]any{"type": "string"},
				"query":   map[string]any{"type": "string", "description": "What context do you need?"},
				"limit":   map[string]any{"type": "number"},
			}, Required: []string{"project", "query"}},
		},
		Handler: func(args map[string]any) (*t.ToolResult, error) {
			slug := h.GetString(args, "project")
			query := h.GetString(args, "query")
			limit := h.GetInt(args, "limit")
			if limit <= 0 {
				limit = 5
			}

			if bridge.UsingEngine() {
				resp, err := bridge.Client.GetContext(slug, query, int32(limit))
				if err == nil {
					return h.JSONResult(resp.Chunks), nil
				}
				logFallback("get_context", err)
			}
			return toonGetContext(ws, slug, query, limit)
		},
	}
}

func saveSession(ws string, bridge *engine.Bridge) t.Tool {
	return t.Tool{
		Definition: t.ToolDefinition{
			Name: "save_session", Description: "Save a session summary to project memory",
			InputSchema: t.InputSchema{Type: "object", Properties: map[string]any{
				"project":    map[string]any{"type": "string"},
				"session_id": map[string]any{"type": "string"},
				"summary":    map[string]any{"type": "string"},
				"events":     map[string]any{"type": "array", "items": map[string]any{"type": "object"}},
			}, Required: []string{"project", "session_id", "summary"}},
		},
		Handler: func(args map[string]any) (*t.ToolResult, error) {
			slug := h.GetString(args, "project")
			sessionID := h.GetString(args, "session_id")
			summary := h.GetString(args, "summary")

			if bridge.UsingEngine() {
				var events []*pb.SessionEvent
				if evts, ok := args["events"].([]any); ok {
					for _, e := range evts {
						if m, ok := e.(map[string]any); ok {
							events = append(events, &pb.SessionEvent{
								Type:      fmt.Sprintf("%v", m["type"]),
								Summary:   fmt.Sprintf("%v", m["summary"]),
								Timestamp: h.Now(),
							})
						}
					}
				}
				resp, err := bridge.Client.StoreSession(slug, sessionID, summary, events)
				if err == nil {
					return h.JSONResult(resp.Session), nil
				}
				logFallback("save_session", err)
			}
			return toonSaveSession(ws, slug, sessionID, summary, args)
		},
	}
}

func listSessions(ws string, bridge *engine.Bridge) t.Tool {
	return t.Tool{
		Definition: t.ToolDefinition{
			Name: "list_sessions", Description: "List recent sessions for a project",
			InputSchema: t.InputSchema{Type: "object", Properties: map[string]any{
				"project": map[string]any{"type": "string"},
				"limit":   map[string]any{"type": "number"},
			}, Required: []string{"project"}},
		},
		Handler: func(args map[string]any) (*t.ToolResult, error) {
			slug := h.GetString(args, "project")
			limit := h.GetInt(args, "limit")
			if limit <= 0 {
				limit = 20
			}

			if bridge.UsingEngine() {
				resp, err := bridge.Client.ListSessions(slug, int32(limit))
				if err == nil {
					return h.JSONResult(resp.Sessions), nil
				}
				logFallback("list_sessions", err)
			}
			return toonListSessions(ws, slug, limit)
		},
	}
}

func getSession(ws string, bridge *engine.Bridge) t.Tool {
	return t.Tool{
		Definition: t.ToolDefinition{
			Name: "get_session", Description: "Get full session details",
			InputSchema: t.InputSchema{Type: "object", Properties: map[string]any{
				"project":    map[string]any{"type": "string"},
				"session_id": map[string]any{"type": "string"},
			}, Required: []string{"project", "session_id"}},
		},
		Handler: func(args map[string]any) (*t.ToolResult, error) {
			slug := h.GetString(args, "project")
			sessionID := h.GetString(args, "session_id")

			if bridge.UsingEngine() {
				resp, err := bridge.Client.GetSession(slug, sessionID)
				if err == nil {
					return h.JSONResult(resp.Session), nil
				}
				logFallback("get_session", err)
			}
			return toonGetSession(ws, slug, sessionID)
		},
	}
}

// --- Helpers ---

func extractTags(args map[string]any) []string {
	var tags []string
	if raw, ok := args["tags"].([]any); ok {
		for _, tag := range raw {
			if s, ok := tag.(string); ok {
				tags = append(tags, s)
			}
		}
	}
	return tags
}

// keywordScore returns a relevance score based on keyword matching.
func keywordScore(text, query string) float64 {
	words := strings.Fields(query)
	if len(words) == 0 {
		return 0
	}
	matches := 0
	for _, w := range words {
		if strings.Contains(text, w) {
			matches++
		}
	}
	return float64(matches) / float64(len(words))
}
