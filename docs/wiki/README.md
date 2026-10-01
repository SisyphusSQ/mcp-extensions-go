# GitHub Wiki source

These files are the version-controlled source for the repository's [GitHub Wiki](https://github.com/SisyphusSQ/mcp-extensions-go/wiki). English and Chinese pages are maintained together. The Chinese pages are an explicit user-requested exception to the project's English documentation convention.

| Source page | Published page |
| --- | --- |
| [Home.md](Home.md) | [Language selection](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/Home) |
| [EN-Overview.md](EN-Overview.md) | [Overview](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/EN-Overview) |
| [ZH-Overview.md](ZH-Overview.md) | [项目概览](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/ZH-Overview) |
| [EN-Implementation-Roadmap.md](EN-Implementation-Roadmap.md) | [Implementation roadmap](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/EN-Implementation-Roadmap) |
| [ZH-Implementation-Roadmap.md](ZH-Implementation-Roadmap.md) | [可实现缺口与后续开发](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/ZH-Implementation-Roadmap) |
| [EN-Validation-and-SDK-Limits.md](EN-Validation-and-SDK-Limits.md) | [Validation and SDK limits](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/EN-Validation-and-SDK-Limits) |
| [ZH-Validation-and-SDK-Limits.md](ZH-Validation-and-SDK-Limits.md) | [验收状态与 SDK 限制](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/ZH-Validation-and-SDK-Limits) |
| [_Sidebar.md](_Sidebar.md) | Wiki navigation |
| [_Footer.md](_Footer.md) | Source and issue links |

## Language layout

The layout follows the owner's [orchestrator Wiki](https://github.com/SisyphusSQ/orchestrator/wiki), reference commit `1ce473c59ddab4558965b80a5762773c8935e25d`. Home only selects a language; English pages use `EN-`, Chinese pages use `ZH-`, the sidebar has separate language groups, and each page links its translation and Wiki Home. The footer links the source and issue tracker. Keep each content page in its own language.

## Publishing

The owner explicitly requested public visibility on 2026-10-01, superseding the previous private-only requirement. GitHub reports `visibility: public`, `private: false` and `has_wiki: true`.

GitHub requires an initial Wiki page created through its website before the Wiki Git repository can be cloned. The owner created Home on 2026-10-01; the initial bilingual publication followed at `54c55ad`. The language-separated layout and footer were then published at `b50998e` on `master`: Home, three English pages, three Chinese pages, sidebar and footer. Public visibility and merging source documentation alone do not initialize or publish the Wiki.

For future updates, clone `git@github.com:SisyphusSQ/mcp-extensions-go.wiki.git` into a separate local directory. Read its current pages and history first. Copy only the nine owned files listed above from `docs/wiki`; preserve unrelated pages and merge any existing navigation. Remove obsolete owned pages when renaming them, and update all incoming links. Do not publish this source README as a Wiki page. Commit and push the Wiki's existing default branch without force, then read back its commit/pages. The Wiki is a separate Git repository; merging the main repository does not publish its pages.

Update both languages and `docs/compatibility.md` when status changes. Planned work must remain labeled as not implemented until a real public API and the relevant acceptance evidence exist. Publishing documentation does not authorize implementing its backlog or rerunning tests during closeout.
