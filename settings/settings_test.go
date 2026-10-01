package settings_test

import (
	"context"
	"errors"
	"maps"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/SisyphusSQ/mcp-extensions-go/internal/testutil"
	"github.com/SisyphusSQ/mcp-extensions-go/settings"
)

func config() settings.Config {
	return settings.Config{Fields: map[string]settings.Field{
		"units": {Type: "string", Title: "Units", Enum: []string{"mm", "in"}},
		"grid":  {Type: "boolean", Title: "Grid"},
	}}
}

func TestSettingsDiscoveryAndProtocol(t *testing.T) {
	for _, version := range []string{"2026-07-28", "2025-11-25"} {
		t.Run(version, func(t *testing.T) {
			cfg := config()
			cfg.Layout = []settings.Group{{Kind: "group", Title: "Display", Items: []settings.Item{{Kind: "property", Property: "units"}}}}
			values := settings.Values{"units": "mm", "grid": false}
			var mu sync.Mutex
			var updates atomic.Int64
			cfg.Read = func(context.Context, *mcp.CallToolRequest) (settings.Values, error) {
				mu.Lock()
				defer mu.Unlock()
				return maps.Clone(values), nil
			}
			cfg.Update = func(_ context.Context, _ *mcp.CallToolRequest, set settings.Values) (settings.Values, error) {
				updates.Add(1)
				mu.Lock()
				defer mu.Unlock()
				maps.Copy(values, set)
				return maps.Clone(values), nil
			}
			options := &mcp.ServerOptions{Capabilities: &mcp.ServerCapabilities{Extensions: map[string]any{"vendor/example": map[string]any{}}}}
			server, err := settings.NewServer(&mcp.Implementation{Name: "settings", Version: "0"}, options, cfg)
			if err != nil {
				t.Fatal(err)
			}
			if _, exists := options.Capabilities.Extensions[settings.CapabilityKey]; exists {
				t.Fatal("mutated caller capabilities")
			}
			cfg.Fields["units"] = settings.Field{Type: "boolean", Title: "Changed"}
			cfg.Layout[0].Items[0].Property = "changed"
			cs := testutil.Connect(t, server, nil, version)
			cap := cs.InitializeResult().Capabilities.Extensions[settings.CapabilityKey].(map[string]any)
			if cap["readTool"] != "settings.read" || cap["updateTool"] != "settings.update" {
				t.Fatalf("bad capability: %#v", cap)
			}
			legacy := cs.InitializeResult().Capabilities.Experimental[settings.CapabilityKey].(map[string]any)
			if legacy["readTool"] != "settings.read" {
				t.Fatal("missing legacy capability")
			}
			tools, err := cs.ListTools(t.Context(), nil)
			if err != nil || len(tools.Tools) != 2 {
				t.Fatalf("tools: %#v, %v", tools, err)
			}
			for _, tool := range tools.Tools {
				if tool.OutputSchema == nil {
					t.Fatal("missing outputSchema")
				}
				if tool.Name == "settings.read" && !tool.Annotations.ReadOnlyHint {
					t.Fatal("read must be read-only")
				}
			}
			call := func(name string, args any) *mcp.CallToolResult {
				t.Helper()
				res, err := cs.CallTool(t.Context(), &mcp.CallToolParams{Name: name, Arguments: args})
				if err != nil {
					t.Fatal(err)
				}
				return res
			}
			read := call("settings.read", map[string]any{})
			if read.IsError || len(read.Content) != 0 {
				t.Fatalf("read = %#v", read)
			}
			got := read.StructuredContent.(map[string]any)
			if got["schema"].(map[string]any)["properties"].(map[string]any)["units"].(map[string]any)["type"] != "string" {
				t.Fatal("schema retained caller aliases")
			}
			if got["layout"].([]any)[0].(map[string]any)["items"].([]any)[0].(map[string]any)["property"] != "units" {
				t.Fatal("layout retained caller aliases")
			}
			updated := call("settings.update", map[string]any{"set": map[string]any{"grid": true}})
			if updated.IsError {
				t.Fatalf("update = %#v", updated)
			}
			all := updated.StructuredContent.(map[string]any)["values"].(map[string]any)
			if all["units"] != "mm" || all["grid"] != true {
				t.Fatalf("partial update lost values: %#v", all)
			}
			for _, args := range []any{
				map[string]any{}, map[string]any{"set": map[string]any{}},
				map[string]any{"set": map[string]any{"unknown": true}},
				map[string]any{"set": map[string]any{"units": "ft"}},
				map[string]any{"set": map[string]any{"grid": "true"}},
				map[string]any{"set": map[string]any{"grid": nil}},
				map[string]any{"set": map[string]any{"grid": false}, "extra": true},
			} {
				if !call("settings.update", args).IsError {
					t.Fatalf("invalid patch accepted: %#v", args)
				}
			}
			if updates.Load() != 1 {
				t.Fatal("invalid input reached storage")
			}
			if !call("settings.read", map[string]any{"extra": true}).IsError {
				t.Fatal("read accepted unknown arguments")
			}
		})
	}
}

