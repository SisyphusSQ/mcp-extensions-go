package forms

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"regexp"
	"strings"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/SisyphusSQ/mcp-extensions-go/internal/schemautil"
	"github.com/SisyphusSQ/mcp-extensions-go/internal/wire"
)

// Form is an immutable validated schema snapshot, safe for concurrent calls.
type Form struct {
	schema Schema
	fields map[string]*jsonschema.Resolved
}

// New validates and snapshots every declaration, option, suggestion and default.
// No I/O, upload, callback or persistence occurs while constructing a form.
func New(schema Schema) (*Form, error) {
	copy, err := wire.Clone(schema)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidDefinition, err)
	}
	if copy.Type != "object" || copy.Properties == nil || copy.AdditionalProperties {
		return nil, fmt.Errorf("%w: requires a closed, flat object schema", ErrInvalidDefinition)
	}
	for _, name := range copy.Required {
		if _, ok := copy.Properties[name]; !ok {
			return nil, fmt.Errorf("%w: unknown required field %q", ErrInvalidDefinition, name)
		}
	}
	for name, field := range copy.Properties {
		if err := prepareField(&field, false); err != nil {
			return nil, fmt.Errorf("%w: field %q: %v", ErrInvalidDefinition, name, err)
		}
		copy.Properties[name] = field
	}
	js, err := wire.Convert[jsonschema.Schema](copy)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidDefinition, err)
	}
	form := &Form{schema: copy, fields: make(map[string]*jsonschema.Resolved)}
	for name, property := range js.Properties {
		// Schema.UnmarshalJSON decodes enum numbers as float64. Validate root
		// enums separately from the exact wire snapshot to retain integer precision.
		property.Enum = nil
		validator, err := property.Resolve(nil)
		if err != nil {
			return nil, fmt.Errorf("%w: field %q: %v", ErrInvalidDefinition, name, err)
		}
		form.fields[name] = validator
	}
	for name, field := range copy.Properties {
		if err := form.checkDeclaredValues(name, field); err != nil {
			return nil, fmt.Errorf("%w: field %q: %v", ErrInvalidDefinition, name, err)
		}
	}
	return form, nil
}

// Parse rejects unknown schema keywords instead of silently losing constraints.
func Parse(data []byte) (*Form, error) {
	var raw any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidDefinition, err)
	}
	if err := validateShape(raw); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidDefinition, err)
	}
	var schema Schema
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&schema); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidDefinition, err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return nil, fmt.Errorf("%w: trailing JSON", ErrInvalidDefinition)
	}
	return New(schema)
}

func validateShape(value any) error {
	root, ok := value.(map[string]any)
	if !ok {
		return fmt.Errorf("form schema must be an object")
	}
	for key := range root {
		switch key {
		case "$schema", "type", "title", "description", "properties", "required", "additionalProperties":
		default:
			return fmt.Errorf("unsupported root keyword %s", key)
		}
	}
	for _, key := range []string{"type", "properties", "required"} {
		if value, exists := root[key]; exists && value == nil {
			return fmt.Errorf("%s cannot be null", key)
		}
	}
	properties, ok := root["properties"].(map[string]any)
	if !ok {
		return fmt.Errorf("properties must be an object")
	}
	if value, present := root["required"]; present {
		if err := stringArray(value); err != nil {
			return err
		}
	}
	for _, raw := range properties {
		if err := fieldShape(raw, false); err != nil {
			return err
		}
	}
	return nil
}

