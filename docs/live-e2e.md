# Private local Codex plugin and acceptance

## Location and installation

PR [#1](https://github.com/SisyphusSQ/mcp-extensions-go/pull/1) was merged into private `main` at `3f45cf66eab5f6fd172763e604d794e12cef5cf1`. This continuation branch is `suqing/live-e2e`.

Plugin: **MCP Extensions Go**, ID `mcp-extensions-go@mcp-extensions-go-local`, local marketplace **MCP Extensions Go Local**, version `0.0.0-dev` (installation metadata only; no release/tag). The package contains actual MCP configuration, a generated executable, App HTML and an icon; no empty skills/hooks/apps are included.

```sh
cd /absolute/path/to/mcp-extensions-go
npm --prefix examples/frontend ci --ignore-scripts
make plugin
codex plugin marketplace add /absolute/path/to/mcp-extensions-go
codex plugin add mcp-extensions-go@mcp-extensions-go-local
```

`make plugin` writes ignored `plugins/mcp-extensions-go/bin/mcp-extensions-stdio` and `assets/app.html`. Go and Node must already be available; no global tool installation or browser download is performed. The official CLI copies these files into its cache and enables the plugin. Rebuild/reinstall after source changes. Existing host processes/pages may retain the previous executable/document until the plugin is reopened or the desktop is restarted.

On sqmc04 the reported cache root is `/Users/suqing/.codex/plugins/cache/mcp-extensions-go-local/mcp-extensions-go/0.0.0-dev`. This installed CLI uses the manifest version in the cache path, rather than the `local` suffix described by the general plugin guide. The supported `.codex-plugin/plugin.json` / `.mcp.json` compatibility format follows upstream Bits & Bolts. The installed CLI accepted portable root manifest metadata but returned no MCP servers for its `mcp.json`; compatibility format resolved `workspace` and loaded the actual server.

The official SDK owns stdio framing/sessions/transports. stdout contains MCP only; logs use stderr. The stdio parent controls local process access. No HTTP credential, listening port, daemon, remote service or public plugin publication is created. The separate HTTP example retains Bearer authentication.

## Settings ownership and restart behavior

The plugin defaults to `os.UserConfigDir()/mcp-extensions-go/live-e2e/settings.json`; on sqmc04 this is `/Users/suqing/Library/Application Support/mcp-extensions-go/live-e2e/settings.json`. An absolute operator-owned `MCP_SETTINGS_FILE` overrides it. This path is not supplied by host tool input.

The directory/lock file is created on startup; the settings file is created only on a successful update. File-backed reads are capped at 64 KiB; saved files are 0600 on Unix. Windows uses inherited user-directory ACLs, which were not separately validated. A cancellable OS lock spans reload, patch and atomic file replacement. Corrupt/unreadable/oversized state fails explicitly rather than resetting it. Values survive normal process restarts. The store is single-owner; all local App instances intentionally share settings. Identity-scoped multi-user storage and power-loss durability are outside this example. Public `settings` APIs still require caller-provided storage.

## Observed evidence on 2026-10-01

| Layer | Evidence | Limit |
| --- | --- | --- |
| Installation | Official CLI reports installed/enabled; plugin/read resolves `workspace` | Does not by itself prove the current desktop loaded it |
| Official CLI protocol | An independent app-server process initializes the plugin, discovers all 5 tools and 3 resources, and sees modern/legacy settings capabilities | Not the running desktop's App bridge |
| Go subprocess protocol | Official CommandTransport exercises two simultaneous stdio processes, unrelated partial updates, resource discovery/read, shutdown and a fresh process reading saved state | Local protocol test, not human host restart acceptance |
| Desktop global navigation | User screenshot shows Open workspace in the sidebar's more menu and the MCP Extensions Go page | Screenshot observation |
| Desktop initialization | User screenshot shows Connected and the initial Welcome to your workspace result | No duplicate-count/instance-isolation trace |
| Desktop settings | User screenshot shows effective values units=in and showGrid=false | Native settings renderer and reopen/restart acceptance pending |
| Local persisted state | Actual user configuration file contains units=in/showGrid=false with mode 0600 on sqmc04 | Confirms a durable write, not a human reopen/restart result |
| Desktop context/message | User screenshot shows context count 1, the explicit question, and a model reply identifying bolt | Replacement/removal/updateId/instance isolation and other targets pending |
| Desktop display | User reported Open fullscreen appears ineffective on the global page | Actual host mode was not captured then; no mode-transition success claimed |

The display feedback fix reads host context and the actual request result. Already-fullscreen and unsupported controls are disabled with clear labels; a retained mode is reported explicitly. The App declares inline/fullscreen support and shows Display mode. Local browser tests cover initial fullscreen, unavailable fullscreen, successful transition and the host retaining inline. Reloaded desktop acceptance of this fix is pending.

The UI tool refused native `com.openai.codex` control. Its earlier MCP Apps inventory was empty, and the running desktop had no default CLI control socket. The user performed actual desktop actions; the agent did not restart the desktop or send a conversation message.

## Human acceptance

1. Close/reopen **Open workspace** after the latest local install. If the desktop retains old plugin files, restart it after this active turn completes. The updated page should show Display mode. It disables Open fullscreen when already fullscreen or unavailable; requesting a transition reports the host's actual outcome.
2. Read settings, change Units/Show grid, save, then close/reopen. Read again and confirm values. Restart the desktop/plugin process and repeat. Invalid or unavailable storage must show an error.
3. Search bolt/washer/empty query in the App. Separately exercise composer mention typeahead and selection; App search alone does not prove native mentions.
4. Exercise the `.txt` file entrypoint if offered, using `examples/frontend/acceptance.txt`. HTML-like text must remain inert. No arbitrary host path may be opened by Go. Host resource reads/permissions and oversized file behavior still need acceptance.
5. Test inline launch from a conversation, global/thread App isolation, context replacement/removal, explicit message targets, and deep-link initialization/navigation. Existing screenshots prove only the stated happy paths.

Full OpenAI extended forms/MRTR remain incomplete for the documented SDK boundaries. No such capability is advertised. Do not mark them accepted through standard elicitation or generated extension fields.

## Development evidence

- `make test vet build` with Go 1.27.0: passed, including real child-process stdio and persistent-store race tests.
- Go 1.25.0: full suite passed earlier in this branch; final modified HTTP/stdio/store race tests plus `make vet build` passed using cache-local `GOTOOLCHAIN=go1.25.0`.
- Go 1.25 Linux/Windows amd64 cross-builds passed; runtime behavior there was not tested.
- Frontend build/typecheck and installed-Chrome browser integration passed, including changed-field patches preserving another App's unrelated update. Display feedback scenarios are covered by the final browser run.
- `go run golang.org/x/vuln/cmd/govulncheck@v1.8.0 ./...` reported No vulnerabilities found with Go 1.27.0. The scanner ran from module/build caches, not a global installation; no module version changed. Its own Go floor is 1.26, separate from this module's verified 1.25 floor.

Security self-review: HTTP auth was preserved; local stdio uses parent process access; host URIs never become filesystem reads; persistent paths are operator-owned with bounded reads, OS locks and private atomic state files; browser output uses text DOM APIs. These findings do not establish multi-user authorization or complete host acceptance.

Commit/push closeout reuses this development evidence without repeating tests or checks. The branch remains available for human acceptance; this does not merge or publish a new release.
