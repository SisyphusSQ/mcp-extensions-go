// The quickstart uses official stdio and public extension APIs only.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/SisyphusSQ/mcp-extensions-go/settings"
)

type preferences struct {
	Units    string `json:"units"`
	ShowGrid bool   `json:"showGrid"`
	Email    string `json:"email"`
	Label    string `json:"display_label"`
}

func run(ctx context.Context) error {
	// This local single-owner demo resets on restart. Replace these callbacks
	// with verified identity, authorization and transactional business storage.
	state := preferences{Units: "mm", ShowGrid: true, Email: "user@example.com", Label: "Workspace"}
	var mu sync.Mutex
	server, err := settings.NewModelServer(&mcp.Implementation{Name: "agent-quickstart", Version: "0.0.2"}, nil, settings.ModelConfig[preferences]{
		Fields: map[string]settings.Field{
			"units":         {Title: "Units", Enum: []string{"mm", "in"}},
			"showGrid":      {Title: "Show grid"},
			"email":         {Title: "Email", Format: "email"},
			"display_label": {Title: "Workspace label"},
		},
		FieldValidators: map[string]settings.FieldValidator{
			"display_label": func(_ context.Context, value any) (any, error) {
				label := strings.TrimSpace(value.(string))
				if label == "" {
					return nil, errors.New("workspace label is required")
				}
				return label, nil
			},
		},
		Read: func(ctx context.Context, _ *mcp.CallToolRequest) (preferences, error) {
			mu.Lock()
			defer mu.Unlock()
			return state, ctx.Err()
		},
		Update: func(ctx context.Context, _ *mcp.CallToolRequest, patch settings.Values) (preferences, error) {
			mu.Lock()
			defer mu.Unlock()
			if err := ctx.Err(); err != nil {
				return preferences{}, err
			}
			next := state
			// Model updates receive exported Go names; effective values retain
			// JSON aliases. Omitted fields stay unchanged, including false.
			for name, value := range patch {
				switch name {
				case "Units":
					next.Units = value.(string)
				case "ShowGrid":
					next.ShowGrid = value.(bool)
				case "Email":
					next.Email = value.(string)
				case "Label":
					next.Label = value.(string)
				default:
					return preferences{}, errors.New("unknown business setting")
				}
			}
			// Validate merged business rules here before committing next.
			state = next
			return state, nil
		},
	})
	if err != nil {
		return err
	}
	// Register other business tools on this same official *mcp.Server.
	return server.Run(ctx, &mcp.StdioTransport{})
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
