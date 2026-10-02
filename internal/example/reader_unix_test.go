//go:build darwin || linux || freebsd || openbsd || netbsd || dragonfly

package example_test

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/SisyphusSQ/mcp-extensions-go/internal/example"
)

func TestReaderRejectsFIFOWithoutBlocking(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "pipe")
	if err := syscall.Mkfifo(path, 0600); err != nil {
		t.Fatal(err)
	}
	reader, err := example.OpenReader(root, 4)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = reader.Close() })
	done := make(chan error, 1)
	go func() { _, err := reader.ReadFile(t.Context(), path); done <- err }()
	select {
	case err := <-done:
		if !errors.Is(err, example.ErrNotRegular) {
			t.Fatalf("pipe error = %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("FIFO open blocked")
	}
}

func TestReaderSymlinkReplacementCannotEscape(t *testing.T) {
	root, outside := t.TempDir(), t.TempDir()
	for path, text := range map[string]string{filepath.Join(root, "part"): "bolt", filepath.Join(outside, "secret"): "evil"} {
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	reader, err := example.OpenReader(root, 4)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = reader.Close() })
	var wg sync.WaitGroup
	wg.Go(func() {
		for i := range 100 {
			target := "part"
			if i%2 == 0 {
				target = filepath.Join(outside, "secret")
			}
			if err := os.Symlink(target, filepath.Join(root, "next")); err != nil {
				t.Error(err)
				return
			}
			if err := os.Rename(filepath.Join(root, "next"), filepath.Join(root, "link")); err != nil {
				t.Error(err)
				return
			}
		}
	})
	for range 200 {
		content, err := reader.ReadFile(t.Context(), filepath.Join(root, "link"))
		if err == nil && string(content) != "bolt" {
			t.Fatalf("read outside the root during replacement: %q", content)
		}
	}
	wg.Wait()
}
