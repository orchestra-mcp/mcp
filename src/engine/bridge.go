package engine

// Bridge provides memory operations with automatic gRPC/TOON fallback.
type Bridge struct {
	Client *Client // nil if engine unavailable
	Ws     string  // workspace root for TOON fallback
}

// NewBridge creates a bridge. If client is nil, all ops use TOON.
func NewBridge(client *Client, ws string) *Bridge {
	return &Bridge{Client: client, Ws: ws}
}

// UsingEngine returns true if the bridge has an active gRPC connection.
func (b *Bridge) UsingEngine() bool {
	return b.Client != nil
}
