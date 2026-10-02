// Package forms builds and validates OpenAI extended form declarations and
// unchanged answers. Choosers, uploads, authorization and continuation storage
// belong to the host or business server. Patterns use the RE2 subset.
package forms

import (
	"encoding/json"
	"errors"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var (
	// ErrInvalidDefinition indicates an unsupported or inconsistent declaration.
	ErrInvalidDefinition = errors.New("invalid form definition")
	// ErrInvalidAnswer indicates an invalid action, field, value or selection.
	ErrInvalidAnswer = errors.New("invalid form answer")
)

// Schema is a flat object schema. Unknown answers are always rejected.
type Schema struct {
	Schema               string           `json:"$schema,omitempty"`
	Type                 string           `json:"type"`
	Title                string           `json:"title,omitempty"`
	Description          string           `json:"description,omitempty"`
	Properties           map[string]Field `json:"properties"`
	Required             []string         `json:"required,omitempty"`
	AdditionalProperties bool             `json:"additionalProperties"`
}

// Option is a labeled choice or suggestion. Preview is the deprecated image
// alias; neither image URL is fetched by this package.
type Option struct {
	Const       string    `json:"const"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	Thumbnail   *mcp.Icon `json:"x-openai-thumbnail,omitempty"`
	Preview     *mcp.Icon `json:"x-openai-preview,omitempty"`
}

// Field describes a primitive or string array. OneOf is a single choice; AnyOf
// is only valid inside Items. Suggestions allow custom values. Default is raw
// JSON so false, zero and empty values remain distinct from an absent default.
type Field struct {
	Type        string          `json:"type,omitempty"`
	Title       string          `json:"title,omitempty"`
	Description string          `json:"description,omitempty"`
	Default     json.RawMessage `json:"default,omitempty"`
	Enum        []any           `json:"enum,omitzero"`
	Examples    []any           `json:"examples,omitzero"`
	Comment     string          `json:"$comment,omitempty"`
	Meta        mcp.Meta        `json:"_meta,omitempty"`
	Deprecated  *bool           `json:"deprecated,omitempty"`
	ReadOnly    *bool           `json:"readOnly,omitempty"`
	WriteOnly   *bool           `json:"writeOnly,omitempty"`
	EnumNames   []string        `json:"enumNames,omitzero"`
	OneOf       []Option        `json:"oneOf,omitzero"`
	AnyOf       []Option        `json:"anyOf,omitzero"`
	Suggestions []Option        `json:"x-openai-suggestions,omitzero"`
	MinLength   *int            `json:"minLength,omitempty"`
	MaxLength   *int            `json:"maxLength,omitempty"`
	Pattern     string          `json:"pattern,omitempty"`
	Format      string          `json:"format,omitempty"`
	Minimum     *float64        `json:"minimum,omitempty"`
	Maximum     *float64        `json:"maximum,omitempty"`
	Items       *Field          `json:"items,omitempty"`
	MinItems    *int            `json:"minItems,omitempty"`
	MaxItems    *int            `json:"maxItems,omitempty"`
	UniqueItems *bool           `json:"uniqueItems,omitempty"`
	Input       *ResourceInput  `json:"x-openai-input,omitempty"`
}

// ResourceInput declares supplied resources and host user selection policy.
// Type is "resource" or the deprecated "file" alias. Selection is array-only:
// "explicit" (the default) or "implicit" (no defaults, permits user selection).
type ResourceInput struct {
	Type        string            `json:"type"`
	Options     []*ResourceOption `json:"options"`
	Selection   string            `json:"selection,omitempty"`
	UserOptions *UserOptions      `json:"userOptions,omitempty"`
}

// UserOptions permits host file or directory selection. Accept contains HTML
// accept tokens; the authorized uploader enforces actual names and media types.
type UserOptions struct {
	Kind   string   `json:"kind,omitempty"`
	Accept []string `json:"accept,omitzero"`
}

// Answer is the OpenAI form result shape, shared with the official SDK.
type Answer = mcp.ElicitResult

// SelectionPolicy optionally restricts declared host user selections further.
// A nil policy accepts valid user URIs when UserOptions is declared, as Python does.
// URI syntax and metadata do not authorize reading a resource.
type SelectionPolicy func(field, uri string) error
