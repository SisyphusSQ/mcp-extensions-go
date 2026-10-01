# Framework handoff

## Location and stage

- GitHub: private repository [SisyphusSQ/mcp-extensions-go](https://github.com/SisyphusSQ/mcp-extensions-go), branch `main`.
- Validated machine: sqmc04.
- Local checkout: `/Users/suqing/coding/golang/00_self/mcp-extensions-go`.
- Stage: SDK framework delivery; further extension work awaits the user's direction.

## Delivered

- Official MCP Go SDK v1.8.0, Go version floor 1.25.0.
- `ui/`: standard MCP Apps binding, OpenAI entrypoint/display metadata, trusted static HTML registration.
- `examples/http/`: authenticated HTTP server with embedded static HTML; no persistence or background installation.
- Makefile, locked dependencies, agent instructions, architecture and compatibility documentation.
- English comments and documentation, with a Chinese companion `README_ZH.md`.
- Development tests covering wire fields, invalid input, snapshots, official SDK resource reads, and authenticated HTTP calls.

Start with the root README and compatibility matrix. Source entrypoints are `ui/metadata.go`, `ui/resource.go`, and `examples/http/main.go`. Add real business handlers using official `mcp.AddTool`.

## Validation

On sqmc04, 2026-10-01, with Go 1.27.0 / darwin arm64:

- `go test -race ./...`: passed for `ui` and `examples/http` after the final code changes.
- `go vet ./...`: passed.
- `go build -o bin/mcp-extensions-http ./examples/http`: passed.
- HTTP integration: official client initialization, tools/list, tools/call, and resources/read passed. Missing/incorrect authentication, cross-origin POSTs, oversized bodies, and cancellation shutdown were exercised.
- Dependency review: OSV version queries found GO-2026-5024 in the SDK's original indirect `x/sys v0.41.0`; it was updated to v0.44.0. A subsequent version query reported no entries for v0.44.0. This was a module-version lookup, not a full reachability scan.
- Security review: runtime-only credentials, constant-time token comparison, authenticated MCP access, request limits, timeouts, cross-origin/localhost protection, no user-controlled file/URL reads, and metadata snapshot ownership were reviewed. No unresolved issue was identified within this scaffold's scope.

These are development results, not Codex/ChatGPT host or production acceptance. Tests were not repeated during commit/push closeout. Go 1.25 matrix validation, host handshake, and remote TLS deployment were not performed.

## Boundaries

Full forms, MRTR, mentions, settings handlers, file context, and the browser App handshake are not implemented. No release or additional task was created. This is a personal private framework, not a claim of complete equivalence with the official extensions.
