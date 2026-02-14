package transport

import (
	"encoding/json"

	"github.com/orchestra-mcp/mcp/src/types"
)

// SSEWriter writes JSON-RPC responses to an SSE session's message channel.
type SSEWriter struct {
	session *SSESession
}

// NewSSEWriter creates an SSE response writer for a session.
func NewSSEWriter(sess *SSESession) *SSEWriter {
	return &SSEWriter{session: sess}
}

func (w *SSEWriter) WriteResult(id, result any) error {
	resp := types.JSONRPCResponse{JSONRPC: "2.0", ID: id, Result: result}
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	return w.send(data)
}

func (w *SSEWriter) WriteError(id any, code int, msg string) error {
	resp := types.JSONRPCResponse{
		JSONRPC: "2.0", ID: id,
		Error: &types.JSONRPCError{Code: code, Message: msg},
	}
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	return w.send(data)
}

func (w *SSEWriter) send(data []byte) error {
	select {
	case w.session.Messages <- data:
		return nil
	case <-w.session.Context().Done():
		return w.session.Context().Err()
	}
}
