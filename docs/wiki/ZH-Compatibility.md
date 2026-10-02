# 兼容性与完整能力表

[English](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/EN-Compatibility) · **中文** · [Wiki 首页](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/Home)

对应源文档：[`docs/compatibility.md`](https://github.com/SisyphusSQ/mcp-extensions-go/blob/main/docs/compatibility.md)。实现状态以该源文档为准。

## 已核对的上游基准

2026-10-01，GitHub commits API 与完整源码 checkout 均确认上游 `main` 为 [`900032d8bd7c1566202d0cb1666986584f932043`](https://github.com/openai/mcp-extensions/tree/900032d8bd7c1566202d0cb1666986584f932043)，与要求的基准没有 diff。下表重新依据[协议](https://github.com/openai/mcp-extensions/blob/900032d8bd7c1566202d0cb1666986584f932043/docs/spec.md)、TypeScript server/app/shared 和 Python 源码建立；此前交接文档的推断未当作当前 SDK 行为证据。

官方 MCP Go SDK 仍为 v1.8.0，Go 下限 1.25.0，`golang.org/x/sys v0.44.0` 保持固定以修复 GO-2026-5024。本项目是个人服务端扩展库，不是 OpenAI 官方 SDK。

## 完整职责对应表

“协议测试”指本机测试，不等于 OpenAI 宿主验收。宿主支持随平台变化；上游 Work 浏览器支持表明确不包含 classic ChatGPT。

| 上游能力 | Go 服务端职责 | 前端/宿主职责 | 当前实现 | 缺口/限制 | 验收入口 |
| --- | --- | --- | --- | --- | --- |
| 标准 MCP 工具、schema、资源、发现、会话、HTTP | 注册业务处理器 | 通过 MCP 连接 | 官方 SDK v1.8.0 | 不实现第二套协议栈 | 官方内存/HTTP 集成测试 |
| MCP Apps 资源绑定和 visibility | `_meta.ui.resourceUri` 与 app/model visibility | 加载 App，执行展示面可见性限制 | `ui.ToolMetadata` | 可见性不是授权 | `ui` wire 测试；已观察桌面全局启动，可见性限制仍待验收 |
| global 入口 | 工具接收 `{}`、title/icons、入口声明 | 侧栏、fullscreen、聊天实例 | 元数据与 `open_workspace` | 其他显示/实例场景待验收 | HTTP `{}` 调用；用户截图确认桌面侧栏启动 |
| global quickAction（SDK 源码） | 目标工具、参数、title/icons | 展示并调用动作 | `ui.QuickAction` | 属于 SDK 源码能力，spec 无独立 quickAction 章节 | wire 契约测试；宿主未验收 |
| thread 入口 | 工具接收 `{}` | 每个聊天独立 App 实例 | 元数据与示例 | 宿主实例隔离未验收 | HTTP 调用；宿主清单 |
| Settings App 入口（SDK 源码） | UI 元数据及可选自定义 App 工具 | Settings App 展示/搜索 | `ui.Settings` 与示例 | 与结构化设置分开 | 元数据测试；宿主未验收 |
| 结构化原生 Settings | 声明真实读写工具、schema/layout、有效值、授权和持久化 | 原生设置渲染、变化字段 patch 调用 | `settings.NewServer`，基础类型、约束、布局、现代/旧版 capability；单所有者持久 stdio 示例 | 业务存储/授权由调用者负责；原生设置 UX 待验收 | 现代 discover/旧版 initialize、局部更新与错误、HTTP/子进程 stdio 测试；截图显示 App 设置值 |
| Settings 工具按钮 | layout 引用同服务接收 `{}` 的工具 | spinner/tooltip 或 App modal | `settings.Item` 工具类型 | 业务工具由调用者提供，不能反查工具注册表 | 声明检查；宿主按钮 UX 未验收 |
| composer Mentions | 查询/结果类型、marker、app visibility、搜索授权 | 输入框 typeahead 与选中引用 | `mentions.AddTool`，`resource_link` 与 SDK `resource` 两种结果 | spec/SDK 不要求单独服务端 capability；平台 UX 未验收 | Mention 协议/错误/visibility 测试；HTTP 搜索与可读资源 |
| 文件扩展启动 | `FileInput {file:{name,resourceUri}}`、类型化工具、入口元数据 | 注入不透明 URI、传递初始输入 | `resources.FileInput`、`open_file` 示例 | 不透明 URI 从不当作路径 | 类型化 SDK 输入测试；真实文件启动未验收 |
| tools/call 的 `openai/resource.path` | 解析元数据，独立授权任何读取 | 附加本地执行路径 | `resources.Path`；读取器仅为私有示例 | 元数据不能证明宿主身份或权限 | 错误元数据、根目录/符号链接/大小/取消测试 |
| 宿主 resources/read 表示 | 对自有资源解析 text/blob 提示 | 拦截已打开文件不透明 URI 的读取 | `ParseReadMetadata`、TypeScript App 资源读取 | 宿主文件读取未验收 | 解析测试；本机 App bridge 与宿主清单 |
| 宿主资源写入提示 | 解析 writable/etag 供业务使用 | 执行写入、条件 ETag、saved/conflict/too-large 结果 | `ParseContentMetadata`；复用官方 TS App 资源 API | 没有 Go 写宿主 RPC，不虚构 writable capability | 解析测试；宿主写入未验收 |
| 资源订阅与写入 | 需要时实现普通自有 MCP 资源 | 拦截已打开文件 URI 的读取/订阅；`openai/resources/write` | 官方 SDK 与官方 TS API | 示例只是只读文件查看器 | 宿主订阅/写入清单未验收 |
| 资源显示模式 | content 的可用/首选模式、tool 的模型偏好 | 当前模式与切换 | `ui.ResourceMetadata`；App 声明 inline/fullscreen，显示实际模式并禁用不支持/已 fullscreen 请求 | 宿主可保持其他模式，App 明确报告 | wire/浏览器测试；桌面曾反馈按钮无明显反应，修复待重开验收 |
| MCP Apps CSP 与权限 | 声明资源安全元数据 | sandbox 与策略执行 | 元数据透传；App 有脚本 hash CSP | 业务域名/权限由调用者声明 | 静态资源测试、前端 CSP 浏览器测试 |
| App `ui/initialize`、工具输入/结果 | 提供可信打包 HTML 和工具 | 标准 App bridge、生命周期、宿主上下文 | 前端使用官方 App，由本地插件打包 | HTTP 默认静态页没有握手；生命周期/隔离待验收 | 本机浏览器握手；截图确认 Connected 和首结果 |
| deep link | 提供 App 入口，不在 Go 保存浏览器状态 | 解析 hostContext、后续变化、平台 URI 导航 | 官方 TS `OpenAIExtensions.deepLink` 示例 | 实际 codex/chatgpt URI 导航未验收 | 浏览器宿主上下文测试；宿主清单 |
| 模型上下文与 updateId | 提供业务工具/资源内容 | 按 App 替换上下文、幂等、移除、通知 | 官方 TS modelContext 示例 | 移除/幂等/实例隔离待宿主验收 | 本机 bridge；截图上下文数量 1，模型回答 bolt |
| 内容 title/thumbnail 与 background audience | 使用官方内容元数据/annotations | composer 附件标签/缩略图、assistant-only 后台上下文 | 标准 MCP 内容透传；App 的带 title 上下文 | 没有独立 Go model-context capability；iOS thumbnail 有限制 | 本机 payload 测试；宿主未验收 |
| `ui/message`、active/new 目标与 send | 返回适合明确 UI 动作的数据 | 宿主发送消息；移动端目标限制 | 官方 TS message API、明确 Send 按钮 | 其他目标/移动端/实例行为待验收 | 本机 payload 测试；截图确认桌面聊天面板提问与模型回复 |
| 本地打开文件 | 业务允许时提供已授权绝对路径 | `openai/files/open` 与原生查看器 | 前端依赖提供官方 TS API | 没有 Go 到宿主自定义 RPC，示例不暴露路径 | 仅宿主清单，未验收 |
| 交互光标（SDK 源码） | 无 | 宿主上下文与 CSS variable | 官方 TS OpenAIExtensions 构造器 | 宿主交互模式未验收 | 本机 App 初始化 |
| 插件 onboarding | 不属于服务端 SDK | manifest `com.openai.onboardingSkill` 与打包 skill | 本地 stdio 插件已安装；没有 onboarding skill | 不添加不必要 onboarding API/skill | 官方 CLI 安装/发现与全局入口截图；未声明 onboarding 流程 |
| 标准 elicitation 与 MRTR | 返回内置 inputRequests/requestState，校验业务续轮 | 完成标准输入并重试 | 官方 SDK、本机调查测试 | 状态所有权、去重、过期、隔离由业务负责 | `internal/sdkcheck` 标准多轮测试 |
| OpenAI 扩展表单 | schema、回答／资源校验、类型模型 | 选择、建议值、预览、上传 UI | forms 声明／解析／模型；JSON enum／注解、丰富选择、资源额外字段 | 直接发送受 SDK 阻碍；正则／邮箱／数值边界；宿主未验收 | Python 51 schema＋86 回答／上传对照，Go 模型／race 测试 |
| 旧版 `openai/elicitation/create` | 发出自定义服务端请求 | 渲染并返回 accept/cancel/decline | 官方类型化 Go SDK 无公开发送 API | 自定义接收注册不能增加自定义发送 | 发送中间件复现 |
| 扩展 MRTR 表单 inputRequests | 当前无公共扩展 API | 执行自定义请求及续轮 | 仅 SDK 调查测试 | 不属 Python Extensions 对齐；旧版 wrapper 不实现 MRTR | 公开边界复现；无宿主支持宣称 |

## 边界与错误语义

Settings 工厂返回 server 前拒绝无效基础类型声明、错误布局引用和重复 capability 配置；同时创建两个工具和两处 capability，保留其他 options 与无关 capability。调用者必须保留设置工具名称。官方 server API 不提供注册表查询，layout 引用工具存在且接收 `{}` 仍由调用者保证。

官方类型化工具辅助通过 `isError: true` 报告无效参数与普通 handler 错误。Settings 返回值校验在成功前拒绝缺字段、未知字段或无效状态。调用者明确返回 `*jsonrpc.Error` 时，保留官方协议错误语义。存储回调必须使用适合给客户端显示的错误，授权请求、保留未提供字段并原子持久化。数值使用 Go 原生类型；`json.Number` 不是 jsonschema-go value validator 理解的数值类型。Pattern 使用现有 Go JSON Schema 校验器的 RE2 语法。

文件解析不授权。私有 `internal/example.Reader` 需要明确的绝对允许根目录和字节限制，打开时使用 `os.Root`，仅允许根目录内符号链接和普通文件，I/O 前后检查取消。它不能逐资源授权、排除调用者根目录内恶意 mount/硬链接，或打断操作系统普通文件 syscall。调用者必须控制根目录并授权每次读取；示例从不依赖宿主路径打开文件。

浏览器能力只能由真实宿主声明。服务端不代替宿主声明 `openai/elicitation`、`openai/files`、`openai/modelContext`、`openai/message` 或资源权限，也不将任何 API 宣称为完整扩展表单。

本地插件已启用，独立官方 app-server 发现 5 个工具、3 个资源和两处设置 capability。用户截图确认全局入口、Connected、首结果及上下文辅助的消息/模型回答，这些观察没有完成剩余宿主矩阵。2026-10-01 源码仓库公开，不代表上架公开插件目录或新增 capability。参阅[插件验收](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/ZH-Live-E2E)和[前端验证](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/ZH-Frontend-Validation)。

本轮只对齐固定版本 Python MCP Extensions。Settings 已增加 format、exclusive 边界、FieldValidators、FieldNames、类型推导。普通回调默认 wire 名称，类型化更新默认 Go 导出业务字段名；返回值仍使用 JSON 别名。只在保存前校验提供的 patch 字段，转换后再次校验。Python before／wrap 装饰器不自动复现；跨字段规则属于业务事务。

通用请求状态加密／恢复、现代表单 MRTR 适配及两轮 review 示例已移除；本地读取器收回 internal/example。已有设置文件存储仍是业务示例。标准 MCP 会话、传输、tasks、EventStore、依赖框架不做对齐。

Go 使用 RE2，不宣称完整 Python email-validator／IDNA 等价；独立整数保留 64 位精度，官方 SDK 的 map 解码可能已损失原始表示或精度。Python 对照样例只证明覆盖场景一致，不等于完整 Pydantic 或真实宿主验收。SDK 直接发送与真实表单宿主验收仍是独立缺口。详见[表单指南](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/ZH-Forms)及[状态清单](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/ZH-Implementation-Roadmap)。Wiki 源修改和已安装插件尚未发布／重载。
