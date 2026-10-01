// Package resources parses OpenAI file context without authorizing file access.
package resources

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/SisyphusSQ/mcp-extensions-go/internal/wire"
)

// MetadataKey identifies host resource metadata.
const MetadataKey = "openai/resource"

// File is the host's opaque reference, not a local filesystem path.
type File struct {
	Name        string `json:"name"`
	ResourceURI string `json:"resourceUri"`
}

// FileInput contains arguments supplied to a file-extension entrypoint.
type FileInput struct {
	File File `json:"file"`
}

// Validate checks the input contract. It does not resolve or read ResourceURI.
func (input FileInput) Validate() error {
	if input.File.Name == "" || strings.ContainsAny(input.File.Name, `/\`) || strings.TrimSpace(input.File.ResourceURI) == "" {
		return fmt.Errorf("file input requires a filename without a path and a non-blank resourceUri")
	}
	return nil
}

// Representation controls host resource reads; unspecified means either form.
type Representation string

const (
	// Text requests UTF-8 text.
	Text Representation = "text"
	// Blob requests base64-encoded bytes.
	Blob Representation = "blob"
)

// ReadMetadata contains a resource read representation hint.
type ReadMetadata struct {
	Representation Representation `json:"representation,omitempty"`
}

// ContentMetadata contains host-owned write permission hints and version tokens.
// Writable is a UI hint; the host must enforce permission on writes.
type ContentMetadata struct {
	Writable *bool  `json:"writable,omitempty"`
	ETag     string `json:"etag,omitempty"`
}

func namespace(meta mcp.Meta) (map[string]json.RawMessage, bool, error) {
	value, present := meta[MetadataKey]
	if !present {
		return nil, false, nil
	}
	object, err := wire.Convert[map[string]json.RawMessage](value)
	if err != nil || object == nil {
		return nil, true, fmt.Errorf("%s metadata must be an object", MetadataKey)
	}
	return object, true, nil
}

// Path returns the path from tool-call metadata, with presence distinct from an
// empty string. Parsing is not authorization or proof of a trusted host origin.
// Only an explicitly authorized reader may open it; this function never does.
func Path(meta mcp.Meta) (string, bool, error) {
	object, present, err := namespace(meta)
	if err != nil || !present {
		return "", present, err
	}
	raw, ok := object["path"]
	if !ok || string(raw) == "null" {
		return "", true, fmt.Errorf("resource tool metadata requires a string path")
	}
	var path string
	if err := json.Unmarshal(raw, &path); err != nil {
		return "", true, fmt.Errorf("resource path must be a string: %w", err)
	}
	return path, true, nil
}

// ParseReadMetadata parses a request hint, allowing unknown future keys.
func ParseReadMetadata(meta mcp.Meta) (*ReadMetadata, error) {
	object, present, err := namespace(meta)
	if err != nil || !present {
		return nil, err
	}
	metadata := new(ReadMetadata)
	if raw, ok := object["representation"]; ok {
		if err := json.Unmarshal(raw, &metadata.Representation); err != nil || metadata.Representation != Text && metadata.Representation != Blob {
			return nil, fmt.Errorf("resource representation must be text or blob")
		}
	}
	return metadata, nil
}

// ParseContentMetadata parses write hints from one resource content item.
func ParseContentMetadata(meta mcp.Meta) (*ContentMetadata, error) {
	object, present, err := namespace(meta)
	if err != nil || !present {
		return nil, err
	}
	metadata := new(ContentMetadata)
	if raw, ok := object["writable"]; ok {
		if string(raw) == "null" || json.Unmarshal(raw, &metadata.Writable) != nil {
			return nil, fmt.Errorf("resource writable must be a boolean")
		}
	}
	if raw, ok := object["etag"]; ok {
		if string(raw) == "null" || json.Unmarshal(raw, &metadata.ETag) != nil {
			return nil, fmt.Errorf("resource etag must be a string")
		}
	}
	return metadata, nil
}
