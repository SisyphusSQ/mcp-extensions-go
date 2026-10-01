# 可实现缺口与后续开发

[English](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/Implementation-Roadmap) · [中文首页](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/Home-ZH)

除明确写为已有基础外，下列事项均为**尚未实现**，是后续开发清单，不是能力支持声明。上游参照为 `openai/mcp-extensions` 的 `900032d8bd7c1566202d0cb1666986584f932043`，2026-10-01 已独立核对它仍是上游 main。

## 可以直接实施的内容

这些服务端类型、构造与校验辅助可以独立于表单传输发挥作用。需要实现真实可用 API 和测试，不预建空包；完整调用链尚未成立时，不宣称完整表单支持。

| 编号 | 未实现内容 | 依赖 | 完成证据 |
| --- | --- | --- | --- |
| D1 | Settings 额外约束 | 现有 Settings 工厂与验证器 | 真实 schema 及错误定义/输入/输出测试 |
| D2 | 表单 schema 和丰富选项 | 上游表单契约、官方 MCP 类型 | 完整支持范围内的构造、序列化与校验 |
| D3 | 建议值、自由输入与字符串数组 | D2 | 建议值和自定义回答遵守同一套规则 |
| D4 | 资源/文件选择器声明 | D2、官方 Resource 类型 | 选项、模式、过滤、默认值及 URI 编码正确 |
| D5 | 表单回答及上传引用校验 | D2–D4 | 不修改原始回答，拒绝非法字段或选择 |
| D6 | Go 模型与 schema 绑定 | D1/D2/D5、公开 schema API | 类型化往返、字段别名和明确验证回调 |

### D1 — Settings 额外约束

在现有基础字段上增加支持范围内的字符串 `format`、数值 `exclusiveMinimum` / `exclusiveMaximum`，对照 Python 生成 schema 和原生设置契约。拒绝不适用、未知或矛盾的约束；保持完整有效值、非空局部 patch 和调用者存储责任。

验收：schema 真实暴露已支持的约束；错误定义在存储执行前被拒绝；非法 patch 不进入更新回调；非法返回值成为工具错误。确认 `format` 真正参与校验，而不只是输出一个验证器忽略的字段。覆盖排他边界上的相等值、整数/数值及非有限数值，不支持的格式须明确报错。当前 Settings pattern 使用 RE2，不等于任意 Python / ECMA-262 正则。

