package ui

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// AddHTMLResource registers static HTML, using the SDK replacement semantics for duplicate URIs.
// resource.Meta is preserved on the descriptor and copied into contents; metadata replaces content openai/ui.
// HTML and JSON metadata are snapshotted at registration. Each read receives independent content metadata.
// HTML must be trusted by the server developer. This function does not sanitize HTML or read files or URLs.
func AddHTMLResource(server *mcp.Server, resource *mcp.Resource, html string, metadata ResourceMetadata) error {
	if server == nil || resource == nil {
		return fmt.Errorf("server and resource are required")
	}
	if err := validateURI(resource.URI); err != nil {
		return err
	}
	if strings.TrimSpace(resource.Name) == "" {
		return fmt.Errorf("resource name is required")
	}
	if resource.MIMEType != "" && resource.MIMEType != MIMEType {
		return fmt.Errorf("app resource must use %s", MIMEType)
	}
	if strings.TrimSpace(html) == "" {
		return fmt.Errorf("app html is required")
	}
	meta, err := metadata.Metadata(resource.Meta)
	if err != nil {
		return err
	}
	registered := *resource
	registered.Meta, err = snapshot(resource.Meta)
	if err != nil {
		return err
	}
	registered.MIMEType = MIMEType
	registered.Size = int64(len(html))
	uri := registered.URI
	server.AddResource(&registered, func(ctx context.Context, _ *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		contentMeta, err := snapshot(meta)
		if err != nil {
			return nil, err
		}
		return &mcp.ReadResourceResult{Contents: []*mcp.ResourceContents{{
			URI: uri, MIMEType: MIMEType, Text: html, Meta: contentMeta,
		}}}, nil
	})
	return nil
}
