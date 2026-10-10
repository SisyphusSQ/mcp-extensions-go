# Server extensions handoff

## Upstream 0.2.0 and module v0.0.2 (2026-10-10)

The owner authorized implementation, a branch/PR into main and the v0.0.2 release.
Branch: `suqing/mcp-extensions-v0.2.0`. The [migration guide](upstream-0.2.0.md)
and [current validation record](validation-2026-10-10.md) supersede historical
claims below excluding a public modern form adapter.

Implemented: public `Form.RequestInput` through official modern MRTR, keyed answer
validation, two-round/concurrent flows, modern/legacy Mention capability options,
optional deprecated Settings tool titles, rejection of removed Settings App
entrypoints, and capability/platform-gated editable frontend drafts. Go SDK v1.8.0
and Go floor 1.25.0 remain pinned. State/auth/transactions remain caller-owned;
legacy custom sending and real host UX remain separate gaps.

Development passed full race/vet/build on Go 1.27.0 and exact 1.25.0, frontend
build/typecheck/browser integration, nine TypeScript 2.2.0 form scenarios, and npm
audit with zero reported vulnerabilities. The frontend uses the original official
Node 0.2.0 tagged workflow archive because npm did not expose that release;
provenance/checksum and license are checked in. No SDK was edited or forked.

Authorized closeout is PR merge, annotated v0.0.2 module tag, GitHub Release
with source archive/checksum, and separate bilingual Wiki synchronization.
This source record does not itself prove those remote steps completed; read the
PR/release/tag and Wiki history. No tests are repeated during closeout. Existing
plugin caches are not reloaded or deployed by this release. No global tooling,
credentials, monitoring baseline or notification automation is changed.

## Agent integration and v0.0.1 release preparation (2026-10-02)

The authorized branch is suqing/agent-guide-v0.0.1 in the existing checkout. Consumer guidance is docs/agent-integration.md, with English/Chinese Wiki pages and README/index navigation. examples/agent-quickstart imports public APIs only; examples/acceptance archives the previously passed 186-call subprocess probe. Source/generated binaries, logs, Node dependencies and private plugin cache remain distinct.

The module release is v0.0.1, separate from the local plugin's 0.0.0-dev installation metadata. The release sequence is reviewed PR, merge into main, annotated module tag, GitHub Release and separate Wiki publication. No tests/lint/acceptance are repeated at this closeout. Prior Go 1.25/1.27 test/vet/build and 186-call results remain evidence; new example/snippet compilation is a build result, not runtime acceptance. See [the consumer guide](agent-integration.md) and [recorded validation](validation-2026-10-02.md). No global tools, dependencies, credentials or runtime resource permissions change.

## Previous extension implementation record

## 2026-10-02 extension-only alignment

Development branch: `suqing/forms-settings-validation`. On 2026-10-02 the owner authorized commit, PR creation/merge into main and publication of the synchronized Wiki pages. This section supersedes the earlier backlog/status paragraphs below. Only OpenAI MCP Extensions Python 0.1.0 at `900032d8bd7c1566202d0cb1666986584f932043` is the alignment target; the underlying MCP Python SDK is excluded.

Retained and expanded: flat form declarations/validation, generic JSON enums, annotation fields, rich choices/images, suggestions/free arrays, resource descriptor extras, declared host user selections/upload references, typed defaults and named-type schema inference. Settings now has format/exclusive-bound validation, supplied-field validators, alias mapping and typed model helpers. UI declaration whitespace and null resource-path context match Python. See [forms usage](forms.md) and [the authority matrix](compatibility.md).

Removed from the uncommitted work: generic `requeststate` encryption/recovery, public modern MRTR adaptation, principal scaffolding, and the two-round `review_parts` tool. Moved the previously public file reader and its tests into `internal/example`; trusted file containment/limits and the existing single-owner settings store remain example business implementation. Source examples expose the original five tools and three resources again. No SDK fork, transport/session/task implementation, new dependency, external service or credential was added.

Python-generated reference fixtures cover 51 schemas and 86 value/upload cases. Focused Go tests pass after these changes. Go 1.27.0 and exact Go 1.25.0 both passed `make test vet build` after the final code changes, including race tests, authenticated HTTP and child-process stdio integration. Local logs are `bin/extensions-alignment-go1.27.log` and `bin/extensions-alignment-go1.25.log`. Legacy custom sending remains blocked by the pinned SDK; real OpenAI host forms/chooser/upload acceptance is unverified. Adversarial/security review preserved HTTP authentication, bounded trusted file reads and resource authorization ownership; no new network/file read or credential entrypoint was added. Numeric enum precision, case-sensitive resource extras and effective typed zero values have focused regression coverage. Wiki publication uses the separate Wiki repository after the main PR merges. The installed plugin cache is not rebuilt/reloaded by this closeout. Existing development results are reused without rerunning tests, lint or verification commands.

## Historical delivery record (through 2026-10-01)

## Previous continuation: public repository and published bilingual Wiki

