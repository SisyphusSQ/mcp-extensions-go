# Upstream 0.2.0 development validation (2026-10-10)

## Scope

Development branch: `suqing/mcp-extensions-v0.2.0`. Module release target:
`v0.0.2`. Alignment uses upstream Python `python-v0.2.0` at
`ebeff485e6d1bf52b6ebf3b4137555e53851d078` and Node `node-v0.2.0` at
`d3a79b765e3ad95eaa03f23b0d3d2f91987b8603`. Official Go SDK v1.8.0 and
the Go 1.25.0 floor are unchanged. See [migration](upstream-0.2.0.md).

## Completed during development

| Command / scenario | Observed result |
| --- | --- |
| `make test vet build`, Go 1.27.0 | Passed full race tests, vet and all example builds |
| `GOTOOLCHAIN=go1.25.0 make test vet build` | Passed full race tests, vet and all example builds at the exact floor |
| `npm run typecheck` and `npm run build` in `examples/frontend` | Passed; generated ignored `dist/app.html` |
| `npm run test:forms` | Nine passed, zero failed using official TypeScript client 2.2.0 and the Go stdio example |
| `npm test` with existing bundled Playwright and installed Chrome | Passed official App/AppBridge handshake and authenticated Go HTTP/browser scenarios |
| `npm audit` on final frontend lockfile | Reported zero vulnerabilities |

New Go tests exercise full schema metadata while the core schema remains an empty
object, independent metadata snapshots, unchanged accepted answers, cancellation,
decline, invalid fields/resources, both capability gates, legacy refusal, and
concurrent two-round flows through the official client. Mention tests inspect
both capability locations and Settings composition; UI/Settings tests cover
removed entrypoints and optional titles. Existing fixture and authenticated HTTP,
stdio, file and persistence tests remain part of the full suite.

The TypeScript client drives actual SDK MRTR retries and inspects extended pattern,
suggestions and resource-array metadata. Accepted/cancelled/declined answers,
invalid pattern/resource/missing fields, missing either capability and legacy
protocol are exercised. No handwritten JSON-RPC or SDK fork is used.

The browser suite covers the actual built App, real Go settings/search, initial
result, changed-field preservation, display/deep-link updates, context/message
payloads, active/new `send:false`, mobile and missing-capability gating, inert
file text and standalone mode. The local host extends its message schema with
the official OpenAI schema through AppBridge's protected extension hook; the
core schema alone strips OpenAI metadata. This is test-host behavior, not a
claim about an OpenAI host implementation.

## Dependencies and limits

The final locked frontend clients are SDK 1.31.0 and client/core 2.2.0, which
include the upstream fix for GHSA-6qxp-vccf-f47h. These fixtures use no OAuth
provider or OAuth credentials. Go dependencies were unchanged; no fresh
`govulncheck` result is claimed. No global package or browser was installed.

The npm registry returned 404 for the official extension's 0.2.0 version.
The original official tagged workflow package is checked in unmodified with
its license, SHA-256 and [provenance](../examples/frontend/vendor/README.md).

This evidence does not establish native host form rendering, chooser/uploads,
preview, draft editing, storage isolation or production deployment. Legacy custom
outbound form sending remains SDK-blocked. RequestState authentication, identity,
authorization, replay handling and mutation transactions belong to the caller.
The installed local plugin was not rebuilt, reloaded or distributed.

Adversarial/security review found no new runtime authentication bypass, arbitrary
file read, server-side URL fetch or secret entrypoint. Resource references grant
no read permission. The public form example is read-only; client-echoed state is
explicitly untrusted. Existing HTTP authentication and resource boundaries remain.

After entering commit/push/release closeout, reuse this development evidence;
do not repeat tests, lint or verification commands.
