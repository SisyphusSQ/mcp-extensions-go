// The example requests an extended form over official stdio without persisting
// answers or performing business mutations. Modern host form support is required.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/SisyphusSQ/mcp-extensions-go/forms"
)

type selection struct {
	Code   string   `json:"code"`
	Images []string `json:"images"`
}

func run(ctx context.Context) error {
	model, err := forms.NewModel[selection](map[string]forms.Field{
		"code":   {Pattern: "^[A-Z]{3}$", Suggestions: []forms.Option{{Const: "ABC", Title: "Example"}}},
		"images": {Items: &forms.Field{Type: "string", Format: "uri"}, Input: &forms.ResourceInput{Type: "resource", Options: []*forms.ResourceOption{{Resource: mcp.Resource{URI: "demo://image", Name: "Demo image"}}}}},
	}, nil)
	if err != nil {
		return err
	}
	server := mcp.NewServer(&mcp.Implementation{Name: "extended-form-example", Version: "0.0.2"}, nil)
	mcp.AddTool(server, &mcp.Tool{Name: "choose", Title: "Choose demo resources", Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true}}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		// In a business tool, authorize every call and validate any caller-owned
		// continuation before requesting input or accepting a retry's answer.
		pending, answer, err := model.Form().RequestInput(ctx, req, forms.RequestOptions{Key: "selection", Message: "Enter a code and select demo images"})
		if err != nil || pending != nil {
			return pending, nil, err
		}
		if answer.Action != "accept" {
			return nil, map[string]string{"action": answer.Action}, nil
		}
		value, err := model.Decode(ctx, answer, nil)
		return nil, value, err
	})
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
