# OpenAI extension forms and Python alignment

**English** · [中文](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/ZH-Forms) · [Wiki home](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/Home)

Source: [`docs/forms.md`](https://github.com/SisyphusSQ/mcp-extensions-go/blob/main/docs/forms.md).

This package follows OpenAI MCP Extensions Python 0.2.0 at [`python-v0.2.0`](https://github.com/openai/mcp-extensions/tree/python-v0.2.0). The reference is the extension package, not the underlying MCP Python SDK. The implementation authority is [compatibility.md](https://github.com/SisyphusSQ/mcp-extensions-go/blob/main/docs/compatibility.md).

## Declarations and unchanged answers

`forms.New` snapshots a flat object with primitive fields or string arrays. `forms.Parse` also rejects unsupported JSON keywords and checks their applicability before decoding. Nested objects, arbitrary unions and references are rejected. `Schema()` returns a separate declaration copy.

Fields support `examples`, `$comment`, `_meta`, `deprecated`, `readOnly`, `writeOnly` and a raw JSON `Default`. Nullable annotations follow the Python declaration shape; explicit null is not a valid submitted value or default. Arbitrary JSON examples and opaque metadata retain nulls and numeric precision.

`Enum` is `[]any`, supporting string, numeric, boolean and whole string-array choices. Every enum value must satisfy its field type and constraints. Top-level duplicate enum values are accepted like Python; array-item choices must be nonempty, unique strings. `enumNames` must match the enum length. Rich single choices use `OneOf`, multi-select item choices use `AnyOf`. Titles may be blank, as Python permits. Rich thumbnails and the deprecated preview alias accept HTTPS or the upstream image data-URI grammar; no image is fetched. Suggestions are hints and can coexist with choices. They never authorize a value outside the field's constraints.

Free arrays use string items, optional suggestions, count bounds and optional uniqueness. Multi-select arrays always enforce uniqueness. Choice items and free string items have different allowed keywords, matching the Python protocol.

Form numeric fields expose only `Minimum` and `Maximum`, as in Python. Exclusive bounds and `multipleOf` are settings features, not form fields. Validation rejects non-finite values, missing/unknown fields, wrong types and invalid choices. `Validate`, `ValidateField` and `ValidateResult` do not coerce values, insert defaults or mutate answers. Accept requires content; cancel/decline must be handled separately.

Supported string formats are `email`, `uri`, `date` and `date-time`. Checks run independently of jsonschema-go because its format keyword is not an assertion. Dates reject year zero and invalid calendar values. URI validation is absolute RFC 3986 syntax with no network access. Email checks cover mailbox syntax and domain-label rules; this is not the complete Python email-validator/IDNA implementation. Go patterns use RE2; unrestricted ECMA-262/Pydantic expression parity is not claimed. Python's standalone `is_valid_value` itself refuses patterns without a separate ECMA-262 validator.

Integer JSON literals preserve signed/unsigned 64-bit precision in standalone validation; out-of-range literals fail instead of rounding. Integral float JSON literals such as `1.0` fail integer fields, like Python. Fractional/exponent numbers use finite float64. The official Go SDK can already have converted map arguments or elicitation content to float64; helpers cannot recover lost source representation/precision. Use strings for identifiers beyond the IEEE-754 safe integer range when going through that SDK boundary.

## Resources and upload references

`ResourceInput` uses canonical `Type: "resource"` or deprecated `"file"`, with `Options: []*forms.ResourceOption`. Each option embeds the public `mcp.Resource`; `Extra` preserves additional JSON descriptor fields, including nulls, without replacing standard keys. All option URIs must be valid and unique. Single selection is a URI string; multiple selection is a string array with URI items.

Array `Selection` is explicit or implicit. Implicit selection forbids defaults and supplies default file user options. Defaults must use supplied options even when user selection is enabled. `UserOptions` supports file/directory and the upstream accept-token rules.

`Validate` accepts user URIs when `UserOptions` is declared, matching Python `validate_file_selections`. An optional `SelectionPolicy` can restrict those host selections further. No callback is required merely to validate the declared selection shape. Without `UserOptions`, references outside the supplied options fail. All URI checks are syntax/selection checks; they grant no permission to read a file.

`PrepareSubmission(name, content, pending, policy)` checks existing selections and the final count, returning copied host upload options. `CompleteSubmission(name, content, uploaded, policy)` combines host-authorized upload references with the existing selection and checks the final field. A policy is optional; denied references, malformed URIs, duplicates and count violations fail. Both leave input maps unchanged. The trusted host must authorize uploads and enforce accept filters against actual files; the business service must authorize any later read. These helpers perform no upload, file access or persistence. Validate the complete answer before business writes.

## Typed Go models

```go
type Selection struct {
    Part string  `json:"part_id"`
    Note *string `json:"note,omitempty"`
}

model, err := forms.NewModel[Selection](map[string]forms.Field{
    "part_id": {OneOf: []forms.Option{
        {Const: "bolt", Title: "Bolt"},
        {Const: "washer", Title: "Washer"},
    }},
}, func(ctx context.Context, value Selection) error {
    return validateBusinessSelection(ctx, value)
})
if err != nil {
    return err
}
value, err := model.Decode(ctx, acceptedAnswer, selectionPolicy)
```

Public jsonschema-go inference supplies JSON aliases, descriptions and omission tags. Pointer fields distinguish omission from zero. Overrides use wire names and replace constraints without changing the model field type. `NewModelWithOptions` accepts `*jsonschema.ForOptions` for named-type enums, string constants, constraints and annotations; unsupported keywords fail rather than disappear. String constants become enum choices. A declared model default makes that property optional and is applied only when producing the typed result, without changing the original answer. Business validation runs on the decoded result before the caller writes anything. Decoding errors, callback errors and context cancellation remain visible.

Settings has `FieldsFor[T]`, `FieldsForWithOptions[T]` and `NewModelServer[T]`. Native settings require non-null primitive fields and prohibit schema defaults. Every effective setting is required; updates remain nonempty partial maps. `ModelConfig.SchemaOptions` supports named-type enums/constraints. Typed update callbacks receive exported Go business field names, while declarations and returned state retain JSON aliases. `FieldNames` can explicitly override this mapping. Plain `settings.Config` uses wire names unless a mapping is provided.

`Config.FieldValidators` (also available on `ModelConfig`) maps wire names to `func(context.Context, any) (any, error)`. Validators run on supplied update fields before persistence and may transform values; the transformed patch is checked again. The wire schema is checked first, so Python before/wrap validators that intentionally accept schema-invalid input require caller-specific adaptation. Omitted fields never run validators. Complete-state/cross-field rules must run in the update callback against merged state, within the caller's transaction. Result validation cannot undo a completed write. No database or generic persistence/recovery API is provided.

## Modern form requests

Use `Form.RequestInput(ctx, req, RequestOptions{Key: "selection", Message: "Choose"})` inside an official tool handler. Return the pending `*mcp.CallToolResult` immediately, with no content or business write. On retry, the helper returns a validated `*Answer`; handle cancel/decline explicitly, or call your model's `Decode` on accept. See [the complete public example](https://github.com/SisyphusSQ/mcp-extensions-go/blob/main/examples/form-mrtr/main.go).

The request must use MCP 2026-07-28 or later and advertise both `elicitation.form` and `extensions["openai/elicitation"].form`. The helper uses the official per-request protocol/capability accessors. Missing support returns `ErrUnsupportedClient`; there is no automatic legacy fallback. It uses standard `elicitation/create`, an empty core object schema, and the independent full schema in `_meta["openai/elicitation"].requestedSchema`. The official SDK sets `resultType` during dispatch; callers detect a pending helper result by its non-nil pointer, not `NeedsInput()` before dispatch.

`RequestOptions.Meta` preserves unrelated metadata. `SelectionPolicy` applies the existing answer-selection contract. Responses for other keys remain caller-owned. `RequestState` is opaque, untrusted client-echoed state: the helper does not sign, encrypt, store, expire, recover or authorize it. Business code must bind continuations to verified identity/tool/arguments and enforce replay/transaction rules before any mutation. The public example performs no writes and keeps no continuation storage.

## Sending boundary and parity evidence

Python's extension `elicit_input` sends legacy `openai/elicitation/create`; its documentation explicitly says it does not implement MRTR. Go SDK v1.8.0 has no public custom outbound request API for this method. Legacy custom sending remains unsupported. Modern `Form.RequestInput` implements the 0.2.0 extension adapter on official MRTR; no encrypted continuation helper or business workflow engine is supplied. Standard sessions, transports, MRTR, tasks and persistence machinery stay with the official SDK or application.

[Python-generated fixtures](https://github.com/SisyphusSQ/mcp-extensions-go/blob/main/forms/testdata/python-parity.json) cover 51 declarations and 86 value/upload cases. The [generator](https://github.com/SisyphusSQ/mcp-extensions-go/blob/main/scripts/generate-form-parity.py) records the pinned source commit and dependency versions. Set `PYTHONPATH` to that upstream Python source and run the generator in an isolated environment with the recorded dependencies; Go tests consume checked-in fixtures without Python/network access. These cases verify selected extension semantics, not complete Pydantic compatibility or real OpenAI UI acceptance. Legacy direct sending remains SDK-blocked; real host chooser/upload/preview acceptance remains unverified.

For a consumer workflow and standalone public-only example, start with the [agent integration guide](https://github.com/SisyphusSQ/mcp-extensions-go/blob/main/docs/agent-integration.md). The [2026-10-02 subprocess record](https://github.com/SisyphusSQ/mcp-extensions-go/blob/main/docs/validation-2026-10-02.md) provides additional backend evidence without claiming host form acceptance.
