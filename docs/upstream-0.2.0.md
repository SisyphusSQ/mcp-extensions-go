# Migrating to OpenAI MCP Extensions 0.2.0

This Go module's v0.0.2 targets Python/Node 0.2.0. The Go SDK stays at v1.8.0
and the Go floor stays at 1.25.0. It remains a personal, unofficial extension
library. See the [upstream changelog](https://github.com/openai/mcp-extensions/blob/python-v0.2.0/CHANGELOG.md)
and [implementation authority](compatibility.md).

## Mention discovery

Prepare capabilities before constructing the server:

```go
options, err := mentions.WithCapability(existingOptions, "search_mentions")
if err != nil {
    return err
}
server := mcp.NewServer(implementation, options)
// Alternatively pass options to settings.NewServer or NewModelServer.
if err := mentions.AddTool(server, &mcp.Tool{Name: "search_mentions"}, searchHandler); err != nil {
    return err
}
```

`openai/mentions: {searchTool}` is advertised in extensions and the legacy
experimental location. The named tool must exist on that server before connection.
The official Go SDK has no public post-construction capability setter, so AddTool
alone retains only the legacy marker. WithCapability preserves unrelated
capabilities and rejects duplicate declarations. The registered helper ensures
readOnlyHint and app visibility; neither grants resource access.

## Settings migration

`settings.Item{Kind: "tool", Tool: "open"}` no longer requires Title. Existing
nonblank Title values remain accepted for source compatibility but are deprecated
and ignored by hosts. Set the referenced tool's Title; hosts fall back to
annotations.title, then name. The tool must accept `{}`.

`type: "settings"` App entrypoints were removed upstream. Go's `ui.Settings`
and Entrypoint.SearchTerms remain deprecated symbols, but metadata construction
now returns a migration error for that variant. Remove that entrypoint. Keep
structured `openai/settings` read/update tools; use a layout tool button to open
a custom settings App. Global/thread/file entrypoints remain supported.

## Modern extended forms

Use `Form.RequestInput` in an ordinary official tool handler. It returns either
a pending tool result or the validated answer for RequestOptions.Key:

```go
pending, answer, err := model.Form().RequestInput(ctx, req, forms.RequestOptions{
    Key: "selection", Message: "Choose a part",
})
if err != nil || pending != nil {
    return pending, nil, err
}
if answer.Action != "accept" {
    return nil, map[string]string{"action": answer.Action}, nil
}
value, err := model.Decode(ctx, answer, selectionPolicy)
return nil, value, err
```

The complete [public-only stdio example](../examples/form-mrtr/main.go) is built
as `bin/form-mrtr` by make build. A host must use MCP 2026-07-28 or later and
advertise both elicitation.form and extensions["openai/elicitation"].form.
Unsupported requests fail before producing a form; no ordinary/legacy fallback
is silently substituted. Accepted answers retain the existing validation and
selection policy. Cancel and decline are explicit business outcomes.

The full schema is carried in ElicitParams.Meta while the core RequestedSchema
is an empty object. Official SDK serialization, MRTR and retry decoding are reused.
The SDK sets resultType during dispatch. Do not call NeedsInput on a newly
constructed helper result to decide whether to return it; test pending != nil.
No JSON-RPC stack, SDK internals, custom result carrier or metadata middleware
is introduced.

RequestState is untrusted, client-echoed, caller-owned continuation state. This
helper provides no identity, storage, signature, encryption, replay protection,
restart recovery or transaction engine. Authorize each request and bind/validate
continuations before using answers for a mutation. Resource URI validation does
not authorize a later file read. Legacy openai/elicitation/create sending remains
blocked by the pinned SDK public API.

## Frontend and editable drafts

The example uses the unmodified official Node 0.2.0 build from its tagged release
workflow, committed as a 33 KiB archive with [provenance and SHA-256](../examples/frontend/vendor/README.md).
The public npm registry returned 404 for 0.2.0 at implementation time. Local
`npm ci --ignore-scripts` uses that file and the lockfile; no global install or
SDK fork is required.

The browser retains official App + OpenAIExtensions. Existing Send keeps its
default immediate-send behavior. Add question to draft uses active/send:false;
Open new draft uses new/send:false. Draft controls require message capability
and a desktop/web hostContext.platform, and are disabled on mobile or an unknown
platform. Only explicit clicks request messaging; there is no automatic send.
Actual editing/rendering belongs to the host.

The old createAppTransport was removed upstream; this example did not use it.
The upstream TypeScript server migration to SDK 2 does not change our Go dependency.
The browser-test client uses SDK 1.31.0 and the modern form interop client uses
2.2.0, fixing [GHSA-6qxp-vccf-f47h](https://github.com/modelcontextprotocol/typescript-sdk/security/advisories/GHSA-6qxp-vccf-f47h).

## Development acceptance

Run make test vet build on the declared Go toolchain and exact Go 1.25 floor.
Run frontend typecheck/build, test:forms, browser tests with an installed browser,
and npm audit while still in development. Test accepted/cancelled/declined and
invalid answers, both capability gates, legacy refusal, two-round and concurrent
forms, discovery, optional titles and removed entrypoints. These are protocol
and local browser proofs; genuine host chooser/upload/preview and draft editing
need separate acceptance. During commit/release closeout, reuse development
outputs and do not repeat these checks.
