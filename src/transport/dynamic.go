package transport

// UnregisterTool removes a tool by flat name and cleans up aliases.
func (s *MCPServer) UnregisterTool(name string) {
	delete(s.tools, name)
	for alias, flat := range s.toolAlias {
		if flat == name {
			delete(s.toolAlias, alias)
		}
	}
}

// UnregisterResource removes a resource by URI.
func (s *MCPServer) UnregisterResource(uri string) {
	delete(s.resources, uri)
}

// UnregisterPrompt removes a prompt by name.
func (s *MCPServer) UnregisterPrompt(name string) {
	delete(s.prompts, name)
}
