package schemautil

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/SisyphusSQ/mcp-extensions-go/internal/wire"
)

// Value snapshots JSON and makes JSON numbers usable by jsonschema-go without
// coercing strings. Integer wire values retain signed/unsigned 64-bit precision.
func Value(value any) (any, error) {
	copy, err := wire.Convert[any](value)
	if err != nil {
		return nil, err
	}
	return numbers(copy)
}

func numbers(value any) (any, error) {
	switch value := value.(type) {
	case json.Number:
		if n, err := strconv.ParseInt(string(value), 10, 64); err == nil {
			return n, nil
		}
		if n, err := strconv.ParseUint(string(value), 10, 64); err == nil {
			return n, nil
		}
		if !strings.ContainsAny(string(value), ".eE") {
			return nil, fmt.Errorf("integer is outside the supported 64-bit range")
		}
		n, err := strconv.ParseFloat(string(value), 64)
		if err != nil || math.IsNaN(n) || math.IsInf(n, 0) {
			return nil, fmt.Errorf("number is outside the supported finite range")
		}
		return n, nil
	case map[string]any:
		for key, item := range value {
			converted, err := numbers(item)
			if err != nil {
				return nil, err
			}
			value[key] = converted
		}
	case []any:
		for i, item := range value {
			converted, err := numbers(item)
			if err != nil {
				return nil, err
			}
			value[i] = converted
		}
	}
	return value, nil
}
