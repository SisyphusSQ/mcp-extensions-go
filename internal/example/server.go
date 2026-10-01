// Package example provides shared business handlers for the runnable examples.
package example

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/SisyphusSQ/mcp-extensions-go/mentions"
	"github.com/SisyphusSQ/mcp-extensions-go/resources"
	"github.com/SisyphusSQ/mcp-extensions-go/settings"
	"github.com/SisyphusSQ/mcp-extensions-go/ui"
)

// AppURI is the shared workspace App resource.
const AppURI = "ui://workspace/home.html"

type openInput struct{}
type openOutput struct {
	Message string `json:"message" jsonschema:"The message shown on the app home page"`
}

// NewServer creates the official server with one caller-owned settings store.
// The example is single-owner. Authentication belongs to the HTTP boundary or
// to the local stdio parent's process control, not to UI visibility metadata.
func NewServer(html string, store *Store) (*mcp.Server, error) {
	if store == nil {
		return nil, fmt.Errorf("settings store is required")
	}
	server, err := settings.NewServer(&mcp.Implementation{Name: "mcp-extensions-go-example", Version: "0.0.0-dev"}, nil, settings.Config{
		Fields: map[string]settings.Field{
			"units":    {Type: "string", Title: "Measurement units", Enum: []string{"mm", "in"}},
			"showGrid": {Type: "boolean", Title: "Show grid"},
		},
		Layout: []settings.Group{{Kind: "group", Title: "Display", Items: []settings.Item{{Kind: "property", Property: "units"}, {Kind: "property", Property: "showGrid"}}}},
		Read:   func(ctx context.Context, _ *mcp.CallToolRequest) (settings.Values, error) { return store.Read(ctx) },
		Update: func(ctx context.Context, _ *mcp.CallToolRequest, set settings.Values) (settings.Values, error) {
			return store.Update(ctx, set)
		},
	})
	if err != nil {
		return nil, err
	}
	if err := ui.AddHTMLResource(server, &mcp.Resource{
		URI: AppURI, Name: "workspace-home", Title: "Workspace", Description: "App home page",
	}, html, ui.ResourceMetadata{
		AvailableDisplayModes: []ui.DisplayMode{ui.Inline, ui.Fullscreen},
		PreferredDisplayMode:  ui.Inline,
	}); err != nil {
		return nil, fmt.Errorf("register app resource: %w", err)
	}
	meta, err := (ui.ToolMetadata{
		ResourceURI:               AppURI,
		Visibility:                []ui.Visibility{ui.App, ui.Model},
		Entrypoints:               []ui.Entrypoint{{Type: ui.Global}, {Type: ui.Thread}, {Type: ui.Settings, SearchTerms: []string{"units", "grid"}}},
		PreferredModelDisplayMode: ui.Inline,
	}).Metadata(nil)
	if err != nil {
		return nil, err
	}
	mcp.AddTool(server, &mcp.Tool{
		Name: "open_workspace", Title: "Open workspace", Description: "Open the app home page", Meta: meta,
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ openInput) (*mcp.CallToolResult, openOutput, error) {
		return nil, openOutput{Message: "Welcome to your workspace"}, ctx.Err()
	})
	if err := mentions.AddTool(server, &mcp.Tool{Name: "search_mentions", Title: "Search parts"}, func(ctx context.Context, _ *mcp.CallToolRequest, input mentions.SearchParams) (mentions.SearchResult, error) {
		result := mentions.SearchResult{Items: []mentions.Item{}}
		for _, name := range []string{"bolt", "washer"} {
			if strings.Contains(name, strings.ToLower(input.Query)) {
				result.Items = append(result.Items, mentions.Item{Link: &mcp.ResourceLink{URI: "parts://" + name, Name: name, MIMEType: "text/plain"}})
			}
		}
		return result, ctx.Err()
	}); err != nil {
		return nil, err
	}
	for _, name := range []string{"bolt", "washer"} {
		server.AddResource(&mcp.Resource{URI: "parts://" + name, Name: name, MIMEType: "text/plain"}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
			return &mcp.ReadResourceResult{Contents: []*mcp.ResourceContents{{URI: req.Params.URI, MIMEType: "text/plain", Text: "Demo part: " + name}}}, ctx.Err()
		})
	}
	fileMeta, err := (ui.ToolMetadata{ResourceURI: AppURI, Visibility: []ui.Visibility{ui.App}, Entrypoints: []ui.Entrypoint{{Type: ui.File, Extensions: []string{".txt"}}}}).Metadata(nil)
	if err != nil {
		return nil, err
	}
	mcp.AddTool(server, &mcp.Tool{Name: "open_file", Title: "Text file viewer", Meta: fileMeta, Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true}}, func(ctx context.Context, _ *mcp.CallToolRequest, input resources.FileInput) (*mcp.CallToolResult, resources.FileInput, error) {
		if err := input.Validate(); err != nil {
			return nil, resources.FileInput{}, err
		}
		// Host resource reads happen through the App, never through this URI.
		return nil, input, ctx.Err()
	})
	return server, nil
}
