# Extension acceptance evidence, 2026-10-02

The new extension APIs at merged commit `89f92dd97be867c68dd3eaf27df093df3824be19` passed 186 actual MCP tool calls on sqmc04 / darwin arm64, built with Go 1.27.0. The official Go SDK client started a separate stdio server process through `CommandTransport`; the server used public `forms` and `settings` APIs with isolated in-memory business callbacks.

The originally ignored probe source is now maintained as [examples/acceptance](../examples/acceptance/main.go). Its acceptance logic is preserved. The archive and new agent quickstart were built during release preparation, but tests were not repeated for commit/release closeout. This record refers to the earlier run, not a new execution of the relocated runner or quickstart.

| Scenario | Prior result |
| --- | --- |
| Pinned official Python extension fixtures | Pass: 51 declarations, 82 values, four upload-reference cases, including rich choices/images, suggestions, arrays, annotations, resources and a precise standalone integer enum |
| Complete answers and upload preparation | Pass: required/unknown fields, accept/cancel/decline, free suggestions, malformed references and final count limits; inputs unchanged |
| Typed forms | Pass: aliases, named-type schema inference, static defaults, annotations, typed results and business rejection; submitted answers unchanged |
| Actual settings tools | Pass: email/date/date-time/URI, exclusive bounds, multipleOf, typed aliases, validator transformation, transformed-value revalidation, partial patches and effective zero values |
| Invalid patch storage boundary | Pass: rejected before the storage callback; isolated state unchanged |
| Invalid complete callback state | Pass: read and update report tool errors; callbacks still own transactions and cannot rely on output checks to roll back writes |

The six recorded output lines were:

```text
PASS: actual MCP subprocess: 51 Python declarations, 82 values and 4 upload-reference cases
PASS: complete answer required/unknown fields, accept/cancel/decline, free suggestions and upload preparation/count limits
PASS: typed form aliases, named-type inference, static defaults, annotations, unchanged answers and business rejection
PASS: actual Settings tools: formats, exclusive bounds, multipleOf, aliases, field transformations, rejection before save, omitted fields and effective zero values
PASS: invalid complete state from read/save is reported as an error; callbacks retain transaction responsibility
PASS: 186 MCP tool calls; no original Workspace state or desktop process changed
```

Before release preparation, the existing browser integration also passed against the official App/AppBridge and authenticated Go HTTP example. Its model-context/message/deep-link/file responses were fixture-host behavior, not native Codex proof. The installed Workspace's actual units-only/grid-only partial updates passed, and the original `units=in`, `showGrid=false` state was restored. The local plugin was rebuilt/reinstalled then; this source release does not reload desktop processes or publish a plugin-directory entry.

Go 1.25.0 and Go 1.27.0 `make test vet build` results from the preceding extension implementation remain the separate library-development evidence. They were not repeated during this documentation/release closeout.

Legacy custom form sending remains SDK-blocked. Native picker/upload/preview rendering is unaccepted. The probe's answer JSON is intentionally preserved; ordinary official SDK map decoding may lose integer representation/precision. Email/IDNA and RE2/Pydantic pattern differences remain documented in [forms.md](forms.md). This evidence does not claim full Python equivalence, production storage/authorization, remote deployment or CI success.
