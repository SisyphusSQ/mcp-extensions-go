package forms

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/jsonschema-go/jsonschema"
	"maps"
	"slices"

	"github.com/SisyphusSQ/mcp-extensions-go/internal/schemautil"
	"github.com/SisyphusSQ/mcp-extensions-go/internal/wire"
)

// Model binds an immutable form to a flat Go struct and an optional business
// validator. Use pointers for optional fields when omitted and zero differ.
type Model[T any] struct {
	form     *Form
	validate func(context.Context, T) error
}

// NewModel generates a form through public jsonschema-go APIs. JSON aliases,
// omitempty/omitzero and jsonschema descriptions are honored. Overrides use JSON
// names and replace field constraints; they cannot change the model field type.
// The callback runs on accepted answers before the caller performs any writes.
func NewModel[T any](overrides map[string]Field, validate func(context.Context, T) error) (*Model[T], error) {
	return NewModelWithOptions[T](overrides, validate, nil)
}

// NewModelWithOptions accepts public schema inference options for named Go types,
// including enums, formats and annotations. Unsupported schema keywords fail.
func NewModelWithOptions[T any](overrides map[string]Field, validate func(context.Context, T) error, options *jsonschema.ForOptions) (*Model[T], error) {
	generated, err := schemautil.Model[T](options, true)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidDefinition, err)
	}
	schema := Schema{Type: "object", Required: generated.Required, Properties: make(map[string]Field)}
	for name, property := range generated.Properties {
		field, err := wire.ConvertStrict[Field](property)
		if err != nil {
			return nil, fmt.Errorf("%w: field %q: %v", ErrInvalidDefinition, name, err)
		}
		if override, exists := overrides[name]; exists {
			if override.Type != "" && override.Type != field.Type {
				return nil, fmt.Errorf("%w: override changes field %q type", ErrInvalidDefinition, name)
			}
			override.Type = cmp.Or(override.Type, field.Type)
			override.Title = cmp.Or(override.Title, field.Title)
			override.Description = cmp.Or(override.Description, field.Description)
			if override.Type == "array" && override.Items == nil {
				override.Items = field.Items
			}
			field = override
		}
		schema.Properties[name] = field
	}
	for name := range overrides {
		if _, exists := schema.Properties[name]; !exists {
			return nil, fmt.Errorf("%w: unknown model field %q", ErrInvalidDefinition, name)
		}
	}
	// Like Python model fields with defaults, these fields are optional on wire.
	schema.Required = slices.DeleteFunc(schema.Required, func(name string) bool { return schema.Properties[name].Default != nil })
	form, err := New(schema)
	if err != nil {
		return nil, err
	}
	return &Model[T]{form: form, validate: validate}, nil
}

// Form returns the immutable form used by this binding.
func (m *Model[T]) Form() *Form { return m.form }

// Decode validates an unchanged accepted answer, then applies declared defaults
// only to the typed result and runs the business callback. There is no coercion;
// the answer is never mutated. Cancel/decline must be handled by the caller.
func (m *Model[T]) Decode(ctx context.Context, answer *Answer, policy SelectionPolicy) (T, error) {
	var value T
	if err := ctx.Err(); err != nil {
		return value, err
	}
	if err := m.form.ValidateResult(answer, policy); err != nil {
		return value, err
	}
	if answer.Action != "accept" {
		return value, fmt.Errorf("%w: result was not accepted", ErrInvalidAnswer)
	}
	content := maps.Clone(answer.Content)
	for name, field := range m.form.schema.Properties {
		if _, exists := content[name]; !exists && field.Default != nil {
			content[name] = field.Default
		}
	}
	encoded, err := json.Marshal(content)
	if err != nil {
		return value, fmt.Errorf("%w: %v", ErrInvalidAnswer, err)
	}
	if err := json.Unmarshal(encoded, &value); err != nil {
		return value, fmt.Errorf("%w: model decoding failed: %v", ErrInvalidAnswer, err)
	}
	if m.validate != nil {
		if err := m.validate(ctx, value); err != nil {
			return value, err
		}
	}
	return value, ctx.Err()
}
