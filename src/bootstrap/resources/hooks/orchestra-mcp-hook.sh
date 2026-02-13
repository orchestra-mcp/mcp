#!/bin/bash
# Orchestra MCP hook — pipes Claude Code events to MCP server
# Called by Claude Code for all configured hook events (async, never blocks)
set -e

INPUT=$(cat)

# Extract fields from Claude Code hook JSON
EVENT_NAME=$(echo "$INPUT" | jq -r '.hook_event_name // "unknown"')
SESSION_ID=$(echo "$INPUT" | jq -r '.session_id // ""')
TOOL_NAME=$(echo "$INPUT" | jq -r '.tool_name // ""')
AGENT_TYPE=$(echo "$INPUT" | jq -r '.agent_type // ""')

# Build JSON-RPC call to orchestra-mcp receive_hook_event tool
echo "$INPUT" | jq -c '{
  jsonrpc: "2.0", id: 1, method: "tools/call",
  params: {
    name: "receive_hook_event",
    arguments: {
      event_type: .hook_event_name,
      session_id: (.session_id // ""),
      tool_name: (.tool_name // ""),
      agent_type: (.agent_type // ""),
      data: .
    }
  }
}' | orchestra-mcp --workspace "${CLAUDE_PROJECT_DIR:-.}" 2>/dev/null | head -1 > /dev/null

exit 0
