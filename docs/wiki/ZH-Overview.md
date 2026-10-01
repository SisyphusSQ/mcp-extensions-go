# 项目概览

[English](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/EN-Overview) · **中文** · [Wiki 首页](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/Home)

这是源码公开的个人 Go 服务端 SDK，不是 OpenAI 官方 SDK。它复用 `github.com/modelcontextprotocol/go-sdk v1.8.0`，Go 下限为 1.25.0。JSON-RPC、传输、会话、工具/资源和标准 MRTR 由官方 SDK 负责；浏览器与宿主行为复用标准 MCP Apps 和 OpenAI TypeScript App SDK。

状态记录日期：2026-10-01。服务端扩展已通过 [PR #1](https://github.com/SisyphusSQ/mcp-extensions-go/pull/1) 合并。本地插件与实际验收支持已通过 [PR #2](https://github.com/SisyphusSQ/mcp-extensions-go/pull/2) 合并，提交为 `4ce5a1f73d1755a72d5b669b7a17c16bfd8721b6`。中英文 Wiki 源文档已通过 [PR #3](https://github.com/SisyphusSQ/mcp-extensions-go/pull/3) 合并；随后所有者明确授权将源码仓库公开。未发布 tag/release，插件也尚未上架公开目录。合并不等于所有宿主验收已经完成。

## 已实现

| 能力 | 实现与边界 |
| --- | --- |
| 原生设置 | 真实读写工具、现代/旧版 capability、基础类型 schema/约束与布局；授权和存储由调用者提供 |
| Mentions | 搜索类型、两种结果类型、app visibility 与元数据；搜索来源由业务提供 |
| 文件 | 文件入口输入、不透明资源引用、路径/表示/写入提示解析，以及供已授权文件使用的根目录与字节限量读取器 |
| UI | MCP Apps 绑定/可见性、global/thread/file/settings 入口、quickAction、显示模式元数据与可信 HTML 注册 |
| 可运行示例 | 带认证的 stateless HTTP 与官方 stdio，共用业务实现 |
| 本地插件 | MCP Extensions Go 已安装启用；单用户持久设置、App 仅保存变化字段、显示模式反馈 |
| 浏览器 App | 官方握手、设置/搜索、文件文本、deep link 上下文，以及宿主支持的模型上下文、消息和显示动作 |

Settings schema 目前覆盖明确的字段集合，尚未与 Pydantic 完全对齐。完整 OpenAI 扩展表单及其 MRTR 流程未实现。仅有定义或元数据时，不能宣称完整表单能力已经支持。

## 后续从哪里开始

[后续开发清单](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/ZH-Implementation-Roadmap) 给出了六项可直接开发的缺口：Settings 额外约束、表单 schema 与丰富选项、建议值、资源选择器、回答/上传引用校验，以及 Go 模型与 schema 绑定。每项都有范围、依赖、验收要求和源码入口，方便逐项实施。

现代扩展表单 MRTR 需要继续验证完整 SDK/宿主链路；旧式自定义表单请求则受当前官方 Go SDK 公开发送接口限制。这两类事项与“代码已经实现、宿主尚未验收”分开记录在[验收状态与限制](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/ZH-Validation-and-SDK-Limits)。

## 运行与文档

```sh
# 使用现有 Go/Node 工具与锁定的本地依赖：
npm --prefix examples/frontend ci --ignore-scripts
make plugin
codex plugin marketplace add /absolute/path/to/mcp-extensions-go
codex plugin add mcp-extensions-go@mcp-extensions-go-local
```

在桌面侧栏更多菜单打开 **Open workspace**。安装更新后重新打开；已有页面/进程可能仍保留旧文档。stdio 插件不监听网络端口，单用户设置写入用户配置目录；它不是多用户生产存储方案。

源码文档：[中文 README](https://github.com/SisyphusSQ/mcp-extensions-go/blob/main/README_ZH.md)、[架构](https://github.com/SisyphusSQ/mcp-extensions-go/blob/main/docs/architecture.md)、[完整能力对应表](https://github.com/SisyphusSQ/mcp-extensions-go/blob/main/docs/compatibility.md)、[SDK 调查](https://github.com/SisyphusSQ/mcp-extensions-go/blob/main/docs/protocol-investigation.md)、[本地插件验收](https://github.com/SisyphusSQ/mcp-extensions-go/blob/main/docs/live-e2e.md)和[交接记录](https://github.com/SisyphusSQ/mcp-extensions-go/blob/main/docs/handoff.md)。
