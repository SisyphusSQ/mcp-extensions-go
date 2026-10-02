# Extensions 对齐状态

[English](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/EN-Implementation-Roadmap) · **中文** · [Wiki 首页](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/Home)

2026-10-02 所有者将范围限定为 OpenAI MCP Extensions Python 0.1.0，提交 `900032d8bd7c1566202d0cb1666986584f932043`，底层 MCP Python SDK 不做对齐。状态以[源能力表](https://github.com/SisyphusSQ/mcp-extensions-go/blob/main/docs/compatibility.md)为准。

| 项目 | 已实现 | 剩余边界 |
| --- | --- | --- |
| D1 Settings | format、含 exclusive 的数值边界、基础类型 schema／模型、保存前字段校验、业务名别名 | 事务／跨字段规则由业务负责；不自动复现 Python before／wrap 装饰器 |
| D2 表单 | 平面类型、JSON enum、完整注解、丰富单选／多选、thumbnail／preview 图片别名 | 明确的 RE2／邮箱／数值边界 |
| D3 建议值和自由输入 | 建议值与 choice 共存，自由字符串／数组，长度／数量／唯一性 | 宿主渲染未验收 |
| D4 资源输入 | resource／file、额外描述字段、explicit／implicit、用户文件／目录和 accept | 宿主选择／上传／读取授权 |
| D5 回答和上传引用 | 保留原值，字段／类型／选择／资源校验，prepare／complete 引用合并 | 不上传字节；后续读取由业务授权 |
| D6 模型 | 公开推导选项、命名类型 enum／约束、JSON 别名、类型化默认值／结果和 Go 回调 | 不实现一般嵌套／union 或 Pydantic 装饰器引擎 |
| 旧版直接发送 | 已复现公开 SDK 边界 | Go v1.8.0 没有自定义发送 API，继续受阻 |
| 真实宿主 E2E | 保留已有 App 的有限证据 | 表单／选择／上传／预览及宿主清单仍未验收 |

排除现代扩展 MRTR 适配、通用加密状态／重启恢复、依赖框架、标准会话／传输、tasks、EventStore。此前公共适配／状态包和两轮 review 示例已移除。根目录读取器收回 internal/example；已有设置存储仍为业务示例。Python 旧版 elicit_input 本身不实现 MRTR。

官方 Python 生成的对照数据覆盖 51 个 schema 和 86 个回答／上传案例。接口、复现与边界见[表单指南](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/ZH-Forms)。这些数据不证明完整 Pydantic 等价或真实产品验收。Wiki 通过独立仓库发布，已安装插件也需要独立构建／重载。开发阶段执行项目 test／vet／build；提交／推送／发版收尾复用已有结果，不重复测试。
