package forms_test

import (
	"encoding/json"
	"errors"
	"math"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/SisyphusSQ/mcp-extensions-go/forms"
)

func ptr[T any](value T) *T { return &value }

func newForm(t *testing.T, field forms.Field) *forms.Form {
	t.Helper()
	form, err := forms.New(forms.Schema{Type: "object", Properties: map[string]forms.Field{"value": field}, Required: []string{"value"}})
	if err != nil {
		t.Fatal(err)
	}
	return form
}

func TestRichChoicesSuggestionsAndUnchangedAnswers(t *testing.T) {
	icon := &mcp.Icon{Source: "https://example.com/bolt.png", Sizes: []string{"48x48"}}
	schema := forms.Schema{Type: "object", Properties: map[string]forms.Field{
		"part":     {Type: "string", OneOf: []forms.Option{{Const: "bolt", Title: "Bolt", Description: "Steel bolt", Thumbnail: icon, Preview: icon}}},
		"parts":    {Type: "array", Items: &forms.Field{AnyOf: []forms.Option{{Const: "bolt", Title: "Bolt"}, {Const: "washer", Title: "Washer"}}}, MinItems: ptr(1), MaxItems: ptr(2)},
		"label":    {Type: "string", MinLength: ptr(2), MaxLength: ptr(12), Pattern: "^[a-z]+$", Suggestions: []forms.Option{{Const: "steel", Title: "Steel"}}},
		"tags":     {Type: "array", Items: &forms.Field{Type: "string", Pattern: "^[a-z]+$", Suggestions: []forms.Option{{Const: "steel", Title: "Steel"}}}, UniqueItems: ptr(true), MaxItems: ptr(2)},
		"legacy":   {Type: "string", Enum: []any{"mm", "in"}, EnumNames: []string{"Millimetres", "Inches"}},
		"optional": {Type: "boolean", Default: json.RawMessage("true")},
	}, Required: []string{"part", "parts", "label", "tags", "legacy"}}
	form, err := forms.New(schema)
	if err != nil {
		t.Fatal(err)
	}
	icon.Source = "http://invalid.example"
	schema.Properties["part"] = forms.Field{Type: "boolean"}
	answer := map[string]any{"part": "bolt", "parts": []string{"bolt", "washer"}, "label": "custom", "tags": []string{"steel", "custom"}, "legacy": "mm"}
	before, _ := json.Marshal(answer)
	if err := form.Validate(answer, nil); err != nil {
		t.Fatal(err)
	}
	after, _ := json.Marshal(answer)
	if string(before) != string(after) {
		t.Fatal("answer mutated or default inserted")
	}
	declared := form.Schema()
	declared.Properties["part"].OneOf[0].Title = "mutated"
	if form.Schema().Properties["part"].OneOf[0].Title != "Bolt" {
		t.Fatal("declaration retained mutable aliases")
	}
	encoded, _ := json.Marshal(form.Schema())
	for _, keyword := range []string{`"oneOf"`, `"anyOf"`, `"x-openai-thumbnail"`, `"x-openai-preview"`, `"x-openai-suggestions"`, `"enumNames"`, `"uniqueItems":true`} {
		if !strings.Contains(string(encoded), keyword) {
			t.Fatalf("missing wire keyword %s: %s", keyword, encoded)
		}
	}
	if _, err := forms.Parse(encoded); err != nil {
		t.Fatal(err)
	}
	for name, values := range map[string][]any{
		"part": {"washer", true, nil}, "parts": {[]string{}, []string{"bolt", "bolt"}, []string{"other"}, "bolt"},
		"label": {"x", "UPPER", 1}, "tags": {[]string{"same", "same"}, []string{"UPPER"}, []any{"ok", 1}}, "legacy": {"ft"},
	} {
		for _, value := range values {
			if !errors.Is(form.ValidateField(name, value, nil), forms.ErrInvalidAnswer) {
				t.Fatalf("accepted %s=%#v", name, value)
			}
		}
	}
	for _, content := range []map[string]any{nil, {}, {"unknown": true}} {
		if form.Validate(content, nil) == nil {
			t.Fatalf("accepted malformed answer %#v", content)
		}
	}
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			for range 10 {
				if err := form.Validate(answer, nil); err != nil {
					t.Error(err)
				}
			}
		})
	}
	wg.Wait()
}

