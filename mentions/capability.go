package mentions

import (
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/SisyphusSQ/mcp-extensions-go/internal/wire"
)

// CapabilityKey identifies modern composer mention discovery.
const CapabilityKey = "openai/mentions"

// Capability names the same-server, read-only, app-visible search tool.
type Capability struct {
	SearchTool string `json:"searchTool"`
}

// WithCapability returns independent server options advertising SearchTool in
// modern and legacy discovery. Use them when constructing the official server,
// including with settings.NewServer, then register that same name with AddTool
// before connecting. This does not register or introspect a tool. Other options
// and capabilities are preserved; duplicate mention declarations are rejected.
func WithCapability(options *mcp.ServerOptions, searchTool string) (*mcp.ServerOptions, error) {
	if strings.TrimSpace(searchTool) == "" {
		return nil, fmt.Errorf("mention search tool name is required")
	}
	var opts mcp.ServerOptions
	if options != nil {
		opts = *options
	}
	caps, err := wire.Clone(opts.Capabilities)
	if err != nil {
		return nil, fmt.Errorf("mention capabilities: %w", err)
	}
	if caps == nil {
		caps = &mcp.ServerCapabilities{Logging: &mcp.LoggingCapabilities{}}
	}
	if _, exists := caps.Extensions[CapabilityKey]; exists {
		return nil, fmt.Errorf("mention capability is already configured")
	}
	if _, exists := caps.Experimental[CapabilityKey]; exists {
		return nil, fmt.Errorf("legacy mention capability is already configured")
	}
	if caps.Extensions == nil {
		caps.Extensions = make(map[string]any)
	}
	if caps.Experimental == nil {
		caps.Experimental = make(map[string]any)
	}
	caps.Extensions[CapabilityKey] = Capability{SearchTool: searchTool}
	caps.Experimental[CapabilityKey] = Capability{SearchTool: searchTool}
	opts.Capabilities = caps
	return &opts, nil
}
