// This acceptance probe uses public extension APIs and the official SDK only.
// It registers only temporary diagnostic tools, never product capabilities.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"os"
	"os/exec"
	"reflect"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/SisyphusSQ/mcp-extensions-go/forms"
	"github.com/SisyphusSQ/mcp-extensions-go/settings"
)

type displayMode string
type preference struct {
	Mode      displayMode `json:"display_mode"`
	Label     string      `json:"display_label"`
	Email     string      `json:"email"`
	Date      string      `json:"date"`
	Timestamp string      `json:"timestamp"`
	URI       string      `json:"uri"`
	Size      float64     `json:"size"`
	Count     int         `json:"count"`
	Enabled   bool        `json:"is_enabled,omitzero"`
}
type counters struct {
	Saves       int      `json:"saves"`
	Validations int      `json:"validations"`
	PatchNames  []string `json:"patchNames"`
}
type formInput struct {
	SchemaJSON  string   `json:"schemaJSON"`
	Operation   string   `json:"operation"`
	ValueJSON   string   `json:"valueJSON,omitempty"`
	ContentJSON string   `json:"contentJSON,omitempty"`
	Uploaded    []string `json:"uploaded,omitempty"`
	Pending     int      `json:"pending,omitzero"`
}
type formOutput struct {
	SchemaJSON string `json:"schemaJSON,omitempty"`
	ValueJSON  string `json:"valueJSON,omitempty"`
	Unchanged  bool   `json:"unchanged"`
}
type part string
type selection struct {
	Part     part    `json:"part_id"`
	Quantity int     `json:"quantity"`
	Note     *string `json:"note,omitempty"`
}
type decodeInput struct {
	AnswerJSON string `json:"answerJSON"`
}
type decodeOutput struct {
	Value      selection `json:"value"`
	SchemaJSON string    `json:"schemaJSON"`
	Unchanged  bool      `json:"unchanged"`
}

func ptr[T any](value T) *T { return &value }
func decode(raw string, into any) error {
	decoder := json.NewDecoder(strings.NewReader(raw))
	decoder.UseNumber()
	return decoder.Decode(into)
}
func encode(value any) string {
	encoded, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return string(encoded)
}

