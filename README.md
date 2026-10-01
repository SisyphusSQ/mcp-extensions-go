# mcp-extensions-go

[中文说明](README_ZH.md)

A private, personal Go library for the server-side features of [OpenAI MCP Extensions](https://github.com/openai/mcp-extensions), built on the official [MCP Go SDK v1.8.0](https://github.com/modelcontextprotocol/go-sdk). This is not an official OpenAI Go SDK. No version or tag has been released.

The official SDK owns JSON-RPC, schemas, tools, resources, discovery, sessions, transports and standard MRTR. This module adds server extension types and helpers. Browser/host behavior uses the official TypeScript App SDKs.

## Implemented

- `settings.NewServer`: real read/update tools, native primitive schemas and constraints, layout, full effective values, nonempty partial patches, and modern/legacy `openai/settings` discovery. The returned server is the ordinary official `*mcp.Server`.
- `mentions.AddTool`: searchable resource links and the upstream SDK's resource variant, `mentions/search` tool metadata, read-only hints and required app visibility.
- `resources`: typed file-entrypoint input, opaque resource references, `openai/resource` path/representation/write-hint parsing, and an optional byte-limited `os.Root` reader for already-authorized local files.
- `ui`: standard MCP Apps resource binding/visibility, OpenAI entrypoints, quick actions/display metadata, and trusted HTML resource registration.
- `examples/http`: authenticated stateless Streamable HTTP with real in-memory settings, searchable/readable demo parts, workspace/file tools, loopback defaults, request limits and graceful shutdown.
- `examples/frontend`: a bundled browser App using standard MCP Apps `App` plus OpenAI TypeScript extensions, with local browser integration tests through the official `AppBridge`.

Settings storage, resource authorization, cross-field validation and transactions are caller-owned. Visibility and metadata never grant permission. The example's one-credential memory store resets on restart.

Full OpenAI extended forms are not implemented. v1.8.0 supports standard MRTR, but its typed input map and custom outbound request API have specific limits. A receiving-middleware result adapter can emit extension fields, but that alone does not establish extended MRTR support. See the [complete responsibility matrix](docs/compatibility.md) and [reproducible SDK investigation](docs/protocol-investigation.md). OpenAI host acceptance remains unverified.

## Run the Go example

Requires Go 1.25.0 or newer. No Node dependency is needed for the default static page.

```sh
export MCP_BEARER_TOKEN="$(openssl rand -hex 32)"
go run ./examples/http
```

The endpoint defaults to `http://127.0.0.1:8080/mcp`. Configure a host with Streamable HTTP and `Authorization: Bearer <runtime credential>`. The token must contain at least 32 bytes without whitespace; never print or commit it. Ctrl+C stops the process. Opening `/mcp` in a browser does not render HTML; MCP `resources/read` supplies App resources.

For the real browser App, build `examples/frontend` and set `MCP_APP_HTML` to the absolute path of its trusted `dist/app.html`; follow [frontend setup and validation](docs/frontend-validation.md). Without that setting, the embedded page is static and performs no `ui/initialize` handshake.

`MCP_LISTEN_ADDR` selects an explicit address. Remote access requires caller-configured HTTPS termination, authentication and network controls. The example does not configure TLS or install a background service.

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

For UI declarations, use `ui.ToolMetadata.Metadata` and `ui.AddHTMLResource` with official tools/resources. Metadata snapshots preserve arbitrary JSON number precision and ownership. For files, `resources.Path(req.Params.Meta)` only parses context; authenticate and authorize independently before calling a root-contained `Reader`. Never open arbitrary host paths or opaque resource URIs directly.

Other consumers of this private module need repository access, Git authentication and their own `GOPRIVATE` configuration. This project changes no global Go settings.

## Development

```sh
make fmt
make test vet build
# Optional exact-floor matrix, using Go's cache-local toolchain mechanism:
GOTOOLCHAIN=go1.25.0 make test vet build
```

`make test` runs race tests with official memory/HTTP transports, settings/search/file behavior, and SDK boundary reproductions. The ignored executable is `bin/mcp-extensions-http`. Frontend build/type/browser commands are documented separately. Tests passed on sqmc04 with Go 1.25.0 and Go 1.27.0; the frontend dependency installation reported zero npm audit vulnerabilities. This is not a full Go vulnerability reachability scan, remote deployment, or OpenAI host acceptance.

- [Documentation index](docs/README.md)
- [Architecture](docs/architecture.md)
- [Compatibility and complete capability inventory](docs/compatibility.md)
- [Forms/MRTR investigation](docs/protocol-investigation.md)
- [Frontend setup and host acceptance checklist](docs/frontend-validation.md)
- [Handoff](docs/handoff.md)
