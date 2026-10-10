# Changelog

## Unreleased

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
