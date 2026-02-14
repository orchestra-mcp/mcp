package transport

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/orchestra-mcp/mcp/src/types"
)

// ResponseWriter abstracts how JSON-RPC responses are sent.
type ResponseWriter interface {
	WriteResult(id, result any) error
	WriteError(id any, code int, msg string) error
}

// StdioWriter writes JSON-RPC responses to stdout (one line per response).
type StdioWriter struct{}

func (w *StdioWriter) WriteResult(id, result any) error {
	resp := types.JSONRPCResponse{JSONRPC: "2.0", ID: id, Result: result}
	data, _ := json.Marshal(resp)
	_, err := fmt.Fprintf(os.Stdout, "%s\n", data)
	return err
}

func (w *StdioWriter) WriteError(id any, code int, msg string) error {
	resp := types.JSONRPCResponse{
		JSONRPC: "2.0", ID: id,
		Error: &types.JSONRPCError{Code: code, Message: msg},
	}
	data, _ := json.Marshal(resp)
	_, err := fmt.Fprintf(os.Stdout, "%s\n", data)
	return err
}
