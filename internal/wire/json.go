// Package wire provides JSON snapshots for extension declarations.
package wire

import (
	"bytes"
	"encoding/json"
)

// Clone copies JSON values without retaining caller-owned maps or slices.
func Clone[T any](value T) (T, error) {
	return Convert[T](value)
}

// Convert snapshots a JSON value into a different wire type.
func Convert[T any](value any) (T, error) {
	var out T
	encoded, err := json.Marshal(value)
	if err != nil {
		return out, err
	}
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.UseNumber()
	err = decoder.Decode(&out)
	return out, err
}

// ConvertStrict converts a JSON snapshot while rejecting unknown object fields.
func ConvertStrict[T any](value any) (T, error) {
	var out T
	data, err := json.Marshal(value)
	if err != nil {
		return out, err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&out)
	return out, err
}
