# Project collaboration rules

- This is a personal Go SDK at `github.com/SisyphusSQ/mcp-extensions-go`, not an official OpenAI SDK.
- Use official `github.com/modelcontextprotocol/go-sdk v1.8.0`; the Go version floor is 1.25.0. Do not duplicate JSON-RPC, sessions, or transports.
- Before editing, read `README.md`, `docs/architecture.md`, `docs/compatibility.md`, and the relevant source.
- For consumer integration, start with `docs/agent-integration.md`; use the public-only `examples/agent-quickstart`. Keep this maintenance file separate from consumer instructions.
- Use English for comments, exported API documentation, agent instructions, and project documentation. Maintain `README_ZH.md` as the Chinese companion to `README.md`.
- Follow Go naming conventions and format changed Go files with gofmt.
- Implementation status is defined by `docs/compatibility.md`. Do not advertise planned extensions as supported.
- During development use `make test vet build`. Report missing tools; do not automatically install global tools.
- Once commit, push, or release closeout starts, do not repeat tests, lint, or verification commands. Reuse existing development evidence.
- The HTTP example binds to loopback by default and requires authentication. Remote access requires caller-configured TLS termination and network access controls.
- Read credentials from the runtime environment. Never commit tokens, cookies, private keys, or real environment configurations.
- Do not create empty future packages, fork the official SDK, or use unsafe/reflection to access SDK internals.
