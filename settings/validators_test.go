package settings_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/SisyphusSQ/mcp-extensions-go/internal/testutil"
	"github.com/SisyphusSQ/mcp-extensions-go/settings"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestFieldValidatorsAliasesAndPresenceBeforeSave(t *testing.T) {
	state := settings.Values{"display_label": "initial", "enabled": true}
	saves, validations := 0, 0
	cfg := settings.Config{
		Fields:     map[string]settings.Field{"display_label": {Type: "string", Title: "Label", MinLength: ptr(2)}, "enabled": {Type: "boolean", Title: "Enabled"}},
		FieldNames: map[string]string{"display_label": "Label", "enabled": "Enabled"},
		FieldValidators: map[string]settings.FieldValidator{"display_label": func(_ context.Context, value any) (any, error) {
			validations++
			if value == "denied" {
				return nil, errors.New("label rejected")
			}
			if value == "invalid" {
				return 42, nil
			}
			return strings.TrimSpace(value.(string)), nil
		}},
		Read: func(context.Context, *mcp.CallToolRequest) (settings.Values, error) { return state, nil },
		Update: func(_ context.Context, _ *mcp.CallToolRequest, set settings.Values) (settings.Values, error) {
			saves++
			if len(set) != 1 {
				return nil, fmt.Errorf("omitted fields included: %#v", set)
			}
			if value, ok := set["Label"]; ok {
				state["display_label"] = value
			} else if value, ok := set["Enabled"]; ok {
				state["enabled"] = value
			} else {
				return nil, fmt.Errorf("wire alias reached save: %#v", set)
			}
			return state, nil
		},
	}
	server, err := settings.NewServer(&mcp.Implementation{Name: "validators", Version: "1"}, nil, cfg)
	if err != nil {
		t.Fatal(err)
	}
	// Registration owns independent maps; caller mutations cannot replace validation.
	cfg.FieldNames["display_label"] = "wrong"
	cfg.FieldValidators["display_label"] = func(context.Context, any) (any, error) { return "bypass", nil }
	cs := testutil.Connect(t, server, nil, "")
	call := func(set settings.Values, wantError bool) {
		t.Helper()
		result, err := cs.CallTool(t.Context(), &mcp.CallToolParams{Name: "settings.update", Arguments: map[string]any{"set": set}})
		if err != nil || result.IsError != wantError {
			t.Fatalf("patch %#v result=%#v err=%v", set, result, err)
		}
	}
	for _, value := range []string{"denied", "invalid", "  "} {
		call(settings.Values{"display_label": value}, true)
	}
	if saves != 0 {
		t.Fatal("invalid field reached storage")
	}
	call(settings.Values{"display_label": "  valid  "}, false)
	if saves != 1 || state["display_label"] != "valid" || state["enabled"] != true {
		t.Fatal("transformed alias patch was not saved")
	}
	before := validations
	call(settings.Values{"enabled": false}, false)
	if validations != before || saves != 2 || state["enabled"] != false {
		t.Fatal("omitted field was validated or zero was dropped")
	}
}

func TestRejectInvalidAliasAndValidatorDefinitions(t *testing.T) {
	for _, change := range []func(*settings.Config){
		func(c *settings.Config) { c.FieldNames = map[string]string{"unknown": "Label"} },
		func(c *settings.Config) { c.FieldNames = map[string]string{"units": "grid"} },
		func(c *settings.Config) { c.FieldNames = map[string]string{"units": ""} },
		func(c *settings.Config) { c.FieldValidators = map[string]settings.FieldValidator{"units": nil} },
		func(c *settings.Config) {
			c.FieldValidators = map[string]settings.FieldValidator{"unknown": func(context.Context, any) (any, error) { return nil, nil }}
		},
	} {
		cfg := config()
		cfg.Read = func(context.Context, *mcp.CallToolRequest) (settings.Values, error) {
			return settings.Values{"units": "mm", "grid": false}, nil
		}
		cfg.Update = func(context.Context, *mcp.CallToolRequest, settings.Values) (settings.Values, error) {
			return settings.Values{"units": "mm", "grid": false}, nil
		}
		change(&cfg)
		if _, err := settings.NewServer(&mcp.Implementation{Name: "test", Version: "1"}, nil, cfg); !errors.Is(err, settings.ErrInvalidDefinition) {
			t.Fatal("invalid registration accepted", err)
		}
	}
}
