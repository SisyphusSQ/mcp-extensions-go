# Forms and MRTR: verified public SDK boundaries

## Sources and method

The OpenAI protocol and both extension implementations were read at commit `900032d8bd7c1566202d0cb1666986584f932043`. Relevant upstream sources:

- [Specification: OpenAI form elicitation](https://github.com/openai/mcp-extensions/blob/900032d8bd7c1566202d0cb1666986584f932043/docs/spec.md#openai-form-elicitation)
- [TypeScript request implementation](https://github.com/openai/mcp-extensions/blob/900032d8bd7c1566202d0cb1666986584f932043/typescript/src/server/forms/elicitation.ts)
- [TypeScript schema/answer validation](https://github.com/openai/mcp-extensions/blob/900032d8bd7c1566202d0cb1666986584f932043/typescript/src/server/forms/schema.ts)
- [Resource picker schema](https://github.com/openai/mcp-extensions/blob/900032d8bd7c1566202d0cb1666986584f932043/typescript/src/server/forms/file-picker.ts)
- [Python request implementation](https://github.com/openai/mcp-extensions/blob/900032d8bd7c1566202d0cb1666986584f932043/python/src/openai_mcp_extensions/form/_elicitation.py)

SDK source was read from the pinned v1.8.0 module, and `internal/sdkcheck/mrtr_test.go` exercises public APIs through official memory/HTTP transports. No SDK source, transport, or session implementation was copied or modified.

## Standard MRTR works

The prior framework's uncertainty is replaced by concrete evidence: v1.8.0 owns SEP-2322 MRTR. [`mrtr.go`](https://github.com/modelcontextprotocol/go-sdk/blob/v1.8.0/mcp/mrtr.go) installs client retry/input fulfillment and server legacy adaptation. `CallToolResult.InputRequests` and `RequestState` are public, and the SDK sets `resultType` to `input_required` or `complete`.

Local protocol tests exercise two actual tool rounds, automatic client fulfillment, manual split-call resume, input cancellation/decline, stale repeated first-round input, replayed completion, expiration, eight concurrent calls, and cross-connection isolation. The test workflow owns an opaque random continuation registry, deadline, owner, round validation, and completion cache. Those business policies are **not SDK guarantees**. The fixture is connection-scoped and not a production persistence API; it neither tests restart recovery nor proves authenticated stateless HTTP continuation isolation. Production continuations must bind to verified identity/tool/arguments and own their lifetime, replay and transaction policy.

## Extended forms are distinct

Upstream uses `openai/elicitation/create`, negotiated by client `extensions["openai/elicitation"].form`. The form supports primitive fields plus string patterns, described options, thumbnails, suggested values, resource selection, and previews. Accepted answers must satisfy the schema and allowed resource-selection policy.

The resource picker is an `x-openai-input` form field, not a Go filesystem picker. A single selection is a URI string; multiple selections are URI arrays. `resource` is canonical and `file` is a deprecated alias. Selection mode is only for arrays. Implicit selection forbids defaults and permits uploads; explicit defaults must name supplied resources. User file/directory selection, uploads, and preview UI belong to the host. Go would validate selected references and authorize subsequent reads independently.

Standard `ServerSession.Elicit` checks the ordinary elicitation capability and uses `elicitation/create`. Its `RequestedSchema any` can serialize extension keywords, but that does not negotiate or invoke the OpenAI extension. Its return path validates standard schema data and may apply defaults; OpenAI's TypeScript validator deliberately validates submitted answers without changing them. These APIs must not be equated.

## Reproduced limitations and usable seams

| Public API / source | Observed result | Consequence |
| --- | --- | --- |
| [`protocol.go:52`](https://github.com/modelcontextprotocol/go-sdk/blob/v1.8.0/mcp/protocol.go#L52), `InputRequest` and `InputRequestMap.MarshalJSON` | Interface implementations are sealed; embedding ElicitParams satisfies the interface but the concrete-type switch rejects the wrapper as `unsupported type` | Cannot put an OpenAI method in the standard typed request map |
| [`protocol.go:102`](https://github.com/modelcontextprotocol/go-sdk/blob/v1.8.0/mcp/protocol.go#L102), `InputRequestMap.UnmarshalJSON` | `openai/elicitation/create` returns `unsupported InputRequest method` | Official Go client cannot fulfill an extended MRTR result through its typed retry machinery |
| [`shared.go:139`](https://github.com/modelcontextprotocol/go-sdk/blob/v1.8.0/mcp/shared.go#L139), default sending handler | Replacing Elicit's method in sending middleware returns `JSON RPC not handled`, with no request reaching the client | Sending middleware cannot register a custom outbound method |
| [`server.go:1729`](https://github.com/modelcontextprotocol/go-sdk/blob/v1.8.0/mcp/server.go#L1729), Elicit | Standard capability checks, method and result schema validation | Not a generic custom server-to-client request function |
| [`server.go:2286`](https://github.com/modelcontextprotocol/go-sdk/blob/v1.8.0/mcp/server.go#L2286), AddReceivingCustomMethod | Handles custom client-to-server methods; rejects standard-method shadowing | Useful for owned custom services, does not solve custom outbound requests |
| [`shared.go:206`](https://github.com/modelcontextprotocol/go-sdk/blob/v1.8.0/mcp/shared.go#L206), receiving dispatch | Method lookup and parameter decoding run before receiving middleware | Cannot intercept an unregistered unknown method or recover already-rejected input-map decoding |
| `ResultBase` + receiving middleware | A test-only result adapter successfully emits an `input_required` result whose input method is `openai/elicitation/create` over official HTTP | Modern server output adaptation is technically possible; it is **not** blocked in all directions |
| Same adapted result + official Go client | Typed result decoding rejects the custom method | Encoding that result alone is not end-to-end extended MRTR support |

Modern raw HTTP probes include protocol metadata/client capabilities and the required `Mcp-Method`/`Mcp-Name` headers. This avoids confusing malformed modern requests with extension limitations.

The receiving-result adapter remains a **test-only experiment**, not a public sending/MRTR API. The separate `forms` package provides schema/answer validation and typed binding; the fixture does not integrate a supported modern form workflow or durable authenticated continuation contract. No `openai/elicitation` capability is advertised. Modern MRTR is outside current Python-extension alignment; this experiment documents the public SDK boundary only. Legacy direct custom elicitation remains blocked by the public sending API.

## Public SDK interfaces needed

For supported typed interoperability, the official SDK would need public registration for outbound custom server requests with parameter/result factories (preserving context, cancellation, request association and capability checks), and extension registration for MRTR input request/response encoders/decoders. This could be a request registry or explicit custom request carrier; an `any` escape hatch that skips required lifecycle handling is insufficient.

Emitting adapted modern fields requires no SDK fork, but it does not establish supported MRTR interoperability. Python Extensions parity does not require this route, and no adapter or continuation recovery API is shipped. Any independently scoped future implementation must check actual per-request client capability and establish host interoperability and continuation ownership.

Reproduce during development with `go test -race ./internal/sdkcheck -v`. This suite is also included in `make test`. It must not be rerun during commit/push closeout.

## Extension-only alignment scope (2026-10-02)

The public library now implements form declarations, validation and model binding; see [forms usage](https://github.com/SisyphusSQ/mcp-extensions-go/blob/main/docs/forms.md). The receiving-result experiment in this investigation remains test-only. Python extension `elicit_input` does not implement MRTR. Generic continuation encryption/recovery and standard MCP SDK features are excluded from alignment; no public modern MRTR adapter or review tool is shipped.
