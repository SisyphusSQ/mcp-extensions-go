# mcp-extensions-go

[English](README.md)

基于官方 [MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk) 的个人扩展库，用 Go 表达 [OpenAI MCP Extensions](https://github.com/openai/mcp-extensions) 的服务端能力。当前交付是 SDK 框架，仓库保持私有，尚未发布版本。

协议、工具、资源、JSON Schema、会话和 HTTP 传输由官方 SDK 负责；本仓库在其上增加扩展元数据和注册辅助。它不是 OpenAI 官方 Go 实现。

## 当前能力

- `ui.ToolMetadata.Metadata`：生成 `_meta.ui.resourceUri`、工具可见性和 `_meta["openai/ui"]`，包含 global/thread/file/settings 入口、快捷动作和模型显示模式。
- `ui.ResourceMetadata.Metadata`：生成资源内容的显示模式元数据，保留调用方提供的标准 MCP Apps CSP 等声明。
- `ui.AddHTMLResource`：把可信静态 HTML 注册为 `text/html;profile=mcp-app` 资源。
- `examples/http`：带 Bearer 认证的 Streamable HTTP 服务，提供 `open_workspace` 工具及其 HTML 资源。

完整表单、MRTR、mentions、文件上下文、settings 能力协商和浏览器端 SDK 尚未实现。准确边界见 [兼容性表](docs/compatibility.md)。

## 运行示例

要求 Go ≥ 1.25.0；官方 SDK 固定为 v1.8.0。示例不需要 Node.js 或前端依赖安装。

```sh
cd /Users/suqing/coding/golang/00_self/mcp-extensions-go
export MCP_BEARER_TOKEN="$(openssl rand -hex 32)"
go run ./examples/http
```

服务默认监听 `127.0.0.1:8080`，MCP 地址为 `http://127.0.0.1:8080/mcp`。使用 Ctrl+C 停止。`MCP_BEARER_TOKEN` 是必须项，至少 32 字节且无空白；不要打印、提交或复制到文档中。

需要私有网络访问时通过 `MCP_LISTEN_ADDR` 指定实际地址；远程客户端应通过配置好的 HTTPS 入口接入，保留认证和网络访问控制。该示例不配置 TLS 证书、反向代理或后台常驻服务。

宿主配置使用 Streamable HTTP，并在认证配置中设置 `Authorization: Bearer <运行环境中的凭据>`。不要把带有真实凭据的配置提交到 Git。单独在浏览器打开 `/mcp` 不会显示网页；HTML 通过 MCP `resources/read` 提供。

示例 HTML 是不含脚本的静态页面；未集成浏览器端 `App` 桥接或 `ui/initialize` 握手，不能据此认为 Codex/ChatGPT 宿主已验收。

## 在自己的服务中使用

```go
meta, err := (ui.ToolMetadata{
    ResourceURI: "ui://my-app/home.html",
    Entrypoints: []ui.Entrypoint{{Type: ui.Global}, {Type: ui.Thread}},
    PreferredModelDisplayMode: ui.Inline,
}).Metadata(nil)
if err != nil {
    return err
}
tool := &mcp.Tool{Name: "open_app", Meta: meta}
// Register your typed business handler with mcp.AddTool(server, tool, handler).
mcp.AddTool(server, tool, handler)

err = ui.AddHTMLResource(server, &mcp.Resource{
    URI: "ui://my-app/home.html", Name: "home",
}, trustedHTML, ui.ResourceMetadata{
    AvailableDisplayModes: []ui.DisplayMode{ui.Inline, ui.Fullscreen},
    PreferredDisplayMode: ui.Inline,
})
if err != nil {
    return err
}
```

导入路径为 `github.com/SisyphusSQ/mcp-extensions-go/ui`。私有模块的其他使用者需要 GitHub 仓库访问权限，并在自己的 Go 环境配置相应 `GOPRIVATE` 与 Git 认证；本仓库不会更改全局 Go 配置。

`ToolMetadata.Metadata` 保留其他顶层键，但替换 `ui` 和 `openai/ui`；`ResourceMetadata.Metadata` 只替换 `openai/ui`。返回值是 JSON 快照，不保留调用方嵌套 map/slice 的别名。元数据表达 UI 声明，不能替代服务端认证、授权或能力协商。

## 开发与交接

```sh
make fmt
make test vet build
```

编译产物为 `bin/mcp-extensions-http`，不进入 Git。测试使用官方 SDK 的内存传输和本机临时 HTTP 端口，不访问外部业务服务。示例没有业务持久化或后台安装行为。

- [文档入口](docs/README.md)
- [架构和后续扩展路线](docs/architecture.md)
- [兼容性与实现范围](docs/compatibility.md)
- [本次交接](docs/handoff.md)
