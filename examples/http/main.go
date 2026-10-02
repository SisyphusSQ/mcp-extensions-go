package main

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	_ "embed"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/SisyphusSQ/mcp-extensions-go/internal/example"
)

const appURI = example.AppURI

//go:embed app.html
var appHTML string

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	addr := os.Getenv("MCP_LISTEN_ADDR")
	if addr == "" {
		addr = "127.0.0.1:8080"
	}
	// Only the server operator may select this trusted, locally built app file.
	if path := os.Getenv("MCP_APP_HTML"); path != "" {
		reader, err := example.OpenReader(filepath.Dir(path), 2<<20)
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
	store, err := example.OpenStore(os.Getenv("MCP_SETTINGS_FILE"))
	if err != nil {
		return nil, err
	}
	server, err := example.NewServer(appHTML, store)
	if err != nil {
		return nil, err
	}
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