func TestPrimitiveFormatsAndNumericRules(t *testing.T) {
	for _, test := range []struct {
		format         string
		valid, invalid []string
	}{
		{"email", []string{"a@example.com"}, []string{"A <a@example.com>", "a", "a@example.com\n"}},
		{"uri", []string{"parts://bolt", "urn:part:bolt", "https://example.com/a?b=c#d", "https://[::1]/a", "file:///a", "urn:part:escaped%20name"}, []string{"relative", "parts://bad uri", "urn:x:%xy", "https://example.com\n", "urn:x:[bad]", "https://[not-ip]/a", "https://example.com/a#bad#fragment"}},
		{"date", []string{"2024-02-29"}, []string{"2025-02-29", "2026-1-01"}},
		{"date-time", []string{"2026-10-02T01:02:03Z", "2026-10-02t01:02:03.1z"}, []string{"2026-10-02", "2026-10-02T01:02:03+24:00", "2026-10-02T01:02:03,1Z"}},
	} {
		t.Run(test.format, func(t *testing.T) {
			form := newForm(t, forms.Field{Type: "string", Format: test.format})
			for _, value := range test.valid {
				if err := form.ValidateField("value", value, nil); err != nil {
					t.Fatalf("rejected %q: %v", value, err)
				}
			}
			for _, value := range test.invalid {
				if form.ValidateField("value", value, nil) == nil {
					t.Fatalf("accepted %q", value)
				}
			}
		})
	}
	form := newForm(t, forms.Field{Type: "integer", Minimum: ptr(1.0), Maximum: ptr(5.0)})
	for _, value := range []any{2, int64(4), json.Number("4")} {
		if err := form.ValidateField("value", value, nil); err != nil {
			t.Fatal(err)
		}
	}
	for _, value := range []any{0, 6, 2.5, "2", true, math.Inf(1), math.NaN()} {
		if form.ValidateField("value", value, nil) == nil {
			t.Fatalf("accepted numeric value %#v", value)
		}
	}
	wide := newForm(t, forms.Field{Type: "integer", Maximum: ptr(math.Exp2(64))})
	if wide.ValidateField("value", json.Number("18446744073709551617"), nil) == nil {
		t.Fatal("out-of-range integer was rounded into an allowed boundary")
	}
	for _, value := range []json.RawMessage{json.RawMessage("false"), json.RawMessage("0"), json.RawMessage(`""`), json.RawMessage(`[]`)} {
		kind := "string"
		var decoded any
		_ = json.Unmarshal(value, &decoded)
		switch decoded.(type) {
		case bool:
			kind = "boolean"
		case float64:
			kind = "number"
		case []any:
			kind = "array"
		}
		field := forms.Field{Type: kind, Default: value}
		if kind == "array" {
			field.Items = &forms.Field{Type: "string"}
		}
		got := newForm(t, field).Schema().Properties["value"].Default
		if string(got) != string(value) {
			t.Fatalf("default changed: %s", got)
		}
	}
}

func resourceField() forms.Field {
	return forms.Field{Type: "array", Items: &forms.Field{Type: "string", Format: "uri"}, MinItems: ptr(1), MaxItems: ptr(3), UniqueItems: ptr(true), Input: &forms.ResourceInput{
		Type: "resource", Options: []*forms.ResourceOption{{Resource: mcp.Resource{URI: "parts://bolt", Name: "Bolt", Meta: mcp.Meta{"vendor/x": "preserved"}}}},
		UserOptions: &forms.UserOptions{Kind: "file", Accept: []string{".txt", "text/plain", "image/*"}},
	}}
}

