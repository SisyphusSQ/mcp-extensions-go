// Package mentions registers app-visible composer search tools.
package mentions

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/SisyphusSQ/mcp-extensions-go/internal/wire"
)

// MetadataKey identifies OpenAI tool extensions. SearchKey marks a mention tool.
const MetadataKey = "openai/extensions"
const SearchKey = "mentions/search"

// SearchParams contains typeahead text; an empty query is valid.
type SearchParams struct {
	Query string `json:"query"`
}

// Resource is the SDK's selectable resource variant. The spec also permits
// standard MCP resource_link items, represented by mcp.ResourceLink below.
type Resource struct {
	Type        string     `json:"type"`
	ResourceURI string     `json:"resourceUri"`
	Title       string     `json:"title"`
	Subtitle    string     `json:"subtitle,omitempty"`
	Icons       []mcp.Icon `json:"icons,omitempty"`
}

// Item holds exactly one supported result variant.
type Item struct {
	Link     *mcp.ResourceLink
	Resource *Resource
}

// MarshalJSON encodes a discriminated mention item, rejecting invalid variants.
func (i Item) MarshalJSON() ([]byte, error) {
	if (i.Link == nil) == (i.Resource == nil) {
		return nil, fmt.Errorf("mention item requires exactly one variant")
	}
	if i.Link != nil {
		if strings.TrimSpace(i.Link.URI) == "" || i.Link.Name == "" {
			return nil, fmt.Errorf("resource link requires uri and name")
		}
		return json.Marshal(i.Link)
	}
	r := i.Resource
	if r.Type != "resource" || strings.TrimSpace(r.ResourceURI) == "" || strings.TrimSpace(r.Title) == "" || r.Subtitle != "" && strings.TrimSpace(r.Subtitle) == "" {
		return nil, fmt.Errorf("mention resource requires type, resourceUri and title")
	}
	for _, icon := range r.Icons {
		if icon.Source == "" {
			return nil, fmt.Errorf("mention icon requires src")
		}
	}
	return json.Marshal(r)
}

// UnmarshalJSON decodes the two upstream result variants.
func (i *Item) UnmarshalJSON(encoded []byte) error {
	var discriminator struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(encoded, &discriminator); err != nil {
		return err
	}
	var item Item
	switch discriminator.Type {
	case "resource_link":
		var result mcp.CallToolResult
		// The official content decoder preserves _meta and all MCP link fields.
		if err := json.Unmarshal([]byte(`{"content":[`+string(encoded)+`]}`), &result); err != nil {
			return err
		}
		item.Link = result.Content[0].(*mcp.ResourceLink)
	case "resource":
		item.Resource = new(Resource)
		if err := json.Unmarshal(encoded, item.Resource); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported mention type %q", discriminator.Type)
	}
	if _, err := item.MarshalJSON(); err != nil {
		return err
	}
	*i = item
	return nil
}

// SearchResult is returned in structuredContent; an empty search returns [].
type SearchResult struct {
	Items []Item `json:"items"`
}

// Handler implements business search and authorization. There is no built-in
// database or retry policy. Errors must be safe for client display.
type Handler func(context.Context, *mcp.CallToolRequest, SearchParams) (SearchResult, error)

// Metadata marks a mention tool and ensures app visibility, preserving other
// fields in ui and openai/extensions. The marker is a deprecated fallback used
// when the server has no openai/mentions capability; prefer WithCapability.
// Visibility is not authorization. The returned value is an independent snapshot.
func Metadata(base mcp.Meta) (mcp.Meta, error) {
	meta, err := wire.Clone(base)
	if err != nil {
		return nil, fmt.Errorf("mention metadata: %w", err)
	}
	if meta == nil {
		meta = make(mcp.Meta)
	}
	namespace := func(key string) (map[string]any, error) {
		if value, exists := meta[key]; exists {
			object, ok := value.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("%s metadata must be an object", key)
			}
			return object, nil
		}
		return make(map[string]any), nil
	}
	app, err := namespace("ui")
	if err != nil {
		return nil, err
	}
	visibility := []string{"app"}
	if value, exists := app["visibility"]; exists {
		if err := convertVisibility(value, &visibility); err != nil {
			return nil, err
		}
	}
	app["visibility"] = visibility
	extensions, err := namespace(MetadataKey)
	if err != nil {
		return nil, err
	}
	extensions[SearchKey] = map[string]any{}
	meta["ui"] = app
	meta[MetadataKey] = extensions
	return meta, nil
}

func convertVisibility(value any, visibility *[]string) error {
	values, err := wire.Convert[[]string](value)
	if err != nil || values == nil {
		return fmt.Errorf("mention visibility must be an array of app/model")
	}
	for _, v := range values {
		if v != "app" && v != "model" {
			return fmt.Errorf("invalid tool visibility %q", v)
		}
	}
	if !slices.Contains(values, "app") {
		values = append(values, "app")
	}
	*visibility = values
	return nil
}

// AddTool registers a real search handler using official schema validation.
// Supply an unused tool name before connecting; SDK replacement semantics apply.
func AddTool(server *mcp.Server, tool *mcp.Tool, handler Handler) error {
	if server == nil || tool == nil || strings.TrimSpace(tool.Name) == "" || handler == nil {
		return fmt.Errorf("server, named tool and search handler are required")
	}
	registered := *tool
	meta, err := Metadata(tool.Meta)
	if err != nil {
		return err
	}
	registered.Meta = meta
	annotations := mcp.ToolAnnotations{ReadOnlyHint: true}
	if tool.Annotations != nil {
		annotations = *tool.Annotations
		annotations.ReadOnlyHint = true
	}
	registered.Annotations = &annotations
	// The result union uses official ResourceLink JSON encoding, not SDK reflection.
	registered.OutputSchema = map[string]any{"type": "object", "properties": map[string]any{
		"items": map[string]any{"type": "array", "items": map[string]any{"type": "object"}},
	}, "required": []string{"items"}, "additionalProperties": false}
	registered.InputSchema = map[string]any{"type": "object", "properties": map[string]any{
		"query": map[string]any{"type": "string"},
	}, "required": []string{"query"}}
	mcp.AddTool(server, &registered, func(ctx context.Context, req *mcp.CallToolRequest, input SearchParams) (*mcp.CallToolResult, SearchResult, error) {
		if err := ctx.Err(); err != nil {
			return nil, SearchResult{}, err
		}
		result, err := handler(ctx, req, input)
		if err != nil {
			return nil, SearchResult{}, err
		}
		if result.Items == nil {
			result.Items = []Item{}
		}
		result, err = wire.Clone(result)
		return &mcp.CallToolResult{Content: []mcp.Content{}}, result, err
	})
	return nil
}
