# AGENTS.md

Specialized agents installed by Orchestra MCP. Each agent handles a specific domain.

## Agent Overview

The scrum-master agent coordinates all others through MCP tools.

```
scrum-master (coordinator)
├── go-architect       → Go backend
├── rust-engineer      → Rust engine
├── frontend-dev       → React/TypeScript
├── ui-ux-designer     → Design system
├── dba                → Database + sync
├── mobile-dev         → React Native
├── devops             → Infrastructure
├── widget-engineer    → Native OS widgets
├── platform-engineer  → OS-level integration
├── extension-architect → Extension system
├── ai-engineer        → AI/LLM features
├── qa-go              → Go tests
├── qa-rust            → Rust tests
├── qa-node            → Node/React tests
└── qa-playwright      → E2E browser tests
```

## MCP-Driven Workflow

All agents follow the Orchestra MCP 13-state workflow:

```
backlog → todo → in-progress → ready-for-testing → in-testing
→ ready-for-docs → in-docs → documented → in-review → done
```

Agents use MCP tools to manage their work:
- `get_next_task` — Pick up work
- `set_current_task` — Start working
- `advance_task` — Progress through stages
- `complete_task` — Finish work
- `save_memory` — Store context for future sessions

## Cross-Agent Communication

Changes in one domain often require updates in others. The scrum-master tracks these dependencies through MCP's epic/story/task hierarchy.

## Adding Agents

New agents are added as `.md` files in `.claude/agents/`. Use `install_agents` to reinstall bundled agents from Orchestra MCP.
