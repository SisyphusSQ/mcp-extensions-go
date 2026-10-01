# Server extensions handoff

## Delivery location

- Private GitHub repository: [SisyphusSQ/mcp-extensions-go](https://github.com/SisyphusSQ/mcp-extensions-go).
- Development branch: `suqing/server-extensions`, based on `f2531afd126b2fd8cb1c79cfb5576b3f8efbfbf5` from `main`.
- Machine: sqmc04, local checkout `/Users/suqing/coding/golang/00_self/mcp-extensions-go`.
- Upstream main/baseline: `900032d8bd7c1566202d0cb1666986584f932043`, independently checked against GitHub and a source checkout on 2026-10-01.
- Official MCP Go SDK remains v1.8.0; Go floor remains 1.25.0; `x/sys` remains v0.44.0.

## Implemented

1. Native settings: primitive field/constraint/schema/layout types; paired read/update registration and capability discovery; effective values, partial patches and real tool errors. Callers own storage, authorization, transactions and cross-field rules. The example uses a single-credential mutex-protected memory store.
2. Mentions: empty-query search input, standard resource links and SDK resource variants, app visibility and search marker, callback registration, result validation and empty-result/error semantics. Business search is caller-owned.
3. Files: file-entrypoint input, opaque resource references, path/read/content metadata parsing, an optional explicit root/byte-limited regular-file reader with os.Root containment, and a read-only file viewer App. No arbitrary host path is opened by the example.
4. SDK investigation: standard multi-round elicitation was verified. Private reproductions show the custom sending and typed MRTR map limitations, and show that receiving result adaptation can emit an extended result over official HTTP. No complete OpenAI form API or unsupported capability is claimed.
5. Real frontend source: standard App and OpenAI TypeScript extensions, initialization, first-result rendering, settings/search UI, supported display/context/message actions, deep-link updates, and host file text rendering. A local AppBridge fixture tests real browser execution against the authenticated Go server.

All project documentation/comments remain English with README_ZH.md maintained as the Chinese companion. The [complete capability inventory](compatibility.md) is the implementation-status authority. Entry points: `settings/settings.go`, `mentions/mentions.go`, `resources/resources.go`, `resources/reader.go`, `examples/http/main.go`, and `examples/frontend/app.ts`.

## Development evidence

After the final code changes on sqmc04, 2026-10-01:

| Check | Result / scope |
| --- | --- |
| `GOTOOLCHAIN=go1.25.0 make test vet build` | Passed: race tests, vet and HTTP executable build on darwin/arm64 |
| `make test vet build` with Go 1.27.0 | Passed: race tests, vet and build on darwin/arm64 |
| Go 1.25 Linux amd64 and Windows amd64 `go build ./...` | Passed cross-compilation; runtime tests on those OSes were not performed |
| `npm run build` | Passed; ignored self-contained `examples/frontend/dist/app.html` with a script hash CSP |
| `npm run typecheck` | Passed for App and local test host |
| `npm test` with installed bundled Playwright 1.62.1 and installed Chrome | Passed: official App/AppBridge handshake, first result, Go HTTP settings/search, display/deep-link changes, model-context/message payloads, inert file text, missing capabilities and standalone mode |
| `npm install --ignore-scripts` | Locked local dependencies installed; npm audit reported 0 vulnerabilities |

Go 1.25.0 was downloaded by Go's `GOTOOLCHAIN` cache mechanism, not installed globally. The existing direct dependency on `jsonschema-go v0.4.3` was promoted from indirect for settings validation; no Go module version changed. `govulncheck` was not installed/available, so no new full Go vulnerability reachability scan was performed. No global tools or browser downloads were installed automatically.

Settings tests exercise modern server/discover and legacy initialize capability locations, declaration snapshots, read-only/schema output, valid partial persistence, invalid patches blocked before storage, primitive constraints, missing returned values and persistence failure. HTTP integration exercises actual authenticated discovery/tool/resource calls, stored settings across stateless requests, and readable mention results, while retaining existing authentication, origin, request-size and shutdown checks.

File tests cover malformed metadata, typed entrypoint input, allowed reads, lexical and symlink escapes, byte limits, directories, cancellation, nonblocking FIFO rejection, and concurrent symlink replacement. MRTR tests cover two real rounds, cancellation/decline, manual resume, repeated/stale inputs, expiration, completion replay and eight concurrent calls with connection isolation. The continuation registry/cache/deadline in these tests is a business fixture, not an SDK guarantee or production store.

Security self-review of this diff found no unresolved issue in the implemented scope: authentication was preserved, file reads require explicit roots and size limits, metadata never grants authorization, browser output uses text DOM APIs, credentials stay out of HTML/source/logs, and concurrency ownership is explicit. This review does not replace the unperformed reachability scan or production/host acceptance.

## Remaining limitations and continuation

- **OpenAI host acceptance unverified:** no MCP Apps page connected to this server was available. Browser inventory reported a request-header policy error. Local AppBridge integration is not Codex/ChatGPT product acceptance. Follow [frontend-validation.md](frontend-validation.md) after a supported authenticated host is connected.
- **Extended forms incomplete:** legacy custom server-to-client requests have no public typed sending API. Typed MRTR input maps reject the OpenAI method. A modern receiving-result adaptation route is technically possible, but a validated extended schema/selection API, bounded identity-scoped continuation design and real host interoperability still need implementation. See exact source links, reproductions and requested SDK interfaces in [protocol-investigation.md](protocol-investigation.md).
- **Persistence and authorization remain caller-owned:** the example resets settings on restart and has one credential/record. It does not establish durable multi-user isolation or transactions for a real application.
- **Files:** parsing is not permission. Reader roots must be trusted; hard links/mounts inside a hostile root and interruptible file syscalls are outside the reader contract. Host file writes/subscriptions and native file opening were not accepted in a live host.
- **Deployment:** no remote HTTPS deployment, persistent service, release, tag or repository visibility change was performed.

Commit/push closeout reuses the development evidence above and does not repeat tests, lint or verification commands. Git state, diff review and remote push/readback remain closeout operations. No additional Issue, PR, automation or memory/TODO record was created.
