package providers

import (
	"bufio"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/orchestra-mcp/mcp/src/transport"
	"github.com/orchestra-mcp/mcp/src/types"
)

// registerSSERoutes adds SSE transport endpoints to the MCP router.
func (p *McpPlugin) registerSSERoutes(mcp fiber.Router) {
	mgr := transport.NewSSESessionManager()

	// GET /api/mcp/sse — SSE event stream connection.
	mcp.Get("/sse", func(c fiber.Ctx) error {
		c.Set("Content-Type", "text/event-stream")
		c.Set("Cache-Control", "no-cache")
		c.Set("Connection", "keep-alive")
		c.Set("X-Accel-Buffering", "no")

		sess := mgr.Create()
		endpoint := fmt.Sprintf("/api/mcp/messages?sessionId=%s", sess.ID)

		return c.SendStreamWriter(func(w *bufio.Writer) {
			defer mgr.Remove(sess.ID)

			// Send endpoint event per MCP SSE protocol.
			fmt.Fprintf(w, "event: endpoint\ndata: %s\n\n", endpoint)
			w.Flush()

			ticker := time.NewTicker(15 * time.Second)
			defer ticker.Stop()

			for {
				select {
				case msg, ok := <-sess.Messages:
					if !ok {
						return
					}
					fmt.Fprintf(w, "event: message\ndata: %s\n\n", msg)
					w.Flush()
				case <-ticker.C:
					fmt.Fprint(w, ": ping\n\n")
					w.Flush()
				case <-sess.Context().Done():
					return
				}
			}
		})
	})

	// POST /api/mcp/messages?sessionId=xxx — Client sends JSON-RPC requests.
	mcp.Post("/messages", func(c fiber.Ctx) error {
		sessionID := c.Query("sessionId")
		if sessionID == "" {
			return c.Status(400).JSON(fiber.Map{"error": "sessionId required"})
		}
		sess, ok := mgr.Get(sessionID)
		if !ok {
			return c.Status(404).JSON(fiber.Map{"error": "session not found"})
		}
		var req types.JSONRPCRequest
		if err := json.Unmarshal(c.Body(), &req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid JSON-RPC request"})
		}
		server := p.createMCPServer()
		writer := transport.NewSSEWriter(sess)
		server.HandleRequest(&req, writer)
		return c.SendStatus(202)
	})
}

// createMCPServer builds a fresh MCPServer with all registered tools/resources/prompts.
func (p *McpPlugin) createMCPServer() *transport.MCPServer {
	server := transport.New("orchestra-mcp", p.Version())
	for _, tool := range p.allTools() {
		server.RegisterTool(tool)
	}
	for _, r := range p.allResources() {
		server.RegisterResource(r)
	}
	for _, pr := range p.allPrompts() {
		server.RegisterPrompt(pr)
	}
	return server
}
