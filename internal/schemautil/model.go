package schemautil

import (
	"bytes"
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/google/jsonschema-go/jsonschema"
)

// Model uses public schema inference for a flat Go struct. Nullable pointer
// properties become optional non-null form values; omission is preserved by Go
// pointer fields. Nested objects, arbitrary unions and non-string arrays fail.
func Model[T any](options *jsonschema.ForOptions, optional bool) (*jsonschema.Schema, error) {
	if reflect.TypeFor[T]().Kind() != reflect.Struct {
		return nil, fmt.Errorf("model must be a flat Go struct")
	}
	if options != nil && options.IgnoreInvalidTypes {
		return nil, fmt.Errorf("extension models cannot silently omit unsupported fields")
	}
	schema, err := jsonschema.For[T](options)
	if err != nil {
		return nil, err
	}
	if schema.Type != "object" {
		return nil, fmt.Errorf("model must encode an object")
	}
	for name, property := range schema.Properties {
		if err := modelField(property, false, optional); err != nil {
			return nil, fmt.Errorf("field %q: %w", name, err)
		}
		if property.Title == "" {
			property.Title = name
		}
	}
	return schema, nil
}

// FieldNames maps JSON aliases back to exported Go business field names.
func FieldNames[T any](properties map[string]*jsonschema.Schema) map[string]string {
	names := make(map[string]string)
	for _, field := range reflect.VisibleFields(reflect.TypeFor[T]()) {
		if !field.IsExported() {
			continue
		}
		name := field.Name
		if tag, ok := field.Tag.Lookup("json"); ok {
			alias, _, _ := strings.Cut(tag, ",")
			if alias == "-" {
				continue
			}
			if alias != "" {
				name = alias
			}
		}
		if _, exists := properties[name]; exists {
			names[name] = field.Name
		}
	}
	return names
}

func modelField(schema *jsonschema.Schema, item, optional bool) error {
	if schema == nil {
		return fmt.Errorf("missing field schema")
	}
	if schema.Const != nil && schema.Type == "string" {
		schema.Enum = []any{*schema.Const}
		schema.Const = nil
	}
	if optional && bytes.Equal(bytes.TrimSpace(schema.Default), []byte("null")) {
		schema.Default = nil
	}
	if len(schema.Types) == 2 && slices.Contains(schema.Types, "null") {
		if !optional {
			return fmt.Errorf("settings require non-null primitive fields")
		}
		for _, kind := range schema.Types {
			if kind != "null" {
				schema.Type = kind
			}
		}
		schema.Types = nil
	}
	if len(schema.Types) != 0 || schema.Ref != "" || len(schema.AnyOf) != 0 || len(schema.OneOf) != 0 || len(schema.AllOf) != 0 {
		return fmt.Errorf("unsupported model union or reference")
	}
	if item && schema.Type != "string" {
		return fmt.Errorf("arrays require string items")
	}
	switch schema.Type {
	case "string", "boolean", "number", "integer":
		return nil
	case "array":
		return modelField(schema.Items, true, optional)
	default:
		return fmt.Errorf("unsupported nested model type %q", schema.Type)
	}
}
