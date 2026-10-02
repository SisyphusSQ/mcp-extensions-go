package forms_test

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/SisyphusSQ/mcp-extensions-go/forms"
	"github.com/google/jsonschema-go/jsonschema"
)

type unit string
type inferredSelection struct {
	Unit unit    `json:"unit_alias"`
	Note *string `json:"note,omitempty"`
}

func TestNamedTypeInferenceDefaultsAndTypedBinding(t *testing.T) {
	constant := any("mm")
	options := &jsonschema.ForOptions{TypeSchemas: map[reflect.Type]*jsonschema.Schema{
		reflect.TypeFor[unit](): {Type: "string", Const: &constant, Default: json.RawMessage(`"mm"`), Title: "Unit", ReadOnly: true, Examples: []any{"mm"}},
	}}
	called := false
	model, err := forms.NewModelWithOptions[inferredSelection](nil, func(_ context.Context, value inferredSelection) error {
		called = true
		if value.Unit != "mm" || value.Note != nil {
			t.Errorf("typed defaults/omission=%#v", value)
		}
		return nil
	}, options)
	if err != nil {
		t.Fatal(err)
	}
	options.TypeSchemas[reflect.TypeFor[unit]()].Enum = []any{"changed"}
	answer := &forms.Answer{Action: "accept", Content: map[string]any{}}
	value, err := model.Decode(t.Context(), answer, nil)
	if err != nil || !called || value.Unit != "mm" || len(answer.Content) != 0 {
		t.Fatal("typed defaults mutated answer or failed", value, err)
	}
	field := model.Form().Schema().Properties["unit_alias"]
	if field.ReadOnly == nil || !*field.ReadOnly || !reflect.DeepEqual(field.Enum, []any{"mm"}) {
		t.Fatal("inference annotations lost", field)
	}
	if _, err := model.Decode(t.Context(), &forms.Answer{Action: "accept", Content: map[string]any{"unit_alias": "in"}}, nil); err == nil {
		t.Fatal("const restriction lost")
	}
	options.TypeSchemas[reflect.TypeFor[unit]()] = &jsonschema.Schema{Type: "string", Not: &jsonschema.Schema{Type: "string"}}
	if _, err := forms.NewModelWithOptions[inferredSelection](nil, nil, options); err == nil {
		t.Fatal("unsupported inferred constraint silently discarded")
	}
}

func TestResourceExtrasAndAnnotationSnapshots(t *testing.T) {
	form, err := forms.Parse([]byte(`{"type":"object","properties":{"resource":{"type":"string","format":"uri","_meta":{"null":null},"x-openai-input":{"type":"resource","options":[{"uri":"parts://a","URI":"parts://other","name":"A","vendor":{"value":null}}]}}}}`))
	if err != nil {
		t.Fatal(err)
	}
	copy := form.Schema()
	copy.Properties["resource"].Input.Options[0].Extra["vendor"] = json.RawMessage(`false`)
	copy.Properties["resource"].Meta["null"] = true
	field := form.Schema().Properties["resource"]
	if field.Input.Options[0].URI != "parts://a" || string(field.Input.Options[0].Extra["URI"]) != `"parts://other"` {
		t.Fatal("opaque descriptor replaced canonical uri")
	}
	if string(field.Input.Options[0].Extra["vendor"]) != `{"value":null}` || field.Meta["null"] != nil {
		t.Fatal("snapshot aliases retained")
	}
	if _, err := forms.New(forms.Schema{Type: "object", Properties: map[string]forms.Field{"resource": {Type: "string", Format: "uri", Input: &forms.ResourceInput{Type: "resource", Options: []*forms.ResourceOption{{Extra: map[string]json.RawMessage{"uri": json.RawMessage(`"parts://a"`)}}}}}}}); err == nil {
		t.Fatal("extra replaced standard resource key")
	}
}

func TestStandaloneEnumIntegerPrecision(t *testing.T) {
	form := newForm(t, forms.Field{Type: "integer", Enum: []any{uint64(9007199254740993)}})
	if err := form.ValidateField("value", json.Number("9007199254740993"), nil); err != nil {
		t.Fatal("declared integer rounded", err)
	}
	if err := form.ValidateField("value", json.Number("9007199254740992"), nil); err == nil {
		t.Fatal("neighboring integer accepted")
	}
}
