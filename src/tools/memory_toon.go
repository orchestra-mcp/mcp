package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	h "github.com/orchestra-mcp/mcp/src/helpers"
	"github.com/orchestra-mcp/mcp/src/toon"
	t "github.com/orchestra-mcp/mcp/src/types"
)

func logFallback(tool string, err error) {
	fmt.Fprintf(os.Stderr, "[memory] %s gRPC failed, TOON fallback: %v\n", tool, err)
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
	_ = toon.ParseFile(chunksPath(ws, slug), &idx) // ignore error: file may not exist yet
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
	_ = toon.ParseFile(indexPath, &idx) // ignore error: file may not exist yet
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
