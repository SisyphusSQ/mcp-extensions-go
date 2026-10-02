package forms

import (
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ResourceOption preserves a standard MCP resource and additional host-owned
// descriptor fields. Extra must not replace standard resource keys. No URI or
// additional field is fetched or evaluated by this package.
type ResourceOption struct {
	mcp.Resource
	Extra map[string]json.RawMessage `json:"-"`
}

var resourceKeys = map[string]bool{
	"uri": true, "name": true, "title": true, "description": true,
	"mimeType": true, "size": true, "icons": true, "annotations": true, "_meta": true,
}

// MarshalJSON merges independent descriptor extensions with standard fields.
func (r ResourceOption) MarshalJSON() ([]byte, error) {
	encoded, err := json.Marshal(r.Resource)
	if err != nil {
		return nil, err
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &object); err != nil {
		return nil, err
	}
	for key, value := range r.Extra {
		if resourceKeys[key] {
			return nil, fmt.Errorf("resource extra replaces standard key %q", key)
		}
		object[key] = value
	}
	return json.Marshal(object)
}

// UnmarshalJSON retains additional descriptors while decoding standard fields.
func (r *ResourceOption) UnmarshalJSON(data []byte) error {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(data, &object); err != nil {
		return err
	}
	if object == nil {
		return fmt.Errorf("resource option must be an object")
	}
	for _, key := range []string{"uri", "name"} {
		if raw, ok := object[key]; !ok || string(raw) == "null" {
			return fmt.Errorf("resource option requires %s", key)
		}
	}
	standard := make(map[string]json.RawMessage)
	extra := make(map[string]json.RawMessage)
	for key, value := range object {
		if resourceKeys[key] {
			standard[key] = value
		} else {
			extra[key] = value
		}
	}
	// encoding/json matches struct fields case-insensitively; descriptor extras
	// such as "URI" must remain opaque rather than replacing the canonical uri.
	encoded, err := json.Marshal(standard)
	if err != nil {
		return err
	}
	var resource mcp.Resource
	if err := json.Unmarshal(encoded, &resource); err != nil {
		return err
	}
	*r = ResourceOption{Resource: resource, Extra: extra}
	return nil
}
