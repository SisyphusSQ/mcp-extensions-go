package forms_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/SisyphusSQ/mcp-extensions-go/forms"
)

type reviewModel struct {
	Part     string   `json:"part_id" jsonschema:"The selected part"`
	Approved bool     `json:"approved"`
	Count    int      `json:"count"`
	Note     *string  `json:"note,omitempty"`
	Tags     []string `json:"tags,omitempty"`
}

func TestModelAliasesOptionalFieldsAndBusinessValidation(t *testing.T) {
	var calls int
	rejected := errors.New("business rule denied")
	model, err := forms.NewModel[reviewModel](map[string]forms.Field{"part_id": {OneOf: []forms.Option{{Const: "bolt", Title: "Bolt"}}}}, func(_ context.Context, value reviewModel) error {
		calls++
		if value.Count < 0 {
			return rejected
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	answer := &forms.Answer{Action: "accept", Content: map[string]any{"part_id": "bolt", "approved": false, "count": 0}}
	value, err := model.Decode(t.Context(), answer, nil)
	if err != nil || value.Part != "bolt" || value.Note != nil || value.Count != 0 || value.Approved || calls != 1 {
		t.Fatalf("decoded = %#v %v calls=%d", value, err, calls)
	}
	answer.Content["note"] = ""
	answer.Content["tags"] = []string{"steel"}
	value, err = model.Decode(t.Context(), answer, nil)
	if err != nil || value.Note == nil || *value.Note != "" {
		t.Fatalf("empty optional value lost: %#v %v", value, err)
	}
	answer.Content["count"] = -1
	if _, err := model.Decode(t.Context(), answer, nil); !errors.Is(err, rejected) {
		t.Fatalf("callback error lost: %v", err)
	}
	before := calls
	for _, invalid := range []any{"0", 0.5, nil} {
		answer.Content["count"] = invalid
		if _, err := model.Decode(t.Context(), answer, nil); err == nil {
			t.Fatalf("bad typed answer accepted %#v", invalid)
		}
	}
	if calls != before {
		t.Fatal("invalid answers reached business callback")
	}
	if _, err := model.Decode(t.Context(), &forms.Answer{Action: "cancel"}, nil); err == nil {
		t.Fatal("cancel decoded as accepted")
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := model.Decode(ctx, answer, nil); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if model.Form().Schema().Properties["part_id"].Description != "The selected part" {
		t.Fatal("description or alias lost")
	}
}

func TestModelRejectsUnsupportedShapesAndPreservesLargeIntegers(t *testing.T) {
	if _, err := forms.NewModel[struct{ Nested struct{ Name string } }](nil, nil); err == nil {
		t.Fatal("nested struct accepted")
	}
	if _, err := forms.NewModel[struct{ Values []int }](nil, nil); err == nil {
		t.Fatal("numeric array accepted")
	}
	if _, err := forms.NewModel[map[string]string](nil, nil); err == nil {
		t.Fatal("map model accepted")
	}
	if _, err := forms.NewModel[reviewModel](map[string]forms.Field{"missing": {Type: "string"}}, nil); err == nil {
		t.Fatal("unknown alias accepted")
	}
	if _, err := forms.NewModel[reviewModel](map[string]forms.Field{"count": {Type: "string"}}, nil); err == nil {
		t.Fatal("incompatible override accepted")
	}
	model, err := forms.NewModel[struct {
		ID uint64 `json:"id"`
	}](nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	value, err := model.Decode(t.Context(), &forms.Answer{Action: "accept", Content: map[string]any{"id": json.Number("18446744073709551615")}}, nil)
	if err != nil || value.ID != ^uint64(0) {
		t.Fatalf("integer precision lost: %#v %v", value, err)
	}
	if _, err := model.Decode(t.Context(), &forms.Answer{Action: "accept", Content: map[string]any{"id": -1}}, nil); err == nil {
		t.Fatal("unsigned model accepted negative number")
	}
}
