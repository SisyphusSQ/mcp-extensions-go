# mcp-extensions-go

[English](README.md)

基于官方 [MCP Go SDK v1.8.0](https://github.com/modelcontextprotocol/go-sdk) 的个人 Go 扩展库，用于实现 [OpenAI MCP Extensions](https://github.com/openai/mcp-extensions) 的服务端职责。仓库保持私有，不是 OpenAI 官方 Go SDK，尚未发布版本或 tag。

JSON-RPC、schema、工具、资源、发现、会话、传输和标准 MRTR 由官方 SDK 负责。本库增加服务端扩展类型与辅助；浏览器和宿主行为复用官方 TypeScript App SDK。

## 已实现

- `settings.NewServer`：真实读取/更新工具、原生设置类型与约束、分组布局、完整有效值、非空局部 patch，以及现代/旧版 `openai/settings` 能力发现。返回普通的官方 `*mcp.Server`。
- `mentions.AddTool`：搜索输入、标准资源链接与上游 SDK 的 resource 结果类型，`mentions/search` 元数据、只读声明和必需的 app 可见性。
- `resources`：文件入口输入、宿主不透明资源引用，`openai/resource` 路径/表示/写入提示解析，以及供已授权本地文件使用的 `os.Root` 限量读取器。
- `ui`：标准 MCP Apps 工具资源绑定和可见性、OpenAI 入口、快捷动作与显示模式元数据、可信 HTML 注册。
- `examples/http`：带 Bearer 认证的 stateless Streamable HTTP，包含真实内存设置、可搜索/读取的示例零件、工作区与文件入口工具、loopback 默认地址、请求限制与优雅退出。
- `examples/frontend`：使用标准 MCP Apps `App` 和 OpenAI TypeScript 扩展的浏览器 App，并通过官方 `AppBridge` 进行本机浏览器集成测试。

设置存储、资源授权、跨字段规则和事务由调用者负责。可见性与元数据不授予权限。示例只有一个运行时凭据，共用带锁的内存记录，重启后恢复默认值。

完整 OpenAI 扩展表单尚未实现。Go SDK v1.8.0 已支持标准 MRTR，但类型化输入映射和自定义发出请求存在具体限制；接收中间件可以输出扩展结果字段，单凭这一点不能声明扩展 MRTR 已实现。详见 [完整能力对应表](docs/compatibility.md) 与 [可复现的 SDK 调查](docs/protocol-investigation.md)。真实 OpenAI 宿主仍未验收。

## 运行 Go 示例

要求 Go ≥ 1.25.0。默认静态页面不需要 Node 依赖。

```sh
cd /Users/suqing/coding/golang/00_self/mcp-extensions-go
export MCP_BEARER_TOKEN="$(openssl rand -hex 32)"
go run ./examples/http
```

端点默认是 `http://127.0.0.1:8080/mcp`。在宿主中配置 Streamable HTTP 和 `Authorization: Bearer <运行时凭据>`。Token 至少 32 字节且无空白，不能打印或提交。Ctrl+C 退出。直接用浏览器打开 `/mcp` 不会显示 HTML；App 资源通过 MCP `resources/read` 提供。

使用真实浏览器 App 时，先构建 `examples/frontend`，再将 `MCP_APP_HTML` 设为可信 `dist/app.html` 的绝对路径，完整步骤见 [前端设置与验证](docs/frontend-validation.md)。未设置时仍使用不包含 `ui/initialize` 握手的静态页面。

`MCP_LISTEN_ADDR` 用于选择明确的地址。远程访问需自行配置 HTTPS 入口、认证和网络访问控制；示例不配置 TLS 或安装后台服务。

## 在业务服务中使用 Settings

```go
server, err := settings.NewServer(
    &mcp.Implementation{Name: "my-server", Version: "1"}, nil,
    settings.Config{
        Fields: map[string]settings.Field{
            "units": {Type: "string", Title: "Units", Enum: []string{"mm", "in"}},
        },
        Read: func(ctx context.Context, req *mcp.CallToolRequest) (settings.Values, error) {
            // Authorize using verified request identity and return all effective values.
            return store.Read(ctx, req)
        },
        Update: func(ctx context.Context, req *mcp.CallToolRequest, set settings.Values) (settings.Values, error) {
            // Authorize, preserve omitted fields, and persist atomically before returning.
            return store.Update(ctx, req, set)
        },
    },
)
if err != nil {
    return err
}
// Register other business tools normally. Reserve settings.read/settings.update.
mcp.AddTool(server, tool, handler)
```

导入 `github.com/SisyphusSQ/mcp-extensions-go/settings` 与官方 `mcp` 包。工厂在启动时生成定义快照并校验，不调用存储；两个工具和对应能力一起注册。完整回调见 HTTP 示例。参数错误、普通存储错误和返回值校验错误使用 MCP 工具错误表达。回调使用 Go 原生数值类型，错误消息须适合给客户端显示。

Mentions 使用 `mentions.AddTool(server, &mcp.Tool{Name: "search_mentions"}, searchHandler)`；空查询有效。辅助保留其他元数据并保证 app 可见性。上游未要求单独的 mentions 服务端能力。

UI 使用 `ui.ToolMetadata.Metadata`、`ui.AddHTMLResource` 和官方工具/资源。元数据快照保留 JSON 数字精度且不持有调用方的可变别名。`resources.Path(req.Params.Meta)` 只解析文件上下文；调用限于允许根目录的 `Reader` 前，必须独立认证和授权，不能直接打开任意宿主路径或不透明资源 URI。

其他私有模块使用者需要仓库权限、Git 认证和自己的 `GOPRIVATE` 配置；本项目不更改全局 Go 设置。

## 开发

```sh
make fmt
make test vet build
# 使用 Go 自带机制下载到缓存并验证版本下限：
GOTOOLCHAIN=go1.25.0 make test vet build
```

`make test` 包括 race、官方内存/HTTP 传输、Settings/Mentions/文件行为及 SDK 边界复现。产物为不进入 Git 的 `bin/mcp-extensions-http`。前端构建、类型检查和浏览器测试见专门文档。sqmc04 已通过 Go 1.25.0 与 Go 1.27.0 验证；前端依赖安装的 npm audit 报告为零漏洞。这不代表 Go 完整漏洞可达性扫描、远程部署或 OpenAI 宿主验收。

- [文档入口](docs/README.md)
- [架构](docs/architecture.md)
- [完整能力与兼容性对应表](docs/compatibility.md)
- [表单/MRTR 调查](docs/protocol-investigation.md)
- [前端与宿主验收清单](docs/frontend-validation.md)
- [交接记录](docs/handoff.md)
