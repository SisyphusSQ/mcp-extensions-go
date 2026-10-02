package sdkcheck

import (
	"encoding/json"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// The public SDK response carrier uses map[string]any with float64 numbers.
// Standalone typed decoding from json.Number cannot fix an earlier wire decode.
func TestInputResponseNumericPrecisionBoundary(t *testing.T) {
	var responses mcp.InputResponseMap
	if err := json.Unmarshal([]byte(`{"pick":{"action":"accept","content":{"id":9007199254740993}}}`), &responses); err != nil {
		t.Fatal(err)
	}
	answer := responses["pick"].(*mcp.ElicitResult)
	if answer.Content["id"] != float64(9007199254740992) {
		t.Fatalf("SDK numeric boundary changed: %#v", answer.Content["id"])
	}
	encoded, err := json.Marshal(answer.Content)
	if err != nil || string(encoded) != `{"id":9007199254740992}` {
		t.Fatalf("boundary encoding = %s %v", encoded, err)
	}
}
