package main

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	_ "embed"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/SisyphusSQ/mcp-extensions-go/mentions"
	"github.com/SisyphusSQ/mcp-extensions-go/resources"
	"github.com/SisyphusSQ/mcp-extensions-go/settings"
	"github.com/SisyphusSQ/mcp-extensions-go/ui"
)

const appURI = "ui://workspace/home.html"

//go:embed app.html
var appHTML string

type openInput struct{}

type openOutput struct {
	Message string `json:"message" jsonschema:"The message shown on the app home page"`
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	addr := os.Getenv("MCP_LISTEN_ADDR")
	if addr == "" {
		addr = "127.0.0.1:8080"
	}
	// Only the server operator may select this trusted, locally built app file.
	if path := os.Getenv("MCP_APP_HTML"); path != "" {
		reader, err := resources.OpenReader(filepath.Dir(path), 2<<20)
		if err != nil {
			slog.Error("open trusted app directory", "error", err)
			os.Exit(1)
		}
		content, readErr := reader.ReadFile(ctx, path)
		closeErr := reader.Close()
		if err := errors.Join(readErr, closeErr); err != nil {
			slog.Error("read trusted app", "error", err)
			os.Exit(1)
		}
		appHTML = string(content)
	}
	if err := run(ctx, addr, os.Getenv("MCP_BEARER_TOKEN")); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

func newHandler(token string) (http.Handler, error) {
	if len(token) < 32 || strings.ContainsAny(token, " \t\r\n") {
		return nil, fmt.Errorf("MCP_BEARER_TOKEN requires at least 32 bytes without whitespace")
	}
	// This is a single-credential demo. Multi-user services must scope storage
	// to verified identity and recheck resource authorization in each handler.
	var mu sync.Mutex
	values := settings.Values{"units": "mm", "showGrid": false}
	server, err := settings.NewServer(&mcp.Implementation{Name: "mcp-extensions-go-example", Version: "0.0.0-dev"}, nil, settings.Config{
		Fields: map[string]settings.Field{
			"units":    {Type: "string", Title: "Measurement units", Enum: []string{"mm", "in"}},
			"showGrid": {Type: "boolean", Title: "Show grid"},
		},
		Layout: []settings.Group{{Kind: "group", Title: "Display", Items: []settings.Item{{Kind: "property", Property: "units"}, {Kind: "property", Property: "showGrid"}}}},
		Read: func(ctx context.Context, _ *mcp.CallToolRequest) (settings.Values, error) {
			mu.Lock()
			defer mu.Unlock()
			return maps.Clone(values), ctx.Err()
		},
		Update: func(ctx context.Context, _ *mcp.CallToolRequest, set settings.Values) (settings.Values, error) {
			mu.Lock()
			defer mu.Unlock()
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			maps.Copy(values, set)
			return maps.Clone(values), nil
		},
	})
	if err != nil {
		return nil, err
	}
	if err := ui.AddHTMLResource(server, &mcp.Resource{
		URI: appURI, Name: "workspace-home", Title: "Workspace", Description: "App home page",
	}, appHTML, ui.ResourceMetadata{
		AvailableDisplayModes: []ui.DisplayMode{ui.Inline, ui.Fullscreen},
		PreferredDisplayMode:  ui.Inline,
	}); err != nil {
		return nil, fmt.Errorf("register app resource: %w", err)
	}
	meta, err := (ui.ToolMetadata{
		ResourceURI:               appURI,
		Visibility:                []ui.Visibility{ui.App, ui.Model},
		Entrypoints:               []ui.Entrypoint{{Type: ui.Global}, {Type: ui.Thread}, {Type: ui.Settings, SearchTerms: []string{"units", "grid"}}},
		PreferredModelDisplayMode: ui.Inline,
	}).Metadata(nil)
	if err != nil {
		return nil, err
	}
	mcp.AddTool(server, &mcp.Tool{
		Name: "open_workspace", Title: "Open workspace", Description: "Open the app home page", Meta: meta,
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ openInput) (*mcp.CallToolResult, openOutput, error) {
		return nil, openOutput{Message: "Welcome to your workspace"}, ctx.Err()
	})
	if err := mentions.AddTool(server, &mcp.Tool{Name: "search_mentions", Title: "Search parts"}, func(ctx context.Context, _ *mcp.CallToolRequest, input mentions.SearchParams) (mentions.SearchResult, error) {
		result := mentions.SearchResult{Items: []mentions.Item{}}
		for _, name := range []string{"bolt", "washer"} {
			if strings.Contains(name, strings.ToLower(input.Query)) {
				result.Items = append(result.Items, mentions.Item{Link: &mcp.ResourceLink{URI: "parts://" + name, Name: name, MIMEType: "text/plain"}})
			}
		}
		return result, ctx.Err()
	}); err != nil {
		return nil, err
	}
	for _, name := range []string{"bolt", "washer"} {
		server.AddResource(&mcp.Resource{URI: "parts://" + name, Name: name, MIMEType: "text/plain"}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
			return &mcp.ReadResourceResult{Contents: []*mcp.ResourceContents{{URI: req.Params.URI, MIMEType: "text/plain", Text: "Demo part: " + name}}}, ctx.Err()
		})
	}
	fileMeta, err := (ui.ToolMetadata{ResourceURI: appURI, Visibility: []ui.Visibility{ui.App}, Entrypoints: []ui.Entrypoint{{Type: ui.File, Extensions: []string{".txt"}}}}).Metadata(nil)
	if err != nil {
		return nil, err
	}
	mcp.AddTool(server, &mcp.Tool{Name: "open_file", Title: "Text file viewer", Meta: fileMeta, Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true}}, func(ctx context.Context, _ *mcp.CallToolRequest, input resources.FileInput) (*mcp.CallToolResult, resources.FileInput, error) {
		if err := input.Validate(); err != nil {
			return nil, resources.FileInput{}, err
		}
		// Host resource reads happen through the App, never through this URI.
		return nil, input, ctx.Err()
	})
	mcpHandler := mcp.NewStreamableHTTPHandler(func(_ *http.Request) *mcp.Server { return server }, &mcp.StreamableHTTPOptions{
		Stateless: true, JSONResponse: true, MaxRequestBodyBytes: 1 << 20,
	})
	expected := sha256.Sum256([]byte(token))
	authenticated := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		value, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		actual := sha256.Sum256([]byte(value))
		if !ok || subtle.ConstantTimeCompare(actual[:], expected[:]) != 1 {
			w.Header().Set("WWW-Authenticate", "Bearer")
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		mcpHandler.ServeHTTP(w, r)
	})
	mux := http.NewServeMux()
	mux.Handle("/mcp", http.NewCrossOriginProtection().Handler(authenticated))
	return mux, nil
}

func run(ctx context.Context, addr, token string) error {
	handler, err := newHandler(token)
	if err != nil {
		return err
	}
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	server := &http.Server{
		Handler: handler, ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout: 15 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: time.Minute,
	}
	stopped := make(chan error, 1)
	go func() { stopped <- server.Serve(listener) }()
	slog.Info("mcp server started", "address", listener.Addr().String(), "path", "/mcp")
	select {
	case err := <-stopped:
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		shutdownErr := server.Shutdown(shutdownCtx)
		if shutdownErr != nil {
			shutdownErr = errors.Join(shutdownErr, server.Close())
		}
		serveErr := <-stopped
		if errors.Is(serveErr, http.ErrServerClosed) {
			serveErr = nil
		}
		return errors.Join(shutdownErr, serveErr)
	}
}
