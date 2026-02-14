package providers

import (
	"fmt"

	"github.com/gofiber/fiber/v3"
	"github.com/orchestra-mcp/framework/app/plugins"
	"github.com/orchestra-mcp/mcp/src/tools"
	t "github.com/orchestra-mcp/mcp/src/types"
)

// builtinResources returns all built-in MCP resources.
func (p *McpPlugin) builtinResources() []t.Resource {
	return tools.Resources(p.workspace)
}

// externalAsResources converts external McpResourceDefinitions to internal types.
func (p *McpPlugin) externalAsResources() []t.Resource {
	ext := p.ExternalResources()
	out := make([]t.Resource, len(ext))
	for i, def := range ext {
		handler := def.Handler
		out[i] = t.Resource{
			Definition: t.ResourceDefinition{
				URI: def.URI, Name: def.Name, Title: def.Title,
				Description: def.Description, MimeType: def.MimeType,
			},
			Handler: func(uri string) ([]t.ResourceContent, error) {
				contents, err := handler(uri)
				if err != nil {
					return nil, err
				}
				result := make([]t.ResourceContent, len(contents))
				for j, c := range contents {
					result[j] = t.ResourceContent{
						URI: c.URI, MimeType: c.MimeType, Text: c.Text, Blob: c.Blob,
					}
				}
				return result, nil
			},
		}
	}
	return out
}

// allResources returns built-in + external resources.
func (p *McpPlugin) allResources() []t.Resource {
	all := p.builtinResources()
	all = append(all, p.externalAsResources()...)
	return all
}

// McpResources bridges internal resources to the plugin system.
func (p *McpPlugin) McpResources() []plugins.McpResourceDefinition {
	internal := p.allResources()
	defs := make([]plugins.McpResourceDefinition, len(internal))
	for i, r := range internal {
		handler := r.Handler
		defs[i] = plugins.McpResourceDefinition{
			URI: r.Definition.URI, Name: r.Definition.Name,
			Title: r.Definition.Title, Description: r.Definition.Description,
			MimeType: r.Definition.MimeType,
			Handler: func(uri string) ([]plugins.McpResourceContent, error) {
				contents, err := handler(uri)
				if err != nil {
					return nil, err
				}
				out := make([]plugins.McpResourceContent, len(contents))
				for j, c := range contents {
					out[j] = plugins.McpResourceContent{
						URI: c.URI, MimeType: c.MimeType, Text: c.Text, Blob: c.Blob,
					}
				}
				return out, nil
			},
		}
	}
	return defs
}

// builtinPrompts returns all built-in MCP prompts.
func (p *McpPlugin) builtinPrompts() []t.Prompt {
	return tools.Prompts(p.workspace)
}

// externalAsPrompts converts external McpPromptDefinitions to internal types.
func (p *McpPlugin) externalAsPrompts() []t.Prompt {
	ext := p.ExternalPrompts()
	out := make([]t.Prompt, len(ext))
	for i, def := range ext {
		handler := def.Handler
		args := make([]t.PromptArgument, len(def.Arguments))
		for j, a := range def.Arguments {
			args[j] = t.PromptArgument{Name: a.Name, Description: a.Description, Required: a.Required}
		}
		out[i] = t.Prompt{
			Definition: t.PromptDefinition{
				Name: def.Name, Title: def.Title,
				Description: def.Description, Arguments: args,
			},
			Handler: func(a map[string]string) (string, []t.PromptMessage, error) {
				desc, msgs, err := handler(a)
				if err != nil {
					return "", nil, err
				}
				result := make([]t.PromptMessage, len(msgs))
				for k, m := range msgs {
					result[k] = t.PromptMessage{
						Role:    m.Role,
						Content: t.ContentBlock{Type: "text", Text: m.Content},
					}
				}
				return desc, result, nil
			},
		}
	}
	return out
}

// allPrompts returns built-in + external prompts.
func (p *McpPlugin) allPrompts() []t.Prompt {
	all := p.builtinPrompts()
	all = append(all, p.externalAsPrompts()...)
	return all
}