func serve(ctx context.Context) error {
	state := preference{Mode: "inline", Label: "initial", Email: "a@example.com", Date: "2024-02-29", Timestamp: "2026-10-02T01:02:03Z", URI: "urn:part:bolt", Size: 2.5, Count: 2}
	stats := counters{PatchNames: []string{}}
	var mu sync.Mutex
	options := &jsonschema.ForOptions{TypeSchemas: map[reflect.Type]*jsonschema.Schema{
		reflect.TypeFor[displayMode](): {Type: "string", Enum: []any{"inline", "fullscreen"}},
	}}
	server, err := settings.NewModelServer(&mcp.Implementation{Name: "new-extension-acceptance", Version: "2026-10-02"}, nil, settings.ModelConfig[preference]{
		SchemaOptions: options,
		Fields: map[string]settings.Field{
			"display_label": {MinLength: ptr(2)}, "email": {Format: "email"}, "date": {Format: "date"},
			"timestamp": {Format: "date-time"}, "uri": {Format: "uri"},
			"size":  {ExclusiveMinimum: ptr(1.0), ExclusiveMaximum: ptr(5.0)},
			"count": {Minimum: ptr(2.0), Maximum: ptr(10.0), MultipleOf: ptr(2.0)},
		},
		FieldValidators: map[string]settings.FieldValidator{"display_label": func(_ context.Context, value any) (any, error) {
			mu.Lock()
			defer mu.Unlock()
			stats.Validations++
			if value == "denied" {
				return nil, errors.New("label rejected")
			}
			if value == "invalid" {
				return 42, nil
			}
			return strings.TrimSpace(value.(string)), nil
		}},
		Read: func(context.Context, *mcp.CallToolRequest) (preference, error) {
			mu.Lock()
			defer mu.Unlock()
			return state, nil
		},
		Update: func(_ context.Context, _ *mcp.CallToolRequest, patch settings.Values) (preference, error) {
			mu.Lock()
			defer mu.Unlock()
			stats.Saves++
			stats.PatchNames = slices.Sorted(maps.Keys(patch))
			for name, value := range patch {
				switch name {
				case "Mode":
					state.Mode = displayMode(value.(string))
				case "Label":
					state.Label = value.(string)
				case "Email":
					state.Email = value.(string)
				case "Date":
					state.Date = value.(string)
				case "Timestamp":
					state.Timestamp = value.(string)
				case "URI":
					state.URI = value.(string)
				case "Size":
					state.Size = value.(float64)
				case "Count":
					state.Count = int(value.(float64))
				case "Enabled":
					state.Enabled = value.(bool)
				default:
					return preference{}, fmt.Errorf("unexpected business field %q", name)
				}
			}
			return state, nil
		},
	})
	if err != nil {
		return err
	}
	mcp.AddTool(server, &mcp.Tool{Name: "acceptance.stats"}, func(context.Context, *mcp.CallToolRequest, struct{}) (*mcp.CallToolResult, counters, error) {
		mu.Lock()
		defer mu.Unlock()
		return nil, stats, nil
	})
	// Inject bad caller-owned data to exercise complete returned-state checks.
	mcp.AddTool(server, &mcp.Tool{Name: "acceptance.corrupt"}, func(context.Context, *mcp.CallToolRequest, struct{}) (*mcp.CallToolResult, counters, error) {
		mu.Lock()
		defer mu.Unlock()
		state.Email = "bad"
		return nil, stats, nil
	})
	mcp.AddTool(server, &mcp.Tool{Name: "forms.check"}, func(_ context.Context, _ *mcp.CallToolRequest, input formInput) (*mcp.CallToolResult, formOutput, error) {
		form, err := forms.Parse([]byte(input.SchemaJSON))
		if err != nil {
			return nil, formOutput{}, err
		}
		output := formOutput{SchemaJSON: encode(form.Schema()), Unchanged: true}
		switch input.Operation {
		case "declaration":
		case "value":
			err = form.ValidateField("value", json.RawMessage(input.ValueJSON), nil)
		case "answer":
			var answer forms.Answer
			if err = decode(input.ContentJSON, &answer); err == nil {
				before := encode(answer)
				err = form.ValidateResult(&answer, nil)
				output.Unchanged = before == encode(answer)
			}
		case "upload", "prepare":
			var content map[string]any
			if err = decode(input.ContentJSON, &content); err == nil {
				before := encode(content)
				var result any
				if input.Operation == "upload" {
					result, err = form.CompleteSubmission("value", content, input.Uploaded, nil)
				} else {
					result, err = form.PrepareSubmission("value", content, input.Pending, nil)
				}
				if err == nil {
					output.ValueJSON = encode(result)
				}
				output.Unchanged = before == encode(content)
			}
		default:
			err = errors.New("unknown acceptance operation")
		}
		return nil, output, err
	})
	constant := any("bolt")
	model, err := forms.NewModelWithOptions[selection](map[string]forms.Field{"quantity": {Default: json.RawMessage(`2`)}}, func(_ context.Context, value selection) error {
		if value.Note != nil && *value.Note == "denied" {
			return errors.New("business selection rejected")
		}
		return nil
	}, &jsonschema.ForOptions{TypeSchemas: map[reflect.Type]*jsonschema.Schema{
		reflect.TypeFor[part](): {Type: "string", Const: &constant, Default: json.RawMessage(`"bolt"`), ReadOnly: true, Examples: []any{"bolt"}},
	}})
	if err != nil {
		return err
	}
	mcp.AddTool(server, &mcp.Tool{Name: "forms.decode"}, func(ctx context.Context, _ *mcp.CallToolRequest, input decodeInput) (*mcp.CallToolResult, decodeOutput, error) {
		var answer forms.Answer
		if err := decode(input.AnswerJSON, &answer); err != nil {
			return nil, decodeOutput{}, err
		}
		before := encode(answer)
		value, err := model.Decode(ctx, &answer, nil)
		return nil, decodeOutput{Value: value, SchemaJSON: encode(model.Form().Schema()), Unchanged: before == encode(answer)}, err
	})
	return server.Run(ctx, &mcp.StdioTransport{})
}

type fixture struct {
	Commit string `json:"upstream_commit"`
	Cases  []struct {
		Name   string
		Schema json.RawMessage
		Valid  bool
		Values []struct {
			Value json.RawMessage
			Valid bool
		}
		Uploads []struct {
			Content  map[string]any
			Uploaded []string
			Valid    bool
			Result   json.RawMessage
		}
	}
}

