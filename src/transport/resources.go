package transport

import (
	"encoding/json"
	"strings"

	"github.com/orchestra-mcp/mcp/src/types"
)

// RegisterResource adds a single resource to the server.
func (s *MCPServer) RegisterResource(r types.Resource) {
	s.resources[r.Definition.URI] = r
}

// RegisterResources adds multiple resources to the server.
func (s *MCPServer) RegisterResources(resources []types.Resource) {
	for _, r := range resources {
		s.RegisterResource(r)
	}
}

// getResourceDefs returns all registered resource definitions.
func (s *MCPServer) getResourceDefs() []types.ResourceDefinition {
	defs := make([]types.ResourceDefinition, 0, len(s.resources))
	for _, r := range s.resources {
		defs = append(defs, r.Definition)
	}
	return defs
}

func (s *MCPServer) handleResourceReadW(req *types.JSONRPCRequest, w ResponseWriter) {
	var params types.ReadResourceParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		w.WriteError(req.ID, -32602, "invalid params")
		return
	}
	resource, ok := s.findResource(params.URI)
	if !ok {
		w.WriteError(req.ID, -32601, "unknown resource: "+params.URI)
		return
	}
	contents, err := resource.Handler(params.URI)
	if err != nil {
		w.WriteError(req.ID, -32000, err.Error())
		return
	}
	_ = w.WriteResult(req.ID, types.ReadResourceResult{Contents: contents})
}

// findResource looks up a resource by exact URI match or template pattern.
func (s *MCPServer) findResource(uri string) (types.Resource, bool) {
	if r, ok := s.resources[uri]; ok {
		return r, true
	}
	for pattern, r := range s.resources {
		if matchPattern(pattern, uri) {
			return r, true
		}
	}
	return types.Resource{}, false
}

// matchPattern checks if a URI matches a pattern with {param} placeholders.
func matchPattern(pattern, uri string) bool {
	pp := strings.Split(pattern, "/")
	up := strings.Split(uri, "/")
	if len(pp) != len(up) {
		return false
	}
	for i, seg := range pp {
		if strings.HasPrefix(seg, "{") && strings.HasSuffix(seg, "}") {
			continue
		}
		if seg != up[i] {
			return false
		}
	}
	return true
}
