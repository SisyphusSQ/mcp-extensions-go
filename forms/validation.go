package forms

import (
	"fmt"
	"math"

	"github.com/google/jsonschema-go/jsonschema"

	"github.com/SisyphusSQ/mcp-extensions-go/internal/schemautil"
	"github.com/SisyphusSQ/mcp-extensions-go/internal/wire"
)

// Validate checks required/unknown fields, primitive/array constraints, formats,
// choices and resources without inserting defaults or changing submitted values.
// A nil policy follows declared host user-selection permissions.
func (f *Form) Validate(content map[string]any, policy SelectionPolicy) error {
	if content == nil {
		return fmt.Errorf("%w: accepted content is required", ErrInvalidAnswer)
	}
	for _, name := range f.schema.Required {
		if _, exists := content[name]; !exists {
			return fmt.Errorf("%w: missing field %q", ErrInvalidAnswer, name)
		}
	}
	for name, value := range content {
		if err := f.ValidateField(name, value, policy); err != nil {
			return err
		}
	}
	return nil
}

// ValidateResult validates accept/cancel/decline. Accepted content is required;
// cancelled or declined answers cannot carry submitted fields.
func (f *Form) ValidateResult(answer *Answer, policy SelectionPolicy) error {
	if answer == nil {
		return fmt.Errorf("%w: missing result", ErrInvalidAnswer)
	}
	switch answer.Action {
	case "accept":
		return f.Validate(answer.Content, policy)
	case "cancel", "decline":
		if len(answer.Content) == 0 {
			return nil
		}
	}
	return fmt.Errorf("%w: invalid action or content", ErrInvalidAnswer)
}

// ValidateField validates a declared field independently, including resource policy.
func (f *Form) ValidateField(name string, value any, policy SelectionPolicy) error {
	field, exists := f.schema.Properties[name]
	if !exists {
		return fmt.Errorf("%w: unknown field %q", ErrInvalidAnswer, name)
	}
	return checkValue(name, field, f.fields[name], value, policy)
}

func checkValue(name string, field Field, validator *jsonschema.Resolved, value any, policy SelectionPolicy) error {
	copy, err := schemautil.Value(value)
	if err != nil || validator.Validate(copy) != nil {
		return fmt.Errorf("%w: field %q does not match its schema", ErrInvalidAnswer, name)
	}
	// JSON Schema permits integral floats, but the Python form protocol requires
	// an integer JSON representation for integer fields.
	if _, floating := copy.(float64); field.Type == "integer" && floating {
		return fmt.Errorf("%w: field %q requires an integer", ErrInvalidAnswer, name)
	}
	if field.Enum != nil {
		matched := false
		for _, option := range field.Enum {
			normalized, err := schemautil.Value(option)
			if err == nil && jsonschema.Equal(copy, normalized) {
				matched = true
				break
			}
		}
		if !matched {
			return fmt.Errorf("%w: field %q is outside its enum", ErrInvalidAnswer, name)
		}
	}
	if err := formats(field, copy); err != nil {
		return fmt.Errorf("%w: field %q: %v", ErrInvalidAnswer, name, err)
	}
	if input := field.Input; input != nil {
		selected := []any{copy}
		if field.Type == "array" {
			selected = copy.([]any)
		}
		for _, item := range selected {
			uri := item.(string)
			allowed := false
			for _, option := range input.Options {
				if option.URI == uri {
					allowed = true
					break
				}
			}
			if !allowed {
				if input.UserOptions == nil || policy != nil && policy(name, uri) != nil {
					return fmt.Errorf("%w: field %q has an unauthorized resource selection", ErrInvalidAnswer, name)
				}
			}
		}
	}
	return nil
}

func formats(field Field, value any) error {
	if field.Type == "string" {
		return schemautil.Format(field.Format, value.(string))
	}
	if field.Type == "array" {
		for _, item := range value.([]any) {
			if err := schemautil.Format(field.Items.Format, item.(string)); err != nil {
				return err
			}
		}
	}
	return nil
}

