# mcp-extensions-go

[中文说明](README_ZH.md)

A public, personal Go library for the server-side features of [OpenAI MCP Extensions](https://github.com/openai/mcp-extensions), built on the official [MCP Go SDK v1.8.0](https://github.com/modelcontextprotocol/go-sdk). This is not an official OpenAI Go SDK. The first module release is [v0.0.1](https://github.com/SisyphusSQ/mcp-extensions-go/releases/tag/v0.0.1).

The official SDK owns JSON-RPC, schemas, tools, resources, discovery, sessions, transports and standard MRTR. This module adds server extension types and helpers. Browser/host behavior uses the official TypeScript App SDKs.

## Install and integrate

```sh
go get github.com/SisyphusSQ/mcp-extensions-go@v0.0.1
```

**For coding agents:** start with the [Agent integration guide](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/EN-Agent-Integration), or its [versioned source](https://github.com/SisyphusSQ/mcp-extensions-go/blob/v0.0.1/docs/agent-integration.md). It covers capability selection, adapting an existing server, business callbacks, public-only recipes and validation boundaries. A complete [stdio quickstart](examples/agent-quickstart/main.go) and the [new-extension acceptance runner](examples/acceptance/main.go) are included.

## Implemented

- `settings.NewServer`: real read/update tools, native primitive schemas and constraints, layout, full effective values, nonempty partial patches, and modern/legacy `openai/settings` discovery. The returned server is the ordinary official `*mcp.Server`.
- `mentions.AddTool`: searchable resource links and the upstream SDK's resource variant, `mentions/search` tool metadata, read-only hints and required app visibility.
- `resources`: typed file-entrypoint input, opaque resource references, `openai/resource` path/representation/write-hint parsing.
- `forms`: Python-extension-compatible flat declarations, rich choices/images, suggestions, resource selections, unchanged-answer/upload-reference validation and typed Go binding.
- `ui`: standard MCP Apps resource binding/visibility, OpenAI entrypoints, quick actions/display metadata, and trusted HTML resource registration.
- `examples/http`: authenticated stateless Streamable HTTP with real in-memory settings, searchable/readable demo parts, workspace/file tools, loopback defaults, request limits and graceful shutdown.
- `examples/frontend`: a bundled browser App using standard MCP Apps `App` plus OpenAI TypeScript extensions, with local browser integration tests through the official `AppBridge`.
- `examples/stdio` and a local Codex plugin: official stdio transport, bundled App and single-owner settings persisted across process restarts.

Settings storage, resource authorization, cross-field validation and transactions are caller-owned. Visibility and metadata never grant permission. HTTP defaults to a one-credential memory store; an absolute operator-owned `MCP_SETTINGS_FILE` enables persistent state. The local plugin uses the durable example store by default, with OS locking and atomic file replacement. This is not multi-user storage.

OpenAI extended form declarations, validation and Go model binding are implemented. Legacy `openai/elicitation/create` sending remains blocked by the pinned official SDK's public outbound API. Alignment covers MCP Extensions only: standard sessions/transports/MRTR, encrypted continuation recovery and tasks are not implemented here. See [forms and settings usage](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/EN-Forms) and [the capability matrix](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/EN-Compatibility). Local protocol tests do not establish real host rendering/interaction acceptance.

## Run the Go example

Requires Go 1.25.0 or newer. No Node dependency is needed for the default static page.

```sh
export MCP_BEARER_TOKEN="$(openssl rand -hex 32)"
go run ./examples/http
```

The endpoint defaults to `http://127.0.0.1:8080/mcp`. Configure a host with Streamable HTTP and `Authorization: Bearer <runtime credential>`. The token must contain at least 32 bytes without whitespace; never print or commit it. Ctrl+C stops the process. Opening `/mcp` in a browser does not render HTML; MCP `resources/read` supplies App resources.

For the real browser App, build `examples/frontend` and set `MCP_APP_HTML` to the absolute path of its trusted `dist/app.html`; follow [frontend setup and validation](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/EN-Frontend-Validation). Without that setting, the embedded page is static and performs no `ui/initialize` handshake.

`MCP_LISTEN_ADDR` selects an explicit address. Remote access requires caller-configured HTTPS termination, authentication and network controls. The example does not configure TLS or install a background service.

## Local Codex plugin

For a ready-to-install local Codex plugin, see [local plugin setup and human acceptance](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/EN-Live-E2E). It packages the official stdio example and opens no listening port. Build/install commands and the observed desktop acceptance boundary are recorded there. The source repository is public; the plugin has not been distributed through a public plugin directory.

## Use settings in a business server

```go
server, err := settings.NewServer(
    &mcp.Implementation{Name: "my-server", Version: "1"}, nil,
    settings.Config{
        Fields: map[string]settings.Field{
            "units": {Type: "string", Title: "Units", Enum: []string{"mm", "in"}},
        },
        Read: func(ctx context.Context, req *mcp.CallToolRequest) (settings.Values, error) {
            // Authorize using verified request identity and return all effective values.
            return store.Read(ctx, req)
        },
        Update: func(ctx context.Context, req *mcp.CallToolRequest, set settings.Values) (settings.Values, error) {
            // Authorize, preserve omitted fields, and persist atomically before returning.
            return store.Update(ctx, req, set)
        },
    },
)
if err != nil {
    return err
}
// Register other business tools normally. Reserve settings.read/settings.update.
mcp.AddTool(server, tool, handler)
```

Import `github.com/SisyphusSQ/mcp-extensions-go/settings` and the official `mcp` package. The factory snapshots declarations, validates definitions without invoking storage, and configures both tools and their capability together. See the runnable HTTP example for complete callbacks. Invalid arguments and ordinary storage/validation failures become MCP tool errors. Callbacks must use Go primitive values and return client-safe errors.

For mention search, use `mentions.AddTool(server, &mcp.Tool{Name: "search_mentions"}, searchHandler)`. Empty queries are valid. It preserves other metadata and ensures app visibility; no separate mentions capability is specified upstream.

For UI declarations, use `ui.ToolMetadata.Metadata` and `ui.AddHTMLResource` with official tools/resources. Metadata snapshots preserve arbitrary JSON number precision and ownership. For files, `resources.Path(req.Params.Meta)` only parses context; authenticate and authorize independently before any business-owned file read. Never open arbitrary host paths or opaque resource URIs directly.

Settings formats and exclusive numeric bounds are enforced before save. `FieldValidators` validate supplied patch fields; `FieldNames` map wire aliases to business names. `settings.NewModelServer[T]` derives fields and aliases from Go models; `forms.NewModelWithOptions[T]` accepts public schema inference options for form models. Storage and complete-state business validation remain caller-owned. See [usage and limits](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/EN-Forms).

The repository became public at the owner's explicit request on 2026-10-01. Public source access does not require private-repository authentication or `GOPRIVATE`; pin the reviewed `v0.0.1` module. This project changes no global Go settings. The local plugin's `0.0.0-dev` installation metadata is separate from the Go module release; this release does not distribute or reload that plugin.

The [GitHub Wiki](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/Home) provides separate English and Chinese guides for architecture, the complete capability matrix, frontend validation, SDK forms/MRTR investigation, local plugin acceptance and future implementation work. Its source remains under `docs/wiki`; Handoff remains in the source repository.

## Development

```sh
make fmt
make test vet build
# Optional exact-floor matrix, using Go's cache-local toolchain mechanism:
GOTOOLCHAIN=go1.25.0 make test vet build
```

`make test` runs race tests with official memory/HTTP/child-process stdio transports, settings/search/file behavior, and SDK boundary reproductions. The ignored executables are `bin/mcp-extensions-http` and `bin/mcp-extensions-stdio`. Frontend build/type/browser commands are documented separately. The 2026-10-02 extension changes passed race tests, vet and build with Go 1.25.0 and Go 1.27.0, including the checked-in Python fixtures. The 2026-10-01 frontend and cache-local `govulncheck v1.8.0` results are historical evidence; no frontend tests or vulnerability scan were repeated for this closeout. These checks do not establish remote deployment or completion of desktop UI acceptance.

- [Documentation (GitHub Wiki)](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/Home)
- [Architecture](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/EN-Architecture)
- [Compatibility and complete capability inventory](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/EN-Compatibility)
- [Forms, models and Settings validators](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/EN-Forms)
- [Forms/MRTR investigation](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/EN-Protocol-Investigation)
- [Frontend setup and host acceptance checklist](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/EN-Frontend-Validation)
- [Local plugin and human acceptance](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/EN-Live-E2E)
- [GitHub Wiki: English](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/EN-Overview) / [中文](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/ZH-Overview)
- [Implementation roadmap](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/EN-Implementation-Roadmap)
- [Validation evidence and SDK limits](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/EN-Validation-and-SDK-Limits)
- [Handoff (source repository)](docs/handoff.md)
- [English/Chinese Wiki source and publication procedure](docs/wiki/README.md)
