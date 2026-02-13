# Bootstrap & Init

The `orchestra-mcp init` command sets up a workspace with all required files and resources.

## What Gets Installed

```
orchestra-mcp init --workspace /path/to/project
```

Creates:

| Path | Description |
|------|-------------|
| `.mcp.json` | MCP server config (orchestra-mcp command + args) |
| `.projects/{name}/` | Project data directory |
| `.projects/.events/` | Hook event storage |
| `.claude/skills/` | 21 bundled skills (auto-activated by context) |
| `.claude/agents/` | 16 bundled agents (specialized subagents) |
| `.claude/hooks/orchestra-mcp-hook.sh` | Hook script for event tracking |
| `.claude/settings.json` | Hook event configuration |
| `CLAUDE.md` | Project instructions for AI assistants |
| `AGENTS.md` | Agent reference documentation |
| `CONTEXT.md` | Project context documentation |

## Project Detection

The init command auto-detects:

- **Project name**: from `package.json` name, `go.mod` module, `Cargo.toml` name, or directory name
- **Project type**: Node.js, Go, Rust, PHP, Python, Ruby, or Unknown

## Embedded Resources

All resources are bundled into the binary via `go:embed`:

```go
//go:embed resources/skills
var bundledSkills embed.FS

//go:embed resources/agents
var bundledAgents embed.FS

//go:embed resources/hooks/orchestra-mcp-hook.sh
var bundledHookScript string

//go:embed resources/docs/CLAUDE.md
var bundledClaudeMD string

//go:embed resources/docs/AGENTS.md
var bundledAgentsMD string

//go:embed resources/docs/CONTEXT.md
var bundledContextMD string
```

## Document Safety

`CLAUDE.md`, `AGENTS.md`, and `CONTEXT.md` are only written if they don't already exist — user customizations are never overwritten.

## Bundled Skills (21)

| Skill | Domain |
|-------|--------|
| `go-backend` | Go API (Fiber v3, GORM) |
| `rust-engine` | Rust engine (Tonic, Tree-sitter, Tantivy) |
| `typescript-react` | React + TypeScript |
| `ui-design` | shadcn/ui + Tailwind |
| `database-sync` | PostgreSQL + SQLite + Redis |
| `proto-grpc` | Protobuf + Buf |
| `chrome-extension` | Chrome Manifest V3 |
| `wails-desktop` | Wails v3 desktop |
| `react-native-mobile` | React Native + WatermelonDB |
| `native-widgets` | macOS/Windows/Linux widgets |
| `macos-integration` | CGo + Spotlight + Keychain |
| `native-extensions` | Extension API |
| `raycast-compat` | Raycast shim |
| `vscode-compat` | VS Code shim |
| `extension-marketplace` | Marketplace |
| `ai-agentic` | AI/LLM integration |
| `gcp-infrastructure` | GCP + Docker |
| `project-manager` | Sprint planning + MCP workflow |
| `docs` | Documentation |
| `qa-testing` | Multi-agent testing |
| `tailwindcss-development` | Tailwind CSS v4 |

## Bundled Agents (16)

`go-architect`, `rust-engineer`, `frontend-dev`, `ui-ux-designer`, `dba`, `mobile-dev`, `devops`, `widget-engineer`, `platform-engineer`, `extension-architect`, `ai-engineer`, `scrum-master`, `qa-go`, `qa-rust`, `qa-node`, `qa-playwright`

## Hook Events

The hook script captures Claude Code events and sends them to MCP:

- `PostToolUse` — after any tool call
- `Notification` — Claude notifications
- `SubagentStart` / `SubagentStop` — subagent lifecycle
- `Stop` — session end
- `SessionStart` — session begin

## MCP Tools for Bootstrap

| Tool | Description |
|------|-------------|
| `install_skills` | Reinstall all bundled skills |
| `install_agents` | Reinstall all bundled agents |
| `install_docs` | Reinstall CLAUDE.md, AGENTS.md, CONTEXT.md |
| `list_skills` | List installed skills |
| `list_agents` | List installed agents |
