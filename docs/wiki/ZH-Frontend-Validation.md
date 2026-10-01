# 前端示例与验证边界

[English](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/EN-Frontend-Validation) · **中文** · [Wiki 首页](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/Home)

对应源文档：[`docs/frontend-validation.md`](https://github.com/SisyphusSQ/mcp-extensions-go/blob/main/docs/frontend-validation.md)。

## 配合 Go 服务端运行

示例直接使用 `@modelcontextprotocol/ext-apps 1.7.5` 和 `@openai/mcp-extensions 0.1.0`，Node 依赖在本地且版本锁定。浏览器/服务端凭据从不打包进 HTML。HTTP 默认内嵌页仍为静态页，使用真实 App 需要明确选择构建产物：

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

`MCP_APP_HTML` 必须是开发者信任的 HTML 绝对路径。服务启动时从明确允许的目录读取，大小上限为 2 MiB。这是运维设置，不能来自宿主 `openai/resource.path`。不会安装后台服务，Ctrl+C 关闭服务。构建 HTML 被 Git 忽略，需要从锁定前端依赖重新构建。

让支持 MCP Apps 的宿主连接带认证 Streamable HTTP：`http://127.0.0.1:8080/mcp`，再启动 `open_workspace` 或对应入口。App 在 `app.connect()` 前注册输入、结果和宿主上下文监听；渲染首次结果时不重复启动调用；读取/更新设置、搜索可读零件；仅在宿主声明支持时提供明确的模型上下文、消息和 fullscreen 动作。

`open_file` 通过宿主资源 API 读取不透明文件引用并显示文本，不要求 Go 打开任意路径。错误以文本 DOM 输出。在宿主外打开 HTML 时显示提示并禁用动作。

HTTP 默认一个运行时 Bearer 凭据对应一份受 mutex 保护的内存记录；运维配置绝对路径 `MCP_SETTINGS_FILE` 可启用共用持久示例 Store。本地 stdio 插件默认持久化。操作系统锁与原子替换保留跨进程不相关的局部 patch。App 初始化读取持久控件值，Save 只发送变化字段。多用户服务仍须自行提供按已验证身份隔离的存储和授权。

## 本机自动化集成

`browser-test.mjs` 在临时 loopback 端口启动已构建 Go 程序，生成临时凭据，连接官方 TypeScript MCP client，读取真实 Go HTML 资源，再在 sandbox iframe 中加载。测试宿主使用官方 `AppBridge` 和 `PostMessageTransport`；本地代理限制方法/资源、检查同源 JSON 请求，把 Bearer 凭据留在浏览器外。运行结束关闭全部监听、服务、MCP client、浏览器与子进程。

以下为后续开发阶段的运行入口；提交收尾不重复运行：

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

没有 `PLAYWRIGHT_MODULE_PATH` 时，套件尝试调用者已有的 `playwright` 包，不自动安装 Playwright 或下载浏览器。sqmc04 使用已有 bundled Playwright 1.62.1 和 `/Applications/Google Chrome.app/Contents/MacOS/Google Chrome`，配合隔离的 headless profile。

覆盖真实 `ui/initialize`、initialized 通知、首结果、带认证 Go Settings 读写、Mention 搜索、显示/deep link 更新、上下文/消息 payload、含 HTML 类文本但保持惰性的宿主资源、缺少 capability 及 standalone 模式。上下文/消息/资源响应属于测试宿主数据，不是 OpenAI 模型行为。

## OpenAI 宿主验收：部分由用户观察

2026-10-01，本地插件通过官方 CLI 安装启用。独立官方 app-server 发现 5 个工具和 3 个资源。用户桌面截图确认全局 Open workspace 入口、Connected、首结果、有效设置，以及上下文数量 1 和明确提问后模型识别 bolt 的回答。

用户反馈 Open fullscreen 没有明显反应；App 已改为显示实际模式，禁用不支持/已 fullscreen 的请求，并明确报告宿主保持其他模式的结果。重新打开后的真实桌面验收仍待完成。

原生桌面控制不可用：UI 工具拒绝 `com.openai.codex`，运行中的桌面没有默认 CLI control socket。代理没有重启桌面或发送聊天消息。本机桥接套件和 CLI 发现不能证明剩余桌面结果。安装与人工验收细节见[本地插件与验收](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/ZH-Live-E2E)。

继续人工验收并记录：

1. global/thread 启动、单次首结果、两个聊天/App 实例隔离，以及支持的显示模式。
2. 原生 Settings 发现/读取、分组布局/工具按钮、合法变化字段 patch、可见校验/持久化错误，以及重启后的存储行为。
3. composer Mention typeahead、空查询、选择、读取选中资源，以及拒绝/不可用搜索。
4. 文件扩展输入、宿主资源读取表示、大文件拒绝和不支持的能力。写入/订阅须先补独立示例，不能宣称已验收。
5. 模型上下文替换/移除/updateId、明确 active/new 聊天消息，以及 deep link 初始与后续导航。
6. 只有解决[协议调查](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/ZH-Protocol-Investigation)中的 SDK/宿主要求后，才验收扩展表单和真实 MRTR 续轮。

不能基于本机协议测试把这些场景标记为已完成。
