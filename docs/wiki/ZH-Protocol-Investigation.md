# 表单与 MRTR：已核对的公开 SDK 边界

[English](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/EN-Protocol-Investigation) · **中文** · [Wiki 首页](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/Home)

对应源文档：[`docs/protocol-investigation.md`](https://github.com/SisyphusSQ/mcp-extensions-go/blob/main/docs/protocol-investigation.md)。

## 来源与方法

OpenAI 协议和两种扩展实现在 commit `900032d8bd7c1566202d0cb1666986584f932043` 下读取。相关上游来源：

- [OpenAI 表单 elicitation 协议](https://github.com/openai/mcp-extensions/blob/900032d8bd7c1566202d0cb1666986584f932043/docs/spec.md#openai-form-elicitation)
- [TypeScript 请求实现](https://github.com/openai/mcp-extensions/blob/900032d8bd7c1566202d0cb1666986584f932043/typescript/src/server/forms/elicitation.ts)
- [TypeScript schema/回答校验](https://github.com/openai/mcp-extensions/blob/900032d8bd7c1566202d0cb1666986584f932043/typescript/src/server/forms/schema.ts)
- [资源选择器 schema](https://github.com/openai/mcp-extensions/blob/900032d8bd7c1566202d0cb1666986584f932043/typescript/src/server/forms/file-picker.ts)
- [Python 请求实现](https://github.com/openai/mcp-extensions/blob/900032d8bd7c1566202d0cb1666986584f932043/python/src/openai_mcp_extensions/form/_elicitation.py)

SDK 源码来自固定 v1.8.0 模块，`internal/sdkcheck/mrtr_test.go` 使用官方内存/HTTP 传输测试公开 API。没有复制或修改 SDK 源码、传输或会话实现。

## 标准 MRTR 可用

此前框架中的不确定性已有具体证据：v1.8.0 实现 SEP-2322 MRTR。[`mrtr.go`](https://github.com/modelcontextprotocol/go-sdk/blob/v1.8.0/mcp/mrtr.go) 安装客户端重试/输入执行和服务端旧版适配。`CallToolResult.InputRequests`、`RequestState` 为公开字段，SDK 将 `resultType` 设为 `input_required` 或 `complete`。

本机协议测试覆盖真实两轮工具调用、客户端自动执行输入、手动分次调用恢复、输入取消/拒绝、过期的首轮重复输入、完成重放、过期、8 个并发调用及跨连接隔离。

测试流程自行管理随机不透明续轮 registry、deadline、owner、轮次校验和完成缓存，这些业务策略**不是 SDK 保证**。测试状态按连接隔离，不是生产持久化 API；没有测试重启恢复，也没有证明带认证 stateless HTTP 的续轮隔离。生产续轮须绑定已验证身份/工具/参数，自行管理生命周期、重放和事务策略。

## 扩展表单是独立契约

上游使用 `openai/elicitation/create`，通过客户端 `extensions["openai/elicitation"].form` 协商。表单支持基础字段，以及 pattern、带描述选项、thumbnail、建议值、资源选择和预览。accepted 回答必须满足 schema 与允许的资源选择策略。

资源选择器是 `x-openai-input` 字段，不是 Go 文件系统选择器。单选为 URI 字符串，多选为 URI 数组；`resource` 是规范类型，`file` 为已弃用别名。selection mode 仅用于数组；implicit 禁止默认值且允许上传，explicit 默认值必须来自给定资源。用户文件/目录选择、上传与预览 UI 属于宿主；Go 校验引用，后续读取还要独立授权。

标准 `ServerSession.Elicit` 检查普通 elicitation capability，调用 `elicitation/create`。`RequestedSchema any` 可以序列化扩展关键词，但不能因此协商或调用 OpenAI 扩展。返回路径执行标准 schema 校验且可能应用默认值；OpenAI TypeScript validator 明确保持提交回答不变。二者不能等同。

## 已复现的限制与可用接口

| 公开 API/源码 | 实际结果 | 含义 |
| --- | --- | --- |
| [`protocol.go:52`](https://github.com/modelcontextprotocol/go-sdk/blob/v1.8.0/mcp/protocol.go#L52)，InputRequest 与 InputRequestMap.MarshalJSON | 接口实现封闭；嵌入 ElicitParams 可满足接口，但具体类型 switch 拒绝包装类型，报 `unsupported type` | 不能将 OpenAI 方法放入标准类型化 request map |
| [`protocol.go:102`](https://github.com/modelcontextprotocol/go-sdk/blob/v1.8.0/mcp/protocol.go#L102)，InputRequestMap.UnmarshalJSON | OpenAI 方法报 `unsupported InputRequest method` | 官方 Go client 的类型化重试流程不能执行扩展 MRTR 结果 |
| [`shared.go:139`](https://github.com/modelcontextprotocol/go-sdk/blob/v1.8.0/mcp/shared.go#L139)，默认发送 handler | 发送中间件替换 Elicit 方法后报 `JSON RPC not handled`，请求未到客户端 | 发送中间件不能注册自定义发出方法 |
| [`server.go:1729`](https://github.com/modelcontextprotocol/go-sdk/blob/v1.8.0/mcp/server.go#L1729)，Elicit | 固定标准 capability、方法和结果 schema 校验 | 不是通用自定义服务端请求函数 |
| [`server.go:2286`](https://github.com/modelcontextprotocol/go-sdk/blob/v1.8.0/mcp/server.go#L2286)，AddReceivingCustomMethod | 处理自定义客户端到服务端方法，拒绝覆盖标准方法 | 可支持自有自定义服务，不能解决自定义发出请求 |
| [`shared.go:206`](https://github.com/modelcontextprotocol/go-sdk/blob/v1.8.0/mcp/shared.go#L206)，接收分发 | 方法查询/参数解码先于接收中间件 | 不能拦截未注册未知方法或恢复已拒绝的 input-map 解码 |
| ResultBase 与接收中间件 | 测试适配器经官方 HTTP 成功输出 input 方法为 OpenAI 方法的 `input_required` 结果 | 现代服务端结果适配技术上可行，不能将所有方向都说成受阻 |
| 同一适配结果与官方 Go client | 类型化结果解码拒绝自定义方法 | 单纯输出字段不是完整扩展 MRTR 支持 |

现代原始 HTTP 复现包含协议元数据/客户端 capability 和必需 `Mcp-Method`、`Mcp-Name` Header，避免把请求本身格式错误误判为扩展限制。

结果适配器仍是**仅供测试的实验**，不是公开发送／MRTR API。独立 forms 包已提供 schema／回答校验与模型绑定；该测试没有集成受支持的现代流程或持久且经过认证的续轮契约，不声明 openai/elicitation capability。现代 MRTR 不在本轮 Python Extensions 对齐范围，该实验仅记录公开 SDK 边界。旧版直接发送仍受 SDK API 阻碍。

## 需要的公开 SDK 接口

要支持类型化互操作，官方 SDK 需要公开注册自定义服务端发出请求，并提供参数/结果工厂，保留 context、取消、请求关联及 capability 检查；还需要 MRTR 输入 request/response 编解码扩展注册。可以是请求 registry 或明确自定义 carrier；绕过生命周期要求的 `any` 逃生口不足以解决问题。

现代适配无需 fork 即可输出字段，但不能据此宣称 MRTR 互操作受支持；Python Extensions 对齐不要求此路径，也不发布适配／恢复 API。未来若单独授权该工作，必须读取实际每次请求 capability，并独立证明宿主互操作和续轮所有权。

开发阶段复现入口为 `go test -race ./internal/sdkcheck -v`，也包含在 `make test` 中。提交/push 收尾不得重复运行。

## 2026-10-02 Extensions 对齐范围

公共表单声明、校验、类型绑定已实现，见[表单指南](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/ZH-Forms)。本文的接收结果适配仍只是调查测试。Python 扩展 elicit_input 不实现 MRTR，通用状态加密／恢复及底层 SDK 功能不做对齐；不发布公共现代 MRTR 适配或两轮 review 工具。
