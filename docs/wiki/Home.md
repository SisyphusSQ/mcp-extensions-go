# mcp-extensions-go

[中文首页](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/Home-ZH) · [Implementation roadmap](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/Implementation-Roadmap) · [Validation and SDK limits](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/Validation-and-SDK-Limits)

This is a private personal Go server SDK, not an official OpenAI SDK. It extends official `github.com/modelcontextprotocol/go-sdk v1.8.0`; the Go floor is 1.25.0. The official SDK owns JSON-RPC, transports, sessions, tools/resources and standard MRTR. Browser and host behavior uses standard MCP Apps and OpenAI's TypeScript App SDK.

Status snapshot: 2026-10-01. Server extensions were merged through [PR #1](https://github.com/SisyphusSQ/mcp-extensions-go/pull/1). The local plugin and live acceptance setup were merged through [PR #2](https://github.com/SisyphusSQ/mcp-extensions-go/pull/2), commit `4ce5a1f73d1755a72d5b669b7a17c16bfd8721b6`. No tag/release or public publication was created. A merge does not imply completion of host acceptance.

## Implemented

| Capability | Implementation / boundary |
| --- | --- |
| Native settings | Read/update tools, modern/legacy capability, primitive schema/constraints and layout; caller-owned authorization/storage |
| Mentions | Search types, both result variants, app visibility and metadata; caller-owned search |
| Files | File-entrypoint input, opaque resource references, path/representation/write-hint parsing, optional authorized root/byte-limited reader |
| UI | MCP Apps binding/visibility; global/thread/file/settings entrypoints, quickAction and display metadata; trusted HTML registration |
| Runnable examples | Authenticated stateless HTTP and official stdio; one shared example implementation |
| Local plugin | Installed/enabled MCP Extensions Go; durable single-owner settings, changed-field App patches and display-mode feedback |
| Browser App | Official handshake, settings/search, file text, deep-link context, supported model context/message/display actions |

Settings schema coverage is explicit, not full Pydantic parity. Extended forms and their complete OpenAI MRTR workflow are not implemented. Definitions/metadata alone must not be advertised as a complete form capability.

## Continue development

The [roadmap](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/Implementation-Roadmap) records six directly implementable gaps: additional settings constraints, form schema/rich choices, suggestions, resource pickers, answer/upload-reference validation, and Go model/schema binding. Each has scope, dependencies, acceptance and source references.

Modern extended-form MRTR is conditional on end-to-end SDK/host interoperability. Legacy custom outbound elicitation is blocked under the current public Go SDK API. [Validation and limits](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/Validation-and-SDK-Limits) keeps those separate from already implemented features awaiting host acceptance.

## Run and read

```sh
# With existing Go/Node tools and locked local frontend dependencies:
npm --prefix examples/frontend ci --ignore-scripts
make plugin
codex plugin marketplace add /absolute/path/to/mcp-extensions-go
codex plugin add mcp-extensions-go@mcp-extensions-go-local
```

Open **Open workspace** in the desktop sidebar's more menu. Reopen after installation updates; an existing page/process may retain an older document. The stdio plugin opens no listening port. Its single-owner settings live under the user configuration directory. It is not multi-user production storage.

Source documents: [README](https://github.com/SisyphusSQ/mcp-extensions-go/blob/main/README.md), [architecture](https://github.com/SisyphusSQ/mcp-extensions-go/blob/main/docs/architecture.md), [complete capability inventory](https://github.com/SisyphusSQ/mcp-extensions-go/blob/main/docs/compatibility.md), [SDK investigation](https://github.com/SisyphusSQ/mcp-extensions-go/blob/main/docs/protocol-investigation.md), [local plugin acceptance](https://github.com/SisyphusSQ/mcp-extensions-go/blob/main/docs/live-e2e.md), and [handoff](https://github.com/SisyphusSQ/mcp-extensions-go/blob/main/docs/handoff.md).
