package example_test

import (
	"context"
	"errors"
	"github.com/SisyphusSQ/mcp-extensions-go/internal/example"
	"os"
	"path/filepath"
	"testing"
)

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
	reader, err := example.OpenReader(root, 4)
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
	if _, err := reader.ReadFile(t.Context(), filepath.Join(root, "large")); !errors.Is(err, example.ErrTooLarge) {
		t.Fatalf("size limit = %v", err)
	}
	if _, err := reader.ReadFile(t.Context(), root); !errors.Is(err, example.ErrNotRegular) {
		t.Fatalf("directory = %v", err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := reader.ReadFile(ctx, filepath.Join(root, "part")); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation = %v", err)
	}
	if _, err := example.OpenReader(root, 0); err == nil {
		t.Fatal("unbounded reader accepted")
	}
}
