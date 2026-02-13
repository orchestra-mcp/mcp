package tools_test

import (
	"encoding/json"
	"testing"

	"github.com/orchestra-mcp/mcp/src/tools"
)

func TestSavePlan(t *testing.T) {
	ws := setupProject(t)
	artTools := tools.Artifacts(ws)
	save := artTools[0]

	res, err := save.Handler(map[string]any{
		"project": "test-app", "title": "Auth Strategy",
		"content": "# Auth\n\nUse JWT tokens", "issue_id": "TA-1",
	})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if res.IsError {
		t.Fatalf("returned error: %s", res.Content[0].Text)
	}
	var data map[string]any
	json.Unmarshal([]byte(res.Content[0].Text), &data)
	if data["file"] != "auth-strategy.md" {
		t.Errorf("file = %v", data["file"])
	}
}

func TestListPlans(t *testing.T) {
	ws := setupProject(t)
	artTools := tools.Artifacts(ws)

	// Empty first
	res, _ := artTools[1].Handler(map[string]any{"project": "test-app"})
	if res.Content[0].Text != "[]" {
		t.Errorf("expected empty, got %s", res.Content[0].Text)
	}

	// Save a plan
	artTools[0].Handler(map[string]any{
		"project": "test-app", "title": "DB Strategy",
		"content": "Use PostgreSQL", "issue_id": "TA-2",
	})

	res, _ = artTools[1].Handler(map[string]any{"project": "test-app"})
	var plans []map[string]any
	json.Unmarshal([]byte(res.Content[0].Text), &plans)
	if len(plans) != 1 {
		t.Fatalf("expected 1 plan, got %d", len(plans))
	}
	if plans[0]["title"] != "DB Strategy" {
		t.Errorf("title = %v", plans[0]["title"])
	}
	if plans[0]["issue_id"] != "TA-2" {
		t.Errorf("issue_id = %v", plans[0]["issue_id"])
	}
}

func TestUsageRecordAndGet(t *testing.T) {
	ws := setupProject(t)
	usageTools := tools.Usage(ws)
	getUsage := usageTools[0]
	record := usageTools[1]
	reset := usageTools[2]

	// Record usage
	res, _ := record.Handler(map[string]any{
		"provider": "anthropic", "model": "claude-opus-4-6",
		"input_tokens": 1000, "output_tokens": 500, "cost": 0.05,
	})
	if res.IsError {
		t.Fatalf("record error: %s", res.Content[0].Text)
	}
	var data map[string]any
	json.Unmarshal([]byte(res.Content[0].Text), &data)
	sessionInput, _ := data["session_input"].(float64)
	if sessionInput != 1000 {
		t.Errorf("session_input = %v", data["session_input"])
	}

	// Record more usage
	record.Handler(map[string]any{
		"input_tokens": 2000, "output_tokens": 1000, "cost": 0.10,
	})

	// Get usage
	res, _ = getUsage.Handler(map[string]any{})
	json.Unmarshal([]byte(res.Content[0].Text), &data)
	totals := data["totals"].(map[string]any)
	totalInput, _ := totals["total_input"].(float64)
	if totalInput != 3000 {
		t.Errorf("total_input = %v, want 3000", totals["total_input"])
	}

	// Reset session
	res, _ = reset.Handler(map[string]any{})
	if res.Content[0].Text != "session ended" {
		t.Errorf("expected 'session ended', got %s", res.Content[0].Text)
	}
}

func TestResetSessionNoOpen(t *testing.T) {
	ws := setupProject(t)
	usageTools := tools.Usage(ws)
	res, _ := usageTools[2].Handler(map[string]any{})
	if res.Content[0].Text != "no open session" {
		t.Errorf("expected 'no open session', got %s", res.Content[0].Text)
	}
}
