# 13-State Workflow

The task lifecycle in Orchestra MCP uses a 13-state machine with validated transitions.

## States

```
backlog → todo → in-progress → ready-for-testing → in-testing
→ ready-for-docs → in-docs → documented → in-review → done
```

### State Categories

| Category | States | Description |
|----------|--------|-------------|
| **Queue** | `backlog`, `todo` | Work not yet started |
| **Active** | `in-progress`, `in-testing`, `in-docs`, `in-review` | Work actively happening |
| **Waiting** | `ready-for-testing`, `ready-for-docs`, `documented` | Ready for next phase |
| **Terminal** | `done`, `rejected`, `cancelled` | Resolved |
| **Special** | `blocked` | Impediment, needs unblocking |

## Transition Rules

| From | Valid Transitions |
|------|-------------------|
| `backlog` | `todo` |
| `todo` | `in-progress`, `backlog` |
| `in-progress` | `ready-for-testing`, `blocked`, `todo` |
| `blocked` | `in-progress`, `todo` |
| `ready-for-testing` | `in-testing`, `in-progress` |
| `in-testing` | `ready-for-docs`, `in-progress` |
| `ready-for-docs` | `in-docs`, `in-testing` |
| `in-docs` | `documented`, `ready-for-docs` |
| `documented` | `in-review` |
| `in-review` | `done`, `rejected`, `documented` |
| `done` | `todo` (reopen) |
| `rejected` | `todo`, `backlog` |
| `cancelled` | `backlog` |

## MCP Tools

### `advance_task` — Happy Path

Moves a task to the next logical state:

```
in-progress       → ready-for-testing
ready-for-testing → in-testing
in-testing        → ready-for-docs
ready-for-docs    → in-docs
in-docs           → documented
documented        → in-review
in-review         → done
```

### `reject_task` — From Review

Moves task to `rejected` and auto-creates a bug under the same story with the rejection reason.

### `update_task` — Any Valid Transition

Validates the transition using `workflow.IsValid(from, to)` before applying.

### `set_current_task` — Start Work

Sets task to `in-progress` and cascades parent story/epic to `in-progress`.

### `complete_task` — Finish Work

Sets task to `done`. If all sibling tasks under the story are done, cascades the story to `done`. If all stories under the epic are done, cascades the epic to `done`.

## Event System

Every transition emits a `TransitionEvent`:

```go
type TransitionEvent struct {
    Project string
    EpicID  string
    StoryID string
    TaskID  string
    Type    string // task, bug, hotfix
    From    string
    To      string
    Time    string
}
```

Events are written to `.projects/.events/` and can be queried via Claude Code hooks.

## Implementation

All workflow logic lives in `src/workflow/workflow.go`:

```go
workflow.IsValid(from, to string) bool
workflow.NextStates(current string) []string
workflow.AdvanceMap[status] string
workflow.CompletedStatuses[status] bool
workflow.ActiveStatuses[status] bool
```