func run(ctx context.Context) error {
	self, err := os.Executable()
	if err != nil {
		return err
	}
	client := mcp.NewClient(&mcp.Implementation{Name: "new-extension-acceptance-client", Version: "2026-10-02"}, nil)
	session, err := client.Connect(ctx, &mcp.CommandTransport{Command: exec.CommandContext(ctx, self, "serve")}, nil)
	if err != nil {
		return err
	}
	defer session.Close()
	requests := 0
	call := func(name string, arguments any, wantError bool) (map[string]any, error) {
		requests++
		result, err := session.CallTool(ctx, &mcp.CallToolParams{Name: name, Arguments: arguments})
		if err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		if result.IsError != wantError {
			return nil, fmt.Errorf("%s: expected isError=%v, got %#v", name, wantError, result)
		}
		if wantError {
			return nil, nil
		}
		value, ok := result.StructuredContent.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("%s: missing structured result", name)
		}
		return value, nil
	}
	encoded, err := os.ReadFile("forms/testdata/python-parity.json")
	if err != nil {
		return err
	}
	var fixtures fixture
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.UseNumber()
	if err := decoder.Decode(&fixtures); err != nil {
		return err
	}
	if fixtures.Commit != "900032d8bd7c1566202d0cb1666986584f932043" {
		return errors.New("unexpected Python fixture baseline")
	}
	values, uploads := 0, 0
	for _, row := range fixtures.Cases {
		input := formInput{SchemaJSON: string(row.Schema), Operation: "declaration"}
		if _, err := call("forms.check", input, !row.Valid); err != nil {
			return fmt.Errorf("declaration %s: %w", row.Name, err)
		}
		if !row.Valid {
			continue
		}
		for _, value := range row.Values {
			input.Operation = "value"
			input.ValueJSON = string(value.Value)
			if _, err := call("forms.check", input, !value.Valid); err != nil {
				return fmt.Errorf("value %s %s: %w", row.Name, value.Value, err)
			}
			values++
		}
		for _, upload := range row.Uploads {
			input.Operation = "upload"
			input.ContentJSON = encode(upload.Content)
			input.Uploaded = upload.Uploaded
			result, err := call("forms.check", input, !upload.Valid)
			if err != nil {
				return fmt.Errorf("upload %s: %w", row.Name, err)
			}
			if upload.Valid {
				var got, want any
				if err := decode(result["valueJSON"].(string), &got); err != nil {
					return err
				}
				if err := decode(string(upload.Result), &want); err != nil {
					return err
				}
				if !reflect.DeepEqual(got, want) || result["unchanged"] != true {
					return fmt.Errorf("upload %s: result mismatch or mutation", row.Name)
				}
			}
			uploads++
		}
	}
	fmt.Printf("PASS: actual MCP subprocess: %d Python declarations, %d values and %d upload-reference cases\n", len(fixtures.Cases), values, uploads)
	fullSchema := `{"type":"object","required":["value"],"properties":{"value":{"type":"string","minLength":2,"x-openai-suggestions":[{"const":"bolt","title":"Bolt"}]}}}`
	for _, row := range []struct {
		answer    string
		wantError bool
	}{
		{`{"action":"accept","content":{"value":"custom"}}`, false},
		{`{"action":"accept","content":{}}`, true},
		{`{"action":"accept"}`, true},
		{`{"action":"accept","content":{"value":"bolt","unknown":true}}`, true},
		{`{"action":"accept","content":{"value":"x"}}`, true},
		{`{"action":"cancel"}`, false},
		{`{"action":"decline"}`, false},
		{`{"action":"cancel","content":{"value":"bolt"}}`, true},
		{`{"action":"other"}`, true},
	} {
		result, err := call("forms.check", formInput{SchemaJSON: fullSchema, Operation: "answer", ContentJSON: row.answer}, row.wantError)
		if err != nil {
			return err
		}
		if !row.wantError && result["unchanged"] != true {
			return errors.New("complete answer validation mutated the result")
		}
	}
	resourceSchema := `{"type":"object","properties":{"value":{"type":"array","items":{"type":"string","format":"uri"},"minItems":1,"maxItems":2,"uniqueItems":true,"x-openai-input":{"type":"resource","options":[{"uri":"parts://a","name":"A"}],"userOptions":{"kind":"file","accept":[".txt"]}}}}}`
	for _, row := range []struct {
		content   string
		pending   int
		wantError bool
	}{
		{`{"value":["parts://a"]}`, 1, false},
		{`{"value":["parts://a"]}`, 2, true},
		{`{"value":["bad uri"]}`, 1, true},
	} {
		result, err := call("forms.check", formInput{SchemaJSON: resourceSchema, Operation: "prepare", ContentJSON: row.content, Pending: row.pending}, row.wantError)
		if err != nil {
			return err
		}
		if !row.wantError && result["unchanged"] != true {
			return errors.New("upload preparation mutated the selection")
		}
	}
	fmt.Println("PASS: complete answer required/unknown fields, accept/cancel/decline, free suggestions and upload preparation/count limits")
	for _, row := range []struct {
		answer    string
		wantError bool
	}{
		{`{"action":"accept","content":{}}`, false},
		{`{"action":"accept","content":{"part_id":"bolt","quantity":3,"note":"ok"}}`, false},
		{`{"action":"accept","content":{"part_id":"washer"}}`, true},
		{`{"action":"accept","content":{"note":"denied"}}`, true},
		{`{"action":"accept","content":{"unknown":true}}`, true},
		{`{"action":"cancel"}`, true},
	} {
		result, err := call("forms.decode", decodeInput{AnswerJSON: row.answer}, row.wantError)
		if err != nil {
			return err
		}
		if !row.wantError {
			if result["unchanged"] != true {
				return errors.New("typed decoding mutated the answer")
			}
			if row.answer == `{"action":"accept","content":{}}` {
				value := result["value"].(map[string]any)
				if value["part_id"] != "bolt" || value["quantity"] != float64(2) {
					return errors.New("typed defaults failed")
				}
				if !strings.Contains(result["schemaJSON"].(string), `"readOnly":true`) {
					return errors.New("inferred annotations lost")
				}
			}
		}
	}
	fmt.Println("PASS: typed form aliases, named-type inference, static defaults, annotations, unchanged answers and business rejection")
	initial, err := call("settings.read", map[string]any{}, false)
	if err != nil {
		return err
	}
	initialValues := initial["values"].(map[string]any)
	if initialValues["is_enabled"] != false {
		return errors.New("typed effective zero was omitted")
	}
	for _, patch := range []map[string]any{
		{"email": "bad"}, {"email": "A <a@example.com>"}, {"date": "2025-02-29"}, {"timestamp": "2026-10-02"}, {"uri": "relative"},
		{"size": 1}, {"size": 5}, {"count": 3}, {"count": 12}, {"display_mode": "unknown"},
		{"display_label": "denied"}, {"display_label": "invalid"}, {"display_label": "  "}, {}, {"unknown": true},
	} {
		if _, err := call("settings.update", map[string]any{"set": patch}, true); err != nil {
			return err
		}
	}
	stats, err := call("acceptance.stats", map[string]any{}, false)
	if err != nil {
		return err
	}
	if stats["saves"] != float64(0) {
		return errors.New("invalid settings reached storage")
	}
	readback, err := call("settings.read", map[string]any{}, false)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(readback["values"], initialValues) {
		return errors.New("invalid settings changed state")
	}
	result, err := call("settings.update", map[string]any{"set": map[string]any{"display_label": "  valid  "}}, false)
	if err != nil {
		return err
	}
	if result["values"].(map[string]any)["display_label"] != "valid" {
		return errors.New("validator transformation failed")
	}
	stats, err = call("acceptance.stats", map[string]any{}, false)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(stats["patchNames"], []any{"Label"}) {
		return errors.New("JSON alias did not map to business field")
	}
	before := stats["validations"]
	for _, patch := range []map[string]any{{"display_mode": "fullscreen"}, {"count": 4}, {"size": 3}, {"is_enabled": true}, {"is_enabled": false}} {
		if _, err := call("settings.update", map[string]any{"set": patch}, false); err != nil {
			return err
		}
	}
	stats, err = call("acceptance.stats", map[string]any{}, false)
	if err != nil {
		return err
	}
	if stats["validations"] != before {
		return errors.New("omitted field triggered a validator")
	}
	result, err = call("settings.read", map[string]any{}, false)
	if err != nil {
		return err
	}
	state := result["values"].(map[string]any)
	if state["display_label"] != "valid" || state["display_mode"] != "fullscreen" || state["count"] != float64(4) || state["size"] != float64(3) || state["is_enabled"] != false || state["email"] != "a@example.com" {
		return errors.New("partial typed settings lost state")
	}
	fmt.Println("PASS: actual Settings tools: formats, exclusive bounds, multipleOf, aliases, field transformations, rejection before save, omitted fields and effective zero values")
	if _, err := call("acceptance.corrupt", map[string]any{}, false); err != nil {
		return err
	}
	if _, err := call("settings.read", map[string]any{}, true); err != nil {
		return err
	}
	if _, err := call("settings.update", map[string]any{"set": map[string]any{"size": 4}}, true); err != nil {
		return err
	}
	if _, err := call("settings.update", map[string]any{"set": map[string]any{"email": "a@example.com"}}, false); err != nil {
		return err
	}
	fmt.Println("PASS: invalid complete state from read/save is reported as an error; callbacks retain transaction responsibility")
	fmt.Printf("PASS: %d MCP tool calls; no original Workspace state or desktop process changed\n", requests)
	return nil
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	var err error
	if len(os.Args) > 1 && os.Args[1] == "serve" {
		err = serve(ctx)
	} else {
		err = run(ctx)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
