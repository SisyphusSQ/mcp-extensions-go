// Package settings registers OpenAI native settings tools on the official SDK.
package settings

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/SisyphusSQ/mcp-extensions-go/internal/wire"
)

// CapabilityKey identifies the structured settings extension.
const CapabilityKey = "openai/settings"

var (
	// ErrInvalidDefinition indicates an invalid schema, layout, or registration.
	ErrInvalidDefinition = errors.New("invalid settings definition")
	// ErrInvalidValues indicates incomplete or invalid values returned by a handler.
	ErrInvalidValues = errors.New("invalid settings values")
)

// Capability names the same-server settings tools.
type Capability struct {
	ReadTool   string `json:"readTool"`
	UpdateTool string `json:"updateTool"`
}

// Field is a native primitive setting. Defaults belong in Read, not the schema.
// String constraints apply only to strings; numeric constraints only to numbers.
type Field struct {
	Type        string   `json:"type"`
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	Enum        []string `json:"enum,omitempty"`
	MinLength   *int     `json:"minLength,omitempty"`
	MaxLength   *int     `json:"maxLength,omitempty"`
	Pattern     string   `json:"pattern,omitempty"`
	Minimum     *float64 `json:"minimum,omitempty"`
	Maximum     *float64 `json:"maximum,omitempty"`
	MultipleOf  *float64 `json:"multipleOf,omitempty"`
}

// Schema describes all settings fields. Every field needs an effective value.
type Schema struct {
	Type       string           `json:"type"`
	Properties map[string]Field `json:"properties"`
	Required   []string         `json:"required,omitempty"`
}

