package workflow_test

import (
	"testing"

	"github.com/orchestra-mcp/mcp/src/workflow"
)

func TestIsValid(t *testing.T) {
	tests := []struct {
		from, to string
		want     bool
	}{
		// Happy-path forward transitions
		{"backlog", "todo", true},
		{"todo", "in-progress", true},
		{"in-progress", "ready-for-testing", true},
		{"ready-for-testing", "in-testing", true},
		{"in-testing", "ready-for-docs", true},
		{"ready-for-docs", "in-docs", true},
		{"in-docs", "documented", true},
		{"documented", "in-review", true},
		{"in-review", "done", true},
		// Backward transitions
		{"todo", "backlog", true},
		{"in-progress", "todo", true},
		{"in-progress", "blocked", true},
		{"blocked", "in-progress", true},
		{"blocked", "todo", true},
		{"ready-for-testing", "in-progress", true},
		{"in-testing", "in-progress", true},
		{"ready-for-docs", "in-testing", true},
		{"in-docs", "ready-for-docs", true},
		{"in-review", "documented", true},
		// Special transitions
		{"in-review", "rejected", true},
		{"done", "todo", true},
		{"rejected", "todo", true},
		{"rejected", "backlog", true},
		{"cancelled", "backlog", true},
		// Invalid transitions
		{"backlog", "done", false},
		{"todo", "done", false},
		{"backlog", "in-progress", false},
		{"done", "backlog", false},
		{"cancelled", "done", false},
		{"unknown", "todo", false},
		{"backlog", "unknown", false},
		{"", "todo", false},
		{"todo", "", false},
	}
	for _, tc := range tests {
		got := workflow.IsValid(tc.from, tc.to)
		if got != tc.want {
			t.Errorf("IsValid(%q,%q) = %v, want %v", tc.from, tc.to, got, tc.want)
		}
	}
}

func TestNextStates(t *testing.T) {
	got := workflow.NextStates("todo")
	if len(got) != 2 {
		t.Fatalf("NextStates(todo) len = %d, want 2", len(got))
	}
	if got[0] != "in-progress" || got[1] != "backlog" {
		t.Errorf("NextStates(todo) = %v", got)
	}
	if workflow.NextStates("nonexistent") != nil {
		t.Error("expected nil for unknown state")
	}
}

func TestNextStatesAllStates(t *testing.T) {
	for _, s := range workflow.AllStatuses {
		got := workflow.NextStates(s)
		if got == nil {
			t.Errorf("NextStates(%q) returned nil", s)
		}
		if len(got) == 0 {
			t.Errorf("NextStates(%q) returned empty", s)
		}
	}
}

func TestCompletedStatuses(t *testing.T) {
	if !workflow.CompletedStatuses["done"] || !workflow.CompletedStatuses["cancelled"] {
		t.Error("done and cancelled should be completed")
	}
	if !workflow.CompletedStatuses["rejected"] {
		t.Error("rejected should be completed")
	}
	if workflow.CompletedStatuses["in-progress"] {
		t.Error("in-progress should not be completed")
	}
	if workflow.CompletedStatuses["backlog"] {
		t.Error("backlog should not be completed")
	}
}

func TestDoneStatuses(t *testing.T) {
	if !workflow.DoneStatuses["done"] {
		t.Error("done should be in DoneStatuses")
	}
	if workflow.DoneStatuses["cancelled"] {
		t.Error("cancelled should not be in DoneStatuses")
	}
	if workflow.DoneStatuses["rejected"] {
		t.Error("rejected should not be in DoneStatuses")
	}
}

func TestActiveStatuses(t *testing.T) {
	for _, s := range []string{"in-progress", "in-testing", "in-docs", "in-review"} {
		if !workflow.ActiveStatuses[s] {
			t.Errorf("%s should be active", s)
		}
	}
	if workflow.ActiveStatuses["todo"] {
		t.Error("todo should not be active")
	}
}

func TestWaitingStatuses(t *testing.T) {
	for _, s := range []string{"ready-for-testing", "ready-for-docs", "documented"} {
		if !workflow.WaitingStatuses[s] {
			t.Errorf("%s should be waiting", s)
		}
	}
	if workflow.WaitingStatuses["in-progress"] {
		t.Error("in-progress should not be waiting")
	}
}

func TestAdvanceMapHappyPath(t *testing.T) {
	chain := []string{"in-progress", "ready-for-testing", "in-testing", "ready-for-docs", "in-docs", "documented", "in-review", "done"}
	for i := 0; i < len(chain)-1; i++ {
		next, ok := workflow.AdvanceMap[chain[i]]
		if !ok {
			t.Fatalf("AdvanceMap[%q] missing", chain[i])
		}
		if next != chain[i+1] {
			t.Errorf("AdvanceMap[%q] = %q, want %q", chain[i], next, chain[i+1])
		}
	}
	if _, ok := workflow.AdvanceMap["done"]; ok {
		t.Error("done should not be in AdvanceMap")
	}
	if _, ok := workflow.AdvanceMap["backlog"]; ok {
		t.Error("backlog should not be in AdvanceMap")
	}
}

func TestAllStatuses(t *testing.T) {
	if len(workflow.AllStatuses) != 13 {
		t.Errorf("AllStatuses count = %d, want 13", len(workflow.AllStatuses))
	}
	for _, s := range workflow.AllStatuses {
		if _, ok := workflow.Transitions[s]; !ok {
			t.Errorf("status %q missing from Transitions map", s)
		}
	}
}

func TestEmitAndListen(t *testing.T) {
	var received []workflow.TransitionEvent
	listener := workflow.TransitionListenerFunc(func(e workflow.TransitionEvent) {
		received = append(received, e)
	})
	workflow.RegisterListener(listener)
	workflow.Emit(workflow.TransitionEvent{
		Project: "test", TaskID: "T-1", From: "todo", To: "in-progress",
	})
	if len(received) == 0 {
		t.Fatal("listener did not receive event")
	}
	if received[0].From != "todo" || received[0].To != "in-progress" {
		t.Errorf("event = %+v", received[0])
	}
}
