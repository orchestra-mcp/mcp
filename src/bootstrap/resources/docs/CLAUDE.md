# CLAUDE.md

This project uses **Orchestra MCP** for AI-driven project management.

## Orchestra MCP

Orchestra MCP provides 56 tools for managing epics, stories, tasks, PRDs, memory, and workflow through the Model Context Protocol. All project management flows through MCP — never manage tasks manually.

### Quick Start

```bash
# See project status
orchestra-mcp get_project_status

# Get next task to work on
orchestra-mcp get_next_task

# Start working on a task
orchestra-mcp set_current_task
```

### Workflow

Tasks follow a 13-state lifecycle:

```
backlog → todo → in-progress → ready-for-testing → in-testing
→ ready-for-docs → in-docs → documented → in-review → done
```

Special states: `blocked`, `rejected` (auto-creates bug), `cancelled`.

Use `advance_task` to move through lifecycle. Use `complete_task` when done.

### Session Flow

Every session should:
1. `get_project_status` — Where are we?
2. `get_next_task` — What's next?
3. `set_current_task` — Start working
4. `advance_task` / `complete_task` — Progress and finish
5. `save_session` — Persist what happened

### Key Tools

| Category | Tools |
|----------|-------|
| Project | `create_project`, `list_projects`, `get_project_status`, `read_prd`, `write_prd` |
| Epic | `create_epic`, `list_epics`, `get_epic`, `update_epic`, `delete_epic` |
| Story | `create_story`, `list_stories`, `get_story`, `update_story`, `delete_story` |
| Task | `create_task`, `list_tasks`, `get_task`, `update_task`, `delete_task` |
| Workflow | `get_next_task`, `set_current_task`, `complete_task`, `search`, `get_workflow_status` |
| Lifecycle | `advance_task`, `reject_task` |
| PRD | `start_prd_session`, `answer_prd_question`, `preview_prd`, `split_prd` |
| Quality | `report_bug`, `log_request` |
| Memory | `save_memory`, `search_memory`, `get_context`, `save_session` |

### Skills & Agents

Skills and agents are installed in `.claude/skills/` and `.claude/agents/`. Use `list_skills` and `list_agents` to discover them.

### Data

Project data is stored in `.projects/` using TOON format (YAML-based). Memory is stored in `.projects/{name}/.memory/`.
