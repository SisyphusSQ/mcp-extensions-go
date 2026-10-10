# Validation and SDK limits

**English** · [中文](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/ZH-Validation-and-SDK-Limits) · [Overview](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/EN-Overview) · [Wiki home](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/Home)

## Current v0.0.2 / upstream 0.2.0 status

Modern Form.RequestInput, Mention capability discovery, optional legacy Settings button titles, rejection of removed Settings App entrypoints and desktop/web drafts are implemented. The 2026-10-01/02 records below are historical; their exclusion of modern MRTR is superseded by this authorized work. Generic recovery and legacy custom sending remain outside support. Real host forms/drafts are unverified. See [current development evidence](https://github.com/SisyphusSQ/mcp-extensions-go/blob/main/docs/validation-2026-10-10.md).

This page separates implementation gaps from host acceptance and business ownership. The [capability matrix](https://github.com/SisyphusSQ/mcp-extensions-go/blob/main/docs/compatibility.md) remains the detailed implementation authority.

## Evidence already available

## Historical evidence (2026-10-01/02)

| Layer | Recorded evidence on sqmc04, 2026-10-01 |
| --- | --- |
| Go 1.27.0 | Full race tests, vet and HTTP/stdio build passed |
| Go 1.25.0 | Full suite before final small changes; final HTTP/stdio/store race tests and vet/build passed |
| Other platforms | Go 1.25 Linux/Windows amd64 cross-builds passed; runtime acceptance unperformed |
| Browser | Build/typecheck and installed-Chrome integration passed: official AppBridge, first result, settings/search, changed-field preservation, display/deep-link/context/message/file behavior and absent capabilities |
| Security scan | Cache-local govulncheck v1.8.0 reported no vulnerabilities using Go 1.27.0; its own floor is 1.26, separate from this module's 1.25 floor |
| Plugin CLI | Installed/enabled; an independent official app-server discovered 5 tools, 3 resources and both settings capability locations |
| Desktop user screenshots | Global Open workspace entry, Connected/initial result, effective settings, context count 1 and a model answer identifying bolt after an explicit question |
| Local saved state | The settings file was read with units=in/showGrid=false and mode 0600; this confirms a write, not human restart acceptance |

The PR merge and Wiki documentation work reuse these outputs. No tests, lint or checks are repeated during closeout. No configured CI checks existed on PR #2; do not label that as CI passing. Full details are in [live acceptance](https://github.com/SisyphusSQ/mcp-extensions-go/blob/main/docs/live-e2e.md).

## Implemented features still awaiting host acceptance

| Scenario | What remains |
| --- | --- |
| Settings | Native settings renderer/groups/tool buttons, visible error handling, close/reopen and actual process restart |
| Mentions | Native composer typeahead, empty query, selecting/reading a reference and denied/unavailable search; App Search alone is insufficient |
| Files | `.txt` entrypoint, opaque resource read/authorization, inert HTML-like text, oversized files and unsupported capabilities |
| App instances | Conversation/thread launch, single initial call, two-instance isolation and resource/tool visibility enforcement |
| Display | Reopened feedback fix, actual reported mode, inline-to-fullscreen transition and a host retaining another mode; the old global-page button was reported ineffective |
| Other UI | Quick action, settings App entrypoint, deep-link initial/subsequent navigation |
| Context/messages | Context replacement/removal/updateId, instance isolation and additional targets; the screenshot proves only the stated happy path |

The example is read-only for host files. Writes, ETag conflicts, subscriptions and native file opening need a suitable frontend example plus host acceptance before being claimed. These are TypeScript App/host responsibilities, not Python-server features missing from Go.

## SDK and capability boundaries

Standard MRTR is implemented by the official Go SDK and locally tested for two rounds, cancel/decline/manual resume, duplicate/stale input, expiration/replay, eight concurrent calls and connection isolation. That continuation fixture is not production persistence or proof of authenticated stateless HTTP recovery.

Extended OpenAI forms are different. The typed Go `InputRequestMap` cannot marshal/decode the OpenAI custom method. The default sender rejects method replacement; custom receiving registration does not add custom sending. A public result adapter can emit extension fields, but its complete MRTR interoperability is unproven. Source locations, reproductions and needed public interfaces are in [protocol-investigation.md](https://github.com/SisyphusSQ/mcp-extensions-go/blob/main/docs/protocol-investigation.md).

Only advertise capabilities actually implemented. Read actual client capabilities, not assumptions from UI metadata or a legacy handshake. Browser model context, messages, deep links, choosers, uploads and native file opening belong to standard MCP Apps/OpenAI TypeScript SDKs and the host.

## Business and security ownership

Callers own verified identity, per-resource authorization, multi-user settings, search sources, cross-field transactions and durable continuation policy. The persistent example is single-owner; its Unix state files use 0600, Windows inherited ACLs were not validated, and power-loss durability is not claimed.

Parsing `openai/resource.path` or receiving an opaque URI grants no access. Local reads require authorization, a trusted allowed root, symlink containment and limits. Roots containing hostile hard links/mounts and interruptible regular-file syscalls remain outside the reader contract. Preserve HTTP authentication, runtime credentials, browser text output and all resource boundaries during future changes. The source repository became public at the owner's explicit request on 2026-10-01; this does not change runtime access controls. At the historical acceptance stage no remote HTTPS deployment, public plugin-directory distribution, tag or release was performed. The v0.0.1 module release does not change runtime access controls or complete host acceptance.

## Extension alignment (2026-10-02)

Checked-in Python fixtures cover 51 form schemas and 86 value/upload cases. Go tests also cover typed defaults/aliases, named-type inference, field validation before storage and preserved partial patches. These are extension contract checks, not full Pydantic/ECMA-262/IDNA equivalence or host UI acceptance. The public modern MRTR adapter and generic request-state recovery were removed from scope. The file reader is private example code. Direct extension sending remains SDK-blocked. See [Forms](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/EN-Forms).

Go 1.27.0 and exact Go 1.25.0 passed the final `make test vet build` matrix, including race tests and the authenticated HTTP/child-process stdio examples. No real OpenAI form host E2E was run.

## New-extension subprocess acceptance (2026-10-02)

The earlier isolated official CommandTransport run passed 186 MCP calls: 51 declarations, 82 values and four upload-reference cases; complete answers, typed aliases/defaults/annotations/business rejection; Settings formats, exclusive bounds, multipleOf, partial patches and validator transformations. Invalid patches never reached storage and invalid complete callback state produced tool errors. [Evidence and reproducible runner](https://github.com/SisyphusSQ/mcp-extensions-go/blob/v0.0.1/docs/validation-2026-10-02.md) preserve that run. Its source is archived as examples/acceptance; it is not rerun during v0.0.1 closeout. The new public-only agent quickstart is built, not runtime-accepted. True host forms remain blocked/unaccepted.
