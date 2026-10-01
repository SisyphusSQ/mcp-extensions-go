# Implementation roadmap

**English** · [中文](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/ZH-Implementation-Roadmap) · [Overview](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/EN-Overview) · [Wiki home](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/Home)

All items below are **not implemented**, unless explicitly described as an existing foundation. This is a future implementation checklist, not a support declaration. Upstream reference: `openai/mcp-extensions` commit `900032d8bd7c1566202d0cb1666986584f932043`, independently checked against upstream main on 2026-10-01.

## Directly implementable work

These server-side types, constructors and validators can be useful independently of an elicitation transport. Implement functional APIs and tests; do not add empty packages or advertise complete form support before a working call path exists.

| ID | Gap | Dependencies | Completion evidence |
| --- | --- | --- | --- |
| D1 | Additional native settings constraints | Existing settings factory and validator | Wire schemas plus invalid-definition/input/result tests |
| D2 | Form schema and rich choices | Upstream form contracts; official MCP types | Construct, serialize and validate complete supported schemas |
| D3 | Suggested/free values and string arrays | D2 | Suggested and custom answers obey the same rules |
| D4 | Resource/file picker declarations | D2; official Resource types | Valid choices/modes/filters/defaults and URI wire shapes |
| D5 | Form answer and upload-reference validation | D2–D4 | Reject malformed/disallowed answers without mutation |
| D6 | Go model/schema binding | D1/D2/D5; public schema APIs | Typed round trips, aliases and explicit validation hooks |

### D1 — Additional settings constraints

Extend the explicit primitive setting fields to cover supported string `format` and numeric `exclusiveMinimum`/`exclusiveMaximum`, comparing Python-generated settings schemas with the native settings contract. Reject inapplicable, unknown or inconsistent constraints. Preserve required effective values, nonempty partial patches and caller-owned persistence.

Acceptance: serialized schema exposes each implemented constraint; invalid declarations fail before storage; invalid patches do not reach the update handler; invalid returned values fail as tool errors. Ensure format validation actually executes rather than emitting an ignored keyword. Test equality at exclusive bounds, integer/number behavior and non-finite values. Unsupported formats must fail explicitly. Document supported pattern syntax; current settings use RE2, not unrestricted Python/ECMA-262 regex parity.

