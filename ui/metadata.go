package ui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"slices"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Metadata returns an independent JSON snapshot, preserving other top-level keys.
// It replaces ui and openai/ui; configure both namespaces through this method.
func (m ToolMetadata) Metadata(base mcp.Meta) (mcp.Meta, error) {
	if err := validateURI(m.ResourceURI); err != nil {
		return nil, err
	}
	for _, v := range m.Visibility {
		if v != App && v != Model {
			return nil, fmt.Errorf("invalid tool visibility %q", v)
		}
	}
	if m.PreferredModelDisplayMode != "" && m.PreferredModelDisplayMode != Inline && m.PreferredModelDisplayMode != Fullscreen {
		return nil, fmt.Errorf("invalid model display mode %q", m.PreferredModelDisplayMode)
	}
	m.Entrypoints = slices.Clone(m.Entrypoints)
	for i, entry := range m.Entrypoints {
		entry.Extensions = slices.Clone(entry.Extensions)
		entry.SearchTerms = slices.Clone(entry.SearchTerms)
		for j, extension := range entry.Extensions {
			entry.Extensions[j] = strings.TrimSpace(extension)
		}
		for j, term := range entry.SearchTerms {
			entry.SearchTerms[j] = strings.TrimSpace(term)
		}
		if err := entry.validate(); err != nil {
			return nil, fmt.Errorf("entrypoint %d: %w", i, err)
		}
		m.Entrypoints[i] = entry
	}
	meta := make(mcp.Meta, len(base)+2)
	for key, value := range base {
		meta[key] = value
	}
	app := map[string]any{"resourceUri": m.ResourceURI}
	if len(m.Visibility) > 0 {
		app["visibility"] = m.Visibility
	}
	openai := map[string]any{}
	if m.Entrypoints != nil {
		openai["entrypoints"] = m.Entrypoints
	}
	if m.PreferredModelDisplayMode != "" {
		openai["preferredModelDisplayMode"] = m.PreferredModelDisplayMode
	}
	meta[AppMetaKey] = app
	meta[OpenAIMetaKey] = openai
	return snapshot(meta)
}

// Metadata returns a resource content snapshot, replacing only openai/ui.
// Standard MCP Apps CSP and permission declarations pass through base["ui"].
func (m ResourceMetadata) Metadata(base mcp.Meta) (mcp.Meta, error) {
	for _, mode := range m.AvailableDisplayModes {
		if !validResourceMode(mode) {
			return nil, fmt.Errorf("invalid resource display mode %q", mode)
		}
	}
	if m.PreferredDisplayMode != "" && !validResourceMode(m.PreferredDisplayMode) {
		return nil, fmt.Errorf("invalid preferred display mode %q", m.PreferredDisplayMode)
	}
	meta := make(mcp.Meta, len(base)+1)
	for key, value := range base {
		meta[key] = value
	}
	meta[OpenAIMetaKey] = m
	return snapshot(meta)
}

func validResourceMode(mode DisplayMode) bool {
	return mode == Inline || mode == Fullscreen || mode == PiP
}

func validateURI(uri string) error {
	u, err := url.Parse(uri)
	if err != nil {
		return fmt.Errorf("invalid app resource uri: %w", err)
	}
	if u.Scheme != "ui" || u.Host == "" || u.User != nil {
		return fmt.Errorf("app resource uri must use ui:// with a host and no userinfo")
	}
	return nil
}

func (e Entrypoint) validate() error {
	switch e.Type {
	case Global:
		if e.Extensions != nil || e.SearchTerms != nil {
			return fmt.Errorf("global entrypoint only allows quickAction")
		}
		if e.QuickAction != nil {
			q := e.QuickAction
			if strings.TrimSpace(q.Title) == "" || len(q.Icons) == 0 || q.Target.Type != "tool" || strings.TrimSpace(q.Target.Name) == "" {
				return fmt.Errorf("quickAction requires title, icons and a tool target name")
			}
			for _, icon := range q.Icons {
				if icon.Source == "" {
					return fmt.Errorf("quickAction icon requires src")
				}
			}
		}
	case Thread:
		if e.Extensions != nil || e.QuickAction != nil || e.SearchTerms != nil {
			return fmt.Errorf("thread entrypoint does not allow additional fields")
		}
	case File:
		if e.Extensions == nil || e.QuickAction != nil || e.SearchTerms != nil {
			return fmt.Errorf("file entrypoint requires extensions and no other fields")
		}
		for _, extension := range e.Extensions {
			if !strings.HasPrefix(extension, ".") {
				return fmt.Errorf("file extension must start with a dot")
			}
		}
	case Settings:
		if e.Extensions != nil || e.QuickAction != nil {
			return fmt.Errorf("settings entrypoint only allows searchTerms")
		}
		for _, term := range e.SearchTerms {
			if strings.TrimSpace(term) == "" {
				return fmt.Errorf("settings search term must not be blank")
			}
		}
	default:
		return fmt.Errorf("invalid entrypoint type %q", e.Type)
	}
	return nil
}

func snapshot(meta mcp.Meta) (mcp.Meta, error) {
	encoded, err := json.Marshal(meta)
	if err != nil {
		return nil, fmt.Errorf("encode ui metadata: %w", err)
	}
	var copied mcp.Meta
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	// Preserve integer precision in arbitrary metadata and quick-action arguments.
	decoder.UseNumber()
	if err := decoder.Decode(&copied); err != nil {
		return nil, fmt.Errorf("decode ui metadata snapshot: %w", err)
	}
	return copied, nil
}