func TestSettingsFailuresAndNumericConstraints(t *testing.T) {
	cfg := config()
	minimum, maximum, multiple := 1.0, 10.0, 2.0
	minLength, maxLength := 2, 8
	cfg.Fields["size"] = settings.Field{Type: "integer", Title: "Size", Minimum: &minimum, Maximum: &maximum, MultipleOf: &multiple}
	cfg.Fields["label"] = settings.Field{Type: "string", Title: "Label", MinLength: &minLength, MaxLength: &maxLength, Pattern: "^[a-z]+$"}
	values := settings.Values{"units": "in", "grid": true, "size": 4, "label": "part"}
	cfg.Read = func(context.Context, *mcp.CallToolRequest) (settings.Values, error) { return values, nil }
	cfg.Update = func(context.Context, *mcp.CallToolRequest, settings.Values) (settings.Values, error) {
		return nil, errors.New("persistence failed")
	}
	server, err := settings.NewServer(&mcp.Implementation{Name: "test", Version: "0"}, nil, cfg)
	if err != nil {
		t.Fatal(err)
	}
	cs := testutil.Connect(t, server, nil, "")
	read := func() *mcp.CallToolResult {
		t.Helper()
		res, err := cs.CallTool(t.Context(), &mcp.CallToolParams{Name: "settings.read", Arguments: map[string]any{}})
		if err != nil {
			t.Fatal(err)
		}
		return res
	}
	if result := read(); result.IsError {
		t.Fatalf("valid numeric state rejected: %#v", result.Content[0])
	}
	for _, n := range []any{0, 3, 12, 2.5} {
		values["size"] = n
		if !read().IsError {
			t.Fatalf("invalid integer state accepted: %v", n)
		}
	}
	values["size"] = 4
	delete(values, "grid")
	if !read().IsError {
		t.Fatal("missing effective value accepted")
	}
	res, err := cs.CallTool(t.Context(), &mcp.CallToolParams{Name: "settings.update", Arguments: map[string]any{"set": map[string]any{"size": 6}}})
	if err != nil || !res.IsError || res.StructuredContent != nil {
		t.Fatalf("persistence error reported as success: %#v %v", res, err)
	}
}

func TestSettingsRejectsBadDefinitions(t *testing.T) {
	for _, mutate := range []func(*settings.Config){
		func(c *settings.Config) { c.UpdateTool, c.ReadTool = "same", "same" },
		func(c *settings.Config) { c.Fields["bad"] = settings.Field{Type: "object", Title: "Bad"} },
		func(c *settings.Config) {
			c.Fields["bad"] = settings.Field{Type: "boolean", Title: "Bad", Enum: []string{"a"}}
		},
		func(c *settings.Config) {
			c.Fields["bad"] = settings.Field{Type: "string", Title: "Bad", Enum: []string{}}
		},
		func(c *settings.Config) { c.Fields["bad"] = settings.Field{Type: "string", Title: " "} },
		func(c *settings.Config) { c.Fields["bad"] = settings.Field{Type: "string", Title: "Bad", Pattern: "["} },
		func(c *settings.Config) {
			c.Layout = []settings.Group{{Kind: "group", Title: "Test", Items: []settings.Item{{Kind: "property", Property: "missing"}}}}
		},
		func(c *settings.Config) {
			c.Layout = []settings.Group{{Kind: "group", Title: "Test", Items: []settings.Item{{Kind: "property", Property: "units"}, {Kind: "property", Property: "units"}}}}
		},
		func(c *settings.Config) {
			c.Layout = []settings.Group{{Kind: "group", Title: "Test", Items: []settings.Item{{Kind: "property", Property: "units", Tool: "bad"}}}}
		},
	} {
		cfg := config()
		cfg.Read = func(context.Context, *mcp.CallToolRequest) (settings.Values, error) { return nil, nil }
		cfg.Update = func(context.Context, *mcp.CallToolRequest, settings.Values) (settings.Values, error) { return nil, nil }
		mutate(&cfg)
		if _, err := settings.NewServer(&mcp.Implementation{Name: "test", Version: "0"}, nil, cfg); !errors.Is(err, settings.ErrInvalidDefinition) {
			t.Fatalf("invalid definition accepted: %v", err)
		}
	}
}
