package workflow

// Transitions defines valid state transitions for the 13-state lifecycle.
var Transitions = map[string][]string{
	"backlog":           {"todo"},
	"todo":              {"in-progress", "backlog"},
	"in-progress":       {"ready-for-testing", "blocked", "todo"},
	"blocked":           {"in-progress", "todo"},
	"ready-for-testing": {"in-testing", "in-progress"},
	"in-testing":        {"ready-for-docs", "in-progress"},
	"ready-for-docs":    {"in-docs", "in-testing"},
	"in-docs":           {"documented", "ready-for-docs"},
	"documented":        {"in-review"},
	"in-review":         {"done", "rejected", "documented"},
	"done":              {"todo"},
	"rejected":          {"todo", "backlog"},
	"cancelled":         {"backlog"},
}

// CompletedStatuses are terminal states that count as resolved.
var CompletedStatuses = map[string]bool{
	"done":      true,
	"rejected":  true,
	"cancelled": true,
}

// DoneStatuses are states where work finished successfully.
var DoneStatuses = map[string]bool{
	"done": true,
}

// ActiveStatuses are states where work is actively happening.
var ActiveStatuses = map[string]bool{
	"in-progress": true,
	"in-testing":  true,
	"in-docs":     true,
	"in-review":   true,
}

// WaitingStatuses are states waiting for the next phase to start.
var WaitingStatuses = map[string]bool{
	"ready-for-testing": true,
	"ready-for-docs":    true,
	"documented":        true,
}

// AdvanceMap defines the happy-path next state for advance_task.
var AdvanceMap = map[string]string{
	"in-progress":       "ready-for-testing",
	"ready-for-testing": "in-testing",
	"in-testing":        "ready-for-docs",
	"ready-for-docs":    "in-docs",
	"in-docs":           "documented",
	"documented":        "in-review",
	"in-review":         "done",
}

// AllStatuses lists every valid status.
var AllStatuses = []string{
	"backlog", "todo", "in-progress", "blocked",
	"ready-for-testing", "in-testing",
	"ready-for-docs", "in-docs", "documented",
	"in-review", "done", "rejected", "cancelled",
}

// IsValid checks whether transitioning from -> to is allowed.
func IsValid(from, to string) bool {
	targets, ok := Transitions[from]
	if !ok {
		return false
	}
	for _, t := range targets {
		if t == to {
			return true
		}
	}
	return false
}

// NextStates returns the valid transitions from a given state.
func NextStates(current string) []string {
	if targets, ok := Transitions[current]; ok {
		return targets
	}
	return nil
}
