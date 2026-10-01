# Architecture

## Goal and current scope

Provide Go types, metadata encoding, and registration helpers for OpenAI MCP Extensions. Business services continue to use the official MCP Go SDK; this repository does not maintain a second JSON-RPC or transport implementation.

This delivery provides a compilable, runnable SDK foundation with protocol integration tests: UI metadata, HTML resources, an authenticated HTTP example, and handoff documentation. Other extensions await concrete requirements. No empty packages or unusable future interfaces are included.

## Layers

```text
Business service / examples/http
  +-- ui: extension types, metadata validation, HTML resource registration
  +-- github.com/modelcontextprotocol/go-sdk/mcp v1.8.0
       +-- tool registration, input validation, output schemas
       +-- resources, sessions, capabilities
       +-- JSON-RPC, Streamable HTTP, standard elicitation

Frontend (future): standard MCP Apps App + OpenAI TypeScript app extensions
  +-- browser DOM, postMessage, host context, deep links
```

OpenAI's Python extension SDK similarly builds on the MCP Python SDK, adding Pydantic models, FastMCP helpers, and extension behavior. Go follows the same layering direction, but Python dynamic objects and TypeScript browser APIs cannot be translated mechanically. See the pinned [upstream Python source](https://github.com/openai/mcp-extensions/tree/900032d8bd7c1566202d0cb1666986584f932043/python).

## Current design

`ui` does not wrap the entire `mcp.Server` or redesign ToolHandler. Callers obtain `mcp.Meta` through Metadata methods and pass it to `mcp.AddTool`, preserving official type inference and lifecycle handling.

Metadata is validated and snapshotted through JSON encoding. Functions, channels, and other non-JSON values are rejected. Entrypoint structs represent a union discriminated by Type; fields belonging to another variant are rejected before encoding. File extensions must already be canonical: `.csv` is accepted, surrounding whitespace is rejected rather than trimmed.

`AddHTMLResource` accepts developer-trusted HTML strings, not paths, URLs, or template input. MIME type, HTML byte size, and URI are fixed at registration. Original resource Meta is preserved on the descriptor, then merged with OpenAI display declarations on resource contents. Re-registering a URI follows official SDK replacement semantics. Icons, Annotations, and other descriptor fields follow SDK ownership conventions and must not be mutated concurrently after registration.

The HTTP example uses stateless JSON responses for its current tools and resources. It does not support server-initiated requests that require a client response. Authentication belongs to the example's access layer; the library does not infer user identity. Visibility, entrypoints, and annotations are declarations only.

## Extension points in v1.8.0

| Official SDK API | Applicable extension work | Boundary |
| --- | --- | --- |
| `mcp.Meta` | UI, mentions, resource paths, other namespaces | Requires extension validation; metadata is not feature implementation |
| `ServerCapabilities.Extensions` | OpenAI capability negotiation | Configure before constructing Server; declare only implemented capabilities |
| `AddReceivingMiddleware` / `AddSendingMiddleware` | Request context, interception, result adaptation | Method parsing precedes receiving middleware; unknown methods do not become supported automatically |
| `AddReceivingCustomMethod` | Client-to-server extension methods | Cannot override built-in methods; not a generic bidirectional RPC channel |
| Standard `ServerSession.Elicit` | Standard elicitation | Standard type and client capability checks apply; not full OpenAI forms |

MRTR and extended forms need a separate protocol investigation. v1.8.0 exposes no generic `ServerSession.CallCustomMethod`; its `InputRequest` interface and encoding paths support specific built-in request types. This framework does not access private connections or modify SDK internals. For nonstandard server-initiated requests, establish a supported extension API or propose an upstream API before choosing an implementation.

For an MRTR flow based on extended multi-round tool results, first verify `tools/call` result adaptation, host follow-up input, and resumed tool calls. Constructing extension fields alone does not establish host acceptance, and standard Elicit is not a substitute. This delivery does not claim that MRTR can be fully implemented without SDK changes.

## Suggested continuation

1. Connect a real frontend using standard MCP Apps App and OpenAI TypeScript app extensions. Validate initialization, tool results, display modes, and entrypoints while keeping the server in Go.
2. Add settings, mentions, and file-context helpers for real business requirements. Implement handlers, negotiation, and errors before updating the matrix. Validate file metadata paths against authorized roots before reading; never trust an arbitrary host path.
3. Verify extended forms and MRTR against a host. Define cancellation, duplicate input, concurrency isolation, expiration, and resume semantics before designing the Go API.
4. Before publishing, decide version policy, license, compatibility guarantees, and release workflow. This delivery creates no release or tag.

These are continuation options, not implemented features or separately created tasks.
