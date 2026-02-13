# Memory & Session System

Orchestra MCP provides persistent project memory and session tracking across AI assistant sessions.

## Architecture

Memory operates in two modes:

1. **Rust Engine (gRPC)** — vector search, persistent storage via `orchestra-engine`
2. **TOON Fallback** — keyword-based search, local YAML files

The bridge pattern auto-selects: gRPC when engine is running, TOON otherwise.

## Memory Tools

### `save_memory` — Store Context

```json
{
  "project": "my-app",
  "content": "We decided to use PostgreSQL for the auth system",
  "summary": "Auth database decision",
  "source": "task",
  "source_id": "TASK-5",
  "tags": ["auth", "database", "decision"]
}
```

### `search_memory` — Find Relevant Context

```json
{
  "project": "my-app",
  "query": "auth database",
  "limit": 10
}
```

Returns scored results by keyword relevance (TOON) or vector similarity (engine).

### `get_context` — Combined Context

Searches both memory chunks and session logs for relevant context. Best used at session start to load previous decisions and context.

```json
{
  "project": "my-app",
  "query": "what auth approach did we choose",
  "limit": 5
}
```

## Session Tools

### `save_session` — Persist Session

```json
{
  "project": "my-app",
  "session_id": "session-2024-01-15-abc",
  "summary": "Implemented auth handlers and wrote tests",
  "events": [
    {"type": "task_completed", "summary": "Auth handler done"},
    {"type": "decision", "summary": "Using JWT with refresh tokens"}
  ]
}
```

### `list_sessions` — Recent Sessions

```json
{
  "project": "my-app",
  "limit": 20
}
```

### `get_session` — Full Session Details

```json
{
  "project": "my-app",
  "session_id": "session-2024-01-15-abc"
}
```

## Data Storage (TOON Fallback)

```
.projects/my-app/.memory/
├── chunks.toon          # All memory chunks (MemoryIndex)
└── sessions/
    ├── index.toon       # Session index (SessionIndex)
    └── {session-id}.toon  # Individual session logs
```

### Memory Chunk

```yaml
id: mem-1
project: my-app
source: task
source_id: TASK-5
summary: Auth database decision
content: We decided to use PostgreSQL for the auth system
tags: [auth, database, decision]
created_at: "2024-01-15T10:00:00Z"
```

### Session Log

```yaml
session_id: session-2024-01-15-abc
project: my-app
summary: Implemented auth handlers and wrote tests
started_at: "2024-01-15T10:00:00Z"
events:
  - type: task_completed
    summary: Auth handler done
    timestamp: "2024-01-15T10:30:00Z"
```

## Search Algorithm (TOON Fallback)

Keyword-based scoring: splits query into words, counts matches in `summary + content + tags`. Score = matches / total words.

## Recommended Session Flow

### Start of Session

```
get_project_status  → Project overview
list_sessions       → What happened recently
get_context         → Load relevant memory for current work
get_next_task       → Pick up where we left off
```

### End of Session

```
save_session        → Persist what was accomplished
save_memory         → Store important decisions and context
```
