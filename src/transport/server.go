package transport

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	h "github.com/orchestra-mcp/mcp/src/helpers"
	"github.com/orchestra-mcp/mcp/src/types"
)

const maxScanSize = 10 * 1024 * 1024 // 10MB

// MCPServer handles JSON-RPC transport for the MCP protocol.
type MCPServer struct {
	name      string
	version   string
	tools     map[string]types.Tool
	toolAlias map[string]string // maps "ns.toolName" -> flat name
	resources map[string]types.Resource
	prompts   map[string]types.Prompt
	writer    ResponseWriter
}

// New creates an MCPServer with the given name and version.
func New(name, version string) *MCPServer {
	return &MCPServer{
		name: name, version: version,
		tools:     make(map[string]types.Tool),
		toolAlias: make(map[string]string),
		resources: make(map[string]types.Resource),
		prompts:   make(map[string]types.Prompt),
		writer:    &StdioWriter{},
	}
}

// RegisterTool adds a single tool to the server.
func (s *MCPServer) RegisterTool(t types.Tool) {
	flat := t.Definition.Name
	s.tools[flat] = t
	if t.Definition.Namespace != "" {
		s.toolAlias[t.Definition.QualifiedName()] = flat
	}
}

// RegisterTools adds multiple tools to the server.
func (s *MCPServer) RegisterTools(tools []types.Tool) {
	for _, t := range tools {
		s.RegisterTool(t)
	}
}

// GetTools returns all registered tool definitions.
func (s *MCPServer) GetTools() []types.ToolDefinition {
	defs := make([]types.ToolDefinition, 0, len(s.tools))
	for _, t := range s.tools {
		defs = append(defs, t.Definition)
	}
	return defs
}

// GetResources returns all registered resource definitions.
func (s *MCPServer) GetResources() []types.ResourceDefinition {
	return s.getResourceDefs()
}

// GetPrompts returns all registered prompt definitions.
func (s *MCPServer) GetPrompts() []types.PromptDefinition {
	return s.getPromptDefs()
}

// Run starts the stdio JSON-RPC loop.
func (s *MCPServer) Run() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 0, maxScanSize), maxScanSize)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var req types.JSONRPCRequest
		if err := json.Unmarshal(line, &req); err != nil {
			fmt.Fprintf(os.Stderr, "parse error: %v\n", err)
			continue
		}
		s.HandleRequest(&req, s.writer)
	}
}

// HandleRequest processes a JSON-RPC request using the given writer.
// Thread-safe: reads server state only, writes via the provided ResponseWriter.
func (s *MCPServer) HandleRequest(req *types.JSONRPCRequest, w ResponseWriter) {
	switch req.Method {
	case "initialize":
		caps := types.ServerCaps{Tools: &types.ToolsCap{}}
		if len(s.resources) > 0 {
			caps.Resources = &types.ResourcesCap{}
		}
		if len(s.prompts) > 0 {
			caps.Prompts = &types.PromptsCap{}
		}
		_ = w.WriteResult(req.ID, types.InitializeResult{
			ProtocolVersion: "2024-11-05",
			Capabilities:    caps,
			ServerInfo:      types.ServerInfo{Name: s.name, Version: s.version},
		})
	case "notifications/initialized":
		// no response for notifications
	case "tools/list":
		w.WriteResult(req.ID, types.ListToolsResult{Tools: s.GetTools()})
	case "tools/call":
		s.handleToolCallW(req, w)
	case "resources/list":
		w.WriteResult(req.ID, types.ListResourcesResult{Resources: s.GetResources()})
	case "resources/read":
		s.handleResourceReadW(req, w)
	case "prompts/list":
		w.WriteResult(req.ID, types.ListPromptsResult{Prompts: s.GetPrompts()})
	case "prompts/get":
		s.handlePromptGetW(req, w)
	case "ping":
		w.WriteResult(req.ID, map[string]any{})
	default:
		w.WriteError(req.ID, -32601, "method not found: "+req.Method)
	}
}

func (s *MCPServer) handleToolCallW(req *types.JSONRPCRequest, w ResponseWriter) {
	var params types.CallToolParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		w.WriteError(req.ID, -32602, "invalid params")
		return
	}
	tool, ok := s.tools[params.Name]
	if !ok {
		if flat, aliased := s.toolAlias[params.Name]; aliased {
			tool, ok = s.tools[flat]
		}
	}
	if !ok {
		w.WriteError(req.ID, -32601, "unknown tool: "+params.Name)
		return
	}
	if err := h.ValidateArgs(params.Arguments,
		tool.Definition.InputSchema.Properties,
		tool.Definition.InputSchema.Required); err != nil {
		w.WriteError(req.ID, -32602, err.Error())
		return
	}
	result, err := tool.Handler(params.Arguments)
	if err != nil {
		w.WriteError(req.ID, -32000, err.Error())
		return
	}
	w.WriteResult(req.ID, result)
}
