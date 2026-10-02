# Integration guide for coding agents

Use this guide when adding OpenAI MCP Extensions to another Go project. Start here, select the required recipe, and read detailed references only as needed. Repository-maintenance instructions remain in [AGENTS.md](../AGENTS.md).

## 1. Inspect the target project and select a capability

Read the target project's instructions, `go.mod`, MCP server constructor, transport, authentication and storage callbacks. Go 1.25.0 or newer is required. Preserve the existing official MCP SDK integration and runtime credential handling.

Install the pinned module from the consuming project:

```sh
go get github.com/SisyphusSQ/mcp-extensions-go@v0.0.1
```

This is a personal extension library, not an official OpenAI SDK. It uses `github.com/modelcontextprotocol/go-sdk v1.8.0`. Review a different existing SDK version before changing it; do not silently downgrade or upgrade the application's dependencies. No global Go configuration or `GOPRIVATE` setting is required for this public module.

| Requirement | Public API | Caller / host responsibility |
| --- | --- | --- |
| Native settings, partial updates | `settings.NewServer`, `settings.NewModelServer[T]` | Identity, authorization, complete merged-state rules and transactional storage |
| Composer resource suggestions | `mentions.AddTool`, `mentions.SearchResult` | Authorized search, readable resources and host typeahead UX |
| App resources and launch entrypoints | `ui.ToolMetadata.Metadata`, `ui.AddHTMLResource` | Trusted HTML, official TypeScript App handshake and host presentation |
| File launch/context metadata | `resources.FileInput`, `resources.Path` | Host resource reads and independent authorization; URI/path metadata grants no access |
| Form schema and accepted-answer validation | `forms.New`, `forms.Parse`, `forms.NewModel[T]` | Obtaining an answer through a supported integration and checking business rules |
| Named-type schema inference | `forms.NewModelWithOptions[T]`, `settings.FieldsForWithOptions[T]` | Explicit public `jsonschema.ForOptions` declarations for custom types |
| Standard tools, sessions, HTTP/stdio, MRTR | Official `mcp` APIs | Existing application's standard MCP integration |

**Sending boundary:** there is no public `elicit_form`, extended-form sending helper or modern form MRTR adapter in this Go library. Legacy `openai/elicitation/create` is blocked by the pinned official SDK's public outbound API. Form declaration/validation support does not mean the host can display a form. Consult [compatibility](compatibility.md) before promising a capability.

## 2. Start a new server or adapt an existing server

For settings, use one settings factory at the existing server-construction point. It returns the ordinary official `*mcp.Server`; retain the original implementation/server options and register the existing business tools/resources on that returned server. There is no `AddSettings(existingServer)` API. Reserve the configured read/update names, which default to `settings.read` and `settings.update`.

For extensions other than settings, keep `mcp.NewServer` and add the relevant helpers. Do not create a second protocol stack, session registry or transport. A settings factory advertises modern and legacy settings capabilities together; callers must not manually advertise host-only capabilities.

The complete [agent quickstart](../examples/agent-quickstart/main.go) imports only public APIs and runs an official stdio settings server. It demonstrates formats, field transformation, JSON aliases, partial patches and concurrency ownership. From this repository:

```sh
go build -o bin/agent-quickstart ./examples/agent-quickstart
```

Configure an MCP client to spawn the absolute path of that binary. It waits for MCP on stdin; a terminal alone is not a client. stdout is protocol-only and errors go to stderr. This local single-owner quickstart keeps state in memory and resets on restart. It opens no network port and supplies no production identity or durable storage.

To reuse it in another project, copy the quickstart into that project's command package, run the pinned `go get` command there, and replace its business callbacks. Do not import this module's `internal/example` packages.

For an existing HTTP service, keep its routing/authentication and expose the returned server through the official `mcp.NewStreamableHTTPHandler`. The [HTTP example](../examples/http/main.go) supplies a complete standalone boundary: runtime Bearer credential, loopback default, cross-origin protection, request limit, timeouts and graceful shutdown. Port those controls into the application's own package when needed. Remote access requires caller-configured TLS and network controls; never copy the MCP handler while discarding its authentication.

## 3. Implement settings callbacks correctly

`Read` returns every effective field, including `false` and `0`. `Update` receives a nonempty partial replacement map; it preserves omitted values rather than replacing the whole record. Authorization, merged-state validation and persistence belong in one caller-owned transaction. An invalid complete return value fails the tool, but cannot undo writes already performed by the callback.

For `NewModelServer[T]`, declarations/results use JSON aliases such as `display_label`; update callbacks receive Go names such as `Label`. Plain `NewServer` uses wire names unless `FieldNames` is configured. Define `FieldValidators` by wire name. They run only on supplied fields after wire validation; transformed values are checked again before `Update`. Python `before/plain/wrap` execution modes and automatic business-model revalidation of read/save results are not reproduced.

Use primitive formats `email`, `uri`, `date`, `date-time` and the available numeric/string constraints. Settings models must be flat non-null primitives; no arrays, nested models or schema defaults. Supply effective defaults through `Read`/storage. See [forms and settings](forms.md) for exact contracts.

## 4. Add only the requested extension recipes

### Forms and typed results

This complete function uses public imports and demonstrates rich choices, aliases and a typed accepted result. The caller must already have obtained the answer; it sends no request.

