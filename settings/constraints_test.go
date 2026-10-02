package settings_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/SisyphusSQ/mcp-extensions-go/internal/testutil"
	"github.com/SisyphusSQ/mcp-extensions-go/settings"
)

func ptr[T any](v T) *T { return &v }

func TestAdditionalConstraintsExecuteBeforeStorageAndOnResults(t *testing.T) {
	cfg := settings.Config{Fields: map[string]settings.Field{
		"email": {Type: "string", Title: "Email", Format: "email"},
		"date":  {Type: "string", Title: "Date", Format: "date"},
		"time":  {Type: "string", Title: "Time", Format: "date-time"},
		"uri":   {Type: "string", Title: "URI", Format: "uri"},
		"size":  {Type: "number", Title: "Size", ExclusiveMinimum: ptr(1.0), ExclusiveMaximum: ptr(5.0)},
	}}
	values := settings.Values{"email": "a@example.com", "date": "2024-02-29", "time": "2026-10-02T01:02:03Z", "uri": "urn:part:bolt", "size": 2.5}
	updates := 0
	cfg.Read = func(context.Context, *mcp.CallToolRequest) (settings.Values, error) { return values, nil }
	cfg.Update = func(_ context.Context, _ *mcp.CallToolRequest, patch settings.Values) (settings.Values, error) {
		updates++
		for name, value := range patch {
			values[name] = value
		}
		return values, nil
	}
	server, err := settings.NewServer(&mcp.Implementation{Name: "test", Version: "1"}, nil, cfg)
	if err != nil {
		t.Fatal(err)
	}
	cs := testutil.Connect(t, server, nil, "2026-07-28")
	call := func(name string, args any) *mcp.CallToolResult {
		t.Helper()
		result, err := cs.CallTool(t.Context(), &mcp.CallToolParams{Name: name, Arguments: args})
		if err != nil {
			t.Fatal(err)
		}
		return result
	}
	read := call("settings.read", map[string]any{})
	if read.IsError {
		t.Fatalf("read failed %#v", read)
	}
	encoded, _ := json.Marshal(read.StructuredContent)
	for _, keyword := range []string{`"format":"email"`, `"exclusiveMinimum":1`, `"exclusiveMaximum":5`} {
		if !strings.Contains(string(encoded), keyword) {
			t.Fatalf("missing schema %s", keyword)
		}
	}
	for name, invalid := range map[string][]any{"email": {"bad", "A <a@example.com>"}, "date": {"2025-02-29"}, "time": {"2026-10-02"}, "uri": {"relative"}, "size": {1, 5}} {
		for _, value := range invalid {
			if !call("settings.update", map[string]any{"set": map[string]any{name: value}}).IsError {
				t.Fatalf("invalid patch accepted %s=%#v", name, value)
			}
		}
	}
	if updates != 0 {
		t.Fatal("invalid constraints reached storage")
	}
	if call("settings.update", map[string]any{"set": map[string]any{"size": 3}}).IsError || updates != 1 {
		t.Fatal("valid partial patch failed")
	}
	values["email"] = "bad"
	if !call("settings.read", map[string]any{}).IsError {
		t.Fatal("invalid returned format accepted")
	}
	if !call("settings.update", map[string]any{"set": map[string]any{"size": 4}}).IsError {
		t.Fatal("invalid returned state accepted")
	}
}

func TestAdditionalConstraintDefinitions(t *testing.T) {
	for _, field := range []settings.Field{
		{Type: "string", Title: "Bad", Format: "unknown"}, {Type: "boolean", Title: "Bad", Format: "email"},
		{Type: "string", Title: "Bad", ExclusiveMinimum: ptr(1.0)},
		{Type: "number", Title: "Bad", Minimum: ptr(1.0), ExclusiveMaximum: ptr(1.0)},
		{Type: "integer", Title: "Bad", ExclusiveMinimum: ptr(2.0), Maximum: ptr(1.0)},
		{Type: "number", Title: "Bad", Maximum: ptr(math.Inf(1))},
		{Type: "string", Title: "Bad", Format: "date", Enum: []string{"2025-02-29"}},
	} {
		cfg := config()
		cfg.Fields["bad"] = field
		cfg.Read = func(context.Context, *mcp.CallToolRequest) (settings.Values, error) {
			t.Fatal("registration invoked storage")
			return nil, nil
		}
		cfg.Update = func(context.Context, *mcp.CallToolRequest, settings.Values) (settings.Values, error) {
			t.Fatal("registration invoked storage")
			return nil, nil
		}
		if _, err := settings.NewServer(&mcp.Implementation{Name: "test", Version: "1"}, nil, cfg); !errors.Is(err, settings.ErrInvalidDefinition) {
			t.Fatalf("definition accepted: %#v %v", field, err)
		}
	}
}

type typedState struct {
	Label   string `json:"display_label"`
	Enabled bool   `json:"enabled"`
	Count   uint64 `json:"count"`
}

func TestTypedSettingsModelAndPartialPatches(t *testing.T) {
	state := typedState{Label: "part", Count: 42}
	server, err := settings.NewModelServer(&mcp.Implementation{Name: "model", Version: "1"}, nil, settings.ModelConfig[typedState]{
		Fields: map[string]settings.Field{"display_label": {MinLength: ptr(2)}},
		Read:   func(context.Context, *mcp.CallToolRequest) (typedState, error) { return state, nil },
		Update: func(_ context.Context, _ *mcp.CallToolRequest, patch settings.Values) (typedState, error) {
			if len(patch) != 1 || patch["Enabled"] != true {
				return typedState{}, fmt.Errorf("patch omitted fields lost: %#v", patch)
			}
			state.Enabled = true
			return state, nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	cs := testutil.Connect(t, server, nil, "2026-07-28")
	for _, call := range []*mcp.CallToolParams{{Name: "settings.read", Arguments: map[string]any{}}, {Name: "settings.update", Arguments: map[string]any{"set": map[string]any{"enabled": true}}}} {
		result, err := cs.CallTool(t.Context(), call)
		if err != nil || result.IsError {
			t.Fatalf("typed state failed: %#v %v", result, err)
		}
		values := result.StructuredContent.(map[string]any)["values"].(map[string]any)
		if values["count"] != float64(42) || values["display_label"] != "part" {
			t.Fatalf("typed values did not round trip: %#v", values)
		}
	}
	if _, err := settings.FieldsFor[struct{ Values []string }](nil); err == nil {
		t.Fatal("array settings accepted")
	}
	fields, err := settings.FieldsFor[struct {
		Optional string `json:"optional,omitempty"`
	}](nil)
	if err != nil || fields["optional"].Type != "string" {
		t.Fatal(fields, err)
	}
	if _, err := settings.FieldsFor[typedState](map[string]settings.Field{"Label": {Type: "string"}}); err == nil {
		t.Fatal("Go name used instead of JSON alias")
	}
}