func TestResourcePolicyUploadsAndSelectionShapes(t *testing.T) {
	field := resourceField()
	form := newForm(t, field)
	policy := func(name, uri string) error {
		if name == "value" && uri == "host://authorized/file" {
			return nil
		}
		return errors.New("denied")
	}
	for _, selection := range []any{[]string{"parts://bolt"}, []string{"host://authorized/file"}} {
		if err := form.ValidateField("value", selection, policy); err != nil {
			t.Fatal(err)
		}
	}
	for _, selection := range []any{[]string{"host://denied"}, []string{"not a URI"}, "parts://bolt", []string{"parts://bolt", "parts://bolt"}} {
		if form.ValidateField("value", selection, policy) == nil {
			t.Fatalf("accepted selection %#v", selection)
		}
	}
	if form.ValidateField("value", []string{"host://authorized/file"}, nil) != nil {
		t.Fatal("declared user selection rejected without optional policy")
	}
	content := map[string]any{"value": []string{"parts://bolt"}}
	options, err := form.PrepareSubmission("value", content, 1, policy)
	if err != nil || !reflect.DeepEqual(options.Accept, field.Input.UserOptions.Accept) {
		t.Fatalf("prepare = %#v %v", options, err)
	}
	options.Accept[0] = "changed"
	value, err := form.CompleteSubmission("value", content, []string{"host://authorized/file"}, policy)
	if err != nil || !reflect.DeepEqual(value, []string{"parts://bolt", "host://authorized/file"}) {
		t.Fatalf("completion = %#v %v", value, err)
	}
	if !reflect.DeepEqual(content["value"], []string{"parts://bolt"}) {
		t.Fatal("upload mutated original content")
	}
	for _, pending := range []int{-1, 3, math.MaxInt} {
		if _, err := form.PrepareSubmission("value", content, pending, policy); err == nil {
			t.Fatalf("accepted pending count %d", pending)
		}
	}
	if _, err := form.PrepareSubmission("value", map[string]any{}, 1, policy); err != nil {
		t.Fatal(err)
	}
	if _, err := form.CompleteSubmission("value", content, []string{"host://authorized/file", "host://authorized/file"}, policy); err == nil {
		t.Fatal("duplicate uploads accepted")
	}
	if _, err := form.CompleteSubmission("value", content, []string{"host://denied"}, policy); err == nil {
		t.Fatal("denied upload accepted")
	}
	if _, err := form.CompleteSubmission("value", content, []string{"host://authorized/file"}, nil); err != nil {
		t.Fatal("host upload rejected without optional policy")
	}
	if _, err := form.CompleteSubmission("value", content, nil, policy); err == nil {
		t.Fatal("empty upload completion accepted")
	}
	implicit := resourceField()
	implicit.Input.Selection, implicit.Input.UserOptions = "implicit", nil
	if newForm(t, implicit).Schema().Properties["value"].Input.UserOptions.Kind != "file" {
		t.Fatal("implicit user selection missing")
	}
	single := forms.Field{Type: "string", Format: "uri", Input: &forms.ResourceInput{Type: "file", Options: []*forms.ResourceOption{}, UserOptions: &forms.UserOptions{Kind: "directory"}}}
	singleForm := newForm(t, single)
	if _, err := singleForm.CompleteSubmission("value", map[string]any{}, []string{"host://authorized/file"}, policy); err != nil {
		t.Fatal(err)
	}
	if _, err := singleForm.PrepareSubmission("value", map[string]any{"value": "parts://bolt"}, 1, policy); err == nil {
		t.Fatal("single upload replaced existing selection")
	}
	if _, err := singleForm.PrepareSubmission("value", map[string]any{}, 2, policy); err == nil {
		t.Fatal("multiple single uploads accepted")
	}
	encoded, _ := json.Marshal(form.Schema())
	if !strings.Contains(string(encoded), `"vendor/x":"preserved"`) {
		t.Fatal("resource metadata lost")
	}
}

