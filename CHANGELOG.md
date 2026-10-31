# Changelog

## Unreleased

### v0.0.3(20261031)

- feature
  - Align with upstream 0.2.1 and add Desktop WebMCP workspace tools using the official example's document.modelContext API.
  - Add grouped model context with host removal/remount synchronization and capability-gated mobile active/new messages and drafts.
  - Refresh Python 0.2.1 fixtures and add ordinary-number/exact-string MRTR interoperability examples.
- bugFix
  - Pin patched TypeScript SDK clients and Zod; update affected transitive dependencies without changing the official extension package.
- note
  - Preserve the zero-value resource metadata declaration; document its distinction from omitted openai/ui metadata.
  - Go SDK v1.8.0, Go 1.25.0 floor, legacy outbound limitations and caller-owned authorization/state remain unchanged. Real OpenAI host acceptance is separate from local browser/protocol validation.

### v0.0.2(20261010)

- feature
  - Align the extension helpers with upstream Python/Node 0.2.0: modern form MRTR input and validated retry answers, Mention capability discovery, and editable frontend drafts.
- bugFix
  - Accept omitted Settings tool-item titles and reject removed Settings App entrypoints with a migration error.
  - Pin the frontend SDK clients to OAuth advisory fixes.
- note
  - Retain deprecated Settings symbols and nonblank legacy button titles for source migration; hosts use the referenced tool's display name.
  - Go SDK v1.8.0 and Go 1.25.0 floor are unchanged. Legacy custom form sending and real host acceptance remain separate limits.
  - The frontend uses the unmodified official 0.2.0 tagged build archive, with provenance, checksum and locked dependencies, while npm has no published 0.2.0 package.

### v0.0.1(20261002)

- feature
  - First personal Go source release: native Settings, Mention search, file metadata, App metadata/resources, extended form declarations/validation/models, and public integration examples.
- note
  - The module uses official MCP Go SDK v1.8.0 and requires Go 1.25.0. See the [original release](https://github.com/SisyphusSQ/mcp-extensions-go/releases/tag/v0.0.1) for its historical evidence and boundaries.
