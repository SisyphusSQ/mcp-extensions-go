package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/SisyphusSQ/mcp-extensions-go/internal/example"
)

// Run the official stdio transport in a real child process, without recompiling.
func TestStdioChild(t *testing.T) {
	if os.Getenv("MCP_STDIO_TEST_CHILD") != "1" {
		return
	}
	if err := run(context.Background()); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

func TestStdioProtocolAndRestart(t *testing.T) {
	directory := t.TempDir()
	html := filepath.Join(directory, "app.html")
	const content = "<!doctype html><title>trusted App</title>"
	if err := os.WriteFile(html, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	state := filepath.Join(directory, "settings.json")
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	connect := func() *mcp.ClientSession {
		command := exec.Command(os.Args[0], "-test.run=^TestStdioChild$")
		command.Env = append(os.Environ(), "MCP_STDIO_TEST_CHILD=1", "MCP_APP_HTML="+html, "MCP_SETTINGS_FILE="+state)
		client := mcp.NewClient(&mcp.Implementation{Name: "stdio-acceptance", Version: "0"}, nil)
		session, err := client.Connect(ctx, &mcp.CommandTransport{Command: command}, nil)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = session.Close() })
		return session
	}
	first := connect()
	tools, err := first.ListTools(ctx, nil)
	if err != nil || len(tools.Tools) != 5 {
		t.Fatalf("discovery: %v %v", tools, err)
	}
	resource, err := first.ReadResource(ctx, &mcp.ReadResourceParams{URI: example.AppURI})
	if err != nil || len(resource.Contents) != 1 || resource.Contents[0].Text != content {
		t.Fatalf("App resource: %v %v", resource, err)
	}
	other := connect()
	var wg sync.WaitGroup
	for _, update := range []struct {
		session *mcp.ClientSession
		patch   map[string]any
	}{
		{first, map[string]any{"units": "in"}}, {other, map[string]any{"showGrid": true}},
	} {
		wg.Go(func() {
			updated, err := update.session.CallTool(ctx, &mcp.CallToolParams{Name: "settings.update", Arguments: map[string]any{"set": update.patch}})
			if err != nil || updated.IsError {
				t.Errorf("update: %v %v", updated, err)
			}
		})
	}
	wg.Wait()
	if err := other.Close(); err != nil {
		t.Fatal(err)
	}
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}
	second := connect()
	read, err := second.CallTool(ctx, &mcp.CallToolParams{Name: "settings.read", Arguments: map[string]any{}})
	if err != nil || read.IsError || read.StructuredContent.(map[string]any)["values"].(map[string]any)["units"] != "in" {
		t.Fatalf("restart lost state: %v %v", read, err)
	}
	if read.StructuredContent.(map[string]any)["values"].(map[string]any)["showGrid"] != true {
		t.Fatal("concurrent process lost unrelated patch")
	}
}
