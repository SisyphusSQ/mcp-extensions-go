# 项目概览

[English](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/EN-Overview) · **中文** · [Wiki 首页](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/Home)

这是源码公开的个人 Go 服务端 SDK，不是 OpenAI 官方 SDK。它复用 `github.com/modelcontextprotocol/go-sdk v1.8.0`，Go 下限为 1.25.0。JSON-RPC、传输、会话、工具/资源和标准 MRTR 由官方 SDK 负责；浏览器与宿主行为复用标准 MCP Apps 和 OpenAI TypeScript App SDK。

历史发布记录：2026-10-01；当前工作状态更新于 2026-10-02。服务端扩展已通过 [PR #1](https://github.com/SisyphusSQ/mcp-extensions-go/pull/1) 合并。本地插件与实际验收支持已通过 [PR #2](https://github.com/SisyphusSQ/mcp-extensions-go/pull/2) 合并，提交为 `4ce5a1f73d1755a72d5b669b7a17c16bfd8721b6`。中英文 Wiki 源文档已通过 [PR #3](https://github.com/SisyphusSQ/mcp-extensions-go/pull/3) 合并；随后所有者明确授权将源码仓库公开。在该历史阶段未发布 tag/release，插件也未上架公开目录。首个 Go 模块版本为 [v0.0.1](https://github.com/SisyphusSQ/mcp-extensions-go/releases/tag/v0.0.1)，本地插件与模块发版分开。合并不等于所有宿主验收已经完成。

## Agent 接入

安装 `github.com/SisyphusSQ/mcp-extensions-go@v0.0.1`，从[编码 Agent 指南](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/ZH-Agent-Integration)开始。提供只依赖公共 API 的接入示例及可复用扩展验收程序。

## 已实现

| 能力 | 实现与边界 |
| --- | --- |
| 原生设置 | 真实读写工具、现代/旧版 capability、基础类型 schema/约束与布局；授权和存储由调用者提供 |
| Mentions | 搜索类型、两种结果类型、app visibility 与元数据；搜索来源由业务提供 |
| 文件 | 文件入口输入、不透明资源引用、路径/表示/写入提示解析 |
| 表单 | 平面 schema、注解和 JSON enum、丰富选择／建议值／资源引用、原始回答校验、类型模型 |
| UI | MCP Apps 绑定/可见性、global/thread/file/settings 入口、quickAction、显示模式元数据与可信 HTML 注册 |
| 可运行示例 | 带认证的 stateless HTTP 与官方 stdio，共用业务实现 |
| 本地插件 | MCP Extensions Go 已安装启用；单用户持久设置、App 仅保存变化字段、显示模式反馈 |
| 浏览器 App | 官方握手、设置/搜索、文件文本、deep link 上下文，以及宿主支持的模型上下文、消息和显示动作 |

Extensions 表单声明、校验与模型已实现；旧版直接发送受 SDK 限制，真实宿主 UX 未验收。范围不包含底层 MCP SDK 或通用状态恢复。详见[表单指南](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/ZH-Forms)。

## 后续从哪里开始

[对齐状态清单](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/ZH-Implementation-Roadmap) 记录已实现的 Settings 约束／校验、表单 schema 与丰富选项、建议值、资源选择、回答／上传引用校验和 Go 类型模型。剩余语言／运行时边界及真实宿主验收单独记录。

现代扩展表单 MRTR 和通用续轮恢复不属于本轮 Python Extensions 对齐；旧式自定义表单请求仍受当前官方 Go SDK 公开发送接口限制。这两类事项与“代码已经实现、宿主尚未验收”分开记录在[验收状态与限制](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/ZH-Validation-and-SDK-Limits)。

## 运行与文档

```sh
# 使用现有 Go/Node 工具与锁定的本地依赖：
npm --prefix examples/frontend ci --ignore-scripts
make plugin
codex plugin marketplace add /absolute/path/to/mcp-extensions-go
codex plugin add mcp-extensions-go@mcp-extensions-go-local
```

在桌面侧栏更多菜单打开 **Open workspace**。安装更新后重新打开；已有页面/进程可能仍保留旧文档。stdio 插件不监听网络端口，单用户设置写入用户配置目录；它不是多用户生产存储方案。

专题文档：[架构](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/ZH-Architecture)、[完整能力对应表](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/ZH-Compatibility)、[前端验证](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/ZH-Frontend-Validation)、[SDK 调查](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/ZH-Protocol-Investigation)和[本地插件验收](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/ZH-Live-E2E)。源码[中文 README](https://github.com/SisyphusSQ/mcp-extensions-go/blob/main/README_ZH.md)仍是模块使用入口。

Settings 还提供 format／exclusive 边界、保存前字段校验、业务别名与类型模型；局部更新和调用者事务语义保留。
