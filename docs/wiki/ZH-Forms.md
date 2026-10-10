# 扩展表单与 Python 对齐

[English](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/EN-Forms) · **中文** · [Wiki 首页](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/Home)

对应源文档：[`docs/forms.md`](https://github.com/SisyphusSQ/mcp-extensions-go/blob/main/docs/forms.md)。实现状态以[源能力表](https://github.com/SisyphusSQ/mcp-extensions-go/blob/main/docs/compatibility.md)为准。

对齐基准是 OpenAI MCP Extensions Python 0.2.0，提交 [`900032d8bd7c1566202d0cb1666986584f932043`](https://github.com/openai/mcp-extensions/tree/900032d8bd7c1566202d0cb1666986584f932043)，不包括底层 MCP Python SDK。

## 声明与原始回答

`forms.New` 校验并快照平面对象，字段支持 string、boolean、integer、number 和字符串数组；`forms.Parse` 还检查未知 JSON 关键字和字段适用性。嵌套对象、任意 union、引用不支持。`Schema()` 返回独立副本。

字段包含 `examples`、`$comment`、`_meta`、`deprecated`、`readOnly`、`writeOnly` 和原始 JSON 默认值。允许为 null 的注解沿用 Python 声明规则，examples 和不透明元数据中的 null 得以保留；回答值和默认值不能为 null。

`Enum` 使用 `[]any`，可表达字符串、数字、布尔值或整个字符串数组。枚举值必须满足字段类型和约束。顶层重复 enum 与 Python 一样允许，数组 item 的多选必须是非空、唯一字符串。`enumNames` 数量与 enum 匹配，标签可以为空。丰富单选使用 `OneOf`，多选 item 使用 `AnyOf`。选项带 title、description、thumbnail 和旧版 preview 图片别名，title 可以为空。图片仅检查 HTTPS／上游 data-URI 声明语法，不会抓取。

建议值是提示，可与 choice 共存；它不让违反约束的值变得有效。自由数组支持字符串 item、建议值、数量和显式唯一性；多选始终执行唯一性。两类 item 的关键字范围不同。

表单数字约束只有 minimum/maximum。exclusive 边界与 `multipleOf` 保留在 Settings，不能用于表单。`Validate`、`ValidateField`、`ValidateResult` 检查字段、类型、格式、选项和资源，不转换类型、不插入默认值、不改变原始回答。accept 必须有 content；cancel/decline 由调用者处理。

支持 email、uri、date、date-time。格式实际执行校验；日期拒绝年份 0000 与无效日历值，URI 只做绝对 RFC 3986 语法检查。邮箱校验覆盖 mailbox 和域标签规则，但不宣称完整 Python email-validator／IDNA 等价。Go pattern 使用 RE2，不宣称完整 ECMA-262／Pydantic 正则等价；Python 独立 `is_valid_value` 遇到 pattern 也明确要求外部 ECMA-262 校验器。

独立校验保留有符号／无符号 64 位整数，超范围直接失败；integer 拒绝 JSON `1.0`。分数／指数数值使用有限 float64。官方 SDK 的 map 解码可能已丢失原始数值表示或精度，扩展无法恢复；经过该边界的大整数标识应使用字符串。

## 资源选择与上传引用

`ResourceInput.Type` 为 resource 或旧版 file；`Options` 使用 `[]*forms.ResourceOption`，嵌入官方 `mcp.Resource`，`Extra` 保留额外 JSON 描述字段且不能覆盖标准字段。URI 必须有效且唯一。单选是 URI 字符串，多选是 URI 字符串数组。

数组 Selection 支持 explicit／implicit。implicit 禁止默认值并默认允许宿主文件选择。默认值只能引用提供的 options。UserOptions 支持 file／directory 和上游 accept 规则。

声明 UserOptions 时，校验允许宿主返回有效用户 URI，与 Python `validate_file_selections` 一致；可选 SelectionPolicy 可进一步限制。未声明 UserOptions 时，options 外的引用失败。URI 校验不会读取文件，也不会授予读取权限。

PrepareSubmission 检查当前选择与最终上传数量，返回独立选项；CompleteSubmission 合并宿主已授权的上传 URI 并校验最终值。额外回调可省略；拒绝引用、错误 URI、重复项和数量违规会失败。原 content 不变。宿主负责实际上传授权、文件类型／accept 检查，业务负责后续读取授权。Go 不上传字节、不读取资源、不持久化；业务写入前仍须检查完整回答。

## Go 模型与 Settings

NewModel 使用公开 jsonschema-go API 推导平面 Go struct、JSON 别名、description、omitempty／omitzero。指针区分未提供与零值。字段覆盖使用 wire 名称，不能改变类型。NewModelWithOptions 接收公开 ForOptions，让命名类型声明 enum、string const、约束和注解；不支持的关键字明确失败。string const 转为 enum。有声明默认值的模型字段在 wire 上可省略；默认值只进入类型化结果，原 answer 不变。解码及业务回调失败明确返回，回调在调用方写入前运行。

Settings 提供 FieldsFor、FieldsForWithOptions 和 NewModelServer。字段要求非 null 基础类型且无 schema 默认值。完整有效值仍要求全部字段；更新是非空局部 map。ModelConfig.SchemaOptions 推导命名类型 enum／约束。类型化 Update 默认接收 Go 导出字段名，wire schema 与返回状态保留 JSON 别名；FieldNames 可覆盖映射。普通 Config 不配置映射时仍使用 wire 名称。

FieldValidators 按 wire 名称配置，签名为 `func(context.Context, any) (any, error)`。只对提供的 patch 字段在保存前执行，可转换值，转换结果再次校验。官方 wire schema 校验在前，不自动复现 Python before／wrap 装饰器接受 schema-invalid 输入的行为。完整状态／跨字段规则必须由 Update 在业务事务内合并后检查；返回值校验不能撤回已发生的写入。本库没有数据库或通用重启恢复 API。

## 现代扩展表单请求

在官方工具 handler 调用 `model.Form().RequestInput(ctx, req, forms.RequestOptions{Key: "selection", Message: "Choose"})`。pending 非 nil 时立即返回该工具结果，不同时返回内容或写入业务状态；否则显式处理 cancel／decline，accept 时用 model.Decode。完整[公共 stdio 示例](https://github.com/SisyphusSQ/mcp-extensions-go/blob/main/examples/form-mrtr/main.go)由 make build 构建为 bin/form-mrtr。

要求协议至少 2026-07-28，并且本次请求同时具有 elicitation.form 和 extensions["openai/elicitation"].form；不支持时返回 ErrUnsupportedClient，不自动替换为普通／旧版请求。核心 requestedSchema 是空对象，完整 schema 在 _meta["openai/elicitation"].requestedSchema。官方 SDK 在分发时设置 resultType，helper 刚构造的 pending 不用 NeedsInput 判断。

Meta 保留无关元数据；SelectionPolicy 沿用既有资源选择策略。RequestState 是客户端回传、不可信、业务所有的状态，本库不签名／加密／存储／授权／去重／过期／恢复。业务写入前需将续轮绑定已验证身份、工具和参数，并检查重放与事务规则。引用语法校验不授予读取权，示例不写入业务数据。

## 发送边界与证据

Python 0.2.0 通过 request_form_input 支持现代 MRTR。Go Form.RequestInput 复用官方标准 elicitation/create／InputRequestMap，承载扩展 schema 并校验重试回答；旧版 elicit_input_legacy 自定义发送仍受公开 API 限制。

0.2.0 现代适配已实现；通用请求状态加密／恢复和业务 review 工作流仍不提供。标准会话、传输、MRTR、tasks 与持久化机制不作为 Extensions 对齐目标。本地读取器收回 internal/example，已有设置文件存储仍是业务示例。

[Python 对照数据](https://github.com/SisyphusSQ/mcp-extensions-go/blob/main/forms/testdata/python-parity.json)包含 51 个 schema 与 86 个回答／上传案例；[生成脚本](https://github.com/SisyphusSQ/mcp-extensions-go/blob/main/scripts/generate-form-parity.py)记录固定源码与依赖版本。Go 测试直接读取数据，无需 Python 或网络。它证明已覆盖案例的语义一致，不证明完整 Pydantic 等价或真实 OpenAI 表单／选择／上传 UX 已验收。

业务接入从 [Agent 指南](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/ZH-Agent-Integration)开始。[2026-10-02 子进程记录](https://github.com/SisyphusSQ/mcp-extensions-go/blob/v0.0.1/docs/validation-2026-10-02.md)补充后端证据，不代表宿主表单验收。