// McpPrompts bridges internal prompts to the plugin system.
func (p *McpPlugin) McpPrompts() []plugins.McpPromptDefinition {
	internal := p.allPrompts()
	defs := make([]plugins.McpPromptDefinition, len(internal))
	for i, pr := range internal {
		handler := pr.Handler
		args := make([]plugins.McpPromptArgument, len(pr.Definition.Arguments))
		for j, a := range pr.Definition.Arguments {
			args[j] = plugins.McpPromptArgument{Name: a.Name, Description: a.Description, Required: a.Required}
		}
		defs[i] = plugins.McpPromptDefinition{
			Name: pr.Definition.Name, Title: pr.Definition.Title,
			Description: pr.Definition.Description, Arguments: args,
			Handler: func(a map[string]string) (string, []plugins.McpPromptMessage, error) {
				desc, msgs, err := handler(a)
				if err != nil {
					return "", nil, err
				}
				out := make([]plugins.McpPromptMessage, len(msgs))
				for k, m := range msgs {
					out[k] = plugins.McpPromptMessage{Role: m.Role, Content: m.Content.Text}
				}
				return desc, out, nil
			},
		}
	}
	return defs
}

// registerResourcePromptRoutes adds REST API endpoints for resources and prompts.
func (p *McpPlugin) registerResourcePromptRoutes(mcp fiber.Router) {
	// GET /api/mcp/resources — list all resources.
	mcp.Get("/resources", func(c fiber.Ctx) error {
		resources := p.allResources()
		defs := make([]t.ResourceDefinition, len(resources))
		for i, r := range resources {
			defs[i] = r.Definition
		}
		return c.JSON(fiber.Map{"resources": defs, "count": len(defs)})
	})

	// POST /api/mcp/resources/read — read a resource by URI.
	mcp.Post("/resources/read", func(c fiber.Ctx) error {
		var req struct {
			URI string `json:"uri"`
		}
		if err := c.Bind().JSON(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
		}
		for _, r := range p.allResources() {
			if r.Definition.URI == req.URI || matchResourceURI(r.Definition.URI, req.URI) {
				contents, err := r.Handler(req.URI)
				if err != nil {
					return c.Status(500).JSON(fiber.Map{"error": err.Error()})
				}
				return c.JSON(fiber.Map{"contents": contents})
			}
		}
		return c.Status(404).JSON(fiber.Map{"error": "unknown resource: " + req.URI})
	})

	// GET /api/mcp/prompts — list all prompts.
	mcp.Get("/prompts", func(c fiber.Ctx) error {
		prompts := p.allPrompts()
		defs := make([]t.PromptDefinition, len(prompts))
		for i, pr := range prompts {
			defs[i] = pr.Definition
		}
		return c.JSON(fiber.Map{"prompts": defs, "count": len(defs)})
	})

	// POST /api/mcp/prompts/get — get a prompt by name.
	mcp.Post("/prompts/get", func(c fiber.Ctx) error {
		var req struct {
			Name      string            `json:"name"`
			Arguments map[string]string `json:"arguments"`
		}
		if err := c.Bind().JSON(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
		}
		for _, pr := range p.allPrompts() {
			if pr.Definition.Name == req.Name {
				desc, msgs, err := pr.Handler(req.Arguments)
				if err != nil {
					return c.Status(500).JSON(fiber.Map{"error": err.Error()})
				}
				return c.JSON(fiber.Map{"description": desc, "messages": msgs})
			}
		}
		return c.Status(404).JSON(fiber.Map{"error": fmt.Sprintf("unknown prompt: %s", req.Name)})
	})
}

// matchResourceURI checks if a concrete URI matches a template pattern.
func matchResourceURI(pattern, uri string) bool {
	pp := split(pattern, "/")
	up := split(uri, "/")
	if len(pp) != len(up) {
		return false
	}
	for i, seg := range pp {
		if len(seg) > 2 && seg[0] == '{' && seg[len(seg)-1] == '}' {
			continue
		}
		if seg != up[i] {
			return false
		}
	}
	return true
}

func split(s, sep string) []string {
	var parts []string
	for s != "" {
		i := indexOf(s, sep)
		if i < 0 {
			parts = append(parts, s)
			break
		}
		parts = append(parts, s[:i])
		s = s[i+len(sep):]
	}
	return parts
}

func indexOf(s, sub string) int {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
