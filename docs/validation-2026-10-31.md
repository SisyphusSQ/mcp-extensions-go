# Upstream 0.2.1 / WebMCP development validation (2026-10-31)

Machine: sqmc04 (darwin/arm64). Branch: `suqing/mcp-extensions-v0.0.3`.
Module target: v0.0.3. References, dependency decisions and exact source hashes:
[migration](upstream-0.2.1.md). This record separates local validation from
OpenAI host acceptance.

## Passed during development

| Command / scenario | Observed result |
| --- | --- |
| `make test vet build`, Go 1.27.0 | Full race tests, vet and all example builds passed |
| `GOTOOLCHAIN=go1.25.0 make test vet build` | Same checks passed at the exact Go floor |
| Frontend `npm run typecheck` and `npm run build` | Passed with the registry package and patched dependency graph |
| Frontend `npm test` | Passed using existing Playwright 1.62.1 and installed Google Chrome, isolated headless profile |
| Frontend `npm run test:forms` | 19 passed, zero failed with official TypeScript client 2.4.0 and real Go stdio/MRTR |
| `bin/mcp-extensions-acceptance` | Six PASS lines; 53 Python declarations, 86 values, four upload cases and 192 actual MCP calls |
| Final `npm audit` | Zero reported vulnerabilities (initial nine dependency entries resolved) |
| Python reference generation | 53 schemas / 90 values and upload cases from the fixed official 0.2.1 source; Go parity passed |

Local ignored logs: `bin/v0.0.3-go1.27-final.log`,
`bin/v0.0.3-go1.25.log`, `bin/v0.0.3-browser-final.log`,
`bin/v0.0.3-forms.log` and `bin/v0.0.3-acceptance.log`.
The first development runs exposed a stale fixture provenance guard and a
test-fixture route missing the WebMCP query. Both were corrected before the
passing runs. No global tools, browsers or Python dependencies were installed.

## New behavior exercised

The real browser/official AppBridge and authenticated Go service cover existing
settings/readable search/file/display behavior, plus named/default context
groups, external item/group/all removal, own-update echoes, unsent selections,
initial restored/cleared state, removal during an outstanding update, and two
independent App instances. Mobile fixture tests inspect all four target/send
combinations. This does not prove mobile product UI.

WebMCP tests inject an isolated test-only `document.modelContext` registry into
the actual App document. They inspect registrations/read-only hints, execute
search through the actual Go tool, change selections through shared UI logic,
read live state after host/UI changes, and reject unknown fields/types/parts.
Missing API and Web/mobile platforms leave the ordinary App usable. This does
not prove OpenAI agent discovery, expanded-view exposure or permissions.

Numeric tests cover standalone integer representation/range, ordinary numbers,
numeric strings rejected as numbers, fractions rejected as integers, exact
large integer/decimal strings preserved, and JavaScript normalization of 1.0
and exponent integers before official TypeScript/Go MRTR. Existing validation
semantics did not need a production change. Both deprecated form aliases are
now explicitly included in regenerated Python fixtures.

## Real host and release boundaries

The UI tool explicitly refused native `com.openai.codex` access; current MCP
Apps inventory contained no page. No live desktop discovery/call, context
removal or new mobile/Work browser acceptance is claimed. Do not use old
screenshots, protocol fixtures or a source release to mark these complete.
Follow [the host checklist](frontend-validation.md) in a supported environment.
The installed plugin/cache was not rebuilt or reloaded by this release.

Source/local-test alignment is 0.2.1 plus the separately pinned WebMCP example.
It is not full OpenAI host/platform certification. Legacy custom outbound
sending remains SDK-blocked; state, authorization, transactions and exact-text
parsing remain caller-owned. Go dependencies and the Go floor are unchanged;
no fresh Go reachability scan is claimed.

Adversarial/security review preserved HTTP authentication, resource ownership,
text DOM output and server-only credentials. New browser tools expose fixed
operations with strict inputs, share the existing busy gate and never accept
arbitrary server tool names, URLs or paths. No new file/network permission or
secret entrypoint was introduced.

Commit/push/release closeout reuses these development results without rerunning
tests, lint or verification commands. Git/build/package/signature/artifact and
remote publication readback remain separate closeout operations.