```go
package integration

import (
    "context"
    "fmt"

    "github.com/SisyphusSQ/mcp-extensions-go/forms"
)

type Selection struct {
    Part string  `json:"part_id"`
    Note *string `json:"note,omitempty"`
}

func DecodeSelection(ctx context.Context, answer *forms.Answer) (Selection, error) {
    if answer == nil {
        return Selection{}, fmt.Errorf("missing form answer")
    }
    if answer.Action != "accept" {
        return Selection{}, fmt.Errorf("selection was not accepted")
    }
    model, err := forms.NewModel[Selection](map[string]forms.Field{
        "part_id": {OneOf: []forms.Option{
            {Const: "bolt", Title: "Bolt"},
            {Const: "washer", Title: "Washer"},
        }},
    }, nil)
    if err != nil {
        return Selection{}, err
    }
    return model.Decode(ctx, answer, nil)
}
```

Handle cancel/decline as explicit business outcomes when applicable. `ValidateResult` preserves the submitted answer; model defaults affect only the typed result. Resource picker `UserOptions` enables host selections; an optional `SelectionPolicy` can restrict them. `PrepareSubmission` and `CompleteSubmission` validate upload references and counts, not file bytes, permissions or an upload service. Validate the complete answer before business writes.

Flat primitive/string-array schemas are supported. Nested objects/arbitrary unions are not. Patterns use RE2; email checks are not full Python email-validator/IDNA parity. Standalone integer JSON is limited to signed/unsigned 64-bit. Ordinary official SDK map decoding may already lose integer representation/precision; no extension helper can recover it. See [exact form limits](forms.md).

### Mentions

Register `mentions.AddTool(server, &mcp.Tool{Name: "search_mentions"}, searchHandler)`. The callback signature is `func(context.Context, *mcp.CallToolRequest, mentions.SearchParams) (mentions.SearchResult, error)`. Return `mentions.Item{Link: &mcp.ResourceLink{URI: ..., Name: ..., MIMEType: ...}}` or the documented resource variant. Empty queries are valid; no matches must produce an empty `Items` slice. The helper adds the search marker and app visibility, not a separate mention capability or authorization. A search link does not automatically register its readable resource. See [shared example registration](../internal/example/server.go) as source reference, not an importable public package.

### App entrypoints

Use `ui.ToolMetadata{ResourceURI: "ui://my-app/home.html", Visibility: []ui.Visibility{ui.App, ui.Model}, Entrypoints: []ui.Entrypoint{{Type: ui.Global}}}.Metadata(nil)` and check its error. Put that metadata on the real official tool and register its trusted built HTML with `ui.AddHTMLResource`. Entrypoint tools accept `{}` unless they are file entrypoints. Follow the [frontend setup](frontend-validation.md) and [official App example](../examples/frontend/app.ts) for handlers-before-connect, `ui/initialize`, initial results and host capability gating. A static HTML resource alone does not implement an App handshake. Model context/messages/deep links use the official TypeScript App SDK, not a Go-to-host RPC invented by the consumer.

### Files

Use `resources.FileInput` for the filename/opaque resource URI and call its `Validate`. Read host-owned contents through the frontend's official resource API when supported. `resources.Path(req.Params.Meta)` parses optional context only; never open it or turn an opaque URI into a filesystem path without independent authorization and a controlled filesystem boundary. `internal/example.Reader` is private example code, not a supported library API.

## 5. Validate the integration and report the actual proof level

During implementation, use the target repository's prescribed checks. For this library's development, the entrypoint is `make test vet build`. Do not rerun tests during commit/release closeout when the applicable collaboration rules require reusing evidence.

For the new server extension scenarios, the reusable [subprocess acceptance runner](../examples/acceptance/main.go) uses official `CommandTransport` and temporary in-memory business state. From a checkout of this repository:

```sh
go run ./examples/acceptance
```

Expect six `PASS` lines and exit zero. Its 186 MCP calls cover the pinned Python fixtures, complete answers, typed models and actual settings read/update errors. It preserves raw answer JSON deliberately. It opens no port, reads no selected file and changes no installed Workspace settings. [Recorded evidence](validation-2026-10-02.md) is not a claim that the runner was repeated during this release.

For the consuming application, inspect initialization/discovery, call its actual tools and check valid/invalid inputs, omitted updates, returned state and storage errors. Test its real callbacks/storage and authorization rather than treating this repository's temporary runner as production acceptance. For UI features, record the real host's initialization and interaction separately from local browser fixtures and protocol calls. Use [the host checklist](live-e2e.md); true form picker/upload/preview acceptance remains pending.

Deliver the chosen capability/API, dependency version, changed integration point, business ownership decisions, reproducible commands/results, and remaining SDK/host limitations. Never advertise all Python extension behavior or live product acceptance from a passing build or fixture suite.

## Copyable task prompt

> Integrate the requested OpenAI MCP Extensions into this Go project using mcp-extensions-go v0.0.1. Read the target repository's instructions and this integration guide, then consult the capability matrix and only the needed API references. Reuse existing authentication, storage and official MCP transports. Preserve partial updates, apply merged-state rules before writes and authorize resource access independently. Do not invent unsupported form sending, MRTR/recovery or private example APIs. Implement the requested path, report actual validation evidence and distinguish backend results from real host acceptance.
