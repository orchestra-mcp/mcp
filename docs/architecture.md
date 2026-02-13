# MCP Plugin Architecture

## Module Structure

The MCP plugin is a standalone Go module (`github.com/orchestra-mcp/mcp`) with a `replace` directive back to the core framework.

```
plugins/mcp/
├── go.mod                    # Standalone module
├── config/                   # Plugin-specific config
├── providers/                # Bridge to core plugin system
│   ├── plugin.go             # McpPlugin struct + lifecycle
│   └── tools.go              # Tool aggregation + REST routes
└── src/                      # All implementation
    ├── cmd/                  # CLI entry point (engine lifecycle)
    ├── types/                # Type definitions
    ├── toon/                 # TOON file format
    ├── workflow/             # 13-state lifecycle machine
    ├── helpers/              # Shared utilities
    ├── transport/            # Stdio JSON-RPC server
    ├── engine/               # Rust engine integration (gRPC)
    │   ├── resolve.go        # Binary discovery
    │   ├── manager.go        # Subprocess lifecycle
    │   ├── client.go         # gRPC client wrapper
    │   └── bridge.go         # gRPC/TOON fallback dispatcher
    ├── gen/memoryv1/         # Generated protobuf code
    ├── tools/                # 57 tool implementations (12 files)
    └── bootstrap/            # Workspace init + embedded resources
        ├── init.go           # Init command
        └── resources/        # go:embed skills, agents, docs, hooks
```

## Dual-Mode Operation

### Standalone (stdio)

```
Claude Code <-> stdio JSON-RPC <-> MCPServer (transport/server.go)
                                        |
                                   tools/* handlers
                                     /        \
                          engine/bridge   .projects/ TOON files
                              |
                    orchestra-engine (gRPC, optional)
```

Entry point: `src/cmd/main.go`. Starts engine subprocess, connects gRPC client, reads stdin JSON-RPC, dispatches to tools.

### Integrated (Go plugin)

```
HTTP Client -> Fiber Router -> /api/mcp/tools/call -> McpPlugin.allTools()
                                                           |
                                                      tool.Handler(args)
                                                        /        \
                                             engine/bridge   .projects/ TOON
```

Entry point: `providers/tools.go`. The `RegisterRoutes()` method adds REST endpoints under `/api/mcp/`.

## Engine Integration

The Rust engine (`orchestra-engine`) provides gRPC-based memory operations. The Go plugin manages it as a subprocess.

### Startup Sequence

```
1. engine.Resolve()     → Find binary (same dir → PATH → not found)
2. manager.Start(ws)    → Check port 50051, spawn if free, wait 3s
3. engine.Dial(addr)    → Connect gRPC client
4. engine.NewBridge()   → Create fallback dispatcher
5. tools.Memory(ws, bridge) → Register memory tools with bridge
```

### Fallback Pattern

Every memory tool follows gRPC-first, TOON-fallback:

```go
func saveMemory(ws string, bridge *engine.Bridge) t.Tool {
    Handler: func(args map[string]any) (*t.ToolResult, error) {
        if bridge.UsingEngine() {
            resp, err := bridge.Client.StoreChunk(...)
            if err == nil {
                return h.JSONResult(resp.Chunk), nil
            }
            logFallback("save_memory", err)
        }
        return toonSaveMemory(ws, slug, args, tags)
    }
}
```

### Shutdown

```
1. defer mgr.Stop()     → SIGTERM engine process (SIGKILL after 3s)
2. defer client.Close()  → Close gRPC connection
```

Process group cleanup via `syscall.SysProcAttr{Setpgid: true}` on unix.

## Tool Architecture

### Tool Definition

Every tool is a `types.Tool` — a definition paired with a handler:

```go
type Tool struct {
    Definition ToolDefinition
    Handler    ToolHandler // func(map[string]any) (*ToolResult, error)
}
```

### Tool Categories (57 tools, 12 files)

