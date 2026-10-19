# skillkit CHANGELOG

> 本文件由 `matrix-tools/scripts/gen_changelogs.py` 从 git 历史生成，按 tag 切段、取每段内的 conventional commit 渲染。

## [v0.8.0] — 2026-10-19

- 杂务(deps)：依赖地板对齐远端最新（guardx v0.5.0 / textx v0.5.0） (13)

## [v0.7.0] — 2026-10-19

- 杂务(deps)：依赖地板对齐远端最新（pkg v0.15.1） (#13)

## [v0.6.1] — 2026-10-19

- 杂务(deps)：依赖地板跟进 guardx v0.4.1 / textx v0.4.0

## [v0.6.0] — 2026-10-18

- 与 v0.5.1 相比无代码差异（版本号/打包变更）

## [v0.5.1] — 2026-10-18

- 与 v0.5.0 相比无代码差异（版本号/打包变更）

## [v0.5.0] — 2026-10-18

### 新增：skillkit 仓建立——技能与资产契约（3 包）

从 agentkit 单仓切出 tier1 能力层模块，module 路径 github.com/oagentkit/skillkit。
收编 3 个包：skill（SKILL.md 渐进披露 + 多根技能库）/ pack（领域包布局契约）/ lineage（装配血缘图）
矩阵内依赖：textx、guardx。外部地板 yaml.v3 + semver + eino。
门禁：go build / go vet / go test -race 全绿，gofmt 干净。

- 杂务(deps)：依赖地板升到「历史最新 + 0.1.0」 (#5)
- 杂务(deps)：跟进最新 tag (#4)
- 杂务(deps)：oekit 依赖升级 (#3)
- 杂务(deps)：guardx v0.1.1 → v0.1.2 (#2)
- 杂务(deps)：跨仓依赖 guardx v0.1.0 → v0.1.1 (#1)
- 杂务(deps)：oekit 三模块升至最新（pkg v0.13.0 / concurrency v0.12.0 / logx v0.7.0）

- CI：接入 org 共享门禁（规则本体在 oagentkit/.github）
