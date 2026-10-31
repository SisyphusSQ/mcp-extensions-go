# Upstream 0.2.1 and Desktop WebMCP migration

Module target: v0.0.3. Track [#9](https://github.com/SisyphusSQ/mcp-extensions-go/issues/9).
The official Go SDK stays v1.8.0 and the Go floor stays 1.25.0.

## Two independent references

- Python/Node 0.2.1: `36f8b38df179144b44d157133bbc5f7c6e353ace`.
- WebMCP example: `9e6d418439cca0790bece62f6cd886ad4c04f133`.
  This main-branch example change is not a newer released SDK version.

The settings/mentions/MRTR/resource wire contracts do not require a new Go
protocol implementation. Modern forms still use the official SDK; legacy
custom outbound sending remains blocked by its public API.

## Frontend inputs and security updates

The frontend now pins `@openai/mcp-extensions` 0.2.1 from the official npm
registry. Its upstream package is unmodified. Registry integrity is
`sha512-z2onrOgTsC0kbe6xtSF+r4+jsEn8x7eKgdUI097nqFq5WfNWyVN4susJPy/Dz+ka4OgOB+zZzYYNonpP0qbLwA==`;
the registry also advertises an npm SLSA provenance attestation. This is a
registry observation, not a claim that an independent attestation verifier ran.
The old 0.2.0 workflow archive is no longer needed and was removed.

An initial audit reported nine vulnerable dependency entries. SDK 1.32.2 and
client 2.4.0 address the current SDK advisories, including
[cross-origin redirects](https://github.com/modelcontextprotocol/typescript-sdk/security/advisories/GHSA-6prh-2h8m-c8cw).
Zod 4.6.7 is direct and overridden throughout this example because the upstream
extension pins vulnerable Zod 4.4.3; the dependency graph override does not edit
the official package. Affected transitives were updated within their declared
ranges. App SDK 1.7.5, esbuild and TypeScript remain unchanged. The final audit
reports zero vulnerabilities; compatibility is exercised by type/build/browser
and MRTR tests. Install locally with `npm ci --ignore-scripts`.

## Grouped model context

Content entries carry a stable example selection ID and optional
`_meta["openai/group"]: { title: "Parts" }`. A note demonstrates a second
group; an ungrouped entry uses the host's default Context group. Titles are
nonempty constants. Metadata is not model-visible business content.

The App restores attached selections at initialization and reads
`extensions.modelContext.getCurrent()` on host-context changes. It identifies
its own updates by updateId, reconciles externally removed items, preserves
unsent local selections, and defers in-flight host notifications until the
update result arrives. Context state belongs to the current App instance;
the example settings store intentionally remains shared between local instances.

## Messages and mentions

Message capability enables active/new targets with either send value on
desktop, Work browser and mobile. The previous desktop/web-only draft gate is
removed. All example messages are text, respecting iOS resource-link limits.
`send: false` must remain editable without submission in the actual host.

Work browser composer mentions are available for connected hosted apps through
the existing `openai/mentions` capability. App search and a local stdio plugin
do not prove hosted Web discovery or composer selection.

## Desktop WebMCP

Register tools through optional `document.modelContext` in a Desktop App.
ChatGPT does not support `App.registerTool()`. Three workspace tools read the
view, search parts, and select/remove a part. They reuse UI functions, strict
JSON Schemas and runtime validation. Results are JSON strings; failures reject
rather than returning a successful-looking MCP result envelope. No Go-to-host
RPC, arbitrary tool proxy or new file-write feature is introduced.

Registration is once per App document; execution reads live state. Missing API
or non-Desktop platforms keep ordinary UI available and display an unavailable
status. The test registry is injected only by browser tests and is not bundled
into the App. Actual discovery/calls require a real expanded Desktop App.

## Numeric and display boundaries

Use integer/number fields for JSON numbers. Request string fields for exact
large integer/decimal text, then parse and validate in business code. Standalone
integer validation rejects raw `1.0`/`1e0`; JavaScript normalizes them to `1`
before the SDK wire call, so the real interop chain accepts those normalized
answers. The helper cannot reconstruct lost spelling or precision.

The Python reference was regenerated from the fixed 0.2.1 source, SHA-256
`01e79d6a3bf1aaa14d7b6cd368dabb6fb53c71f56f330ee0875608f3d72a7c66`,
using the existing isolated dependencies recorded in the fixture. It covers
53 declarations and 90 values/upload cases. Deprecated file and preview form
aliases remain accepted. Resource descriptor `openai/preview` is separate.

`ui.ResourceMetadata{}` still emits `openai/ui: {}`, allowing all host display
modes. Completely omitting that key preserves legacy inline behavior. Use
explicit available/preferred modes when presentation matters; the helper does
not synthesize host defaults or change its historical zero-value output.

See [development evidence and real-host limitations](validation-2026-10-31.md).