| File | Count | Function | Signature |
|------|-------|----------|-----------|
| `project.go` | 5 | `Project(ws)` | Project CRUD + PRD |
| `epic.go` | 5 | `Epic(ws)` | Epic CRUD |
| `story.go` | 5 | `Story(ws)` | Story CRUD |
| `task.go` | 5 | `Task(ws)` | Task CRUD |
| `workflow.go` | 5 | `Workflow(ws)` | Next task, current, complete, search, status |
| `lifecycle.go` | 2 | `Lifecycle(ws)` | Advance + reject |
| `prd.go` | 9 | `Prd(ws)` | PRD session, phases |
| `bugfix.go` | 2 | `Bugfix(ws)` | Bug report, feature request |
| `memory.go` | 6 | `Memory(ws, bridge)` | Memory + sessions (gRPC/TOON) |
| `usage.go` | 3 | `Usage(ws)` | Token tracking |
| `artifacts.go` | 2 | `Artifacts(ws)` | Plans |
| `claude.go` | 7 | `Claude(ws)` | Skills, agents, docs, hooks |
| `readme.go` | 1 | `Readme(ws)` | README generation |

Tools are registered in `src/cmd/main.go`:

```go
s.RegisterTools(tools.Project(ws))
s.RegisterTools(tools.Epic(ws))
// ... 12 tool groups
s.RegisterTools(tools.Memory(ws, bridge))  // bridge for engine fallback
```

### External Tools

Other plugins push tools via `RegisterExternalTools()`. These are stored in `McpPlugin.externalTools` (thread-safe via `sync.RWMutex`). The `allTools()` method combines built-in + external.

## 13-State Workflow

```
backlog → todo → in-progress → ready-for-testing → in-testing
→ ready-for-docs → in-docs → documented → in-review → done

Special: blocked (from in-progress), rejected (from in-review), cancelled (terminal)
```

Defined in `src/workflow/workflow.go`:
- `Transitions` — valid state transitions map
- `AdvanceMap` — happy-path next state for `advance_task`
- `CompletedStatuses` — terminal states (done, rejected, cancelled)
- `ActiveStatuses` — work-in-progress states
- `IsValid(from, to)` — validates transitions

## Data Storage

### Project Data (TOON)

```
.projects/
├── my-app/
│   ├── project-status.toon      # Project metadata + issue index
│   ├── prd.md                   # Generated PRD
│   ├── .memory/
│   │   ├── chunks.toon          # Memory chunks (TOON fallback)
│   │   └── sessions/
│   │       ├── index.toon       # Session index
│   │       └── {id}.toon        # Individual sessions
│   ├── .plans/
│   │   └── {slug}.md            # Saved plans
│   └── epics/
│       └── {epic-id}/
│           ├── epic.toon
│           └── stories/
│               └── {story-id}/
│                   ├── story.toon
│                   └── tasks/
│                       └── {task-id}.toon
├── .events/
│   └── hook-events.toon         # Claude Code hook events
└── .usage/
    └── usage.toon               # Token usage tracking
```

### Bootstrap Resources (go:embed)

```
bootstrap/resources/
├── skills/          # 21 skill directories (SKILL.md each)
├── agents/          # 16 agent markdown files
├── docs/            # CLAUDE.md, AGENTS.md, CONTEXT.md templates
└── hooks/           # orchestra-mcp-hook.sh
```

Installed to workspace on `orchestra-mcp init`.

## Protocol

The stdio server implements MCP protocol version `2024-11-05`:

| Method | Description |
|--------|-------------|
| `initialize` | Handshake, returns capabilities |
| `tools/list` | Returns all 57 tool definitions |
| `tools/call` | Executes a tool by name |
| `ping` | Health check |

Messages are newline-delimited JSON (JSON Lines), not Content-Length framed.

## REST API

When integrated with the Go server:

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/mcp/tools` | List all tools (built-in + external) |
| `POST` | `/api/mcp/tools/call` | Call a tool by name |
| `GET` | `/health` | Server health + active plugin count |
