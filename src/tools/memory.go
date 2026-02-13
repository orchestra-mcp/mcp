package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/orchestra-mcp/mcp/src/engine"
	pb "github.com/orchestra-mcp/mcp/src/gen/memoryv1"
	h "github.com/orchestra-mcp/mcp/src/helpers"
	"github.com/orchestra-mcp/mcp/src/toon"
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

// --- TOON fallback implementations ---

func toonSaveMemory(ws, slug string, args map[string]any, tags []string) (*t.ToolResult, error) {
	dir := memoryDir(ws, slug)
	_ = os.MkdirAll(dir, 0o755)
	cp := chunksPath(ws, slug)
	var idx t.MemoryIndex
	_ = toon.ParseFile(cp, &idx)
	chunk := t.MemoryChunk{
		ID: fmt.Sprintf("mem-%d", len(idx.Chunks)+1), Project: slug,
		Source: h.GetString(args, "source"), SourceID: h.GetString(args, "source_id"),
		Summary: h.GetString(args, "summary"), Content: h.GetString(args, "content"),
		Tags: tags, CreatedAt: h.Now(),
	}
	idx.Chunks = append(idx.Chunks, chunk)
	if err := toon.WriteFile(cp, &idx); err != nil {
		return h.ErrorResult(err.Error()), nil
	}
	return h.JSONResult(chunk), nil
}

func toonSearchMemory(ws, slug, query string, limit int) (*t.ToolResult, error) {
	query = strings.ToLower(query)
	var idx t.MemoryIndex
	if err := toon.ParseFile(chunksPath(ws, slug), &idx); err != nil {
		return h.JSONResult([]any{}), nil
	}
	type scored struct {
		Chunk t.MemoryChunk `json:"chunk"`
		Score float64       `json:"score"`
	}
	var results []scored
	for _, c := range idx.Chunks {
		text := strings.ToLower(c.Summary + " " + c.Content + " " + strings.Join(c.Tags, " "))
		score := keywordScore(text, query)
		if score > 0 {
			results = append(results, scored{Chunk: c, Score: score})
		}
	}
	sort.Slice(results, func(i, j int) bool { return results[i].Score > results[j].Score })
	if len(results) > limit {
		results = results[:limit]
	}
	return h.JSONResult(results), nil
}

func toonGetContext(ws, slug, query string, limit int) (*t.ToolResult, error) {
	query = strings.ToLower(query)
	var idx t.MemoryIndex
	_ = toon.ParseFile(chunksPath(ws, slug), &idx)
	var sessions t.SessionIndex
	sp := filepath.Join(sessionsDir(ws, slug), "index.toon")
	_ = toon.ParseFile(sp, &sessions)

	type contextItem struct {
		Type    string  `json:"type"`
		Summary string  `json:"summary"`
		Content string  `json:"content"`
		Score   float64 `json:"score"`
	}
	var items []contextItem
	for _, c := range idx.Chunks {
		text := strings.ToLower(c.Summary + " " + c.Content + " " + strings.Join(c.Tags, " "))
		score := keywordScore(text, query)
		if score > 0 {
			items = append(items, contextItem{Type: "memory", Summary: c.Summary, Content: c.Content, Score: score})
		}
	}
	for _, s := range sessions.Sessions {
		text := strings.ToLower(s.Summary)
		score := keywordScore(text, query)
		if score > 0 {
			items = append(items, contextItem{Type: "session", Summary: s.Summary, Content: s.SessionID, Score: score})
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Score > items[j].Score })
	if len(items) > limit {
		items = items[:limit]
	}
	return h.JSONResult(items), nil
}

func toonSaveSession(ws, slug, sessionID, summary string, args map[string]any) (*t.ToolResult, error) {
	dir := sessionsDir(ws, slug)
	_ = os.MkdirAll(dir, 0o755)
	session := t.SessionLog{SessionID: sessionID, Project: slug, Summary: summary, StartedAt: h.Now()}
	if evts, ok := args["events"].([]any); ok {
		for _, e := range evts {
			if m, ok := e.(map[string]any); ok {
				session.Events = append(session.Events, t.SessionEvent{
					Type: fmt.Sprintf("%v", m["type"]), Summary: fmt.Sprintf("%v", m["summary"]),
					Timestamp: h.Now(),
				})
			}
		}
	}
	_ = toon.WriteFile(filepath.Join(dir, sessionID+".toon"), &session)
	indexPath := filepath.Join(dir, "index.toon")
	var idx t.SessionIndex
	_ = toon.ParseFile(indexPath, &idx)
	idx.Sessions = append(idx.Sessions, session)
	_ = toon.WriteFile(indexPath, &idx)
	return h.JSONResult(session), nil
}

func toonListSessions(ws, slug string, limit int) (*t.ToolResult, error) {
	indexPath := filepath.Join(sessionsDir(ws, slug), "index.toon")
	var idx t.SessionIndex
	if err := toon.ParseFile(indexPath, &idx); err != nil {
		return h.JSONResult([]any{}), nil
	}
	sessions := idx.Sessions
	if len(sessions) > limit {
		sessions = sessions[len(sessions)-limit:]
	}
	return h.JSONResult(sessions), nil
}

func toonGetSession(ws, slug, sessionID string) (*t.ToolResult, error) {
	sessionPath := filepath.Join(sessionsDir(ws, slug), sessionID+".toon")
	var session t.SessionLog
	if err := toon.ParseFile(sessionPath, &session); err != nil {
		return h.ErrorResult(err.Error()), nil
	}
	return h.JSONResult(session), nil
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

func logFallback(tool string, err error) {
	fmt.Fprintf(os.Stderr, "[memory] %s gRPC failed, TOON fallback: %v\n", tool, err)
}

// keywordScore returns a simple relevance score based on keyword matching.
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
