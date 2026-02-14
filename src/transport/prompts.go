package transport

import (
	"encoding/json"

	"github.com/orchestra-mcp/mcp/src/types"
)

// RegisterPrompt adds a single prompt to the server.
func (s *MCPServer) RegisterPrompt(p types.Prompt) {
	s.prompts[p.Definition.Name] = p
}

// RegisterPrompts adds multiple prompts to the server.
func (s *MCPServer) RegisterPrompts(prompts []types.Prompt) {
	for _, p := range prompts {
		s.RegisterPrompt(p)
	}
}

// getPromptDefs returns all registered prompt definitions.
func (s *MCPServer) getPromptDefs() []types.PromptDefinition {
	defs := make([]types.PromptDefinition, 0, len(s.prompts))
	for _, p := range s.prompts {
		defs = append(defs, p.Definition)
	}
	return defs
}

func (s *MCPServer) handlePromptGetW(req *types.JSONRPCRequest, w ResponseWriter) {
	var params types.GetPromptParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		_ = w.WriteError(req.ID, -32602, "invalid params")
		return
	}
	prompt, ok := s.prompts[params.Name]
	if !ok {
		_ = w.WriteError(req.ID, -32601, "unknown prompt: "+params.Name)
		return
	}
	description, messages, err := prompt.Handler(params.Arguments)
	if err != nil {
		_ = w.WriteError(req.ID, -32000, err.Error())
		return
	}
	_ = w.WriteResult(req.ID, types.GetPromptResult{
		Description: description,
		Messages:    messages,
	})
}
