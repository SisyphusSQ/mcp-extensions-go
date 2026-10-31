# Extension alignment status

**English** · [中文](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/ZH-Implementation-Roadmap) · [Wiki home](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/Home)

## Current v0.0.3 / upstream 0.2.1 and WebMCP

Source/local-test alignment is 0.2.1 (`36f8b38`); the separately pinned WebMCP example is `9e6d418`. The App implements named/default context groups, external item/group/all removal with updateId/remount/in-flight synchronization, mobile active/new send/draft combinations, and optional Desktop `read_view`, `search_parts`, `select_part` tools via `document.modelContext`.

The registry extension is unchanged. SDK 1.32.2/client 2.4.0 and a Zod 4.6.7 override address current dependency advisories; App SDK remains 1.7.5. Python 0.2.1 fixtures cover 53 declarations and 90 values/upload cases. Local Go 1.25/1.27 race/vet/build, frontend type/build/browser, 19 MRTR scenarios, 192 acceptance calls and npm audit passed. Exact numeric text uses strings; empty resource metadata remains an explicit all-mode declaration.

Native Desktop UI access was refused on 2026-10-31. Actual WebMCP discovery/calls, host context deletion, mobile drafts and Work browser hosted mentions remain unaccepted. The installed plugin/cache is unchanged. See [migration](https://github.com/SisyphusSQ/mcp-extensions-go/blob/main/docs/upstream-0.2.1.md) and [current evidence](https://github.com/SisyphusSQ/mcp-extensions-go/blob/main/docs/validation-2026-10-31.md); older observations below retain their historical scope.


## Historical v0.0.2 / upstream 0.2.0 status

Modern Form.RequestInput, Mention capability discovery, optional legacy Settings button titles, rejection of removed Settings App entrypoints and desktop/web drafts are implemented. The 2026-10-01/02 records below are historical; their exclusion of modern MRTR is superseded by this authorized work. Generic recovery and legacy custom sending remain outside support. Real host forms/drafts are unverified. See [current development evidence](https://github.com/SisyphusSQ/mcp-extensions-go/blob/main/docs/validation-2026-10-10.md).

## Historical evidence (2026-10-01/02)

The owner's 2026-10-02 scope targets OpenAI MCP Extensions Python 0.1.0 at `900032d8bd7c1566202d0cb1666986584f932043`, excluding the underlying MCP Python SDK. [The source capability matrix](https://github.com/SisyphusSQ/mcp-extensions-go/blob/main/docs/compatibility.md) is authoritative.

| Area | Implemented | Remaining limit |
| --- | --- | --- |
| D1 Settings | Formats, inclusive/exclusive numeric bounds, primitive schema/model helpers, pre-save field validation, business-name aliases | Caller transactions/cross-field rules; no automatic Python before/wrap decorator engine |
| D2 Forms | Flat types, generic JSON enum, annotations, rich single/multi choices, thumbnails/preview aliases | Documented RE2/email/number limits |
| D3 Suggestions/free input | Hints alongside choices, free strings/string arrays, length/count/uniqueness | Host rendering unverified |
| D4 Resource inputs | Resource/file declarations, extra descriptors, explicit/implicit selection, user file/directory options and accept filters | Host chooser/uploads/read authorization |
| D5 Answers/uploads | Unchanged values, fields/types/choices/resources, prepare/complete upload references | No actual bytes uploaded; caller authorizes subsequent reads |
| D6 Models | Public inference options for enums/constraints, JSON aliases, typed defaults/results and Go business callbacks | No general nested/unions/Pydantic decorator implementation |
| Legacy direct sending | Public SDK boundary reproduced | Go v1.8.0 custom outbound API missing; remains blocked |
| Real host E2E | Existing App evidence remains bounded | Forms/chooser/upload/preview and the remaining host checklist are unverified |

Modern form MRTR adaptation, generic encrypted request-state/restart recovery, dependency frameworks, standard sessions/transports, tasks and EventStore are excluded. The previous public adapter/state package and two-round review example were removed. The root-contained reader moved into private example code; the existing settings store remains a business example. Python's legacy `elicit_input` does not implement MRTR.

Python-generated reference fixtures cover 51 schema and 86 value/upload cases. See [Forms](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/EN-Forms) for APIs, reproducibility and precise limits. Do not interpret fixtures as complete Pydantic equivalence or live product acceptance. Source Wiki updates are published separately from the main repository; installed plugin caches require a separate rebuild/reload. During development use repository test/vet/build; at commit/push/release closeout reuse evidence without rerunning checks.
