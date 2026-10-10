package ui

import "github.com/modelcontextprotocol/go-sdk/mcp"

const (
	// MIMEType is the media type for MCP App HTML resources.
	MIMEType = "text/html;profile=mcp-app"
	// OpenAIMetaKey identifies OpenAI UI extension metadata.
	OpenAIMetaKey = "openai/ui"
	// AppMetaKey identifies standard MCP Apps metadata.
	AppMetaKey = "ui"
)

// DisplayMode describes a resource display mode; support depends on the host.
type DisplayMode string

const (
	// Inline requests inline display.
	Inline DisplayMode = "inline"
	// Fullscreen requests fullscreen display.
	Fullscreen DisplayMode = "fullscreen"
	// PiP is the upstream picture-in-picture value; host support is not guaranteed.
	PiP DisplayMode = "pip"
)

// Visibility describes app or model tool visibility, not access control.
type Visibility string

const (
	// App makes a tool visible to the app.
	App Visibility = "app"
	// Model makes a tool visible to the model.
	Model Visibility = "model"
)

// EntrypointType identifies an app launch surface in the host.
type EntrypointType string

const (
	// Global identifies the global sidebar entrypoint.
	Global EntrypointType = "global"
	// Thread identifies the conversation entrypoint.
	Thread EntrypointType = "thread"
	// File identifies the file entrypoint.
	File EntrypointType = "file"
	// Settings identifies the removed settings App entrypoint.
	// Deprecated: use structured settings or a tool button in its layout.
	Settings EntrypointType = "settings"
)

// Entrypoint is discriminated by Type. Metadata rejects fields from other variants.
// File requires non-nil Extensions. An empty slice encodes as [] per upstream.
type Entrypoint struct {
	Type        EntrypointType `json:"type"`
	Extensions  []string       `json:"extensions,omitzero"`
	QuickAction *QuickAction   `json:"quickAction,omitempty"`
	// Deprecated: settings App entrypoints were removed upstream in v0.2.0.
	SearchTerms []string `json:"searchTerms,omitempty"`
}

// QuickAction describes a shortcut tool call on a global entrypoint.
type QuickAction struct {
	Title  string            `json:"title"`
	Icons  []mcp.Icon        `json:"icons"`
	Target QuickActionTarget `json:"target"`
}

// QuickActionTarget identifies the tool to invoke; Type must be "tool".
// Arguments must encode as JSON; this package does not evaluate expressions.
type QuickActionTarget struct {
	Type      string         `json:"type"`
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments,omitempty"`
}

// ToolMetadata combines standard MCP Apps bindings with OpenAI launch metadata.
type ToolMetadata struct {
	ResourceURI               string
	Visibility                []Visibility
	Entrypoints               []Entrypoint
	PreferredModelDisplayMode DisplayMode
}

// ResourceMetadata describes OpenAI display hints on HTML resource contents.
type ResourceMetadata struct {
	AvailableDisplayModes []DisplayMode `json:"availableDisplayModes,omitempty"`
	PreferredDisplayMode  DisplayMode   `json:"preferredDisplayMode,omitempty"`
}
