package forms_test

import (
	"encoding/json"
	"testing"

	"github.com/SisyphusSQ/mcp-extensions-go/forms"
)

func TestNumericAnswerBoundaries(t *testing.T) {
	for _, tt := range []struct {
		name  string
		field forms.Field
		value any
		valid bool
	}{
		{"integer", forms.Field{Type: "integer"}, json.Number("42"), true},
		{"integral decimal representation", forms.Field{Type: "integer"}, json.Number("1.0"), false},
		{"integer exponent representation", forms.Field{Type: "integer"}, json.Number("1e0"), false},
		{"standalone exact integer", forms.Field{Type: "integer"}, json.Number("9007199254740993"), true},
		{"integer outside 64 bits", forms.Field{Type: "integer"}, json.Number("18446744073709551616"), false},
		{"number exponent", forms.Field{Type: "number"}, json.Number("1e2"), true},
		{"fraction", forms.Field{Type: "number"}, json.Number("0.125"), true},
		{"numeric string is not number", forms.Field{Type: "number"}, "1.25", false},
		{"exact large integer text", forms.Field{Type: "string", Pattern: `^-?[0-9]+(\.[0-9]+)?$`}, "9007199254740993", true},
		{"exact decimal text", forms.Field{Type: "string", Pattern: `^-?[0-9]+(\.[0-9]+)?$`}, "0.12345678901234567890123456789", true},
		{"number is not exact text", forms.Field{Type: "string"}, 1.25, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			form, err := forms.New(forms.Schema{Type: "object", Properties: map[string]forms.Field{"value": tt.field}, Required: []string{"value"}})
			if err != nil {
				t.Fatal(err)
			}
			if err := form.Validate(map[string]any{"value": tt.value}, nil); (err == nil) != tt.valid {
				t.Fatalf("valid = %v, want %v: %v", err == nil, tt.valid, err)
			}
		})
	}
}
