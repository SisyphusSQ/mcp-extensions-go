package settings

import (
	"cmp"
	"context"
	"fmt"
	"maps"
	"reflect"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/SisyphusSQ/mcp-extensions-go/internal/schemautil"
	"github.com/SisyphusSQ/mcp-extensions-go/internal/wire"
)

// FieldsFor generates primitive settings fields from a flat Go struct using
// public schema inference and JSON aliases. Every effective value is required,
// including fields tagged omitempty. Overrides replace constraints by JSON name.
func FieldsFor[T any](overrides map[string]Field) (map[string]Field, error) {
	return FieldsForWithOptions[T](overrides, nil)
}

// FieldsForWithOptions allows named Go types to declare enums and constraints
// through public jsonschema-go inference options. Unsupported keywords fail.
func FieldsForWithOptions[T any](overrides map[string]Field, options *jsonschema.ForOptions) (map[string]Field, error) {
	schema, err := schemautil.Model[T](options, false)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidDefinition, err)
	}
	fields := make(map[string]Field)
	for name, property := range schema.Properties {
		if property.Type == "array" {
			return nil, fmt.Errorf("%w: settings do not support arrays", ErrInvalidDefinition)
		}
		field, err := wire.ConvertStrict[Field](property)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrInvalidDefinition, err)
		}
		if override, exists := overrides[name]; exists {
			if override.Type != "" && override.Type != field.Type {
				return nil, fmt.Errorf("%w: override changes field %q type", ErrInvalidDefinition, name)
			}
			override.Type = cmp.Or(override.Type, field.Type)
			override.Title = cmp.Or(override.Title, field.Title)
			override.Description = cmp.Or(override.Description, field.Description)
			field = override
		}
		fields[name] = field
	}
	for name := range overrides {
		if _, exists := fields[name]; !exists {
			return nil, fmt.Errorf("%w: unknown model field %q", ErrInvalidDefinition, name)
		}
	}
	schemaCopy, _, _, err := prepare(fields, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidDefinition, err)
	}
	return schemaCopy.Properties, nil
}

// ModelConfig binds typed effective values to caller-owned storage. Update keeps
// a map patch so omitted fields remain distinct from zero. It must authorize,
// validate the merged state and persist in one transaction before returning.
// Result validation cannot undo writes already performed by the callback.
type ModelConfig[T any] struct {
	ReadTool        string
	UpdateTool      string
	Fields          map[string]Field
	SchemaOptions   *jsonschema.ForOptions
	FieldNames      map[string]string
	FieldValidators map[string]FieldValidator
	Layout          []Group
	Read            func(context.Context, *mcp.CallToolRequest) (T, error)
	Update          func(context.Context, *mcp.CallToolRequest, Values) (T, error)
}

// NewModelServer registers settings using a flat typed effective-state model.
// It preserves ordinary NewServer partial-patch and authorization semantics.
func NewModelServer[T any](info *mcp.Implementation, options *mcp.ServerOptions, config ModelConfig[T]) (*mcp.Server, error) {
	if config.Read == nil || config.Update == nil {
		return nil, fmt.Errorf("%w: both handlers are required", ErrInvalidDefinition)
	}
	fields, err := FieldsForWithOptions[T](config.Fields, config.SchemaOptions)
	if err != nil {
		return nil, err
	}
	generated, err := schemautil.Model[T](config.SchemaOptions, false)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidDefinition, err)
	}
	wireNames := schemautil.FieldNames[T](generated.Properties)
	names := maps.Clone(wireNames)
	maps.Copy(names, config.FieldNames)
	values := func(value T, err error) (Values, error) {
		if err != nil {
			return nil, err
		}
		copy, err := schemautil.Value(value)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrInvalidValues, err)
		}
		object, ok := copy.(map[string]any)
		if !ok || object == nil {
			return nil, fmt.Errorf("%w: model must encode an object", ErrInvalidValues)
		}
		// Effective settings include zero values even when JSON omission tags are
		// useful elsewhere. Go model fields remain the source of those missing values.
		model := reflect.ValueOf(value)
		for name := range fields {
			if _, exists := object[name]; exists {
				continue
			}
			field := model.FieldByName(wireNames[name])
			if !field.IsValid() {
				return nil, fmt.Errorf("%w: missing model field %q", ErrInvalidValues, name)
			}
			encoded, err := schemautil.Value(field.Interface())
			if err != nil {
				return nil, fmt.Errorf("%w: field %q: %v", ErrInvalidValues, name, err)
			}
			object[name] = encoded
		}
		return object, nil
	}
	return NewServer(info, options, Config{
		ReadTool: config.ReadTool, UpdateTool: config.UpdateTool, Fields: fields, Layout: config.Layout, FieldNames: names, FieldValidators: config.FieldValidators,
		Read: func(ctx context.Context, req *mcp.CallToolRequest) (Values, error) {
			return values(config.Read(ctx, req))
		},
		Update: func(ctx context.Context, req *mcp.CallToolRequest, set Values) (Values, error) {
			return values(config.Update(ctx, req, set))
		},
	})
}
