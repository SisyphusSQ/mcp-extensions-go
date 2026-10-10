# Agent 接入指南

[English](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/EN-Agent-Integration) · **中文** · [Wiki 首页](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/Home)

面向在其他 Go 项目接入 SDK 的编码 agent。先选择所需能力，再按需查阅 API。仓库维护规则保留在 [AGENTS.md](https://github.com/SisyphusSQ/mcp-extensions-go/blob/v0.0.2/AGENTS.md)，英文权威说明为 [docs/agent-integration.md](https://github.com/SisyphusSQ/mcp-extensions-go/blob/v0.0.2/docs/agent-integration.md)。

## 1. 检查项目与选择能力

先读目标项目规则、go.mod、服务构造处、传输、认证和存储回调。要求 Go 1.25.0 或更新版本，保留已有官方 MCP 集成及运行时凭据。在业务项目安装：

```sh
go get github.com/SisyphusSQ/mcp-extensions-go@v0.0.2
```

这是个人扩展库，不是 OpenAI 官方 SDK，使用官方 `github.com/modelcontextprotocol/go-sdk v1.8.0`。已有其他 SDK 版本时先评估，不静默升降级；公开模块不需要全局 Go 配置或 GOPRIVATE。

| 需求 | 公共 API | 业务／宿主责任 |
| --- | --- | --- |
| 原生设置、局部更新 | `settings.NewServer`、`settings.NewModelServer[T]` | 身份授权、合并状态规则和事务存储 |
| Composer 资源建议 | `mentions.AddTool`、`mentions.SearchResult` | 授权搜索、可读资源和宿主 UI |
| App 资源／入口 | `ui.ToolMetadata.Metadata`、`ui.AddHTMLResource` | 可信 HTML、官方 TypeScript 握手和呈现 |
| 文件入口／上下文 | `resources.FileInput`、`resources.Path` | 宿主读取及独立授权，元数据不授予访问权 |
| 表单声明／回答校验 | `forms.New`、`forms.Parse`、`forms.NewModel[T]` | 取得回答并执行业务规则 |
| 命名类型 schema 推导 | `forms.NewModelWithOptions[T]`、`settings.FieldsForWithOptions[T]` | 公共 `jsonschema.ForOptions` 类型声明 |
| 标准工具、会话、传输、MRTR | 官方 `mcp` API | 应用现有 MCP 集成 |

**现代表单：** Form.RequestInput 提供公共现代 MRTR 适配，要求协议至少 2026-07-28 及两种 form capability。pending 非 nil 立即返回；否则处理取消／拒绝或 Decode 回答。旧式 `openai/elicitation/create` 受当前官方 SDK 公开发送接口限制。声明和校验不等于能显示表单，承诺前查[兼容性矩阵](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/ZH-Compatibility)。

## 2. 新建或接入已有服务

需要 Settings 时，在现有构造位置使用一个设置工厂，保留 implementation／ServerOptions，再在返回的官方 `*mcp.Server` 注册已有工具和资源。没有 `AddSettings(existingServer)`；保留 settings.read／settings.update 或配置的工具名称。其他扩展可保留 mcp.NewServer 后追加辅助，不新建协议、会话或传输。工厂同时声明现代和旧版设置能力，不手写宿主 capability。

[完整接入示例](https://github.com/SisyphusSQ/mcp-extensions-go/blob/v0.0.2/examples/agent-quickstart/main.go)仅使用公共 API，演示格式、转换、别名、局部更新和并发。在本仓库构建：

```sh
go build -o bin/agent-quickstart ./examples/agent-quickstart
```

MCP 客户端运行二进制绝对路径；程序等待 stdin，终端本身不是客户端。stdout 仅协议，stderr 输出错误。该示例单用户、内存状态、重启重置，不监听端口，不提供生产身份或持久存储。用于其他项目时复制到业务 command 包，安装固定版本并替换回调，不导入 internal/example。

已有 HTTP 服务保留路由和认证，通过官方 mcp.NewStreamableHTTPHandler 暴露返回实例。[HTTP 示例](https://github.com/SisyphusSQ/mcp-extensions-go/blob/v0.0.2/examples/http/main.go)包含运行时 Bearer、loopback、跨来源防护、请求限制、超时和优雅退出。移植到业务自己的包时保留这些边界；远程 TLS／网络控制由业务配置。

## 3. 正确实现 Settings 回调

Read 返回所有有效值，包括 false／0。Update 接收非空局部替换 map，保留缺省值；授权、合并状态校验和保存放在业务事务内。返回值校验不能撤销已完成写入。

类型化声明／结果使用 JSON 别名 display_label，更新回调收到 Go 名称 Label。普通 NewServer 默认 wire 名称，可配置 FieldNames。FieldValidators 按 wire 名定义，仅对提供字段执行：先校验 wire，再转换，再校验转换结果，然后调用 Update。不自动复现 Python before/plain/wrap，也不自动执行业务模型对读取／保存结果的再验证。

格式支持 email／uri／date／date-time，加已有字符串和数值约束。Settings 为非 null 平面基础类型，不支持数组、嵌套或 schema 默认值；有效默认值由 Read／存储提供。详见[表单和设置](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/ZH-Forms)。

## 4. 按需追加扩展

**Forms：** 英文指南提供完整 DecodeSelection 函数和公开 imports，演示 JSON 别名、bolt／washer 丰富选项和 Decode。调用者必须已取得回答，函数不发送请求。取消／拒绝按业务需要处理。ValidateResult 不修改提交，默认值只影响类型化结果。UserOptions 允许宿主选择，SelectionPolicy 可额外限制；PrepareSubmission／CompleteSubmission 校验引用和数量，不负责字节、权限或上传服务。写入前验证完整回答。

平面基础类型／字符串数组可用，嵌套和任意 union 不支持。正则为 RE2，邮箱不是完整 Python email-validator／IDNA；独立整数限 signed／unsigned 64 位，SDK 普通 map 解码可能已丢失表示／精度，扩展不能恢复。

**Mentions：** 用 `mentions.AddTool(server, &mcp.Tool{Name: "search_mentions"}, searchHandler)`，回调为 `func(context.Context, *mcp.CallToolRequest, mentions.SearchParams) (mentions.SearchResult, error)`。返回 `mentions.Item{Link: &mcp.ResourceLink{...}}` 或已记录 resource 变体。空查询有效，无匹配返回空 Items slice。构造前用 mentions.WithCapability 配置 openai/mentions，将返回 options 传给官方／设置工厂，再注册同名 AddTool。辅助保留废弃 marker、readOnlyHint 和 app visibility；授权和可读资源注册仍由业务负责。[共享源码](https://github.com/SisyphusSQ/mcp-extensions-go/blob/v0.0.2/internal/example/server.go)仅供参考，不是公共 import。

**App：** 检查 ui.ToolMetadata.Metadata 错误并把资源绑定／visibility／入口放到真实工具，使用 ui.AddHTMLResource 注册可信构建 HTML。非文件入口接收 `{}`。按[前端指南](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/ZH-Frontend-Validation)和[App 源码](https://github.com/SisyphusSQ/mcp-extensions-go/blob/v0.0.2/examples/frontend/app.ts)在连接前注册监听，完成官方初始化、首结果和 capability gating。静态 HTML 不能代替握手；上下文、消息、deep link 使用官方 TypeScript SDK。

**Files：** 使用 resources.FileInput 并调用 Validate，宿主支持时通过前端官方资源 API 读取。resources.Path(req.Params.Meta) 仅解析可选上下文，没有独立授权和受控文件系统边界时不打开路径／opaque URI。私有示例 Reader 不是公共 API。

## 5. 验证与交付

开发时遵守目标仓库；本库入口 make test vet build。收尾规则要求复用证据时不重复测试。仓库根目录可运行[验收程序](https://github.com/SisyphusSQ/mcp-extensions-go/blob/v0.0.2/examples/acceptance/main.go)：

```sh
go run ./examples/acceptance
```

预期六行 PASS、退出零，186 次 MCP 调用覆盖 Python 对照、完整回答、模型和设置工具错误。官方 CommandTransport 启动真实子进程，使用临时内存状态并保留原回答 JSON；不监听端口、不读取选择文件、不改变已安装 Workspace。[2026-10-02 记录](https://github.com/SisyphusSQ/mcp-extensions-go/blob/v0.0.2/docs/validation-2026-10-02.md)是已有证据，不表示发版重新运行。

业务项目仍需实际初始化／发现、合法与非法工具调用、缺省字段、完整结果、真实存储和授权。临时程序不代替生产验收；UI 的宿主初始化／交互与浏览器 fixture／协议调用分别记录。真实 picker／upload／preview 仍未验收。

交付报告所选 API、依赖版本、构造位置、业务责任、复现命令和结果、剩余 SDK／宿主边界。构建或 fixture 通过不等于完整 Python 等价或产品验收。

## 可复制的任务提示

> 使用 mcp-extensions-go v0.0.2 在当前 Go 项目实现所需 MCP Extensions。先读目标仓库规则与本指南，再查能力矩阵及所需 API。复用认证、存储和官方传输；保留局部更新，写入前验证合并状态，独立授权资源。需要现代表单时使用公共 RequestInput，不编造旧式发送、通用恢复或私有示例 API。完成所需链路，报告实际证据并区分后端与宿主验收。
