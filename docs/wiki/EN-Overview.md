# Overview

**English** · [中文](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/ZH-Overview) · [Wiki home](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/Home)

This is a public personal Go server SDK, not an official OpenAI SDK. It extends official `github.com/modelcontextprotocol/go-sdk v1.8.0`; the Go floor is 1.25.0. The official SDK owns JSON-RPC, transports, sessions, tools/resources and standard MRTR. Browser and host behavior uses standard MCP Apps and OpenAI's TypeScript App SDK.

Historical merge/plugin record: 2026-10-01; current extension documentation: 2026-10-02. Server extensions were merged through [PR #1](https://github.com/SisyphusSQ/mcp-extensions-go/pull/1). The local plugin and live acceptance setup were merged through [PR #2](https://github.com/SisyphusSQ/mcp-extensions-go/pull/2), commit `4ce5a1f73d1755a72d5b669b7a17c16bfd8721b6`. Bilingual Wiki source was merged through [PR #3](https://github.com/SisyphusSQ/mcp-extensions-go/pull/3). The owner then explicitly authorized making the source repository public. No tag/release or public plugin-directory distribution was created at that historical stage. The first Go module release is [v0.0.1](https://github.com/SisyphusSQ/mcp-extensions-go/releases/tag/v0.0.1); the local plugin remains separate. A merge does not imply completion of host acceptance.

## Agent integration

Install `github.com/SisyphusSQ/mcp-extensions-go@v0.0.1` and start with the [coding-agent guide](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/EN-Agent-Integration). It includes a public-only quickstart and the reusable extension acceptance runner.

## Implemented

| Capability | Implementation / boundary |
| --- | --- |
| Native settings | Read/update tools, modern/legacy capability, primitive schema/constraints and layout; caller-owned authorization/storage |
| Mentions | Search types, both result variants, app visibility and metadata; caller-owned search |
| Files | File-entrypoint input, opaque resource references, path/representation/write-hint parsing |
| Forms | Flat schemas, annotations/JSON enums, rich choices/suggestions/resources, unchanged-answer validation and typed models |
| UI | MCP Apps binding/visibility; global/thread/file/settings entrypoints, quickAction and display metadata; trusted HTML registration |
| Runnable examples | Authenticated stateless HTTP and official stdio; one shared example implementation |
| Local plugin | Installed/enabled MCP Extensions Go; durable single-owner settings, changed-field App patches and display-mode feedback |
| Browser App | Official handshake, settings/search, file text, deep-link context, supported model context/message/display actions |

Settings and forms follow the Python extension contract with documented language/runtime limits. Form declarations, validation and typed models are implemented; direct custom sending remains SDK-blocked. Standard MCP SDK capabilities and generic continuation recovery are excluded. See [Forms](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/EN-Forms).

## Continue development

The [alignment status](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/EN-Implementation-Roadmap) records implemented Settings constraints/validators, form schemas/rich choices, suggestions, resource selections, answer/upload-reference validation and typed Go models. Remaining language/runtime limits and real host acceptance are listed separately.

Modern extended-form MRTR and generic continuation recovery are outside this Python extension alignment scope. Legacy custom outbound elicitation is blocked under the current public Go SDK API. [Validation and limits](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/EN-Validation-and-SDK-Limits) keeps those separate from already implemented features awaiting host acceptance.

## Run and read

```sh
# With existing Go/Node tools and locked local frontend dependencies:
npm --prefix examples/frontend ci --ignore-scripts
make plugin
codex plugin marketplace add /absolute/path/to/mcp-extensions-go
codex plugin add mcp-extensions-go@mcp-extensions-go-local
```

Open **Open workspace** in the desktop sidebar's more menu. Reopen after installation updates; an existing page/process may retain an older document. The stdio plugin opens no listening port. Its single-owner settings live under the user configuration directory. It is not multi-user production storage.

Guides: [architecture](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/EN-Architecture), [complete compatibility matrix](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/EN-Compatibility), [frontend validation](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/EN-Frontend-Validation), [SDK investigation](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/EN-Protocol-Investigation), and [local acceptance](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/EN-Live-E2E). The source [README](https://github.com/SisyphusSQ/mcp-extensions-go/blob/main/README.md) remains the module usage entrypoint.

Settings also provides formats/exclusive bounds, pre-save field validators, business aliases and typed models; partial patch and caller transaction semantics are preserved.
