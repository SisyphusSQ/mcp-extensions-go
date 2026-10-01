// Package testutil connects official SDK transports for protocol tests.
package testutil

import (
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Connect creates an in-memory client and registers transport cleanup.
func Connect(t *testing.T, server *mcp.Server, options *mcp.ClientOptions, version string) *mcp.ClientSession {
	t.Helper()
	clientTransport, serverTransport := mcp.NewInMemoryTransports()
	ss, err := server.Connect(t.Context(), serverTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ss.Close() })
	client := mcp.NewClient(&mcp.Implementation{Name: "protocol-test", Version: "0"}, options)
	cs, err := client.Connect(t.Context(), clientTransport, &mcp.ClientSessionOptions{ProtocolVersion: version})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cs.Close() })
	return cs
}
