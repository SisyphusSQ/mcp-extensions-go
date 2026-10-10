package mentions_test

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/SisyphusSQ/mcp-extensions-go/internal/testutil"
	"github.com/SisyphusSQ/mcp-extensions-go/mentions"
	"github.com/SisyphusSQ/mcp-extensions-go/settings"
)

func TestCapabilityComposesWithSettings(t *testing.T) {
	for _, version := range []string{"2026-07-28", "2025-11-25"} {
		t.Run(version, func(t *testing.T) {
			base := &mcp.ServerOptions{Capabilities: &mcp.ServerCapabilities{Extensions: map[string]any{"vendor/example": map[string]any{"value": "original"}}}}
			opts, err := mentions.WithCapability(base, "parts.search")
			if err != nil {
				t.Fatal(err)
			}
			base.Capabilities.Extensions["vendor/example"].(map[string]any)["value"] = "changed"
			server, err := settings.NewServer(&mcp.Implementation{Name: "test", Version: "0"}, opts, settings.Config{
				Fields: map[string]settings.Field{"grid": {Type: "boolean", Title: "Grid"}},
				Read: func(context.Context, *mcp.CallToolRequest) (settings.Values, error) {
					return settings.Values{"grid": false}, nil
				},
				Update: func(context.Context, *mcp.CallToolRequest, settings.Values) (settings.Values, error) {
					return settings.Values{"grid": true}, nil
				},
			})
			if err != nil {
				t.Fatal(err)
			}
			if err := mentions.AddTool(server, &mcp.Tool{Name: "parts.search"}, func(context.Context, *mcp.CallToolRequest, mentions.SearchParams) (mentions.SearchResult, error) {
				return mentions.SearchResult{}, nil
			}); err != nil {
				t.Fatal(err)
			}
			cs := testutil.Connect(t, server, nil, version)
			caps := cs.InitializeResult().Capabilities
			for _, declarations := range []map[string]any{caps.Extensions, caps.Experimental} {
				if declarations[mentions.CapabilityKey].(map[string]any)["searchTool"] != "parts.search" || declarations[settings.CapabilityKey] == nil {
					t.Fatalf("capabilities = %#v", declarations)
				}
			}
			if caps.Extensions["vendor/example"].(map[string]any)["value"] != "original" {
				t.Fatal("retained caller-owned capabilities")
			}
			if _, exists := base.Capabilities.Extensions[mentions.CapabilityKey]; exists {
				t.Fatal("mutated caller options")
			}
			tools, err := cs.ListTools(t.Context(), nil)
			if err != nil {
				t.Fatal(err)
			}
			for _, tool := range tools.Tools {
				if tool.Name == "parts.search" && (!tool.Annotations.ReadOnlyHint || tool.Meta["ui"].(map[string]any)["visibility"].([]any)[0] != "app") {
					t.Fatal("missing search annotations")
				}
			}
		})
	}
}

func TestInvalidCapability(t *testing.T) {
	if _, err := mentions.WithCapability(nil, " "); err == nil {
		t.Fatal("blank search tool accepted")
	}
	for _, capabilities := range []*mcp.ServerCapabilities{
		{Extensions: map[string]any{mentions.CapabilityKey: map[string]any{}}},
		{Experimental: map[string]any{mentions.CapabilityKey: map[string]any{}}},
		{Extensions: map[string]any{"vendor": make(chan int)}},
	} {
		if _, err := mentions.WithCapability(&mcp.ServerOptions{Capabilities: capabilities}, "search"); err == nil {
			t.Fatal("invalid capabilities accepted")
		}
	}
}
