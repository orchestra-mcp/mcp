# CONTEXT.md

Project context managed by Orchestra MCP.

## Orchestra MCP Integration

This project uses Orchestra MCP for AI-driven project management with 56 tools covering:

- **Project hierarchy**: Epics → Stories → Tasks
- **13-state workflow**: backlog through done with testing, docs, and review stages
- **PRD generation**: Guided PRD creation with phase splitting
- **Memory system**: Persistent project knowledge base across sessions
- **Session tracking**: What happened, what's next
- **Usage tracking**: Token and cost monitoring

## Workflow States

```
backlog → todo → in-progress → blocked
                             → ready-for-testing → in-testing
                             → ready-for-docs → in-docs → documented
                             → in-review → done / rejected / cancelled
```

## Data Storage

- `.projects/` — Project data (epics, stories, tasks) in TOON format
- `.projects/{name}/.memory/` — Memory chunks and session logs
- `.projects/{name}/.plans/` — Saved implementation plans
- `.projects/.events/` — Hook event log
- `.mcp.json` — MCP server configuration

## Skills & Agents

- `.claude/skills/` — Domain-specific skill patterns (auto-activated by context)
- `.claude/agents/` — Specialized agent configurations
- `.claude/hooks/` — Claude Code hook scripts

Use `list_skills` and `list_agents` MCP tools to discover what's installed.

## Session Protocol

1. Start: `get_project_status` → `get_next_task` → `set_current_task`
2. Work: `advance_task` to progress, `update_task` for changes
3. End: `complete_task` → `save_session` → `save_memory`
