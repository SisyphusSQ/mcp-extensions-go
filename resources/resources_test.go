package resources_test

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/SisyphusSQ/mcp-extensions-go/internal/testutil"
	"github.com/SisyphusSQ/mcp-extensions-go/resources"
)

func TestResourceMetadata(t *testing.T) {
	if _, present, err := resources.Path(nil); err != nil || present {
		t.Fatal("missing context should be absent")
	}
	if _, present, err := resources.Path(mcp.Meta{resources.MetadataKey: nil}); err != nil || present {
		t.Fatal("null Python file context should be absent")
	}
	for _, path := range []string{"", "/workspace/file", "relative"} {
		got, present, err := resources.Path(mcp.Meta{resources.MetadataKey: map[string]any{"path": path, "future": true}})
		if err != nil || !present || got != path {
			t.Fatalf("parse path = %q %v %v", got, present, err)
		}
	}
	for _, value := range []any{1, []string{}, map[string]any{}, map[string]any{"path": nil}, map[string]any{"path": true}} {
		if _, _, err := resources.Path(mcp.Meta{resources.MetadataKey: value}); err == nil {
			t.Fatalf("bad path metadata accepted: %#v", value)
		}
	}
	read, err := resources.ParseReadMetadata(mcp.Meta{resources.MetadataKey: map[string]any{"representation": "blob"}})
	if err != nil || read.Representation != resources.Blob {
		t.Fatalf("read metadata = %#v %v", read, err)
	}
	for _, value := range []any{"auto", nil, 1} {
		if _, err := resources.ParseReadMetadata(mcp.Meta{resources.MetadataKey: map[string]any{"representation": value}}); err == nil {
			t.Fatal("invalid representation accepted")
		}
	}
	content, err := resources.ParseContentMetadata(mcp.Meta{resources.MetadataKey: map[string]any{"writable": false, "etag": "v1"}})
	if err != nil || content.Writable == nil || *content.Writable || content.ETag != "v1" {
		t.Fatalf("content metadata = %#v %v", content, err)
	}
	for _, value := range []map[string]any{{"writable": "true"}, {"writable": nil}, {"etag": 1}, {"etag": nil}} {
		if _, err := resources.ParseContentMetadata(mcp.Meta{resources.MetadataKey: value}); err == nil {
			t.Fatal("invalid content metadata accepted")
		}
	}
}

func TestFileInputThroughSDK(t *testing.T) {
	server := mcp.NewServer(&mcp.Implementation{Name: "file", Version: "0"}, nil)
	mcp.AddTool(server, &mcp.Tool{Name: "open_file"}, func(_ context.Context, _ *mcp.CallToolRequest, input resources.FileInput) (*mcp.CallToolResult, any, error) {
		return nil, nil, input.Validate()
	})
	cs := testutil.Connect(t, server, nil, "")
	for _, args := range []any{map[string]any{}, map[string]any{"file": map[string]any{"name": "part.stl"}}, map[string]any{"file": map[string]any{"name": "  ", "resourceUri": "host-resource://part"}}} {
		res, err := cs.CallTool(t.Context(), &mcp.CallToolParams{Name: "open_file", Arguments: args})
		if err != nil || !res.IsError {
			t.Fatalf("invalid file input accepted: %#v %v", res, err)
		}
	}
	res, err := cs.CallTool(t.Context(), &mcp.CallToolParams{Name: "open_file", Arguments: map[string]any{"file": map[string]any{"name": "part.stl", "resourceUri": "host-resource://part"}}})
	if err != nil || res.IsError {
		t.Fatalf("valid file input rejected: %#v %v", res, err)
	}
	if err := (resources.FileInput{File: resources.File{Name: "../part.stl", ResourceURI: "host-resource://part"}}).Validate(); err != nil {
		t.Fatal("opaque host filename treated as a local path", err)
	}
}
