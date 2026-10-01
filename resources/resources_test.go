package resources_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/SisyphusSQ/mcp-extensions-go/internal/testutil"
	"github.com/SisyphusSQ/mcp-extensions-go/resources"
)

func TestResourceMetadata(t *testing.T) {
	if _, present, err := resources.Path(nil); err != nil || present {
		t.Fatal("missing context should be absent")
	}
	for _, path := range []string{"", "/workspace/file", "relative"} {
		got, present, err := resources.Path(mcp.Meta{resources.MetadataKey: map[string]any{"path": path, "future": true}})
		if err != nil || !present || got != path {
			t.Fatalf("parse path = %q %v %v", got, present, err)
		}
	}
	for _, value := range []any{nil, 1, []string{}, map[string]any{}, map[string]any{"path": nil}, map[string]any{"path": true}} {
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
	for _, args := range []any{map[string]any{}, map[string]any{"file": map[string]any{"name": "part.stl"}}, map[string]any{"file": map[string]any{"name": "../part.stl", "resourceUri": "host-resource://part"}}} {
		res, err := cs.CallTool(t.Context(), &mcp.CallToolParams{Name: "open_file", Arguments: args})
		if err != nil || !res.IsError {
			t.Fatalf("invalid file input accepted: %#v %v", res, err)
		}
	}
	res, err := cs.CallTool(t.Context(), &mcp.CallToolParams{Name: "open_file", Arguments: map[string]any{"file": map[string]any{"name": "part.stl", "resourceUri": "host-resource://part"}}})
	if err != nil || res.IsError {
		t.Fatalf("valid file input rejected: %#v %v", res, err)
	}
}

func TestReaderContainmentAndLimits(t *testing.T) {
	root := t.TempDir()
	out := t.TempDir()
	write := func(path, content string) {
		t.Helper()
		if err := os.WriteFile(path, []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
	}
	write(filepath.Join(root, "part"), "bolt")
	write(filepath.Join(root, "large"), "012345")
	write(filepath.Join(out, "secret"), "outside")
	for name, target := range map[string]string{"inside": "part", "escape": filepath.Join(out, "secret")} {
		if err := os.Symlink(target, filepath.Join(root, name)); err != nil {
			t.Fatal(err)
		}
	}
	reader, err := resources.OpenReader(root, 4)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = reader.Close() })
	for _, name := range []string{"part", "inside"} {
		content, err := reader.ReadFile(t.Context(), filepath.Join(root, name))
		if err != nil || string(content) != "bolt" {
			t.Fatalf("read = %q %v", content, err)
		}
	}
	for _, path := range []string{"part", filepath.Join(out, "secret"), filepath.Join(root, "escape")} {
		if _, err := reader.ReadFile(t.Context(), path); err == nil {
			t.Fatal("path escape accepted")
		}
	}
	if _, err := reader.ReadFile(t.Context(), filepath.Join(root, "large")); !errors.Is(err, resources.ErrTooLarge) {
		t.Fatalf("size limit = %v", err)
	}
	if _, err := reader.ReadFile(t.Context(), root); !errors.Is(err, resources.ErrNotRegular) {
		t.Fatalf("directory = %v", err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := reader.ReadFile(ctx, filepath.Join(root, "part")); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation = %v", err)
	}
	if _, err := resources.OpenReader(root, 0); err == nil {
		t.Fatal("unbounded reader accepted")
	}
}
