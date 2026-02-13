# Rust Engine Integration

The MCP plugin optionally integrates with the `orchestra-engine` Rust binary for gRPC-based memory operations.

## Overview

```
orchestra-mcp (Go)
    │
    ├── engine/resolve.go    → Find engine binary
    ├── engine/manager.go    → Start/stop engine subprocess
    ├── engine/client.go     → gRPC client (6 RPCs)
    └── engine/bridge.go     → gRPC-first, TOON-fallback
         │
         ▼
orchestra-engine (Rust, gRPC on port 50051)
    ├── MemoryService.StoreChunk
    ├── MemoryService.SearchMemory
    ├── MemoryService.GetContext
    ├── MemoryService.StoreSession
    ├── MemoryService.ListSessions
    └── MemoryService.GetSession
```

## Binary Discovery

`engine/resolve.go` finds the engine binary in this order:

1. Same directory as the running `orchestra-mcp` binary (`os.Executable()` dir)
2. `exec.LookPath("orchestra-engine")` (system PATH)
3. Returns `""` if not found (not fatal — TOON fallback)

## Subprocess Lifecycle

`engine/manager.go` manages the engine process:

### Start

```go
mgr := engine.NewManager()
err := mgr.Start(workspace)
```

1. Resolves engine binary path
2. Checks if port 50051 is already occupied (200ms TCP dial)
3. If port is free, spawns engine as subprocess
4. Waits up to 3s for port to become reachable
5. Redirects engine stdout/stderr to MCP stderr (not stdout — that's JSON-RPC)

### Stop

```go
defer mgr.Stop()
```

1. Sends SIGTERM on unix (Process.Kill on Windows)
2. Kills process group (`syscall.SysProcAttr{Setpgid: true}`)
3. 3-second grace period

### Port Configuration

Default port: `50051`. Override via `ORCHESTRA_ENGINE_PORT` env var.

## gRPC Client

`engine/client.go` wraps the generated protobuf client:

```go
client, err := engine.Dial(mgr.Addr())  // "localhost:50051"
defer client.Close()

// 6 methods with 5-second timeout:
client.StoreChunk(project, source, sourceID, summary, content, tags)
client.SearchMemory(project, query, limit)
client.GetContext(project, query, limit)
client.StoreSession(project, sessionID, summary, events)
client.ListSessions(project, limit)
client.GetSession(project, sessionID)
```

Uses `google.golang.org/grpc` with `insecure.NewCredentials()` (local only).

## Bridge Pattern

`engine/bridge.go` provides the fallback abstraction:

```go
bridge := engine.NewBridge(client, ws)  // client can be nil

if bridge.UsingEngine() {
    // Use gRPC
} else {
    // Use TOON fallback
}
```

Every memory tool handler follows this pattern in `tools/memory.go`.

## Proto Definition

The gRPC service is defined in `proto/memory.proto`. Generated Go code is committed at `src/gen/memoryv1/`:

- `memory.pb.go` — message types
- `memory_grpc.pb.go` — client stub

Regenerate with: `make proto-mcp`

## Distribution

Both binaries ship together in release archives:

- **GoReleaser**: `.goreleaser.yaml` includes engine binary in archives
- **install.sh**: Installs both `orchestra-mcp` and `orchestra-engine` to `/usr/local/bin`
- **npm**: `install.js` chmods both binaries after extraction
- **Makefile**: `make build-engine` builds Rust, `make build` builds everything

## Modes

```
[Orchestra MCP] Engine: running on localhost:50051
[Orchestra MCP] Server v1.0.0 running with 57 tools | Memory: Rust engine (gRPC on localhost:50051)
```

or without engine:

```
[Orchestra MCP] Engine: orchestra-engine binary not found (using TOON fallback)
[Orchestra MCP] Server v1.0.0 running with 57 tools | Memory: TOON fallback
```
