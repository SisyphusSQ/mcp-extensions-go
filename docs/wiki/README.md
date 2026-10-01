# GitHub Wiki source

These files are the version-controlled source for the repository's [GitHub Wiki](https://github.com/SisyphusSQ/mcp-extensions-go/wiki). English and Chinese pages are maintained together. The Chinese pages are an explicit user-requested exception to the project's English documentation convention.

| Source page | Published page |
| --- | --- |
| [Home.md](Home.md) | [Language selection](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/Home) |
| [EN-Overview.md](EN-Overview.md) | [Overview](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/EN-Overview) |
| [ZH-Overview.md](ZH-Overview.md) | [项目概览](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/ZH-Overview) |
| [EN-Architecture.md](EN-Architecture.md) | [Architecture](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/EN-Architecture) |
| [ZH-Architecture.md](ZH-Architecture.md) | [架构](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/ZH-Architecture) |
| [EN-Compatibility.md](EN-Compatibility.md) | [Compatibility and capabilities](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/EN-Compatibility) |
| [ZH-Compatibility.md](ZH-Compatibility.md) | [兼容性与完整能力表](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/ZH-Compatibility) |
| [EN-Frontend-Validation.md](EN-Frontend-Validation.md) | [Frontend validation](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/EN-Frontend-Validation) |
| [ZH-Frontend-Validation.md](ZH-Frontend-Validation.md) | [前端验证](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/ZH-Frontend-Validation) |
| [EN-Live-E2E.md](EN-Live-E2E.md) | [Local plugin and live acceptance](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/EN-Live-E2E) |
| [ZH-Live-E2E.md](ZH-Live-E2E.md) | [本地插件与实际验收](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/ZH-Live-E2E) |
| [EN-Protocol-Investigation.md](EN-Protocol-Investigation.md) | [Forms and MRTR investigation](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/EN-Protocol-Investigation) |
| [ZH-Protocol-Investigation.md](ZH-Protocol-Investigation.md) | [表单与 MRTR 调查](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/ZH-Protocol-Investigation) |
| [EN-Implementation-Roadmap.md](EN-Implementation-Roadmap.md) | [Implementation roadmap](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/EN-Implementation-Roadmap) |
| [ZH-Implementation-Roadmap.md](ZH-Implementation-Roadmap.md) | [可实现缺口与后续开发](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/ZH-Implementation-Roadmap) |
| [EN-Validation-and-SDK-Limits.md](EN-Validation-and-SDK-Limits.md) | [Validation and SDK limits](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/EN-Validation-and-SDK-Limits) |
| [ZH-Validation-and-SDK-Limits.md](ZH-Validation-and-SDK-Limits.md) | [验收状态与 SDK 限制](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/ZH-Validation-and-SDK-Limits) |
| [_Sidebar.md](_Sidebar.md) | Wiki navigation |
| [_Footer.md](_Footer.md) | Source and issue links |

## Language layout

The layout follows the owner's [orchestrator Wiki](https://github.com/SisyphusSQ/orchestrator/wiki), reference commit `1ce473c59ddab4558965b80a5762773c8935e25d`. Home only selects a language; English pages use `EN-`, Chinese pages use `ZH-`, the sidebar has separate language groups, and each page links its translation and Wiki Home. The footer links the source and issue tracker. Keep each content page in its own language.

## Documentation coverage

Architecture, the complete compatibility matrix, frontend validation, SDK protocol investigation and local plugin/live acceptance are published as full English pages and Chinese companions. Preserve runnable instructions, error/security ownership, SDK source references, existing test evidence and unverified host scenarios. Handoff is intentionally excluded from the Wiki at the owner's request and remains a source-repository continuation record.

## Publishing

The owner explicitly requested public visibility on 2026-10-01, superseding the previous private-only requirement. GitHub reports `visibility: public`, `private: false` and `has_wiki: true`.

GitHub requires an initial Wiki page created through its website before the Wiki Git repository can be cloned. The owner created Home on 2026-10-01; the initial bilingual publication followed at `54c55ad`. The language-separated layout and footer were then published at `b50998e` on `master`: Home, three English pages, three Chinese pages, sidebar and footer. Public visibility and merging source documentation alone do not initialize or publish the Wiki.

The full user-facing document expansion was published at `17c3278` on `master`: Home, eight English pages, eight Chinese pages, sidebar and footer. The public Home page displays 17 content pages and the expanded language groups. Handoff was not published.

For future updates, clone `git@github.com:SisyphusSQ/mcp-extensions-go.wiki.git` into a separate local directory. Read its current pages and history first. Copy only the nineteen owned files listed above from `docs/wiki`; preserve unrelated pages and merge any existing navigation. Remove obsolete owned pages when renaming them, and update all incoming links. Do not publish this source README as a Wiki page. Commit and push the Wiki's existing default branch without force, then read back its commit/pages. The Wiki is a separate Git repository; merging the main repository does not publish its pages.

Update each owning source document and both Wiki languages when behavior or status changes; `docs/compatibility.md` remains the implementation authority. Planned work must remain labeled as not implemented until a real public API and the relevant acceptance evidence exist. Publishing documentation does not authorize implementing its backlog or rerunning tests during closeout.
