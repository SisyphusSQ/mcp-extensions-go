package mentions_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/SisyphusSQ/mcp-extensions-go/internal/testutil"
	"github.com/SisyphusSQ/mcp-extensions-go/mentions"
)

func TestMentionProtocolAndVisibility(t *testing.T) {
	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0"}, nil)
	base := mcp.Meta{"ui": map[string]any{"resourceUri": "ui://app/home", "visibility": []string{"model"}}, "openai/extensions": map[string]any{"vendor": true}}
	tool := &mcp.Tool{Name: "search", Meta: base}
	err := mentions.AddTool(server, tool, func(_ context.Context, _ *mcp.CallToolRequest, args mentions.SearchParams) (mentions.SearchResult, error) {
		switch args.Query {
		case "fail":
			return mentions.SearchResult{}, errors.New("search unavailable")
		case "none":
			return mentions.SearchResult{}, nil
		case "bad":
			return mentions.SearchResult{Items: []mentions.Item{{}}}, nil
		default:
			return mentions.SearchResult{Items: []mentions.Item{
				{Link: &mcp.ResourceLink{URI: "parts://bolt", Name: "bolt", Meta: mcp.Meta{"openai/title": "Bolt"}, Icons: []mcp.Icon{{Source: "https://example.com/bolt.svg"}}}},
				{Resource: &mentions.Resource{Type: "resource", ResourceURI: "parts://washer", Title: "Washer", Subtitle: "M6"}},
			}}, nil
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	base["ui"].(map[string]any)["resourceUri"] = "changed"
	cs := testutil.Connect(t, server, nil, "")
	tools, err := cs.ListTools(t.Context(), nil)
	if err != nil || len(tools.Tools) != 1 {
		t.Fatalf("tools: %#v %v", tools, err)
	}
	meta := tools.Tools[0].Meta
	app := meta["ui"].(map[string]any)
	if app["resourceUri"] != "ui://app/home" || len(app["visibility"].([]any)) != 2 || app["visibility"].([]any)[1] != "app" {
		t.Fatalf("metadata = %#v", meta)
	}
	if meta[mentions.MetadataKey].(map[string]any)[mentions.SearchKey] == nil || !tools.Tools[0].Annotations.ReadOnlyHint {
		t.Fatal("missing search marker or read-only annotation")
	}
	if _, exists := cs.InitializeResult().Capabilities.Extensions["openai/mentions"]; exists {
		t.Fatal("invented mention capability")
	}
	for _, query := range []string{"", "bolt", "none", "fail", "bad"} {
		res, err := cs.CallTool(t.Context(), &mcp.CallToolParams{Name: "search", Arguments: map[string]any{"query": query}})
		if err != nil {
			t.Fatal(err)
		}
		if query == "fail" || query == "bad" {
			if !res.IsError {
				t.Fatalf("failure hidden: %#v", res)
			}
			continue
		}
		if res.IsError || len(res.Content) != 0 {
			t.Fatalf("search: %#v", res)
		}
		items := res.StructuredContent.(map[string]any)["items"].([]any)
		if query == "none" {
			if len(items) != 0 {
				t.Fatal("empty results were not []")
			}
		} else if len(items) != 2 || items[0].(map[string]any)["type"] != "resource_link" || items[0].(map[string]any)["_meta"].(map[string]any)["openai/title"] != "Bolt" || items[1].(map[string]any)["type"] != "resource" {
			t.Fatalf("unexpected result variants: %#v", items)
		}
	}
	res, err := cs.CallTool(t.Context(), &mcp.CallToolParams{Name: "search", Arguments: map[string]any{}})
	if err != nil || !res.IsError {
		t.Fatalf("missing query accepted: %#v %v", res, err)
	}
}

func TestMentionInvalidMetadataAndItems(t *testing.T) {
	for _, meta := range []mcp.Meta{
		{"ui": false}, {"openai/extensions": []string{}},
		{"ui": map[string]any{"visibility": []string{"public"}}},
		{"ui": map[string]any{"visibility": nil}}, {"bad": make(chan int)},
	} {
		if _, err := mentions.Metadata(meta); err == nil {
			t.Fatalf("bad metadata accepted: %#v", meta)
		}
	}
	for _, item := range []mentions.Item{
		{}, {Link: &mcp.ResourceLink{}}, {Resource: &mentions.Resource{Type: "file"}},
		{Link: &mcp.ResourceLink{URI: "x", Name: "x"}, Resource: &mentions.Resource{Type: "resource", ResourceURI: "x", Title: "x"}},
	} {
		if _, err := json.Marshal(item); err == nil {
			t.Fatal("invalid item accepted")
		}
	}
	var item mentions.Item
	if err := json.Unmarshal([]byte(`{"type":"unknown"}`), &item); err == nil {
		t.Fatal("unknown result variant accepted")
	}
}
