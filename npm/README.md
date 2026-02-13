# @orchestra-mcp/cli

AI-powered project management via [Model Context Protocol](https://modelcontextprotocol.io). 40 built-in tools for managing projects, epics, stories, tasks, PRDs, workflows, and more — directly from your AI assistant.

## Install

```bash
# Global install
npm install -g @orchestra-mcp/cli

# Or use directly with npx
npx @orchestra-mcp/cli --help
```

### Other install methods

```bash
# Homebrew (macOS/Linux)
brew install orchestra-mcp/tap/orchestra-mcp

# Direct download
curl -fsSL https://raw.githubusercontent.com/orchestra-mcp/mcp/master/scripts/install.sh | sh
```

## Quick Start

```bash
# Initialize workspace (creates .mcp.json, .projects/)
orchestra-mcp init --workspace /path/to/project

# Start MCP server (stdio JSON-RPC)
orchestra-mcp --workspace /path/to/project
```

## Claude Code Integration

Add to your `.mcp.json`:

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

Or with npx (no global install needed):

```json
{
  "mcpServers": {
    "orchestra": {
      "command": "npx",
      "args": ["-y", "@orchestra-mcp/cli", "--workspace", "."]
    }
  }
}
```

## Tools (40 Built-in)

| Category | Tools |
|----------|-------|
| **Project** | `list_projects`, `create_project`, `get_project_status`, `read_prd`, `write_prd` |
| **Epic** | `list_epics`, `create_epic`, `get_epic`, `update_epic`, `delete_epic` |
| **Story** | `list_stories`, `create_story`, `get_story`, `update_story`, `delete_story` |
| **Task** | `list_tasks`, `create_task`, `get_task`, `update_task`, `delete_task` |
| **Workflow** | `get_next_task`, `set_current_task`, `complete_task`, `search`, `get_workflow_status` |
| **PRD** | `start_prd_session`, `answer_prd_question`, `get_prd_session`, `abandon_prd_session`, `skip_prd_question`, `back_prd_question`, `preview_prd` |
| **Bugfix** | `report_bug`, `log_request` |
| **Usage** | `get_usage`, `record_usage`, `reset_session_usage` |
| **Readme** | `regenerate_readme` |
| **Artifacts** | `save_plan`, `list_plans` |

## How It Works

This npm package downloads the pre-built `orchestra-mcp` binary for your platform on install. The binary is a standalone Go program with zero runtime dependencies.

Supported platforms:
- macOS (Intel & Apple Silicon)
- Linux (x64 & ARM64)
- Windows (x64 & ARM64)

## Links

- [GitHub](https://github.com/orchestra-mcp/mcp)
- [Documentation](https://github.com/orchestra-mcp/mcp#readme)
- [Issue Tracker](https://github.com/orchestra-mcp/mcp/issues)
- [Changelog](https://github.com/orchestra-mcp/mcp/blob/master/CHANGELOG.md)

## License

MIT
