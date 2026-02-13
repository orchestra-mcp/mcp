package engine

import (
	"context"
	"time"

	pb "github.com/orchestra-mcp/mcp/src/gen/memoryv1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client wraps the gRPC MemoryServiceClient.
type Client struct {
	conn   *grpc.ClientConn
	memory pb.MemoryServiceClient
}

// Dial connects to the engine at the given address.
func Dial(addr string) (*Client, error) {
	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}
	return &Client{conn: conn, memory: pb.NewMemoryServiceClient(conn)}, nil
}

// Close shuts down the gRPC connection.
func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

func ctx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}

// StoreChunk stores a memory chunk via gRPC.
func (c *Client) StoreChunk(project, source, sourceID, summary, content string, tags []string) (*pb.StoreChunkResponse, error) {
	cx, cancel := ctx()
	defer cancel()
	return c.memory.StoreChunk(cx, &pb.StoreChunkRequest{
		Project: project, Source: source, SourceId: sourceID,
		Summary: summary, Content: content, Tags: tags,
	})
}

// SearchMemory searches memory via gRPC.
func (c *Client) SearchMemory(project, query string, limit int32) (*pb.SearchResponse, error) {
	cx, cancel := ctx()
	defer cancel()
	return c.memory.SearchMemory(cx, &pb.SearchRequest{
		Project: project, Query: query, Limit: limit,
	})
}

// GetContext returns relevant context chunks via gRPC.
func (c *Client) GetContext(project, query string, limit int32) (*pb.ContextResponse, error) {
	cx, cancel := ctx()
	defer cancel()
	return c.memory.GetContext(cx, &pb.ContextRequest{
		Project: project, Query: query, Limit: limit,
	})
}

// StoreSession stores a session log via gRPC.
func (c *Client) StoreSession(project, sessionID, summary string, events []*pb.SessionEvent) (*pb.StoreSessionResponse, error) {
	cx, cancel := ctx()
	defer cancel()
	return c.memory.StoreSession(cx, &pb.StoreSessionRequest{
		Project: project, SessionId: sessionID, Summary: summary, Events: events,
	})
}

// ListSessions returns recent sessions via gRPC.
func (c *Client) ListSessions(project string, limit int32) (*pb.ListSessionsResponse, error) {
	cx, cancel := ctx()
	defer cancel()
	return c.memory.ListSessions(cx, &pb.ListSessionsRequest{
		Project: project, Limit: limit,
	})
}

// GetSession retrieves a full session via gRPC.
func (c *Client) GetSession(project, sessionID string) (*pb.GetSessionResponse, error) {
	cx, cancel := ctx()
	defer cancel()
	return c.memory.GetSession(cx, &pb.GetSessionRequest{
		Project: project, SessionId: sessionID,
	})
}