PR [#1](https://github.com/SisyphusSQ/mcp-extensions-go/pull/1) merged into main at `3f45cf66eab5f6fd172763e604d794e12cef5cf1` while the repository was private. PR [#2](https://github.com/SisyphusSQ/mcp-extensions-go/pull/2) merged `suqing/live-e2e` at `4ce5a1f73d1755a72d5b669b7a17c16bfd8721b6`. PR [#3](https://github.com/SisyphusSQ/mcp-extensions-go/pull/3) merged the bilingual Wiki source at `fe96dc077b2b13f6a236d9e3bc563aec7b38f372`. The original local-plugin stop below was superseded by the user's explicit PR/merge and bilingual GitHub Wiki request. Host acceptance status remains defined by [live-e2e.md](live-e2e.md).

On 2026-10-01 the owner explicitly requested public repository visibility, overriding the earlier private-only requirement. GitHub readback confirmed `private: false`, `visibility: public` and `has_wiki: true`. Before the visibility change, 96 unique historical Git blobs were reviewed for common credential literals and sensitive configuration filenames; no candidates were found. This publication did not change runtime authentication, resource authorization or local plugin distribution.

Wiki source is under `docs/wiki`. D1–D6 are directly implementable gaps; S1 is conditional modern extended MRTR; S2 is blocked legacy custom sending. The owner initialized Home through GitHub, then the bilingual pages were published on top of the initial history at `54c55ad`. At the owner's request, the layout was aligned with [orchestrator Wiki](https://github.com/SisyphusSQ/orchestrator/wiki), reference commit `1ce473c59ddab4558965b80a5762773c8935e25d`: Home selects a language, three English pages use `EN-`, three Chinese pages use `ZH-`, sidebar groups are separate, and a shared footer links source/issues. This layout was published at `b50998e` on `master`. See [Wiki Home](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/Home), [中文概览](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/ZH-Overview), [English overview](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/EN-Overview) and the [future publication procedure](wiki/README.md). No feature implementation is authorized by merely listing its backlog.

The owner subsequently requested complete Docs content in the Wiki, excluding Handoff. Architecture, the full compatibility matrix, frontend validation, protocol investigation and local plugin/live acceptance were published in both languages at Wiki commit `17c3278`. The public Home displays 17 content pages: a language selector and eight pages per language, plus shared sidebar/footer. This handoff remains only in the source repository. The copied instructions, SDK blockers and pending host outcomes retain their original evidence boundaries.

The official stdio example and local plugin are implemented and installed/enabled on sqmc04. Shared example business handlers retain HTTP authentication. Single-owner file settings use bounded reads, OS locks and atomic replacement; real simultaneous stdio processes preserve unrelated patches and a restarted process reads persisted values. The App initializes saved controls, sends changed fields only, and displays actual host mode with appropriate fullscreen feedback.

User screenshots confirm global Open workspace navigation, Connected and the initial result, effective settings values, and a model answer using bolt context after an explicit message. The user reported an ineffective Open fullscreen button; the display feedback fix is installed, with reopened-page human acceptance pending. The actual settings file was read locally: mode 0600, units=in, showGrid=false. This confirms persistence occurred; it is not a human restart acceptance result.

Development evidence: Go 1.27 full race/vet/build passed; Go 1.25 full suite passed before the final small changes, then final HTTP/stdio/store race tests and vet/build passed. Linux/Windows amd64 Go 1.25 cross-builds passed. Final frontend build/typecheck/browser suite passed, including changed-field concurrency and fullscreen/already-fullscreen/unsupported/retained-mode scenarios. Cache-local `govulncheck v1.8.0` reported no vulnerabilities with Go 1.27; no global tool was installed. `x/sys v0.44.0` is now direct for the example's Windows OS locking, with no version change.

Native desktop UI control and its default CLI control socket were unavailable. The agent did not restart the desktop or send a user conversation message; host evidence came from user-provided screenshots. Native mentions/settings/file routing, restart/isolation, deep links, additional context/message cases and extended forms remain unaccepted. Full OpenAI forms/MRTR still have the documented SDK blockers; no capability is fabricated.

Security self-review preserved authentication and resource boundaries, operator-controlled bounded file access, private local settings and explicit concurrency ownership; no unresolved issue was found in this local single-owner scope. Multi-user and Windows/Linux runtime acceptance remain outside the evidence.

Human acceptance remains pending after merge. No tests/checks are repeated during PR/merge or documentation closeout; this documentation-only continuation reuses the recorded development evidence. The source repository and Wiki are public by explicit authorization; no tag, release, public plugin-directory listing, automation or memory/TODO update is created. Security review of the Wiki additions found no credentials or expanded runtime resource permissions; all new capability entries remain explicitly unimplemented.

## Initial server extension delivery (historical)

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

All project documentation/comments remain English with README_ZH.md maintained as the Chinese companion. The [complete capability inventory](compatibility.md) is the implementation-status authority. Entry points: `settings/settings.go`, `mentions/mentions.go`, `resources/resources.go`, `internal/example/reader.go`, `examples/http/main.go`, and `examples/frontend/app.ts`.

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
- **Files:** parsing is not permission. Private example reader roots must be trusted; hard links/mounts inside a hostile root and interruptible file syscalls are outside the reader contract. Host file writes/subscriptions and native file opening were not accepted in a live host.
- **Deployment:** no remote HTTPS deployment, persistent service, release, tag or repository visibility change was performed.

Commit/push closeout reuses the development evidence above and does not repeat tests, lint or verification commands. Git state, diff review and remote push/readback remain closeout operations. No additional Issue, PR, automation or memory/TODO record was created.
