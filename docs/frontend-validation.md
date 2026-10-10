# Frontend example and validation boundary

## Run with the Go server

The example uses `@modelcontextprotocol/ext-apps 1.7.5` and `@openai/mcp-extensions 0.2.0` (unmodified official artifact, pinned in [vendor](../examples/frontend/vendor/README.md)) directly. Node dependencies are local and locked. Browser/server credentials are never bundled into HTML. The default embedded HTTP page remains static; select the real built App explicitly:

```sh
cd /Users/suqing/coding/golang/00_self/mcp-extensions-go/examples/frontend
npm ci --ignore-scripts
npm run build
npm run typecheck
cd /Users/suqing/coding/golang/00_self/mcp-extensions-go
export MCP_BEARER_TOKEN="$(openssl rand -hex 32)"
export MCP_APP_HTML="/Users/suqing/coding/golang/00_self/mcp-extensions-go/examples/frontend/dist/app.html"
go run ./examples/http
```

`MCP_APP_HTML` must be an absolute path to developer-trusted HTML. The example reads it from an explicitly allowed directory with a 2 MiB limit at startup. This is an operator setting, never a host-provided `openai/resource.path`. Nothing installs a background service. Ctrl+C closes the server. The built HTML is ignored by Git; rebuild from the locked frontend inputs.

Connect an MCP Apps-capable host to authenticated Streamable HTTP at `http://127.0.0.1:8080/mcp`, then launch `open_workspace` or its entrypoint. The App registers input/result/host-context listeners before `app.connect()`, renders the initial result without repeating the launch call, reads/updates settings, searches readable demo parts, and offers explicit model-context/message/fullscreen actions when host capabilities allow them. The `open_file` entrypoint reads its opaque file reference through the host resource API and displays text; it never asks Go to read an arbitrary path. Errors are shown using text DOM output. Opening the HTML outside a host shows a notice and disables actions.

HTTP defaults to one mutex-protected in-memory record for one runtime bearer credential. An absolute operator-owned `MCP_SETTINGS_FILE` enables the shared durable example store. The local stdio plugin uses that durable store by default. OS locking and atomic replacement preserve unrelated partial patches across processes. The App reads persisted controls on initialization and saves changed fields only. Multi-user services must still supply identity-scoped storage and authorization.

## Local automated integration

`browser-test.mjs` starts the built Go executable on a temporary loopback port with an ephemeral credential, connects an official TypeScript MCP client, reads the actual Go HTML resource, and loads it in a sandboxed iframe. The test host uses the official `AppBridge` and `PostMessageTransport`. Its local proxy limits methods/resources, checks same-origin JSON requests, and keeps the bearer credential out of the browser. All listeners, servers, the MCP client, browser and child process are closed after the run.

```sh
# Build Go before running the browser suite.
make build
cd examples/frontend
npm run build
npm run typecheck
# Supply an installed Playwright module and an installed Chrome/Chromium.
PLAYWRIGHT_MODULE_PATH=/absolute/path/to/playwright/index.mjs \
PLAYWRIGHT_CHROMIUM_EXECUTABLE_PATH=/absolute/path/to/chrome npm test
```

Without `PLAYWRIGHT_MODULE_PATH`, the suite tries a caller-installed `playwright` package. It never installs Playwright or downloads browsers automatically. On sqmc04 the existing bundled Playwright 1.62.1 and `/Applications/Google Chrome.app/Contents/MacOS/Google Chrome` were used with an isolated headless profile.

The suite covers actual `ui/initialize`, initialized notification, initial tool result, authenticated Go settings read/update, mention search, display/deep-link updates, context/message/current-draft/new-draft payloads and mobile draft gating, host resource text with inert HTML-like contents, missing host capabilities, and standalone mode. Context/message/resource responses are test-host fixtures, not OpenAI model behavior.

## OpenAI host acceptance: partial, user-observed

On 2026-10-01 the local plugin was installed and enabled using the official CLI. A separate official app-server process discovered all five tools and three resources. User-provided desktop screenshots confirmed the global Open workspace entrypoint, Connected state, initial tool result, effective setting values, and an explicit question/model response with context count 1 and an answer identifying bolt. The user also reported that Open fullscreen appeared ineffective; the App now displays actual mode, disables unavailable/already-fullscreen requests, and reports a host response that retains another mode. Reloaded desktop acceptance of that feedback fix is pending.

Native desktop control was unavailable: `com.openai.codex` was rejected by the UI tool, and the running desktop exposes no default CLI control socket. The agent did not restart it or send a conversation message. The local bridge suite and CLI discovery do not establish the remaining desktop outcomes. See [live-e2e.md](live-e2e.md) for installation and manual acceptance details.

Continue human acceptance of the installed plugin and record:

1. Global/thread launch, single initial result, isolation of two conversation/App instances, and supported display modes.
2. Native settings discover/read, grouped layout/tool button, valid changed-field patches, visible validation/persistence errors, and storage behavior after a restart.
3. Composer mention typeahead, empty query, selection, readable selected resource, and denied/unavailable search.
4. File-extension input, host-owned resource read representations, large-file rejection, and unsupported capabilities. Writes/subscriptions need a separate example before claiming their acceptance.
5. Model-context replacement/removal/updateId, explicit active/new conversation messages, and deep-link initialization plus subsequent navigation.
6. Run the public form-mrtr example on a modern extension-capable host and accept/cancel/decline forms, resource arrays, chooser/uploads and preview UI; SDK interoperability does not prove rendering.
7. Desktop/Work current/new draft editing without submission; mobile draft controls disabled.

Do not mark any of these complete based on the local protocol fixture.

## Cross-SDK form interoperability

After `make build`, run `npm run test:forms` from `examples/frontend`. The locked TypeScript client 2.2.0 drives the public Go stdio example, inspecting extended schema metadata and valid/invalid answers and rejecting missing capabilities/legacy protocol. TypeScript SDK 1.31.0 fixes GHSA-6qxp-vccf-f47h in the older browser-test client; the App SDK remains 1.7.5. No OAuth credentials or auth provider are used by these fixtures.