func (f *Form) checkDeclaredValues(name string, field Field) error {
	// Defaults and declared enum values may reference only supplied resources.
	declared := field
	if field.Input != nil {
		input := *field.Input
		input.UserOptions = nil
		declared.Input = &input
	}
	if field.Default != nil {
		if err := checkValue(name, declared, f.fields[name], field.Default, nil); err != nil {
			return fmt.Errorf("invalid default")
		}
	}
	for _, value := range field.Enum {
		if err := checkValue(name, declared, f.fields[name], value, nil); err != nil {
			return fmt.Errorf("invalid enum value")
		}
	}
	if field.Items != nil && field.Items.Default != nil {
		js, err := wire.Convert[jsonschema.Schema](field.Items)
		if err != nil {
			return err
		}
		validator, err := js.Resolve(nil)
		if err != nil || checkValue("item default", *field.Items, validator, field.Items.Default, nil) != nil {
			return fmt.Errorf("invalid item default")
		}
	}
	return nil
}

// PrepareSubmission checks a field before uploads and returns an independent
// user-selection policy. Pending uploads count toward array bounds. The uploader
// must enforce Accept against actual files; this helper performs no upload.
func (f *Form) PrepareSubmission(name string, content map[string]any, pending int, policy SelectionPolicy) (*UserOptions, error) {
	field, exists := f.schema.Properties[name]
	if !exists || pending < 0 {
		return nil, fmt.Errorf("%w: unknown field or invalid upload count", ErrInvalidAnswer)
	}
	value, present := content[name]
	if pending == 0 {
		if present {
			return nil, f.ValidateField(name, value, policy)
		}
		return nil, nil
	}
	if field.Input == nil || field.Input.UserOptions == nil {
		return nil, fmt.Errorf("%w: field does not permit uploads", ErrInvalidAnswer)
	}
	if field.Type == "string" {
		if pending != 1 || present {
			return nil, fmt.Errorf("%w: single upload requires an empty selection", ErrInvalidAnswer)
		}
	} else {
		if !present {
			value = []string{}
		}
		copy, err := schemautil.Value(value)
		selected, ok := copy.([]any)
		if err != nil || !ok || pending > math.MaxInt-len(selected) {
			return nil, fmt.Errorf("%w: invalid array selection", ErrInvalidAnswer)
		}
		count := len(selected) + pending
		if field.MinItems != nil && count < *field.MinItems || field.MaxItems != nil && count > *field.MaxItems {
			return nil, fmt.Errorf("%w: upload count violates array bounds", ErrInvalidAnswer)
		}
		partial := field
		partial.MinItems, partial.MaxItems = nil, nil
		partial.Enum = nil // Pending uploads are checked against enum only after merge.
		js, err := wire.Convert[jsonschema.Schema](partial)
		if err != nil {
			return nil, err
		}
		validator, err := js.Resolve(nil)
		if err != nil {
			return nil, err
		}
		if err := checkValue(name, partial, validator, value, policy); err != nil {
			return nil, err
		}
	}
	copy, _ := wire.Clone(field.Input.UserOptions)
	return copy, nil
}

// CompleteSubmission combines selections with uploaded references, requiring
// an optional additional policy, then validates the final field. It does
// not mutate content or authorize any subsequent resource read.
func (f *Form) CompleteSubmission(name string, content map[string]any, uploaded []string, authorize SelectionPolicy) (any, error) {
	if len(uploaded) == 0 {
		return nil, fmt.Errorf("%w: upload references are required", ErrInvalidAnswer)
	}
	if _, err := f.PrepareSubmission(name, content, len(uploaded), authorize); err != nil {
		return nil, err
	}
	for _, uri := range uploaded {
		if !schemautil.URI(uri) || authorize != nil && authorize(name, uri) != nil {
			return nil, fmt.Errorf("%w: unauthorized upload reference", ErrInvalidAnswer)
		}
	}
	var value any = uploaded[0]
	if f.schema.Properties[name].Type == "array" {
		selected := []string{}
		if current, exists := content[name]; exists {
			copy, _ := schemautil.Value(current)
			for _, item := range copy.([]any) {
				selected = append(selected, item.(string))
			}
		}
		value = append(selected, uploaded...)
	}
	if err := f.ValidateField(name, value, authorize); err != nil {
		return nil, err
	}
	return value, nil
}
