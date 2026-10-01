# Compatibility and implementation scope

## Baseline

- Official Go SDK: pinned [v1.8.0](https://github.com/modelcontextprotocol/go-sdk/releases/tag/v1.8.0), requiring Go 1.25.0 or newer.
- Indirect `golang.org/x/sys` is pinned to v0.44.0 to address [GO-2026-5024](https://pkg.go.dev/vuln/GO-2026-5024) in upstream's v0.41.0; the Go version floor remains 1.25.0.
- OpenAI MCP Extensions: reference commit [`900032d8bd7c1566202d0cb1666986584f932043`](https://github.com/openai/mcp-extensions/tree/900032d8bd7c1566202d0cb1666986584f932043). This is a source baseline, not a promise to track floating main.
- UI field sources: [TypeScript definitions](https://github.com/openai/mcp-extensions/blob/900032d8bd7c1566202d0cb1666986584f932043/typescript/src/server/ui.ts) and [Python definitions](https://github.com/openai/mcp-extensions/blob/900032d8bd7c1566202d0cb1666986584f932043/python/src/openai_mcp_extensions/ui.py). No upstream implementation files were copied.

## Matrix

| Feature | Status | API / boundary |
| --- | --- | --- |
| Tool-to-resource binding | Implemented | `ToolMetadata.ResourceURI` maps to `_meta.ui.resourceUri`; requires `ui://`, a host, and no userinfo |
| Tool visibility | Implemented | `_meta.ui.visibility` accepts app/model; not authorization |
| OpenAI entrypoints | Metadata implemented | global, thread, file, settings; host affordances not validated |
| Global quickAction | Metadata implemented | title, nonempty icons, tool target, JSON arguments; caller registers the target tool |
| Model display mode | Metadata implemented | Tool `_meta["openai/ui"].preferredModelDisplayMode`: inline/fullscreen |
| Resource display modes | Metadata implemented | Content `_meta["openai/ui"]`: availableDisplayModes and preferredDisplayMode; upstream types include pip, example declares inline/fullscreen only; host support needs validation |
| Static HTML | Implemented | `AddHTMLResource`, MIME `text/html;profile=mcp-app`; no file reads or URL downloads |
| MCP Apps CSP / permissions | Passed through | Resource `_meta.ui`; no specialized types or generators |
| HTTP and Bearer authentication | Example implemented | Official stateless Streamable HTTP, 1 MiB body limit, timeouts, cross-origin protection, default localhost protection, graceful shutdown |
| Frontend App handshake | Not implemented | Static HTML has no `ui/initialize` or browser bridge; use frontend SDKs in future |
| settings read/update | Not implemented | A settings UI entrypoint does not implement `openai/settings` capabilities or handlers |
| mentions/search | Not implemented | No mentions types, search handlers, or metadata helpers |
| File picker / `openai/resource.path` | Not implemented | No parsing, path authorization, or file handlers |
| Extended forms and MRTR | Not implemented | SDK and host request/result adaptation need separate verification |
| Model context, model message, deep links | Not implemented | Browser-to-host interactions; Go does not replace the frontend SDK |

UI validation covers enums, entrypoint variants, required fields, and JSON encodability. It is not a complete port of every Zod/Pydantic validator. For example, it does not resolve or fetch icon URLs or validate the business meaning of arbitrary base metadata. Display-list consistency remains the caller's responsibility; no constraint absent from upstream was added.

## Important boundaries

Successful encoding does not establish host support for an entrypoint or display mode. Successful resource registration does not complete the host handshake. This framework does not advertise `io.modelcontextprotocol/ui` as a server capability or fabricate client/host capabilities. Unimplemented OpenAI features have no capability declarations.

The HTTP example supports stateless request-response calls. Server-initiated standard elicitation and nonstandard input requests are outside its scope. Adding forms requires revisiting the transport/host flow, timeouts, and lifecycle.

The Go version floor follows the dependency contract. Local validation used Go 1.27.0; no Go 1.25 toolchain was downloaded for matrix testing.
