# Tool Development Guide

How to add new tools to the MCP plugin.

## Adding a Built-in Tool

### 1. Create the Tool File

```go
// src/tools/my_category.go
package tools

import (
    h "github.com/orchestra-mcp/mcp/src/helpers"
    t "github.com/orchestra-mcp/mcp/src/types"
)

func MyCategory(workspace string) []t.Tool {
    return []t.Tool{
        {
            Definition: t.ToolDefinition{
                Name:        "my_tool",
                Description: "What this tool does",
                InputSchema: t.InputSchema{
                    Type: "object",
                    Properties: map[string]any{
                        "name": map[string]any{
                            "type":        "string",
                            "description": "The name parameter",
                        },
                    },
                    Required: []string{"name"},
                },
            },
            Handler: func(args map[string]any) (*t.ToolResult, error) {
                name := h.GetString(args, "name")
                if name == "" {
                    return h.ErrorResult("name is required"), nil
                }
                return h.JSONResult(map[string]string{
                    "status": "done",
                    "name":   name,
                }), nil
            },
        },
    }
}
```

### 2. Register in `cmd/main.go`

```go
s.RegisterTools(tools.MyCategory(ws))
```

### 3. Write Tests

```go
// tests/unit/tools/my_category_test.go
package tools_test

import (
    "testing"
    "github.com/orchestra-mcp/mcp/src/tools"
)

func TestMyTool(t *testing.T) {
    ws := t.TempDir()
    toolList := tools.MyCategory(ws)
    res, err := toolList[0].Handler(map[string]any{"name": "test"})
    if err != nil {
        t.Fatalf("error: %v", err)
    }
    if res.IsError {
        t.Fatalf("tool error: %s", res.Content[0].Text)
    }
}
```

## Adding an Engine-Aware Tool

Tools that need the Rust engine follow the bridge pattern — gRPC first, TOON fallback.

### 1. Accept the Bridge

```go
func MyEngineTools(ws string, bridge *engine.Bridge) []t.Tool {
    return []t.Tool{myEngineTool(ws, bridge)}
}
```

### 2. gRPC-First Pattern

```go
func myEngineTool(ws string, bridge *engine.Bridge) t.Tool {
    return t.Tool{
        Definition: t.ToolDefinition{
            Name: "my_engine_tool", Description: "Uses engine when available",
            InputSchema: t.InputSchema{Type: "object", Properties: map[string]any{
                "project": map[string]any{"type": "string"},
                "query":   map[string]any{"type": "string"},
            }, Required: []string{"project", "query"}},
        },
        Handler: func(args map[string]any) (*t.ToolResult, error) {
            slug := h.GetString(args, "project")
            query := h.GetString(args, "query")

            // Try gRPC first
            if bridge.UsingEngine() {
                resp, err := bridge.Client.SearchMemory(slug, query, 10)
                if err == nil {
                    return h.JSONResult(resp.Results), nil
                }
                fmt.Fprintf(os.Stderr, "[my_tool] gRPC failed, TOON fallback: %v\n", err)
            }

            // TOON fallback
            return toonFallback(ws, slug, query)
        },
    }
}
```

### 3. Register with Bridge

In `cmd/main.go`:

```go
bridge := engine.NewBridge(client, ws)
s.RegisterTools(tools.MyEngineTools(ws, bridge))
```

## Using the Workflow

### Validate Transitions

```go
import "github.com/orchestra-mcp/mcp/src/workflow"

if !workflow.IsValid(task.Status, newStatus) {
    return h.ErrorResult("invalid transition"), nil
}
```

### Advance Happy Path

```go
next, ok := workflow.AdvanceMap[task.Status]
if !ok {
    return h.ErrorResult("cannot advance from " + task.Status), nil
}
task.Status = next
```

### Emit Events

```go
workflow.Emit(workflow.TransitionEvent{
    Project: slug, EpicID: epicID, StoryID: storyID, TaskID: taskID,
    Type: task.Type, From: oldStatus, To: newStatus, Time: h.Now(),
})
```

## Result Helpers

Use helpers from `src/helpers/results.go`:

```go
h.TextResult("plain text response")
h.JSONResult(map[string]any{"key": "value"})
h.ErrorResult("something went wrong")
```

## Argument Helpers

Use helpers from `src/helpers/args.go`:

```go
name := h.GetString(args, "name")     // string or ""
count := h.GetInt(args, "count")      // int or 0 (handles float64)
val := h.GetFloat64(args, "val")      // float64 or 0
exists := h.Has(args, "key")          // bool
```

## File Helpers

Use helpers from `src/helpers/paths.go`:

```go
dir := h.ProjectsDir(workspace)              // .projects/
dir := h.ProjectDir(workspace, "my-app")     // .projects/my-app/
exists := h.FileExists(path)                 // bool
```

## TOON Files

Read and write TOON (YAML) files via `src/toon/`:

```go
// Write a struct to a TOON file
toon.WriteFile(path, &myStruct)

// Parse a TOON file into a struct
var data MyStruct
toon.ParseFile(path, &data)
```

## Adding Bootstrap Resources

To bundle new resources that get installed on `orchestra-mcp init`:

1. Add files to `src/bootstrap/resources/` (skills, agents, docs, or hooks)
2. Resources are auto-embedded via `//go:embed` directives in `init.go`
3. Skills go in `resources/skills/{name}/SKILL.md`
4. Agents go in `resources/agents/{name}.md`
5. Docs go in `resources/docs/{NAME}.md` (installed to project root)

## Adding External Tools (From Another Plugin)

See [Creating Plugins — Push Tools to MCP](../../docs/guides/creating-plugins.md).
