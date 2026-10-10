# 验收状态与 SDK 限制

[English](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/EN-Validation-and-SDK-Limits) · **中文** · [项目概览](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/ZH-Overview) · [Wiki 首页](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/Home)

## v0.0.2／上游 0.2.0 当前状态

当前已提供现代 Form.RequestInput、Mention capability、可省略的 Settings 按钮标题、移除 Settings App 入口的迁移错误和 desktop／web 草稿示例。下面 2026-10-01／02 记录属于历史，其中排除现代 MRTR 的范围已被本轮授权取代；通用恢复和旧式发送限制仍在。真实宿主表单／草稿仍未验收。当前证据见[开发验证记录](https://github.com/SisyphusSQ/mcp-extensions-go/blob/main/docs/validation-2026-10-10.md)。

这里区分实现缺口、宿主验收和业务责任。[完整能力对应表](https://github.com/SisyphusSQ/mcp-extensions-go/blob/main/docs/compatibility.md)仍是详细实现状态的权威记录。

## 已有验证证据

## 历史证据（2026-10-01／02）

| 层级 | sqmc04，2026-10-01 的已有记录 |
| --- | --- |
| Go 1.27.0 | 全量 race、vet 与 HTTP/stdio 构建通过 |
| Go 1.25.0 | 最终小改动前全量测试通过；最终 HTTP/stdio/store race 与 vet/build 通过 |
| 其他平台 | Go 1.25 Linux/Windows amd64 交叉构建通过；未做运行验收 |
| 浏览器 | 构建、类型检查和现有 Chrome 集成通过：官方 AppBridge、首结果、设置/搜索、变化字段保留、显示/deep link/上下文/消息/文件和缺少 capability |
| 漏洞扫描 | Go 1.27.0 下缓存内 govulncheck v1.8.0 未发现漏洞；扫描工具自身 Go 下限为 1.26，与本模块 1.25 下限分开 |
| 插件 CLI | 已安装启用；独立官方 app-server 发现 5 个工具、3 个资源和两种设置能力位置 |
| 用户桌面截图 | 全局 Open workspace 入口、Connected/首结果、有效设置、上下文数量 1，以及明确发送问题后模型识别 bolt 的回答 |
| 本地持久状态 | 实际文件为 units=in/showGrid=false，权限 0600；证明发生持久写入，不等于用户已完成重启验收 |

PR 合并和 Wiki 文档工作复用上述输出，收尾不重复测试、lint 或检查。PR #2 没有配置 CI checks，不能描述为 CI 通过。完整证据见[实际验收记录](https://github.com/SisyphusSQ/mcp-extensions-go/blob/main/docs/live-e2e.md)。

## 已实现、仍需真实宿主验收的场景

| 场景 | 还需要确认的内容 |
| --- | --- |
| Settings | 原生设置渲染、分组/工具按钮、错误显示、关闭重开与实际进程重启 |
| Mentions | 原生 composer 搜索、空查询、选择/读取引用、拒绝或不可用搜索；App 内 Search 不足以证明原生 mentions |
| 文件 | `.txt` 入口、不透明引用的读取与授权、HTML 类文本不执行、大文件和不支持的能力 |
| App 实例 | 聊天/thread 入口、单次首调用、两个实例隔离和工具/资源 visibility 的实际限制 |
| 显示模式 | 重开后的反馈修复、实际模式、inline→fullscreen、宿主保持其他模式；旧版全局页按钮曾被报告无明显反应 |
| 其他 UI | quickAction、Settings App 入口、deep link 初始与后续导航 |
| 上下文/消息 | 替换/移除/updateId、实例隔离、其他消息目标；截图只证明已有成功路径 |

当前示例对宿主文件只读。写回、ETag 冲突、订阅和原生打开文件，需要补充合适的前端示例并实际验收后才能宣称支持。它们属于 TypeScript App/宿主责任，不是“Python 服务端已实现而 Go 缺失”的功能。

## SDK 与 capability 边界

标准 MRTR 由官方 Go SDK 实现，本机已验证两轮、取消/拒绝/手动恢复、重复/过期输入、过期/完成重放、八个并发调用和连接隔离。这些测试中的续轮状态存储不是生产持久化，也不是已认证 stateless HTTP 恢复的证明。

OpenAI 扩展表单是另一条契约。Go 的类型化 `InputRequestMap` 无法编码/解码 OpenAI 自定义方法；默认发送器拒绝方法替换，自定义接收注册也不增加发送能力。公开结果适配可输出扩展字段，但完整 MRTR 互操作尚未验证。具体源码位置、复现与所需公开接口见 [protocol-investigation.md](https://github.com/SisyphusSQ/mcp-extensions-go/blob/main/docs/protocol-investigation.md)。

只声明已实现的 capability，读取真实客户端能力，不从 UI 元数据或旧握手推断。浏览器模型上下文、消息、deep link、选择/上传界面和原生文件打开，由标准 MCP Apps/OpenAI TypeScript SDK 与宿主负责。

## 业务与安全责任

调用者负责已验证身份、每个资源授权、多用户设置、搜索来源、跨字段事务和持久续轮策略。持久设置示例只有一个所有者；Unix 状态文件使用 0600，Windows 继承 ACL 未验证，不宣称断电耐久性。

解析 `openai/resource.path` 或收到不透明 URI 都不授予读取权。读取本地文件要独立授权、配置可信允许根目录、限制符号链接与资源大小。根目录内的恶意硬链接/mount 和可中断普通文件系统调用仍在 私有示例读取器契约之外。后续修改保持 HTTP 认证、运行时凭据、浏览器文本输出和资源权限边界。源码仓库已按所有者在 2026-10-01 的明确要求公开，运行时访问控制没有因此改变。该历史验收阶段未进行远程 HTTPS 部署、公开插件目录上架、tag 或 release。v0.0.1 模块发版不改变运行时权限，也不完成宿主验收。

## 2026-10-02 扩展对齐

Python 对照覆盖 51 个 schema 与 86 个回答／上传案例，Go 另验证模型默认值／别名、命名类型推导、保存前字段校验和局部 patch。它不代表完整 Pydantic／ECMA-262／IDNA 等价或宿主 UX 验收。公共 MRTR 适配与通用恢复已移除，读取器收回私有示例，直接发送仍受 SDK 限制。见[表单指南](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/ZH-Forms)。

最终 Go 1.27.0 与最低 Go 1.25.0 均通过 make test／vet／build，包括 race、带认证 HTTP 与真实 stdio 子进程集成。未运行真实 OpenAI 表单宿主 E2E。

## 新增扩展子进程验收（2026-10-02）

此前隔离的官方 CommandTransport 实际调用 186 次 MCP：51 声明、82 值、4 上传引用；完整回答、模型别名／默认值／注解／业务拒绝；Settings 格式、排他边界、multipleOf、局部 patch 和字段转换。非法 patch 未到存储，非法完整回调状态返回工具错误。[证据与复现入口](https://github.com/SisyphusSQ/mcp-extensions-go/blob/v0.0.1/docs/validation-2026-10-02.md)记录此前运行；源码归档为 examples/acceptance，v0.0.1 收尾不重跑。新增公共 API 接入示例仅完成构建，未做运行验收；真实宿主表单仍受限／未验收。
