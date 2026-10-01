# 架构

[English](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/EN-Architecture) · **中文** · [Wiki 首页](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/Home)

对应源文档：[`docs/architecture.md`](https://github.com/SisyphusSQ/mcp-extensions-go/blob/main/docs/architecture.md)。

## 目标与职责分层

为 OpenAI MCP Extensions 提供经过校验的 Go 服务端辅助，协议、schema、传输与会话继续由官方 MCP Go SDK v1.8.0 负责。Go 下限为 1.25.0。不访问 SDK 内部，不使用 unsafe/反射技巧，不 fork SDK，也不建立第二套 JSON-RPC 实现。

```text
业务服务 / internal/example（HTTP 与 stdio）
  +-- settings：原生设置工具、capability、声明校验
  +-- mentions：app 可见的搜索工具、两种结果类型
  +-- resources：文件输入/上下文解析、可选已授权根目录读取器
  +-- ui：App 元数据、可信 HTML 注册
  +-- 官方 mcp.Server
       +-- schema、工具、资源、发现、会话与传输
       +-- 标准 elicitation 与标准 MRTR

examples/frontend
  +-- 标准 MCP Apps App
  +-- OpenAI TypeScript OpenAIExtensions
       +-- 宿主上下文、文件资源、模型上下文、消息与 deep link

测试
  +-- 官方内存 / 带认证 HTTP / 子进程 stdio 集成
  +-- 仅供测试的 SDK 边界复现与业务续轮流程
  +-- 官方 AppBridge、真实浏览器与 Go HTTP 服务
```

实施顺序为 Settings、Mentions、文件上下文、SDK 表单/MRTR 调查，再到真实前端和验证。[完整能力表](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/ZH-Compatibility) 逐项区分服务端、前端与宿主责任。

## Settings

`settings.NewServer` 是返回官方 `*mcp.Server` 的构造辅助。它先校验并快照静态字段/布局，再同时创建读取、更新工具及现代/旧版 capability。它保留其他 server options 与 capability，注册期间不调用存储。调用者正常注册其他 MCP 工具，并保留设置工具名称；官方 SDK 的同名替换语义仍然适用。

直接复用 SDK 已有依赖 `jsonschema-go v0.4.3`，没有修改版本。基础类型及字符串/数值约束生成完整值 schema 和非空局部 patch schema；官方类型化 `mcp.AddTool` 处理输入/输出 schema。返回成功前检查完整回调状态：每个字段必须有有效值，未知字段被拒绝，值单独快照。数值回调使用 Go 原生数值类型，因为该校验器把 `json.Number` 视为字符串。

Read 必须只读。Update 必须独立授权、保留未提供字段、检查跨字段规则并原子持久化。SDK 辅助不提供数据库、隐式重试、锁或租户模型。`internal/example.Store` 是单所有者示例代码，不是公共存储 API。HTTP 默认内存存储，运维配置绝对路径 `MCP_SETTINGS_FILE` 可启用文件存储；stdio 插件默认使用用户配置目录中的文件。

文件读取上限为 64 KiB，Unix 文件权限为 0600。操作系统锁覆盖重新读取、patch 和原子替换，获取锁支持取消且最多等待五秒。无效或不可读状态明确失败。它支持正常进程重启，不宣称断电耐久性或按身份隔离的多用户事务。文件后端支持 Windows 和源码列出的 flock Unix 平台；其他平台明确拒绝文件后端，内存存储仍可用。

## Mentions

辅助注册普通工具，并添加 `openai/extensions.mentions/search` 与 app visibility。已有 ui/extension 字段保留在独立快照中。搜索由调用者提供，空查询有效。结果使用官方 `mcp.ResourceLink` 编码，或上游 SDK 源码中的 `resource` 类型；无效结果类型成为工具错误，空结果编码为 `items: []`。上游没有要求单独的 mentions 服务端 capability，不能自行虚构。

## 文件上下文

`FileInput` 包含文件名和不透明宿主资源 URI。App 使用官方 TypeScript 资源 API 读取该 URI，Go 不将它映射为文件系统路径。路径、表示、writable/etag 元数据解析器允许未来无关字段，但拒绝格式错误的已知字段。

可选 `resources.Reader` 不执行认证或逐资源授权。调用者必须配置绝对允许根目录和正数大小上限。`os.Root` 在打开期间保证根目录约束，允许根目录内的相对符号链接并拒绝逃逸。Unix 非阻塞打开与普通文件检查拒绝 FIFO、设备和目录；stat 大小检查及限量读取处理检查后文件增长，I/O 前后检查取消。

可信根目录必须排除恶意 mount/硬链接；路径按调用者配置的根目录命名空间解释，包括 macOS `/var` 别名。HTTP 示例只读取运维明确选择的可信 App 构建文件，从不读取宿主传入的任意路径。

## UI 与前端

`ui` 保留元数据快照/校验与可信静态资源 API。工具绑定 `ui://` 资源。可见性、入口和显示模式只是声明，不代表授权或宿主验收。

浏览器示例将官方 App 和 OpenAI 扩展打包为带脚本 hash CSP 的可信 HTML。先安装处理器再连接；渲染首次工具结果时不重复调用启动工具；不支持的宿主能力禁用对应动作。用户主动触发的模型上下文和消息动作通过宿主执行，没有自建 Go 宿主桥接。文件、错误与结果以文本 DOM 输出，Go 凭据留在服务端。持久设置初始化控件但不替换首次结果；Save 只发送变化字段，保留其他 App 实例的不相关更新。

本地插件打包官方 stdio 可执行文件和 App，仓库 marketplace 提供一个本地插件。使用 `.codex-plugin/plugin.json` 与 `.mcp.json` 兼容格式：已安装 CLI 能识别便携 root manifest，却没有加载对应 MCP 服务，兼容格式解决了加载问题。官方 CLI 安装生成缓存副本并启用插件。stdio 父进程控制访问，HTTP 仍强制认证。未创建网络监听、daemon、凭据或公开插件目录条目。详见[本地插件与验收](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/ZH-Live-E2E)。

测试桥接使用官方 AppBridge/PostMessageTransport、隔离的 headless Chrome profile、受限同源代理和临时 Go Bearer 凭据。上下文、模型消息与文件宿主行为来自测试数据，这是本机集成证据，不是 OpenAI 产品验收。

## 表单与 MRTR 决策

官方 SDK 实现标准 MRTR，本机已验证真实多轮调用。业务续轮的所有权、过期、重放和事务仍由应用负责。

旧版 OpenAI 自定义服务端发出 elicitation 受公开发送方法注册表限制。类型化 `InputRequestMap` 无法编码/解码 OpenAI 方法。接收中间件配合 ResultBase 可以通过官方 HTTP 输出适配后的扩展结果，不能将所有方向都说成不可实现。

该路径仍是仅供测试的实验，没有真实扩展宿主、完整扩展字段/选择校验或生产续轮契约。未添加公共占位 API 或不支持的 capability。具体复现、SDK 源码位置和需要的公开接口见[协议调查](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/ZH-Protocol-Investigation)。

## 安全与生命周期

HTTP 示例保留运行时凭据、常量时间摘要比较、loopback 默认地址、跨 Origin/localhost 防护、1 MiB 请求限制、超时与优雅退出。库处理器接收官方 request/context，所属服务可注入已验证身份与授权。回调错误不能包含凭据；文件系统错误可能含路径，业务工具须适当映射。

源码仓库已按所有者在 2026-10-01 的明确要求公开。仓库可见性不改变运行时认证、文件授权或本地插件分发。未安装全局工具、持久服务，未部署远程 TLS，未发布 tag/release。Go 1.25 矩阵使用缓存内 `GOTOOLCHAIN`；Node 依赖安装在前端示例本地，锁定版本并禁用生命周期脚本。