入口：[Python Settings](https://github.com/openai/mcp-extensions/blob/900032d8bd7c1566202d0cb1666986584f932043/python/src/openai_mcp_extensions/settings.py) 与当前 `settings/settings.go`。跨字段规则、授权和事务仍由业务更新回调在持久化前负责。

### D2 — 表单 schema 与丰富选项

实现支持范围内的对象/字段 schema：基础类型与数组、必填/可选字段、默认值、字符串/数值/数组约束，带标签的单选/多选、`enumNames`，以及选项标题、描述、缩略图和预览图元数据。原生设置与表单的规则分别处理。

验收：正确编码 `oneOf` 与数组 `anyOf` 选项；选项非空且唯一，标签对应正确；约束适用于字段类型，边界不矛盾，默认值有效，必填字段引用已声明属性。按上游规则校验图片 URL/data URI，但不主动下载这些 URL。明确拒绝不支持的嵌套或 union schema。Pattern 使用符合契约的验证器或明确受限的子集，不能静默把 RE2 当成完整 ECMA-262。

入口：[Python 表单协议](https://github.com/openai/mcp-extensions/blob/900032d8bd7c1566202d0cb1666986584f932043/python/src/openai_mcp_form_protocol/__init__.py)、[TypeScript schema](https://github.com/openai/mcp-extensions/blob/900032d8bd7c1566202d0cb1666986584f932043/typescript/src/server/forms/schema.ts)。

### D3 — 建议值与自由输入

实现字符串字段和字符串数组项的 `x-openai-suggestions`。建议值不是排他的枚举列表：允许自由输入时，自定义值应有效，并与建议值遵守相同的长度、pattern、format 规则。数组还须检查数量和唯一性。

验收：覆盖列表内/列表外值、错误建议值、错误默认值、空值/必填、重复与数量边界。校验用户实际提交的回答，不插入默认值或偷偷转换类型。复用 D2/D5，不再实现另一套表单协议。

入口：[Python 建议值示例](https://github.com/openai/mcp-extensions/blob/900032d8bd7c1566202d0cb1666986584f932043/python/README.md#suggested-values)。

### D4 — 资源/文件选择器声明

实现规范的 `x-openai-input.type=resource` 与已弃用的 `file` 别名、资源选项、数组的 explicit/implicit 选择、用户文件/目录选项和 accept 过滤。单选返回 URI 字符串，多选返回 URI 数组；selection 模式仅用于数组。隐式选择不允许默认值，并允许用户选择；显式默认值必须引用允许的资源。

验收：覆盖非法/重复 URI、错误/重复 accept token、不适用的字段类型、不支持模式、默认值及单选/多选编码；保留官方资源元数据。选择结果不能变成授权或任意文件系统读取。选择、上传和预览界面属于宿主；该任务不新增 Go 浏览器选择器、数据库或上传服务。

入口：[Python picker](https://github.com/openai/mcp-extensions/blob/900032d8bd7c1566202d0cb1666986584f932043/python/src/openai_mcp_extensions/form/_resource_picker.py)、[TypeScript picker](https://github.com/openai/mcp-extensions/blob/900032d8bd7c1566202d0cb1666986584f932043/typescript/src/server/forms/file-picker.ts)。

### D5 — 回答与上传引用完成校验

校验必填/未知字段、JSON 基础类型、支持的格式、选项、资源 URI、数组、默认值和用户文件策略。提供真实可用的提交准备/完成辅助：检查待上传数量和 accept 规则，将宿主已授权上传得到的 URI 与现有选择合并，再校验最终值。保留提交的值，不自动补默认值或把字符串转成数字。

验收：拒绝非法选项/上传、不允许的资源、缺失的 accepted content、错误类型、单值/数组形状、数组大小/唯一性、非有限数值及非法 schema。上传引用必须来自拥有该资源权限的宿主/业务上下文；客户端自报 URI 不等于权限。这些辅助不执行上传，也不授予文件访问权。

入口：[Python 校验辅助](https://github.com/openai/mcp-extensions/blob/900032d8bd7c1566202d0cb1666986584f932043/python/src/openai_mcp_form_protocol/__init__.py)。其中独立 pattern 校验明确要求 ECMA-262 验证器，不能宣称不受限制的正则等价性。

### D6 — Go 模型与 schema 绑定

复用官方 SDK/schema 的公开 API，支持 Go 类型生成 schema、平面 JSON 字段别名、类型化 accepted 结果和明确验证回调。提供可用的模型辅助，不复制 Pydantic 装饰器，也不探查 SDK 私有注册表；保留当前显式字段/map API。

验收：平面别名、可选表单字段、未提交与零值区别、完整有效设置、不支持的模型结构、类型化 JSON 往返和回调错误。局部 patch 校验不能要求补齐未修改字段。完整业务状态规则和授权须由调用者在持久化事务内先检查；回调返回后的校验不能撤销已经发生的写入。不承诺逐字节复刻 Pydantic 行为。

入口：[Python schema 生成](https://github.com/openai/mcp-extensions/blob/900032d8bd7c1566202d0cb1666986584f932043/python/src/openai_mcp_extensions/form/_schema.py)、[Python Settings](https://github.com/openai/mcp-extensions/blob/900032d8bd7c1566202d0cb1666986584f932043/python/src/openai_mcp_extensions/settings.py)。

## 有条件实施：现代 OpenAI 表单 MRTR

S1 不能归入“现在即可完整对齐”。已有私有测试证明：公开接收中间件与 ResultBase 适配可以经官方 HTTP 输出扩展 `input_required` 结果。完整传输/续轮链路仍未验证，官方 Go 类型化客户端会拒绝这个自定义输入方法。

完成 D2–D5 后，在真实扩展宿主中按每次请求的实际客户端能力验证。先核对固定 SDK 的参数/结果解码和中间件顺序，再设计公开 API。续轮状态绑定已验证身份、工具、参数和轮次，限制容量/生命周期，处理重放、过期、完成、取消/拒绝、恢复及并发隔离。Stateless HTTP 必须验证真实身份隔离；如宣称持久恢复，还须验证进程重启后的所有权。事务和恢复由调用者存储负责。

验收必须有真实多轮宿主记录，并覆盖不支持的客户端、非法选择、过期/重复输入、取消/恢复/过期及并发用户/实例。仅能输出字段或通过标准 Elicit 测试都不够。如果公开 SDK 的解码、发送或能力读取入口阻塞，记录源码位置和复现，向上游提出公开接口需求，继续完成 D1–D6。不 fork，不用 unsafe/反射访问内部，不复制 JSON-RPC、会话或传输。

## 当前阻碍：旧式直接自定义表单请求

S2：Python `elicit_form` / `elicit_input` 可以通过公开请求接口发送 `openai/elicitation/create`，并校验 accept/decline/cancel。Go SDK v1.8.0 没有对应的自定义发出方法注册接口；发送中间件不能突破固定方法映射。需要上游提供合适的公开 API，或未来独立验证过的新 SDK；当前固定版本没有这些条件。

Python 文档也明确说明，该 `elicit_input` 旧式包装不实现 MRTR，不能据此认为 Python 扩展自身已完整覆盖 OpenAI MRTR。

参见 [SDK 调查](https://github.com/SisyphusSQ/mcp-extensions-go/blob/main/docs/protocol-investigation.md)和 [Python 请求实现](https://github.com/openai/mcp-extensions/blob/900032d8bd7c1566202d0cb1666986584f932043/python/src/openai_mcp_extensions/form/_elicitation.py)。

## 实施顺序与同步要求

建议依次推进 D1 → D2 → D3/D4 → D5 → D6，再验证 S1；S2 保持阻碍状态，直到需要的公开 API 存在。后续实施同步中英文 Wiki、能力对应表、架构、示例和交接；开发阶段用官方传输、真实前端和可访问的宿主做相关验证。进入 commit/push 收尾后复用证据，不重复测试/检查。不声明尚未实现的 capability，不添加占位 API。
