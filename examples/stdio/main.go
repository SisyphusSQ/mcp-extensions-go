// The local Codex plugin uses official stdio transport under its parent process.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/SisyphusSQ/mcp-extensions-go/internal/example"
	"github.com/SisyphusSQ/mcp-extensions-go/resources"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		slog.Error("stdio server stopped", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	path, err := filepath.Abs(os.Getenv("MCP_APP_HTML"))
	if err != nil || os.Getenv("MCP_APP_HTML") == "" {
		return fmt.Errorf("MCP_APP_HTML must identify a trusted built App")
	}
	reader, err := resources.OpenReader(filepath.Dir(path), 2<<20)
	if err != nil {
		return err
	}
	content, readErr := reader.ReadFile(ctx, path)
	if err := errors.Join(readErr, reader.Close()); err != nil {
		return err
	}
	statePath := os.Getenv("MCP_SETTINGS_FILE")
	if statePath == "" {
		directory, err := os.UserConfigDir()
		if err != nil {
			return err
		}
		statePath = filepath.Join(directory, "mcp-extensions-go", "live-e2e", "settings.json")
	}
	store, err := example.OpenStore(statePath)
	if err != nil {
		return err
	}
	server, err := example.NewServer(string(content), store)
	if err != nil {
		return err
	}
	// stdout is reserved for the official MCP transport; all logs use stderr.
	return server.Run(ctx, &mcp.StdioTransport{})
}
