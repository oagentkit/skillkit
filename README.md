# skillkit

技能与资产契约——SKILL.md 渐进披露/领域包布局契约/装配血缘图 —— oagentkit 模块矩阵 **tier 2**（集成层，依赖 tier0/tier1 + 重第三方地板）。

```
go get github.com/oagentkit/skillkit
```

| 包 | 职责 |
|---|---|
| `skill` | SKILL.md 渐进披露：元数据常驻、正文按需加载；多根技能库与版本判定 |
| `pack` | 领域包（expert/pack 目录布局）的契约面 |
| `lineage` | 装配血缘：expert → skill → MCP server 三类资产节点的依赖图 |

## 定位

**tier 2** — 集成层，依赖 tier0/tier1 + 重第三方地板。

- 本仓依赖的矩阵内仓：`guardx`、`textx`。

## 外部依赖地板

- `gopkg.in/yaml.v3`
- `github.com/Masterminds/semver/v3`

## 与 oekit 的关系

本矩阵（oagentkit，Agent/LLM 编排工具链）**依赖** oekit（Open Easy Kit，Go 微服务工具链）：
`oekit/pkg` / `oekit/concurrency` / `oekit/logx` 等。oekit 是另一套独立矩阵，不是子品牌。

## 开发

```bash
go test ./... -race
go vet ./...
gofmt -l .
```

规范见 [AGENTS.md](AGENTS.md)。

## License

MIT
- SKILL.md 的**渐进披露**是硬要求：元数据常驻、正文按需加载。
  不得为了「一次读全」把正文塞进元数据 —— 那会让常驻上下文无谓膨胀。
- `pack` 的目录布局是契约面，改动布局等于 breaking，必须升 major。