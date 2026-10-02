package settings_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/SisyphusSQ/mcp-extensions-go/internal/testutil"
	"github.com/SisyphusSQ/mcp-extensions-go/settings"
	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type displayMode string
type inferredSettings struct {
	Mode    displayMode `json:"display_mode"`
	Enabled bool        `json:"is_enabled,omitempty"`
	Count   int         `json:"item_count,omitzero"`
}

func TestSettingsNamedTypesAndEffectiveZeroValues(t *testing.T) {
	options := &jsonschema.ForOptions{TypeSchemas: map[reflect.Type]*jsonschema.Schema{
		reflect.TypeFor[displayMode](): {Type: "string", Enum: []any{"inline", "fullscreen"}},
	}}
	state := inferredSettings{Mode: "inline"}
	server, err := settings.NewModelServer(&mcp.Implementation{Name: "model", Version: "1"}, nil, settings.ModelConfig[inferredSettings]{
		SchemaOptions: options,
		Read:          func(context.Context, *mcp.CallToolRequest) (inferredSettings, error) { return state, nil },
		Update: func(_ context.Context, _ *mcp.CallToolRequest, set settings.Values) (inferredSettings, error) {
			if len(set) != 1 || set["Mode"] != "fullscreen" {
				return inferredSettings{}, fmt.Errorf("alias/presence=%#v", set)
			}
			state.Mode = "fullscreen"
			return state, nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	cs := testutil.Connect(t, server, nil, "")
	for _, params := range []*mcp.CallToolParams{{Name: "settings.read", Arguments: map[string]any{}}, {Name: "settings.update", Arguments: map[string]any{"set": map[string]any{"display_mode": "fullscreen"}}}} {
		result, err := cs.CallTool(t.Context(), params)
		if err != nil || result.IsError {
			t.Fatal("effective model failed", result, err)
		}
		values := result.StructuredContent.(map[string]any)["values"].(map[string]any)
		if values["is_enabled"] != false || values["item_count"] != float64(0) {
			t.Fatal("effective zero omitted", values)
		}
	}
	if _, err := settings.FieldsFor[struct{ Optional *string }](nil); err == nil {
		t.Fatal("nullable setting accepted")
	}
	if _, err := settings.FieldsForWithOptions[inferredSettings](nil, &jsonschema.ForOptions{IgnoreInvalidTypes: true}); err == nil {
		t.Fatal("silent field omission accepted")
	}
	options.TypeSchemas[reflect.TypeFor[displayMode]()] = &jsonschema.Schema{Type: "string", Default: []byte(`"inline"`)}
	if _, err := settings.FieldsForWithOptions[inferredSettings](nil, options); err == nil {
		t.Fatal("settings schema default accepted")
	}
}