// Item is a property reference or a same-server tool accepting {}.
type Item struct {
	Kind        string `json:"kind"`
	Property    string `json:"property,omitempty"`
	Tool        string `json:"tool,omitempty"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
}

// MarshalJSON preserves an empty property key in the property variant.
func (i Item) MarshalJSON() ([]byte, error) {
	if i.Kind == "property" {
		return json.Marshal(map[string]string{"kind": i.Kind, "property": i.Property})
	}
	type plain Item
	return json.Marshal(plain(i))
}

// Group is an ordered, non-nested settings section.
type Group struct {
	Kind  string `json:"kind"`
	Title string `json:"title"`
	Items []Item `json:"items"`
}

// Values contains effective values or a partial replacement patch.
type Values = map[string]any

// ReadResult is the structured content of the read tool.
type ReadResult struct {
	Schema Schema  `json:"schema"`
	Values Values  `json:"values"`
	Layout []Group `json:"layout,omitempty"`
}

// UpdateArguments replaces only supplied fields; it does not reset omitted fields.
type UpdateArguments struct {
	Set Values `json:"set"`
}

// UpdateResult contains all values after successful persistence.
type UpdateResult struct {
	Values Values `json:"values"`
}

// Config binds a static declaration to caller-owned storage and authorization.
// Read must be read-only. Update must authorize, preserve omitted fields, check
// cross-field rules, and persist atomically before returning all effective values.
// Handlers may run concurrently and must return errors safe for client display.
type Config struct {
	ReadTool   string
	UpdateTool string
	Fields     map[string]Field
	Layout     []Group
	Read       func(context.Context, *mcp.CallToolRequest) (Values, error)
	Update     func(context.Context, *mcp.CallToolRequest, Values) (Values, error)
}

// NewServer constructs an official MCP server with both settings tools and their
// capability in modern and legacy discovery. It registers before connecting.
// Other tools may be added normally; callers must not replace the settings names.
// It copies definitions and options without calling storage handlers at startup.
func NewServer(info *mcp.Implementation, options *mcp.ServerOptions, config Config) (*mcp.Server, error) {
	if info == nil || config.Read == nil || config.Update == nil {
		return nil, fmt.Errorf("%w: implementation and both handlers are required", ErrInvalidDefinition)
	}
	if config.ReadTool == "" {
		config.ReadTool = "settings.read"
	}
	if config.UpdateTool == "" {
		config.UpdateTool = "settings.update"
	}
	if strings.TrimSpace(config.ReadTool) == "" || strings.TrimSpace(config.UpdateTool) == "" || config.ReadTool == config.UpdateTool {
		return nil, fmt.Errorf("%w: distinct non-blank tool names are required", ErrInvalidDefinition)
	}
	schema, layout, resolved, err := prepare(config.Fields, config.Layout)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidDefinition, err)
	}
	var opts mcp.ServerOptions
	if options != nil {
		opts = *options
	}
	caps, err := wire.Clone(opts.Capabilities)
	if err != nil {
		return nil, fmt.Errorf("%w: capabilities: %v", ErrInvalidDefinition, err)
	}
	if caps == nil {
		caps = &mcp.ServerCapabilities{Logging: &mcp.LoggingCapabilities{}}
	}
	if _, exists := caps.Extensions[CapabilityKey]; exists {
		return nil, fmt.Errorf("%w: settings capability is already configured", ErrInvalidDefinition)
	}
	if _, exists := caps.Experimental[CapabilityKey]; exists {
		return nil, fmt.Errorf("%w: legacy settings capability is already configured", ErrInvalidDefinition)
	}
	capability := Capability{ReadTool: config.ReadTool, UpdateTool: config.UpdateTool}
	if caps.Extensions == nil {
		caps.Extensions = make(map[string]any)
	}
	if caps.Experimental == nil {
		caps.Experimental = make(map[string]any)
	}
	caps.Extensions[CapabilityKey] = capability
	caps.Experimental[CapabilityKey] = capability
	opts.Capabilities = caps
	server := mcp.NewServer(info, &opts)
	valuesSchema := resolved.Schema()
	readOutput := envelope(map[string]any{"schema": map[string]any{"type": "object"}, "values": valuesSchema, "layout": map[string]any{"type": "array", "items": map[string]any{"type": "object"}}}, []string{"schema", "values"})
	patch := *valuesSchema
	patch.Required = nil
	minProperties := 1
	patch.MinProperties = &minProperties
	updateInput := envelope(map[string]any{"set": &patch}, []string{"set"})
	updateOutput := envelope(map[string]any{"values": valuesSchema}, []string{"values"})
	validateValues := func(values Values) (Values, error) {
		// Validate Go primitives before snapshotting: jsonschema-go treats
		// json.Number as a string, while snapshots preserve its wire precision.
		if err := resolved.Validate(values); err != nil {
			return nil, fmt.Errorf("%w: returned state does not match the settings schema", ErrInvalidValues)
		}
		copy, err := wire.Clone(values)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrInvalidValues, err)
		}
		return copy, nil
	}
	mcp.AddTool(server, &mcp.Tool{
		Name: config.ReadTool, InputSchema: envelope(map[string]any{}, nil), OutputSchema: readOutput,
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, ReadResult, error) {
		if err := ctx.Err(); err != nil {
			return nil, ReadResult{}, err
		}
		values, err := config.Read(ctx, req)
		if err == nil {
			values, err = validateValues(values)
		}
		// Each call receives a fresh declaration, independent of returned results.
		declaration, cloneErr := wire.Clone(ReadResult{Schema: schema, Layout: layout, Values: values})
		return &mcp.CallToolResult{Content: []mcp.Content{}}, declaration, errors.Join(err, cloneErr)
	})
	mcp.AddTool(server, &mcp.Tool{
		Name: config.UpdateTool, InputSchema: updateInput, OutputSchema: updateOutput,
	}, func(ctx context.Context, req *mcp.CallToolRequest, args UpdateArguments) (*mcp.CallToolResult, UpdateResult, error) {
		if err := ctx.Err(); err != nil {
			return nil, UpdateResult{}, err
		}
		values, err := config.Update(ctx, req, args.Set)
		if err == nil {
			values, err = validateValues(values)
		}
		return &mcp.CallToolResult{Content: []mcp.Content{}}, UpdateResult{Values: values}, err
	})
	return server, nil
}

func envelope(properties map[string]any, required []string) map[string]any {
	schema := map[string]any{"type": "object", "properties": properties, "additionalProperties": false}
	if len(required) > 0 {
		schema["required"] = required
	}
	return schema
}

func prepare(fields map[string]Field, groups []Group) (Schema, []Group, *jsonschema.Resolved, error) {
	schema, err := wire.Clone(Schema{Type: "object", Properties: fields, Required: slices.Sorted(maps.Keys(fields))})
	if err != nil || fields == nil {
		return Schema{}, nil, nil, fmt.Errorf("fields must be a JSON object: %v", err)
	}
	for name, field := range schema.Properties {
		if strings.TrimSpace(field.Title) == "" {
			return Schema{}, nil, nil, fmt.Errorf("field %q requires a title", name)
		}
		hasString := fields[name].Enum != nil || field.MinLength != nil || field.MaxLength != nil || field.Pattern != ""
		hasNumber := field.Minimum != nil || field.Maximum != nil || field.MultipleOf != nil
		if field.Type != "boolean" && field.Type != "string" && field.Type != "number" && field.Type != "integer" || hasString && field.Type != "string" || hasNumber && field.Type != "number" && field.Type != "integer" {
			return Schema{}, nil, nil, fmt.Errorf("field %q has unsupported primitive type or constraints", name)
		}
		if fields[name].Enum != nil && len(fields[name].Enum) == 0 || field.MinLength != nil && *field.MinLength < 0 || field.MaxLength != nil && *field.MaxLength < 0 || field.MinLength != nil && field.MaxLength != nil && *field.MinLength > *field.MaxLength || field.Minimum != nil && field.Maximum != nil && *field.Minimum > *field.Maximum || field.MultipleOf != nil && *field.MultipleOf <= 0 {
			return Schema{}, nil, nil, fmt.Errorf("field %q has invalid bounds or enum", name)
		}
	}
	layout, err := wire.Clone(groups)
	if err != nil {
		return Schema{}, nil, nil, err
	}
	seen := make(map[string]bool)
	for _, group := range groups {
		if group.Kind != "group" || strings.TrimSpace(group.Title) == "" || group.Items == nil {
			return Schema{}, nil, nil, fmt.Errorf("groups require kind, title and items")
		}
		for _, item := range group.Items {
			switch item.Kind {
			case "property":
				if _, ok := fields[item.Property]; !ok || seen[item.Property] || item.Tool != "" || item.Title != "" || item.Description != "" {
					return Schema{}, nil, nil, fmt.Errorf("unknown, duplicate or malformed property %q", item.Property)
				}
				seen[item.Property] = true
			case "tool":
				if item.Property != "" || strings.TrimSpace(item.Tool) == "" || strings.TrimSpace(item.Title) == "" {
					return Schema{}, nil, nil, fmt.Errorf("tool items require tool and title")
				}
			default:
				return Schema{}, nil, nil, fmt.Errorf("unknown layout item kind %q", item.Kind)
			}
		}
	}
	var valuesSchema jsonschema.Schema
	encoded, err := wire.Clone[any](schema)
	if err != nil {
		return Schema{}, nil, nil, err
	}
	encoded.(map[string]any)["additionalProperties"] = false
	valuesSchema, err = wire.Convert[jsonschema.Schema](encoded)
	if err != nil {
		return Schema{}, nil, nil, err
	}
	resolved, err := valuesSchema.Resolve(nil)
	return schema, layout, resolved, err
}
