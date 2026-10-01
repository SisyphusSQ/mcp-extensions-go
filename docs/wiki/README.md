# GitHub Wiki source

These files are the version-controlled source for the private repository's [GitHub Wiki](https://github.com/SisyphusSQ/mcp-extensions-go/wiki). English and Chinese pages are maintained together. The Chinese pages are an explicit user-requested exception to the project's English documentation convention.

| Source page | Published page |
| --- | --- |
| [Home.md](Home.md) | [Home](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/Home) |
| [Home-ZH.md](Home-ZH.md) | [中文首页](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/Home-ZH) |
| [Implementation-Roadmap.md](Implementation-Roadmap.md) | [Implementation roadmap](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/Implementation-Roadmap) |
| [Implementation-Roadmap-ZH.md](Implementation-Roadmap-ZH.md) | [可实现缺口与后续开发](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/Implementation-Roadmap-ZH) |
| [Validation-and-SDK-Limits.md](Validation-and-SDK-Limits.md) | [Validation and SDK limits](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/Validation-and-SDK-Limits) |
| [Validation-and-SDK-Limits-ZH.md](Validation-and-SDK-Limits-ZH.md) | [验收状态与 SDK 限制](https://github.com/SisyphusSQ/mcp-extensions-go/wiki/Validation-and-SDK-Limits-ZH) |
| [_Sidebar.md](_Sidebar.md) | Wiki navigation |

## Publishing

GitHub requires an initial Wiki page created through its website before the Wiki Git repository can be cloned. The Wiki feature is enabled on this private repository. At source preparation time, the first page and authenticated browser access were not available, so publication was pending. Keep the repository private; do not create a public mirror to work around availability or access limitations.

After initialization, clone `git@github.com:SisyphusSQ/mcp-extensions-go.wiki.git` into a separate local directory. Read its current pages and history first. Copy only the seven owned files listed above from `docs/wiki`; preserve unrelated pages and merge any existing sidebar navigation. Do not publish this source README as a Wiki page. Commit and push the Wiki's existing default branch without force, then read back its commit/pages. The Wiki is a separate Git repository; merging the main repository does not publish its pages.

Update both languages and `docs/compatibility.md` when status changes. Planned work must remain labeled as not implemented until a real public API and the relevant acceptance evidence exist. Publishing documentation does not authorize implementing its backlog or rerunning tests during closeout.
