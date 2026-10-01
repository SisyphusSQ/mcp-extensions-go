# mcp-extensions-go

[中文说明](README_ZH.md)

A private Go SDK framework for the server-side features of [OpenAI MCP Extensions](https://github.com/openai/mcp-extensions), built on the official [MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk). This is a personal project, not an official OpenAI Go SDK. No version has been released.

The official SDK owns JSON-RPC, tools, resources, schemas, sessions, and HTTP transports. This module adds extension metadata and registration helpers.

## Current features

- `ui.ToolMetadata.Metadata`: standard MCP Apps resource binding and visibility, plus OpenAI global/thread/file/settings entrypoints, quick actions, and model display preferences.
- `ui.ResourceMetadata.Metadata`: OpenAI resource display hints, preserving caller-provided standard MCP Apps CSP and other metadata.
- `ui.AddHTMLResource`: registers trusted static HTML as `text/html;profile=mcp-app`.
- `examples/http`: an authenticated Streamable HTTP server exposing `open_workspace` and its HTML resource.

Full forms, MRTR, mentions, file context, settings capability negotiation, and a browser-side SDK are not implemented. See the [compatibility matrix](docs/compatibility.md) for exact boundaries.

## Run the example

Requires Go 1.25.0 or newer. The official MCP Go SDK is pinned to v1.8.0. No Node.js or frontend dependency installation is needed.

```sh
# Run from the repository root.
export MCP_BEARER_TOKEN="$(openssl rand -hex 32)"
go run ./examples/http
```

The default endpoint is `http://127.0.0.1:8080/mcp`. Press Ctrl+C to stop. `MCP_BEARER_TOKEN` is required and must contain at least 32 bytes without whitespace. Never print or commit credentials.

Set `MCP_LISTEN_ADDR` to bind a specific private address when needed. Remote clients should connect through a configured HTTPS entrypoint with authentication and network access controls. The example does not configure TLS certificates, a reverse proxy, or a persistent background service.

Configure the host to use Streamable HTTP and supply `Authorization: Bearer <runtime credential>` through its authentication settings. Never commit a configuration containing the actual credential. Opening `/mcp` in a browser does not render the page: HTML is served by MCP `resources/read`.

The example HTML is static and contains no JavaScript. It does not implement the browser `App` bridge or the `ui/initialize` handshake, and is not evidence of Codex or ChatGPT host compatibility.

## Use in your server

Import `github.com/SisyphusSQ/mcp-extensions-go/ui` alongside `github.com/modelcontextprotocol/go-sdk/mcp`.

```go
meta, err := (ui.ToolMetadata{
    ResourceURI: "ui://my-app/home.html",
    Entrypoints: []ui.Entrypoint{{Type: ui.Global}, {Type: ui.Thread}},
    PreferredModelDisplayMode: ui.Inline,
}).Metadata(nil)
if err != nil {
    return err
}
tool := &mcp.Tool{Name: "open_app", Meta: meta}
// Register your typed business handler with mcp.AddTool(server, tool, handler).
mcp.AddTool(server, tool, handler)

err = ui.AddHTMLResource(server, &mcp.Resource{
    URI: "ui://my-app/home.html", Name: "home",
}, trustedHTML, ui.ResourceMetadata{
    AvailableDisplayModes: []ui.DisplayMode{ui.Inline, ui.Fullscreen},
    PreferredDisplayMode: ui.Inline,
})
if err != nil {
    return err
}
```

Other consumers of this private module need GitHub repository access, appropriate `GOPRIVATE` settings, and Git authentication in their own environment. This repository does not modify global Go settings.

`ToolMetadata.Metadata` preserves unrelated top-level keys but replaces `ui` and `openai/ui`. `ResourceMetadata.Metadata` replaces only `openai/ui`. Both return JSON snapshots without retaining aliases to caller-owned nested maps or slices. Metadata is a UI declaration, not authentication, authorization, or capability negotiation.

## Development and handoff

```sh
make fmt
make test vet build
```

The executable is written to `bin/mcp-extensions-http`, which is ignored by Git. Tests use official SDK in-memory transports and temporary local HTTP ports; they do not access external business services. The example has no business persistence or background installation behavior.

- [Documentation index](docs/README.md)
- [Architecture and extension path](docs/architecture.md)
- [Compatibility and scope](docs/compatibility.md)
- [Handoff record](docs/handoff.md)
