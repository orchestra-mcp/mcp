# Orchestra MCP Plugin

Model Context Protocol server for AI-powered project management. Pure Go, 57 built-in tools, Rust engine integration, extensible by other plugins.

## Overview

The MCP plugin provides a complete project management toolkit that AI assistants (Claude Code, Cursor, etc.) can use via the Model Context Protocol. It works in two modes:

- **Standalone CLI** — any project, no Go server needed
- **Integrated plugin** — registered with Orchestra's plugin system, tools available via REST API

Features:
- **57 MCP tools** — project hierarchy, 13-state workflow, PRD generation, memory/RAG, session tracking
- **Rust engine** — optional gRPC engine for vector search and persistent memory (auto-starts/stops)
- **TOON fallback** — works without the engine using local YAML-based storage
- **Bundled skills & agents** — installs 21 skills, 16 agents, and CLAUDE.md/AGENTS.md/CONTEXT.md on init
- **Claude Code hooks** — event pipeline for tool use tracking

## Install

### npm / npx

```bash
npx @orchestra-mcp/cli init
# or install globally
npm install -g @orchestra-mcp/cli
```

### Direct Download

```bash
curl -fsSL https://raw.githubusercontent.com/orchestra-mcp/mcp/master/scripts/install.sh | sh
```

### From Source

```bash
cd plugins/mcp && go build -o orchestra-mcp ./src/cmd/
```

## Usage

### Standalone CLI

```bash
# Initialize a workspace (creates .projects/, .mcp.json, .claude/)
./orchestra-mcp init --workspace /path/to/project

# Start stdio MCP server
./orchestra-mcp --workspace /path/to/project
```

### What `init` Installs

```
.mcp.json                      # MCP server config
.projects/{name}/              # Project data (TOON format)
.claude/skills/                # 21 bundled skills
.claude/agents/                # 16 bundled agents
.claude/hooks/                 # Hook scripts
.claude/settings.json          # Hook event config
CLAUDE.md                      # Project instructions for AI
AGENTS.md                      # Agent reference
CONTEXT.md                     # Project context
```

### Integrated Plugin

```go
import mcpproviders "github.com/orchestra-mcp/mcp/providers"

pm := plugins.NewPluginManager(cfg)
pm.Register(mcpproviders.NewMcpPlugin())
pm.Boot()

// Tools are now available via:
// - REST: GET /api/mcp/tools, POST /api/mcp/tools/call
// - CollectMcpTools() aggregation
// - Stdio (standalone binary)
```

### Claude Code Integration

Add to `.mcp.json`:

```json
{
  "mcpServers": {
    "orchestra": {
      "command": "orchestra-mcp",
      "args": ["--workspace", "."]
    }
  }
}
```

## Structure

```
plugins/mcp/
├── go.mod                          # Standalone module
├── .goreleaser.yaml                # Multi-platform build config
├── config/mcp.go                   # McpConfig (Enabled, Binary)
├── providers/
│   ├── plugin.go                   # McpPlugin — Go plugin registration
│   └── tools.go                    # Tool bridge + REST API routes
├── src/
│   ├── cmd/main.go                 # CLI entry point (engine lifecycle)
│   ├── version/version.go          # Build-time version (ldflags)
│   ├── types/                      # Protocol, tool, data types
│   ├── toon/toon.go                # TOON file read/write (YAML)
│   ├── workflow/workflow.go         # 13-state lifecycle machine
│   ├── helpers/                     # Path, string, args, result utilities
│   ├── transport/server.go          # Stdio JSON-RPC server
│   ├── engine/                      # Rust engine integration
│   │   ├── resolve.go              # Binary discovery
│   │   ├── manager.go              # Subprocess lifecycle
│   │   ├── client.go               # gRPC client wrapper
│   │   └── bridge.go               # gRPC/TOON fallback dispatcher
│   ├── gen/memoryv1/               # Generated protobuf code
│   ├── tools/                       # 57 tool implementations (12 files)
│   └── bootstrap/
│       ├── init.go                  # Workspace init command
│       └── resources/               # go:embed bundled skills, agents, docs
├── tests/
│   └── unit/                        # Unit tests by package
├── npm/                             # npm wrapper (@orchestra-mcp/cli)
├── scripts/install.sh               # Curl one-liner installer
└── docs/                            # Plugin documentation
```

## Tools (57 Built-in)

