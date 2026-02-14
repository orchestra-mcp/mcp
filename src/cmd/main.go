package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/orchestra-mcp/discord/src/notifier"
	h "github.com/orchestra-mcp/mcp/src/helpers"
	"github.com/orchestra-mcp/mcp/src/bootstrap"
	"github.com/orchestra-mcp/mcp/src/engine"
	"github.com/orchestra-mcp/mcp/src/toon"
	"github.com/orchestra-mcp/mcp/src/tools"
	"github.com/orchestra-mcp/mcp/src/transport"
	t "github.com/orchestra-mcp/mcp/src/types"
	"github.com/orchestra-mcp/mcp/src/version"
	"github.com/orchestra-mcp/mcp/src/workflow"
)

const cmdInit = "init"

func main() {
	ws := "."
	var cmd string

	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--version", "-v":
			fmt.Printf("orchestra-mcp %s (commit %s, built %s)\n",
				version.Version, version.Commit, version.Date)
			return
		case "--help", "-h":
			printUsage()
			return
		case "--workspace":
			if i+1 < len(args) {
				ws = args[i+1]
				i++
			}
		case cmdInit:
			cmd = cmdInit
		}
	}

	if cmd == cmdInit {
		if err := bootstrap.Run(ws); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(1)
		}
		return
	}

	// Start Rust engine (non-fatal if binary missing)
	mgr := engine.NewManager()
	if err := mgr.Start(ws); err != nil {
		fmt.Fprintf(os.Stderr, "[Orchestra MCP] Engine: %v (using TOON fallback)\n", err)
	} else {
		fmt.Fprintf(os.Stderr, "[Orchestra MCP] Engine: running on %s\n", mgr.Addr())
	}
	defer mgr.Stop()

	// Connect gRPC client if engine is running
	var client *engine.Client
	if mgr.IsRunning() {
		c, err := engine.Dial(mgr.Addr())
		if err != nil {
			fmt.Fprintf(os.Stderr, "[Orchestra MCP] Engine dial failed: %v\n", err)
		} else {
			client = c
		}
	}
	if client != nil {
		defer client.Close()
	}
	bridge := engine.NewBridge(client, ws)

	s := transport.New("orchestra-mcp", version.Version)
	s.RegisterTools(tools.Project(ws))
	s.RegisterTools(tools.Epic(ws))
	s.RegisterTools(tools.Story(ws))
	s.RegisterTools(tools.Task(ws))
	s.RegisterTools(tools.Workflow(ws))
	s.RegisterTools(tools.Prd(ws))
	s.RegisterTools(tools.Bugfix(ws))
	s.RegisterTools(tools.Usage(ws))
	s.RegisterTools(tools.Readme(ws))
	s.RegisterTools(tools.Artifacts(ws))
	s.RegisterTools(tools.Lifecycle(ws))
	s.RegisterTools(tools.Claude(ws))
	s.RegisterTools(tools.Memory(ws, bridge))

	// Register Discord notifier for workflow transitions
	if dn := notifier.New(); dn != nil {
		workflow.RegisterListener(workflow.TransitionListenerFunc(func(e workflow.TransitionEvent) {
			ne := enrichEvent(ws, e)
			dn.OnTransition(ne)
		}))
		fmt.Fprintf(os.Stderr, "[Orchestra MCP] Discord notifier: enabled\n")
	}

	memMode := "TOON fallback"
	if bridge.UsingEngine() {
		memMode = fmt.Sprintf("Rust engine (gRPC on %s)", mgr.Addr())
	}
	fmt.Fprintf(os.Stderr, "[Orchestra MCP] Server v%s running with %d tools | Memory: %s\n",
		version.Version, len(s.GetTools()), memMode)
	s.Run()
}

func enrichEvent(ws string, e workflow.TransitionEvent) notifier.TransitionEvent {
	ne := notifier.TransitionEvent{
		Project: e.Project, EpicID: e.EpicID, StoryID: e.StoryID,
		TaskID: e.TaskID, Type: e.Type, From: e.From, To: e.To, Time: e.Time,
	}
	projDir := h.ProjectDir(ws, e.Project)

	// Load task title + priority
	if e.TaskID != "" && e.StoryID != "" && e.EpicID != "" {
		taskPath := filepath.Join(projDir, "epics", e.EpicID, "stories", e.StoryID, "tasks", e.TaskID+".toon")
		var task t.IssueData
		if toon.ParseFile(taskPath, &task) == nil {
			ne.TaskTitle = task.Title
			ne.Priority = task.Priority
		}
	}
	// Load story title
	if e.StoryID != "" && e.EpicID != "" {
		storyPath := filepath.Join(projDir, "epics", e.EpicID, "stories", e.StoryID, "story.toon")
		var story t.IssueData
		if toon.ParseFile(storyPath, &story) == nil {
			ne.StoryTitle = story.Title
		}
	}
	// Load epic title
	if e.EpicID != "" {
		epicPath := filepath.Join(projDir, "epics", e.EpicID, "epic.toon")
		var epic t.IssueData
		if toon.ParseFile(epicPath, &epic) == nil {
			ne.EpicTitle = epic.Title
		}
	}
	// Load project completion stats
	statusPath := filepath.Join(projDir, "project-status.toon")
	var ps t.ProjectStatus
	if toon.ParseFile(statusPath, &ps) == nil {
		total := len(ps.Tasks)
		done := 0
		for _, tk := range ps.Tasks {
			if tk.Status == "done" {
				done++
			}
		}
		ne.TotalCount = total
		ne.DoneCount = done
		if total > 0 {
			ne.CompletionPct = fmt.Sprintf("%.1f", float64(done)/float64(total)*100)
		}
	}
	return ne
}

func printUsage() {
	fmt.Print(`orchestra-mcp — AI-powered project management via Model Context Protocol

Usage:
  orchestra-mcp [flags]
  orchestra-mcp init [--workspace <path>]

Commands:
  init              Initialize MCP workspace (.mcp.json, .projects/)

Flags:
  --workspace <path>  Set workspace directory (default: ".")
  --version, -v       Print version and exit
  --help, -h          Print this help message

Examples:
  orchestra-mcp                          Start MCP server (stdio JSON-RPC)
  orchestra-mcp --workspace /my/project  Start with custom workspace
  orchestra-mcp init                     Initialize workspace in current dir
  orchestra-mcp init --workspace /path   Initialize workspace at path
`)
}
