# 本地 Codex 插件与实际验收

[English](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/EN-Live-E2E) · **中文** · [Wiki 首页](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/Home)

对应源文档：[`docs/live-e2e.md`](https://github.com/SisyphusSQ/mcp-extensions-go/blob/main/docs/live-e2e.md)。

## 位置与安装

[PR #1](https://github.com/SisyphusSQ/mcp-extensions-go/pull/1) 在仓库仍为私有时合并到 main，提交 `3f45cf66eab5f6fd172763e604d794e12cef5cf1`。[PR #2](https://github.com/SisyphusSQ/mcp-extensions-go/pull/2) 将 `suqing/live-e2e` 合并为 `4ce5a1f73d1755a72d5b669b7a17c16bfd8721b6`。2026-10-01，所有者明确要求公开源码；插件仍是本地安装，没有上架公开插件目录。合并不代表完成人工验收。

插件为 **MCP Extensions Go**，ID `mcp-extensions-go@mcp-extensions-go-local`，本地 marketplace 为 **MCP Extensions Go Local**，安装版本 `0.0.0-dev` 仅是安装元数据，不是 release/tag。包内包含真实 MCP 配置、生成可执行文件、App HTML 与图标，没有空 skills/hooks/apps。

```sh
cd /absolute/path/to/mcp-extensions-go
npm --prefix examples/frontend ci --ignore-scripts
make plugin
codex plugin marketplace add /absolute/path/to/mcp-extensions-go
codex plugin add mcp-extensions-go@mcp-extensions-go-local
```

`make plugin` 写入被忽略的 `plugins/mcp-extensions-go/bin/mcp-extensions-stdio` 和 `assets/app.html`。Go、Node 必须已有，不自动安装全局工具或下载浏览器。官方 CLI 把文件复制到缓存并启用；源码改变后重新构建/安装。已有进程/页面可能保留旧程序或旧 HTML，直到重新打开插件或重启桌面。

sqmc04 的缓存目录为 `/Users/suqing/.codex/plugins/cache/mcp-extensions-go-local/mcp-extensions-go/0.0.0-dev`。已安装 CLI 使用 manifest version 作为缓存路径，不是通用插件指南中的 `local` 后缀。`.codex-plugin/plugin.json` 与 `.mcp.json` 兼容格式参考上游 Bits & Bolts。CLI 接受便携 root manifest，但未从其 `mcp.json` 返回 MCP 服务；兼容格式正确解析 `workspace` 并加载真实服务。

stdio framing、会话、传输由官方 SDK 负责，stdout 只输出 MCP，日志写 stderr。父进程控制本机访问。不创建 HTTP 凭据、监听端口、daemon、远程服务或公开插件条目；独立 HTTP 示例继续强制 Bearer 认证。

## 设置所有权与重启行为

插件默认使用 `os.UserConfigDir()/mcp-extensions-go/live-e2e/settings.json`；sqmc04 实际路径为 `/Users/suqing/Library/Application Support/mcp-extensions-go/live-e2e/settings.json`。运维绝对路径 `MCP_SETTINGS_FILE` 可覆盖，不接受宿主工具输入提供该路径。

启动时创建目录/锁文件，只有成功更新才创建设置文件。读取限制 64 KiB，Unix 保存文件为 0600；Windows 继承用户目录 ACL，未独立验证。可取消操作系统锁覆盖重新读取、patch 和原子替换。损坏、不可读、过大状态明确报错，不静默重置。

设置保留于正常进程重启；这是单所有者存储，本地所有 App 实例有意共享设置。按身份隔离的多用户存储和断电耐久性不在示例范围；公共 Settings API 仍要求调用者提供存储。

## 2026-10-01 的实际证据

| 层级 | 证据 | 限制 |
| --- | --- | --- |
| 安装 | 官方 CLI 报告 installed/enabled，plugin/read 解析 workspace | 不能单凭此证明当前桌面已加载 |
| 官方 CLI 协议 | 独立 app-server 初始化插件，发现 5 个工具、3 个资源、现代/旧版设置能力 | 不是运行桌面的 App bridge |
| Go 子进程协议 | 官方 CommandTransport 验证双 stdio 进程、不相关局部更新、资源发现/读取、退出与新进程读取保存状态 | 本机协议测试，不是人工宿主重启验收 |
| 桌面全局入口 | 用户截图显示更多菜单 Open workspace 和 MCP Extensions Go 页面 | 截图观察 |
| 桌面初始化 | 用户截图显示 Connected 与 Welcome 首结果 | 没有重复调用计数/实例隔离 trace |
| 桌面设置 | 截图显示 units=in、showGrid=false | 原生 Settings renderer 和重开/重启待验收 |
| 本地持久状态 | sqmc04 实际用户配置为上述值、权限 0600 | 证明写入发生，不是人工重开/重启结果 |
| 桌面上下文/消息 | 截图显示上下文数量 1、明确问题和识别 bolt 的模型回答 | 替换/移除/updateId/实例隔离与其他消息目标待验收 |
| 桌面显示 | 用户反馈全局页 Open fullscreen 无明显反应 | 当时未捕获实际模式，不宣称切换成功 |

显示反馈修复读取宿主上下文与请求实际结果。已经 fullscreen 或不支持时禁用按钮并明确标注；宿主保留模式时明确报告。App 声明 inline/fullscreen 并显示 Display mode。本机浏览器测试覆盖初始 fullscreen、不支持、成功切换与宿主保持 inline；重开后的真实桌面验收待完成。

UI 工具拒绝原生 `com.openai.codex` 控制；此前 MCP Apps inventory 为空，运行桌面无默认 CLI control socket。真实桌面动作由用户完成，代理没有重启桌面或发送聊天消息。

## 人工验收

1. 最新安装后关闭/重开 Open workspace。若保留旧文件，在当前活动轮次结束后重启桌面。新页面应显示 Display mode；已 fullscreen/不可用时禁用按钮，切换请求报告实际结果。
2. 读取设置、改 Units/Show grid、保存、关闭/重开后再读取；重启桌面/插件进程再重复。无效/不可用存储必须显示错误。
3. App 搜索 bolt/washer/空查询；另外单独验收 composer Mention typeahead 与选择。App Search 不证明原生 Mentions。
4. 若宿主提供 `.txt` 入口，使用 `examples/frontend/acceptance.txt`。HTML 类文本必须不执行，Go 不得打开任意宿主路径；宿主引用读取/权限与大文件仍需验收。
5. 从聊天 inline 启动，验收 global/thread 实例隔离、上下文替换/移除、明确消息目标、deep link 初始/后续导航。截图只证明已列出的成功路径。

完整 OpenAI 扩展表单/MRTR 仍受已记录 SDK 边界影响，没有声明这些 capability。标准 elicitation 或生成扩展字段不能当作已验收证据。

## 开发证据

- Go 1.27.0 的 `make test vet build` 通过，包含真实子进程 stdio 和持久 Store race 测试。
- Go 1.25.0 完整套件在最终小改动前通过；最终修改的 HTTP/stdio/Store race 测试与 `make vet build` 使用缓存 `GOTOOLCHAIN=go1.25.0` 通过。
- Go 1.25 Linux/Windows amd64 交叉构建通过，没有运行对应系统验收。
- 前端 build/typecheck 和已安装 Chrome 集成通过，包括变化字段 patch 保留另一实例不相关更新；最终浏览器运行覆盖显示反馈场景。
- Go 1.27.0 下 `go run golang.org/x/vuln/cmd/govulncheck@v1.8.0 ./...` 报告 No vulnerabilities found。工具从 module/build cache 运行，未全局安装，未改变模块版本；扫描工具自身 Go 下限 1.26，与本模块 1.25 下限分开。

安全自查保留 HTTP 认证、stdio 父进程访问控制；宿主 URI 不变成文件系统读取；持久路径由运维配置，限制读取、使用 OS 锁和私有原子状态文件；浏览器使用文本 DOM。这不证明多用户授权或完整宿主验收。

commit/push/merge 收尾复用开发证据，不重复测试/检查。人工验收继续针对合并代码进行；未发布新 release/tag。可实施缺口见[后续开发清单](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/ZH-Implementation-Roadmap)。