func TestInvalidDefinitions(t *testing.T) {
	for name, field := range map[string]forms.Field{
		"nested": {Type: "object"}, "missing items": {Type: "array"}, "number items": {Type: "array", Items: &forms.Field{Type: "number"}},
		"inapplicable": {Type: "boolean", MinLength: ptr(1)}, "unknown format": {Type: "string", Format: "uuid"}, "bad regex": {Type: "string", Pattern: "(?=x)"},
		"empty choices":   {Type: "string", Enum: []any{}},
		"labels":          {Type: "string", Enum: []any{"a"}, EnumNames: []string{}},
		"http image":      {Type: "string", OneOf: []forms.Option{{Const: "a", Title: "A", Thumbnail: &mcp.Icon{Source: "http://example.com/a.png"}}}},
		"bad data image":  {Type: "string", OneOf: []forms.Option{{Const: "a", Title: "A", Thumbnail: &mcp.Icon{Source: "data:image/png;base64,%%%"}}}},
		"invalid default": {Type: "string", Enum: []any{"a"}, Default: json.RawMessage(`"b"`)}, "null default": {Type: "string", Default: json.RawMessage("null")},
		"empty range": {Type: "number", Minimum: ptr(2.0), Maximum: ptr(1.0)}, "nan": {Type: "number", Minimum: ptr(math.NaN())},
		"invalid array bounds": {Type: "array", Items: &forms.Field{Type: "string"}, MinItems: ptr(2), MaxItems: ptr(1)},
		"nonunique multi":      {Type: "array", Items: &forms.Field{Enum: []any{"a"}, Type: "string"}, UniqueItems: ptr(false)},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := forms.New(forms.Schema{Type: "object", Properties: map[string]forms.Field{"value": field}}); !errors.Is(err, forms.ErrInvalidDefinition) {
				t.Fatalf("definition accepted: %v", err)
			}
		})
	}
	for _, mutate := range []func(*forms.Field){
		func(f *forms.Field) { f.Input.Options = append(f.Input.Options, f.Input.Options[0]) },
		func(f *forms.Field) { f.Input.Options[0].URI = "relative" },
		func(f *forms.Field) { f.Input.Selection = "unknown" },
		func(f *forms.Field) { f.Input.Selection = "implicit"; f.Default = json.RawMessage(`[]`) },
		func(f *forms.Field) { f.Default = json.RawMessage(`["parts://unknown"]`) },
		func(f *forms.Field) { f.Input.UserOptions.Kind = "device" },
		func(f *forms.Field) { f.Input.UserOptions.Accept = []string{".TXT", ".txt"} },
		func(f *forms.Field) { f.Input.UserOptions.Accept = []string{".txt,.png"} },
		func(f *forms.Field) { f.Input.UserOptions.Accept = []string{" text/plain"} },
	} {
		field := resourceField()
		mutate(&field)
		if _, err := forms.New(forms.Schema{Type: "object", Properties: map[string]forms.Field{"value": field}}); err == nil {
			t.Fatal("invalid resource definition accepted")
		}
	}
	for _, schema := range []forms.Schema{
		{Type: "object"}, {Type: "object", Properties: map[string]forms.Field{}, Required: []string{"missing"}},
		{Type: "object", Properties: map[string]forms.Field{}, AdditionalProperties: true},
	} {
		if _, err := forms.New(schema); err == nil {
			t.Fatal("invalid object definition accepted")
		}
	}
	for _, schema := range []string{`{"type":"object","properties":{"x":{"type":"string","unknown":true}}}`, `{"type":"object","properties":{}} {}`, `{"type":"object","properties":{"x":{"type":["string","null"]}}}`} {
		if _, err := forms.Parse([]byte(schema)); err == nil {
			t.Fatalf("unsupported JSON schema accepted %s", schema)
		}
	}
	image := newForm(t, forms.Field{Type: "string", OneOf: []forms.Option{{Const: "a", Title: "A", Thumbnail: &mcp.Icon{Source: "data:image/png;base64,YQ=="}}}})
	if image.ValidateResult(&forms.Answer{Action: "accept", Content: map[string]any{"value": "a"}}, nil) != nil {
		t.Fatal("accepted result failed")
	}
	for _, answer := range []*forms.Answer{nil, {Action: "accept"}, {Action: "unknown"}, {Action: "cancel", Content: map[string]any{"value": "a"}}} {
		if image.ValidateResult(answer, nil) == nil {
			t.Fatal("invalid result accepted")
		}
	}
	for _, action := range []string{"cancel", "decline"} {
		if image.ValidateResult(&forms.Answer{Action: action}, nil) != nil {
			t.Fatal("terminal action rejected")
		}
	}
}