| Category | Count | Tools |
|----------|-------|-------|
| Project | 5 | `list_projects`, `create_project`, `get_project_status`, `read_prd`, `write_prd` |
| Epic | 5 | `list_epics`, `create_epic`, `get_epic`, `update_epic`, `delete_epic` |
| Story | 5 | `list_stories`, `create_story`, `get_story`, `update_story`, `delete_story` |
| Task | 5 | `list_tasks`, `create_task`, `get_task`, `update_task`, `delete_task` |
| Workflow | 5 | `get_next_task`, `set_current_task`, `complete_task`, `search`, `get_workflow_status` |
| Lifecycle | 2 | `advance_task`, `reject_task` |
| PRD | 9 | `start_prd_session`, `answer_prd_question`, `get_prd_session`, `abandon_prd_session`, `skip_prd_question`, `back_prd_question`, `preview_prd`, `split_prd`, `list_prd_phases` |
| Quality | 2 | `report_bug`, `log_request` |
| Memory | 6 | `save_memory`, `search_memory`, `get_context`, `save_session`, `list_sessions`, `get_session` |
| Usage | 3 | `get_usage`, `record_usage`, `reset_session_usage` |
| Artifacts | 2 | `save_plan`, `list_plans` |
| Claude | 7 | `list_skills`, `list_agents`, `install_skills`, `install_agents`, `install_docs`, `receive_hook_event`, `get_hook_events` |
| Docs | 1 | `regenerate_readme` |

## 13-State Workflow

```
backlog → todo → in-progress → ready-for-testing → in-testing
→ ready-for-docs → in-docs → documented → in-review → done

Special: blocked (from in-progress), rejected (from in-review), cancelled (terminal)
```

Use `advance_task` for happy-path progression. Use `reject_task` to reject from review (auto-creates bug). Transitions are validated — invalid transitions return an error with valid next states.

## Rust Engine Integration

The MCP binary optionally auto-starts the `orchestra-engine` Rust binary for:
- Vector search memory (gRPC)
- Persistent session storage
- Full-text search indexing

If the engine binary is not found, all memory tools fall back to TOON (local YAML files). The engine is discovered by:
1. Same directory as the `orchestra-mcp` binary
2. `PATH` lookup
3. Falls back to TOON if not found

## Extensibility

Other plugins can push tools into the MCP server:

```go
mcpPlugin.RegisterExternalTools([]plugins.McpToolDefinition{
    {
        Name:        "my_custom_tool",
        Description: "A tool from my plugin",
        InputSchema: map[string]any{"type": "object"},
        Handler:     func(input map[string]any) (any, error) { return "ok", nil },
    },
})
```

External tools appear in all channels: stdio, REST API, and `CollectMcpTools()`.

## REST API

When integrated with the Go server:

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/mcp/tools` | List all tools (built-in + external) |
| `POST` | `/api/mcp/tools/call` | Call a tool by name |
| `GET` | `/health` | Server health + active plugin count |

### Call a tool via REST

```bash
curl -X POST http://localhost:8080/api/mcp/tools/call \
  -H 'Content-Type: application/json' \
  -d '{"name": "list_projects", "arguments": {}}'
```

## Configuration

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `enabled` | bool | `true` | Enable the MCP plugin |
| `binary` | string | `orchestra-mcp` | Path to the standalone binary |
| `workspace` | string | `.` | Default workspace root |

## Testing & Code Quality

```bash
# Run tests
cd plugins/mcp && go test ./tests/... -v

# Run linter
cd plugins/mcp && golangci-lint run ./...

# Format code
gofumpt -w config/ providers/ src/ tests/
```

Or from the repo root:

```bash
make test      # runs all tests (framework + plugin)
make lint      # runs linter on framework + plugin
make fmt       # formats all Go code
make check     # full pipeline: format check + lint + tests
```

## Documentation

- [Architecture](docs/architecture.md) — internal architecture, dual-mode operation, data storage
- [Tool Development](docs/tool-development.md) — adding new tools, engine-aware tools, workflow API
- [13-State Workflow](docs/workflow.md) — lifecycle states, transitions, advance/reject/complete
- [Engine Integration](docs/engine-integration.md) — Rust gRPC engine, bridge pattern, distribution
- [Bootstrap & Init](docs/bootstrap.md) — what gets installed, embedded resources, skills/agents
- [Memory & Sessions](docs/memory.md) — memory system, session tracking, search, recommended flow
- [Plugin System Guide](../../docs/guides/plugin-system.md) — framework docs