func fieldShape(value any, item bool) error {
	field, ok := value.(map[string]any)
	if !ok {
		return fmt.Errorf("field must be an object")
	}
	kind, _ := field["type"].(string)
	for key, value := range field {
		if key == "format" && value != nil && value == "" {
			return fmt.Errorf("format cannot be empty")
		}
		if key == "enumNames" && value != nil {
			if err := stringArray(value); err != nil {
				return err
			}
		}
		switch key {
		case "type", "title", "description", "default", "enum", "enumNames", "oneOf", "anyOf", "x-openai-suggestions", "minLength", "maxLength", "pattern", "format", "minimum", "maximum", "items", "minItems", "maxItems", "uniqueItems", "x-openai-input", "examples", "$comment", "_meta", "deprecated", "readOnly", "writeOnly":
		default:
			return fmt.Errorf("unsupported field keyword %s", key)
		}
		stringOnly := key == "minLength" || key == "maxLength" || key == "pattern" || key == "format" || key == "oneOf" || key == "x-openai-suggestions" || key == "enumNames"
		numberOnly := key == "minimum" || key == "maximum"
		arrayOnly := key == "items" || key == "minItems" || key == "maxItems" || key == "uniqueItems"
		if stringOnly && kind != "string" || numberOnly && kind != "number" && kind != "integer" || arrayOnly && kind != "array" {
			return fmt.Errorf("constraint %s does not apply to field type", key)
		}
		if value == nil && (key == "type" && (!item || field["enum"] == nil && field["anyOf"] == nil) || key == "pattern" || key == "x-openai-input") {
			return fmt.Errorf("%s cannot be null", key)
		}
		if key == "items" && value != nil {
			if err := fieldShape(value, true); err != nil {
				return err
			}
		}
		if item {
			choices := field["enum"] != nil || field["anyOf"] != nil
			allowed := key == "type" || key == "enum" || key == "anyOf"
			if !choices {
				allowed = key == "type" || key == "title" || key == "description" || key == "default" || key == "minLength" || key == "maxLength" || key == "pattern" || key == "format" || key == "x-openai-suggestions"
			}
			if !allowed {
				return fmt.Errorf("unsupported array item keyword %s", key)
			}
		}
		if key == "oneOf" || key == "anyOf" || key == "x-openai-suggestions" {
			if options, ok := value.([]any); ok {
				for _, raw := range options {
					option, ok := raw.(map[string]any)
					if !ok {
						return fmt.Errorf("option must be an object")
					}
					for key := range option {
						switch key {
						case "const", "title", "description", "x-openai-thumbnail", "x-openai-preview":
						default:
							return fmt.Errorf("unsupported option keyword %s", key)
						}
					}
					for _, key := range []string{"const", "title"} {
						if _, ok := option[key].(string); !ok {
							return fmt.Errorf("option requires string %s", key)
						}
					}
					for _, key := range []string{"x-openai-thumbnail", "x-openai-preview"} {
						if icon, exists := option[key]; exists && icon == nil {
							return fmt.Errorf("icon cannot be null")
						}
					}
				}
			}
		}
		if key == "x-openai-input" {
			input, ok := value.(map[string]any)
			if !ok {
				return fmt.Errorf("resource input must be an object")
			}
			for key := range input {
				switch key {
				case "type", "options", "selection", "userOptions":
				default:
					return fmt.Errorf("unsupported resource input keyword %s", key)
				}
			}
			if mode, exists := input["selection"]; exists && mode == "" {
				return fmt.Errorf("selection cannot be empty")
			}
			if input["type"] == nil || input["options"] == nil {
				return fmt.Errorf("resource input requires type and options")
			}
			if user, exists := input["userOptions"]; exists && user != nil {
				object, ok := user.(map[string]any)
				if !ok {
					return fmt.Errorf("user options must be an object")
				}
				for key := range object {
					switch key {
					case "kind", "accept":
					default:
						return fmt.Errorf("unsupported user options keyword %s", key)
					}
				}
				if kind, exists := object["kind"]; exists && (kind == nil || kind == "") {
					return fmt.Errorf("user kind cannot be null")
				}
				if accept, exists := object["accept"]; exists {
					if err := stringArray(accept); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func stringArray(value any) error {
	array, ok := value.([]any)
	if !ok {
		return fmt.Errorf("requires a string array")
	}
	for _, value := range array {
		if _, ok := value.(string); !ok {
			return fmt.Errorf("requires string array elements")
		}
	}
	return nil
}

// Schema returns an independent declaration. Defaults are never applied.
func (f *Form) Schema() Schema {
	copy, _ := wire.Clone(f.schema) // New has already verified JSON encoding.
	return copy
}

func prepareField(field *Field, item bool) error {
	if item && field.Type == "" && (field.AnyOf != nil || field.Enum != nil) {
		field.Type = "string"
	}
	if field.Type != "string" && field.Type != "boolean" && field.Type != "number" && field.Type != "integer" && field.Type != "array" || item && field.Type != "string" {
		return fmt.Errorf("unsupported field type")
	}
	stringOnly := field.EnumNames != nil || field.OneOf != nil || field.AnyOf != nil || field.Suggestions != nil || field.MinLength != nil || field.MaxLength != nil || field.Pattern != "" || field.Format != ""
	numberOnly := field.Minimum != nil || field.Maximum != nil
	arrayOnly := field.Items != nil || field.MinItems != nil || field.MaxItems != nil || field.UniqueItems != nil
	if stringOnly && field.Type != "string" || numberOnly && field.Type != "integer" && field.Type != "number" || arrayOnly && field.Type != "array" || item && (field.Input != nil || field.OneOf != nil || field.EnumNames != nil) || !item && field.AnyOf != nil {
		return fmt.Errorf("constraint does not apply to this field type")
	}
	if err := schemautil.String(field.Format, field.Pattern, field.MinLength, field.MaxLength); err != nil {
		return err
	}
	if err := schemautil.Number(field.Minimum, field.Maximum, nil, nil, nil); err != nil {
		return err
	}
	if item {
		if field.Examples != nil || field.Comment != "" || field.Meta != nil || field.Deprecated != nil || field.ReadOnly != nil || field.WriteOnly != nil {
			return fmt.Errorf("array items do not support field annotations")
		}
		if field.Enum != nil || field.AnyOf != nil {
			if (field.Enum == nil) == (field.AnyOf == nil) || field.Title != "" || field.Description != "" || field.Default != nil || field.MinLength != nil || field.MaxLength != nil || field.Pattern != "" || field.Format != "" || field.Suggestions != nil {
				return fmt.Errorf("multi-select items require exactly one set of choices")
			}
		}
	}
	if field.Enum != nil {
		if len(field.Enum) == 0 {
			return fmt.Errorf("choices must be nonempty")
		}
		seen := make(map[string]bool)
		for _, value := range field.Enum {
			encoded, err := json.Marshal(value)
			if err != nil {
				return err
			}
			if item {
				if _, ok := value.(string); !ok || seen[string(encoded)] {
					return fmt.Errorf("multi-select choices must be unique strings")
				}
			}
			seen[string(encoded)] = true
		}
	}
	if field.EnumNames != nil {
		if field.Enum == nil || len(field.EnumNames) != len(field.Enum) {
			return fmt.Errorf("labels must match enum choices")
		}
	}
	for index, options := range [][]Option{field.OneOf, field.AnyOf, field.Suggestions} {
		if options == nil {
			continue
		}
		if index != 2 && len(options) == 0 {
			return fmt.Errorf("options must be nonempty")
		}
		seen := make(map[string]bool)
		for _, option := range options {
			if index != 2 && seen[option.Const] {
				return fmt.Errorf("options require unique values")
			}
			seen[option.Const] = true
			for _, icon := range []*mcp.Icon{option.Thumbnail, option.Preview} {
				if icon != nil && !image(icon.Source) {
					return fmt.Errorf("option image must be HTTPS or a base64 image data URI")
				}
			}
		}
	}
	if field.Type == "array" {
		if field.Items == nil {
			return fmt.Errorf("array requires string items")
		}
		if err := prepareField(field.Items, true); err != nil {
			return err
		}
		if field.MinItems != nil && *field.MinItems < 0 || field.MaxItems != nil && *field.MaxItems < 0 || field.MinItems != nil && field.MaxItems != nil && *field.MinItems > *field.MaxItems {
			return fmt.Errorf("inconsistent array bounds")
		}
		if field.Items.Enum != nil || field.Items.AnyOf != nil {
			if field.UniqueItems != nil && !*field.UniqueItems {
				return fmt.Errorf("multiple choices must be unique")
			}
			unique := true
			field.UniqueItems = &unique
		}
	}
	if field.Input != nil {
		return prepareResource(field)
	}
	return nil
}

var dataImage = regexp.MustCompile(`^data:image/[a-zA-Z0-9.+-]+;base64,([a-zA-Z0-9+/]+={0,2})$`)

func image(src string) bool {
	if dataImage.MatchString(src) {
		return true
	}
	u, err := url.Parse(src)
	return err == nil && schemautil.URI(src) && u.Scheme == "https" && u.Hostname() != ""
}

func prepareResource(field *Field) error {
	input := field.Input
	text := field
	if field.Type == "array" {
		text = field.Items
	}
	if text.Type != "string" || text.Format != "uri" || input.Type != "resource" && input.Type != "file" || input.Options == nil {
		return fmt.Errorf("resource input requires URI strings and resource options")
	}
	if input.Selection != "" && (field.Type != "array" || input.Selection != "explicit" && input.Selection != "implicit") {
		return fmt.Errorf("invalid resource selection mode")
	}
	if input.Selection == "implicit" {
		if field.Default != nil {
			return fmt.Errorf("implicit selection forbids defaults")
		}
		if input.UserOptions == nil {
			input.UserOptions = &UserOptions{Kind: "file"}
		}
	}
	seen := make(map[string]bool)
	for _, option := range input.Options {
		if option == nil || !schemautil.URI(option.URI) || seen[option.URI] {
			return fmt.Errorf("invalid or duplicate resource option")
		}
		seen[option.URI] = true
	}
	if input.UserOptions != nil {
		if input.UserOptions.Kind == "" {
			input.UserOptions.Kind = "file"
		}
		if input.UserOptions.Kind != "file" && input.UserOptions.Kind != "directory" {
			return fmt.Errorf("unknown user resource kind")
		}
		seen = make(map[string]bool)
		for _, token := range input.UserOptions.Accept {
			normalized := strings.ToLower(token)
			if !acceptToken(token) || seen[normalized] {
				return fmt.Errorf("invalid or duplicate accept token")
			}
			seen[normalized] = true
		}
	}
	return nil
}

var mimeToken = regexp.MustCompile("^[!#$%&'*+.^_`|~0-9A-Za-z-]+/([!#$%&'*+.^_`|~0-9A-Za-z-]+|\\*)$")

func acceptToken(token string) bool {
	if token == "" || strings.Trim(token, " \t\n\r\f") != token {
		return false
	}
	if strings.HasPrefix(token, ".") {
		return !strings.Contains(token, ",")
	}
	return mimeToken.MatchString(token)
}
