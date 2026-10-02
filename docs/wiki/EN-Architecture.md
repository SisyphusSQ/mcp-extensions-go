# Architecture

**English** · [中文](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/ZH-Architecture) · [Wiki home](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/Home)

Source: [`docs/architecture.md`](https://github.com/SisyphusSQ/mcp-extensions-go/blob/main/docs/architecture.md).

## Goal and responsibility layers

Provide validated Go server helpers for OpenAI MCP Extensions while the official MCP Go SDK v1.8.0 owns the protocol, schemas, transports and sessions. The public Go version floor remains 1.25.0. No SDK internals, unsafe/reflection tricks, SDK fork or second JSON-RPC stack are used.

```text
Business service / internal/example (HTTP and stdio)
  +-- settings: native settings tools + capability + declaration validation
  +-- mentions: app-visible search tool + two result variants
  +-- resources: file input/context parsing
  +-- ui: App metadata + trusted HTML registration
  +-- forms: extension schema, unchanged answers, resources and typed models
  +-- official mcp.Server
       +-- schemas, tools, resources, discovery, sessions and transports
       +-- standard elicitation and standard MRTR

examples/frontend
  +-- standard MCP Apps App
  +-- OpenAI TypeScript OpenAIExtensions
       +-- host context, file resources, model context, messages and deep links

Tests
  +-- official in-memory / authenticated HTTP / child-process stdio integration
  +-- test-only SDK-boundary probes and business continuation fixture
  +-- official AppBridge + real browser + Go HTTP server
```

The implementation order was settings, mentions, file context, SDK forms/MRTR investigation, then a real frontend and validation. The [capability inventory](https://github.com/SisyphusSQ/mcp-extensions-go/blob/main/docs/compatibility.md) separates server, frontend and host responsibilities for every upstream feature.

## Settings

`settings.NewServer` is a construction helper returning the official `*mcp.Server`. It validates and snapshots static field/layout definitions before constructing both read/update tools and their modern/legacy capability locations. It preserves other server options and capabilities and never calls persistence during registration. Consumers register other MCP tools as usual. Settings tool names are reserved by the caller; public SDK replacement semantics still apply.

The existing SDK dependency `jsonschema-go` is used directly, at the unchanged version v0.4.3. Primitive types and string/numeric constraints define both complete value schemas and a nonempty partial patch schema. Official typed `mcp.AddTool` handles input/output schema processing. Complete callback state is checked before a result succeeds; every field is required as an effective value, and unknown fields are rejected. Values are independently snapshotted. Numeric callbacks use Go numeric primitives because the validator regards `json.Number` as a string.

Read must be read-only. Update must independently authorize, preserve omitted fields, check cross-field constraints and persist atomically. The SDK helper supplies no database, implicit retries, locks or tenant model. `internal/example.Store` is single-owner example code, not a public storage API. HTTP defaults to memory; an absolute operator-owned `MCP_SETTINGS_FILE` enables a file backend. The stdio plugin defaults to a file under the user configuration directory. Reads are bounded to 64 KiB, files use mode 0600, and OS locking spans reload/patch/atomic replacement with a cancellable five-second acquisition limit. Invalid or unreadable state fails explicitly. This supports normal process restarts, not power-loss durability or identity-scoped multi-user transactions. The file backend supports Windows and the listed flock-capable Unix platforms; other platforms fail explicitly while memory storage remains usable.

## Mentions

The helper registers an ordinary tool with `openai/extensions.mentions/search` and app visibility. Existing ui/extension fields are preserved in an independent snapshot. Search remains caller-owned and accepts an empty query. Results use standard official `mcp.ResourceLink` encoding or the SDK-source `resource` variant. Invalid result variants become tool errors; empty results encode as `items: []`. There is no upstream standalone mention server capability to fabricate.

## File context

`FileInput` carries a filename and an opaque host resource URI. The App reads the resource using the official TypeScript resource API; Go does not map that URI to a filesystem path. Path, representation, and writable/etag metadata parsers allow future unrelated keys but reject malformed known fields.

`internal/example.Reader` is private example code, not a public extension API. It performs no authentication or per-resource authorization. The caller configures an absolute allowed root and positive byte limit. `os.Root` enforces containment during open, allowing relative in-root symlinks and rejecting escapes. Nonblocking open on Unix and regular-file checks reject FIFO/device/directory reads. Stat limits plus a bounded read handle files that grow after stat. Cancellation is checked around I/O. A trusted root must exclude hostile mounts/hard links; paths are interpreted within the caller's configured root namespace, including macOS `/var` aliases. The HTTP example reads only an operator-selected trusted App build; it never reads a host-provided path.

## UI and frontend

`ui` retains the existing metadata snapshot/validation and trusted static resource APIs. Tools bind to `ui://` resources. Visibility, entrypoints and display choices are declarations, never permission or host acceptance.

The browser example bundles the official App and OpenAI extensions into trusted HTML with a script hash CSP. Handlers are installed before connection, the initial tool result is rendered without repeating its launch tool, and unsupported host capabilities disable actions. User-triggered model context/message actions use the host; no Go host bridge was invented. File contents, errors and results enter the DOM as text. Go credentials remain server-side. Persisted settings initialize controls without replacing the initial result; Save sends changed fields only to preserve unrelated updates from other App instances.

The local plugin packages the generated official stdio executable and App. The repo marketplace exposes one local plugin. The supported `.codex-plugin/plugin.json` and `.mcp.json` compatibility format is used because the installed CLI recognized portable plugin metadata but did not load its MCP server. Official CLI installation materializes the cache copy and enables it. The stdio parent controls access; HTTP authentication remains mandatory. No network listener, daemon, credential or public plugin-directory entry is created. See [live-e2e.md](https://github.com/SisyphusSQ/mcp-extensions-go/blob/main/docs/live-e2e.md).

The test-only bridge fixture uses official AppBridge/PostMessageTransport, an isolated headless Chrome profile, a same-origin bounded local proxy and an ephemeral Go bearer credential. Model/context/file host behavior is fixture data. It is local integration evidence, not OpenAI product acceptance.

## Extension forms and SDK boundary

`forms` implements the Python extension's flat declarations, annotation fields, rich choices, hints, primitive/string-array constraints and resource/file picker rules. Validators preserve submitted JSON; host user selections follow declared user options, with an optional extra policy. Additional resource descriptors are snapshotted as opaque JSON. Public schema inference supplies aliases, enums/constraints for named types and typed results; defaults affect only typed decoding. No chooser, upload service or file access is implemented. See [forms.md](https://github.com/SisyphusSQ/mcp-extensions-go/blob/main/docs/forms.md) for exact limits and Python-generated fixtures.

Settings formats are enforced before persistence. Per-field validators run only for supplied patch fields and their transformed results are checked again. Update callbacks can receive business field names while the wire schema/effective values retain aliases. Typed model helpers preserve partial-map updates; merged-state rules and transactions remain application-owned.

Parity targets OpenAI MCP Extensions, not the MCP Python SDK. Standard MRTR, sessions, transports, tasks, generic request-state encryption/recovery and identity frameworks are excluded. No modern form MRTR adapter or review workflow is public. The SDK-boundary result adapter remains only an investigation fixture. Python's legacy `elicit_input` explicitly does not implement MRTR; direct custom sending is blocked by Go SDK v1.8.0's public outbound registry. No SDK internals or protocol implementation is copied. See [reproductions and source locations](https://github.com/SisyphusSQ/mcp-extensions-go/blob/main/docs/protocol-investigation.md).

## Security and lifecycle

The HTTP example retains runtime-only credentials, constant-time digest comparison, loopback defaults, cross-origin/localhost protection, a 1 MiB request limit, timeouts and graceful shutdown. Library handlers receive the official request/context so verified identity and authorization can be supplied by the owning server. Callback error messages must not contain credentials. Root filesystem errors may contain paths and must be mapped appropriately by a business tool.

The source repository became public at the owner's explicit request on 2026-10-01. Repository visibility does not change runtime authentication, filesystem authorization or the local plugin's distribution. No global tooling, persistent service, remote TLS deployment, tag or release is created. Go 1.25 matrix testing uses the cache-local `GOTOOLCHAIN` mechanism. Node installs are local to the frontend example with locked versions and lifecycle scripts disabled.
