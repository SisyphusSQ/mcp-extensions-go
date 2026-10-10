# OpenAI MCP Extensions 0.2.0 package

`openai-mcp-extensions-0.2.0.tgz` is the unmodified official build artifact for
[`node-v0.2.0`](https://github.com/openai/mcp-extensions/tree/node-v0.2.0), commit
`d3a79b765e3ad95eaa03f23b0d3d2f91987b8603`.

- Source: [official release workflow run 38008407230](https://github.com/openai/mcp-extensions/actions/runs/38008407230), artifact `packages-node`, `node/package.tgz`.
- SHA-256: `969f406ad38856f4dda4cf7ddefdd06f82e1ff4e796e4dea69c2e2a8478f411d`.
- License: Apache-2.0; the original LICENSE is included in the archive.
- Reason: the GitHub release existed on 2026-10-10, but the public npm registry returned 404 for this version. The checked-in archive and npm lockfile make the frontend example reproducible without an expiring workflow download or an unreviewed source rebuild.

Install with `npm ci --ignore-scripts` from the frontend directory. This archive
is not a fork and no SDK code has been edited. A later migration to the registry
package should verify its provenance and content before changing the lockfile.