Source: [Python settings](https://github.com/openai/mcp-extensions/blob/900032d8bd7c1566202d0cb1666986584f932043/python/src/openai_mcp_extensions/settings.py); current `settings/settings.go`. Cross-field validation, authorization and transactions remain in the business update callback before persistence.

### D2 — Form schema and rich choices

Provide supported object/field schema types for primitive fields and arrays, required/optional fields, defaults, string/numeric/array constraints, labeled single/multi choices, `enumNames`, and option title/description/thumbnail/preview image metadata. Keep native settings and form semantics distinct.

Acceptance: compatible JSON for supported `oneOf` / array `anyOf` choices; nonempty unique choices; valid labels; correct constraint applicability; consistent bounds; valid defaults; required fields refer to declared properties. Validate upstream image URL/data-URI rules without fetching those URLs. Reject unsupported nested/union schema shapes explicitly. Pattern semantics require a suitable validator or an explicit supported subset; do not silently treat RE2 as full ECMA-262.

Sources: [shared Python form protocol](https://github.com/openai/mcp-extensions/blob/900032d8bd7c1566202d0cb1666986584f932043/python/src/openai_mcp_form_protocol/__init__.py), [TypeScript form schema](https://github.com/openai/mcp-extensions/blob/900032d8bd7c1566202d0cb1666986584f932043/typescript/src/server/forms/schema.ts).

### D3 — Suggestions and free input

Implement `x-openai-suggestions` for string fields and string-array items. Suggested values are not an exclusive enumeration: custom values must remain valid when permitted, with the same length/pattern/format rules as suggestions. Apply array count/uniqueness constraints.

Acceptance: both listed and custom values, invalid suggestions, invalid defaults, empty/required values and duplicate/count boundaries. Validate the submitted answer without inserting defaults or coercing values. Reuse D2/D5 rather than a separate form implementation.

Source: [Python examples](https://github.com/openai/mcp-extensions/blob/900032d8bd7c1566202d0cb1666986584f932043/python/README.md#suggested-values).

### D4 — Resource/file picker declarations

Implement canonical `x-openai-input.type=resource` plus the deprecated `file` alias, resource options, explicit/implicit array selection, user file/directory options and accept filters. Single selection is a URI string; multiple selection is a URI array. Selection mode is array-only. Implicit selection forbids defaults and permits user selection. Explicit defaults must reference allowed resources.

Acceptance: malformed/duplicate URIs, invalid/duplicate accept tokens, invalid field types, unsupported modes, defaults and single/multiple encodings. Preserve official resource metadata. Picked references must never become authorization or arbitrary filesystem reads. The host owns chooser/upload/preview interactions; this task adds no Go browser chooser, database or upload service.

Sources: [Python resource picker](https://github.com/openai/mcp-extensions/blob/900032d8bd7c1566202d0cb1666986584f932043/python/src/openai_mcp_extensions/form/_resource_picker.py), [TypeScript file picker](https://github.com/openai/mcp-extensions/blob/900032d8bd7c1566202d0cb1666986584f932043/typescript/src/server/forms/file-picker.ts).

### D5 — Answers and upload-reference completion

Validate required/unknown fields, JSON primitive types, supported formats, choices, resource URI selections, arrays, defaults and allowed user-file policy. Provide useful equivalents of prepare/complete submission helpers: check pending-upload counts/accept filters, merge host-authorized uploaded URIs with existing selection, then validate the final value. Preserve submitted values; do not add defaults or silently convert strings to numbers.

Acceptance: denied choices/uploads, invalid resource references, missing accepted content, wrong types, scalar versus array shapes, array size/uniqueness, non-finite numbers and malformed schemas. Upload references must come from the owning authorized host/business context; a client-provided URI is not permission. These helpers do not perform uploads or grant file access.

Source: [Python form protocol helpers](https://github.com/openai/mcp-extensions/blob/900032d8bd7c1566202d0cb1666986584f932043/python/src/openai_mcp_form_protocol/__init__.py). Its standalone pattern validation explicitly requires an ECMA-262 validator; do not claim unrestricted regex equivalence.

### D6 — Go model/schema binding

Use public official SDK/schema APIs for Go type-to-schema generation, flat JSON field aliases, typed accepted results and explicit validation callbacks. Provide a usable model helper rather than copying Pydantic decorators or inspecting private SDK registries. Retain the existing explicit field/map APIs.

Acceptance: flat aliases, optional form fields, omitted versus zero values, required effective settings, unsupported model shapes, typed JSON round trips and callback failures. Partial-patch validation must not require omitted fields. Whole-state business rules and authorization must run before persistence in the caller's transaction; validation after a callback cannot undo its writes. Do not promise byte-for-byte Pydantic behavior.

Sources: [Python form schema generation](https://github.com/openai/mcp-extensions/blob/900032d8bd7c1566202d0cb1666986584f932043/python/src/openai_mcp_extensions/form/_schema.py), [Python settings](https://github.com/openai/mcp-extensions/blob/900032d8bd7c1566202d0cb1666986584f932043/python/src/openai_mcp_extensions/settings.py).

## Conditional work: modern OpenAI form MRTR

S1 is not a directly implementable parity claim. A public receiving-middleware/ResultBase adapter already emitted an extended `input_required` result over official HTTP in a local protocol test. Complete transport/continuation interoperability remains unproven; the official typed Go client rejects the custom input method.

After D2–D5, verify an extension-capable real host with the actual per-request client capability. Trace parameter/result decoding and middleware order in the pinned SDK before designing a public API. Bind continuation state to verified identity, tool, arguments and round; bound capacity/lifetime; validate replay, expiration, completion, cancellation/decline, resume and concurrent isolation. For stateless HTTP, prove authenticated identity isolation and ownership after restart if durability is claimed. Caller-provided storage owns transactions and recovery.

Acceptance requires an actual multi-round host transcript, rejected unsupported clients, invalid selections, stale/repeated input, cancellation/resume/expiration and concurrent users/instances. Encoding extension fields or passing standard Elicit tests alone is insufficient. If a public SDK decode/send/capability seam blocks the path, document its source location and a reproduction, request a public extension API, and continue D1–D6. Do not fork, use unsafe/reflection into internals, or duplicate JSON-RPC/sessions/transports.

## Blocked work: legacy direct custom elicitation

S2: Python `elicit_form` / `elicit_input` can send `openai/elicitation/create` through its SDK's public request API and validate accept/decline/cancel. Go SDK v1.8.0 exposes no corresponding custom outbound method registration. Sending middleware cannot overcome its fixed method map. This path requires a suitable upstream public API or a later independently verified SDK version; neither is present in this pinned delivery.

Python's `elicit_input` documentation explicitly says that this legacy wrapper does not implement MRTR. It must not be used as evidence that the Python extension alone provides complete OpenAI MRTR.

See [SDK investigation](https://github.com/SisyphusSQ/mcp-extensions-go/blob/main/docs/protocol-investigation.md) and [Python request implementation](https://github.com/openai/mcp-extensions/blob/900032d8bd7c1566202d0cb1666986584f932043/python/src/openai_mcp_extensions/form/_elicitation.py).

## Implementation order and maintenance

Proceed D1 → D2 → D3/D4 → D5 → D6, then investigate S1; keep S2 blocked until its public API exists. During future implementation update both Wiki languages, the capability matrix, architecture, examples and handoff. Run relevant development tests using official transports and the real frontend/host when available. At commit/push closeout reuse evidence and do not repeat tests/checks. Do not advertise any unimplemented capability or create placeholder APIs.
