package settings_test

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/SisyphusSQ/mcp-extensions-go/internal/testutil"
	"github.com/SisyphusSQ/mcp-extensions-go/settings"
)

func TestToolLayoutTitleIsOptional(t *testing.T) {
	for _, title := range []string{"", "Legacy title"} {
		t.Run(title, func(t *testing.T) {
			cfg := config()
			cfg.Read = func(context.Context, *mcp.CallToolRequest) (settings.Values, error) {
				return settings.Values{"units": "mm", "grid": false}, nil
			}
			cfg.Update = func(context.Context, *mcp.CallToolRequest, settings.Values) (settings.Values, error) {
				return settings.Values{"units": "mm", "grid": false}, nil
			}
			cfg.Layout = []settings.Group{{Kind: "group", Title: "Display", Items: []settings.Item{{Kind: "tool", Tool: "open", Title: title}}}}
			server, err := settings.NewServer(&mcp.Implementation{Name: "test", Version: "0"}, nil, cfg)
			if err != nil {
				t.Fatal(err)
			}
			mcp.AddTool(server, &mcp.Tool{Name: "open", Title: "Tool title"}, func(context.Context, *mcp.CallToolRequest, struct{}) (*mcp.CallToolResult, any, error) {
				return nil, nil, nil
			})
			cs := testutil.Connect(t, server, nil, "2026-07-28")
			result, err := cs.CallTool(t.Context(), &mcp.CallToolParams{Name: "settings.read", Arguments: map[string]any{}})
			if err != nil || result.IsError {
				t.Fatalf("read: %#v %v", result, err)
			}
			item := result.StructuredContent.(map[string]any)["layout"].([]any)[0].(map[string]any)["items"].([]any)[0].(map[string]any)
			if item["tool"] != "open" {
				t.Fatal("lost tool reference")
			}
			if _, exists := item["title"]; title == "" && exists {
				t.Fatal("empty deprecated title emitted")
			}
		})
	}
}
