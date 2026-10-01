package ui_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/SisyphusSQ/mcp-extensions-go/ui"
)

func TestToolMetadataWireContract(t *testing.T) {
	base := mcp.Meta{"vendor": map[string]any{"value": "original"}}
	metadata := ui.ToolMetadata{
		ResourceURI: "ui://test/home.html", Visibility: []ui.Visibility{ui.App, ui.Model},
		PreferredModelDisplayMode: ui.Fullscreen,
		Entrypoints: []ui.Entrypoint{
			{Type: ui.Global, QuickAction: &ui.QuickAction{
				Title: "打开", Icons: []mcp.Icon{{Source: "data:image/png;base64,AA=="}},
				Target: ui.QuickActionTarget{Type: "tool", Name: "open", Arguments: map[string]any{"tab": "home"}},
			}},
			{Type: ui.Thread}, {Type: ui.File, Extensions: []string{}},
			{Type: ui.Settings, SearchTerms: []string{"工作台"}},
		},
	}
	got, err := metadata.Metadata(base)
	if err != nil {
		t.Fatal(err)
	}
	wantJSON := `{"vendor":{"value":"original"},"ui":{"resourceUri":"ui://test/home.html","visibility":["app","model"]},"openai/ui":{"preferredModelDisplayMode":"fullscreen","entrypoints":[{"type":"global","quickAction":{"title":"打开","icons":[{"src":"data:image/png;base64,AA=="}],"target":{"type":"tool","name":"open","arguments":{"tab":"home"}}}},{"type":"thread"},{"type":"file","extensions":[]},{"type":"settings","searchTerms":["工作台"]}]}}`
	var want mcp.Meta
	if err := json.Unmarshal([]byte(wantJSON), &want); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("wire metadata = %#v, want %#v", got, want)
	}
	base["vendor"].(map[string]any)["value"] = "changed"
	metadata.Entrypoints[0].QuickAction.Target.Arguments["tab"] = "changed"
	if !reflect.DeepEqual(got, want) {
		t.Fatal("metadata retained caller-owned nested values")
	}
}

func TestInvalidToolMetadata(t *testing.T) {
	tests := []struct {
		name string
		meta ui.ToolMetadata
	}{
		{"http uri", ui.ToolMetadata{ResourceURI: "https://example.com/ui"}},
		{"missing host", ui.ToolMetadata{ResourceURI: "ui:///home.html"}},
		{"unknown visibility", ui.ToolMetadata{ResourceURI: "ui://app/home", Visibility: []ui.Visibility{"public"}}},
		{"pip model mode", ui.ToolMetadata{ResourceURI: "ui://app/home", PreferredModelDisplayMode: ui.PiP}},
		{"unknown entry", ui.ToolMetadata{ResourceURI: "ui://app/home", Entrypoints: []ui.Entrypoint{{Type: "other"}}}},
		{"thread fields", ui.ToolMetadata{ResourceURI: "ui://app/home", Entrypoints: []ui.Entrypoint{{Type: ui.Thread, Extensions: []string{}}}}},
		{"file missing extensions", ui.ToolMetadata{ResourceURI: "ui://app/home", Entrypoints: []ui.Entrypoint{{Type: ui.File}}}},
		{"file invalid extension", ui.ToolMetadata{ResourceURI: "ui://app/home", Entrypoints: []ui.Entrypoint{{Type: ui.File, Extensions: []string{"csv"}}}}},
		{"empty search term", ui.ToolMetadata{ResourceURI: "ui://app/home", Entrypoints: []ui.Entrypoint{{Type: ui.Settings, SearchTerms: []string{" "}}}}},
		{"empty quick action", ui.ToolMetadata{ResourceURI: "ui://app/home", Entrypoints: []ui.Entrypoint{{Type: ui.Global, QuickAction: &ui.QuickAction{}}}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := tt.meta.Metadata(nil); err == nil {
				t.Fatal("invalid metadata accepted")
			}
		})
	}
	if _, err := (ui.ToolMetadata{ResourceURI: "ui://app/home"}).Metadata(mcp.Meta{"bad": make(chan int)}); err == nil {
		t.Fatal("non-JSON metadata accepted")
	}
}

func TestResourceMetadataWireContract(t *testing.T) {
	got, err := (ui.ResourceMetadata{
		AvailableDisplayModes: []ui.DisplayMode{ui.Inline, ui.Fullscreen, ui.PiP}, PreferredDisplayMode: ui.Inline,
	}).Metadata(mcp.Meta{"ui": map[string]any{"csp": map[string]any{"connectDomains": []string{}}}})
	if err != nil {
		t.Fatal(err)
	}
	var want mcp.Meta
	if err := json.Unmarshal([]byte(`{"ui":{"csp":{"connectDomains":[]}},"openai/ui":{"availableDisplayModes":["inline","fullscreen","pip"],"preferredDisplayMode":"inline"}}`), &want); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("metadata = %#v, want %#v", got, want)
	}
	if _, err := (ui.ResourceMetadata{PreferredDisplayMode: "other"}).Metadata(nil); err == nil {
		t.Fatal("invalid mode accepted")
	}
}

func TestMetadataPreservesLargeJSONNumbers(t *testing.T) {
	meta, err := (ui.ToolMetadata{ResourceURI: "ui://app/home"}).Metadata(mcp.Meta{
		"vendor": map[string]any{"id": uint64(9007199254740993)},
	})
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(meta["vendor"])
	if err != nil {
		t.Fatal(err)
	}
	if string(encoded) != `{"id":9007199254740993}` {
		t.Fatalf("integer precision lost: %s", encoded)
	}
}
