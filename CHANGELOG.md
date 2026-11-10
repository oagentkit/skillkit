# skillkit CHANGELOG

> 本文件由 `matrix-tools/scripts/gen_changelogs.py` 从 git 历史生成，按 tag 切段、取每段内的 conventional commit 渲染。

## [v0.10.5] — 2026-11-11

### 修复(skill)：Describe miss 口径统一 + Name 两 Provider 口径文档化（#83 #84）

- #84 Library.Describe 正常路径 miss 补「仅含名字的零值」返回（一行）——
  与 nil 接收者同口径，注释承诺「未知名字返回仅含名字的零值」对两条 miss
  路径同时成立
- #83 Skill.Name 字段文档写明两 Provider 口径差异（FileProvider=frontmatter
  展示名、Library=规范引用名）与「身份比对请用 Ref.Name」守则；双 Provider
  对照测试补 Name 断言钉住现状。统一口径是破坏式变更（FileProvider 侧被
  TestFrontmatterParsed 钉住），文档里给出走 breaking 变更的路径
测试按 issue 号命名新增 2 条回归用例；go test ./... 全绿。
Co-authored-by: Murphy-Yi <zhangyi@Murphy-YideMac-Mini.local>

### 修复(skill)：FileProvider 指纹先于读、RewriteBody 文档头与解析器同口径（#77 #78）

skillkit#77：FileProvider.Resolve 此前在读文件之后才取 mtime+size 指纹——
写落在「读结束→指纹 stat」窗内时，缓存条目持有旧内容却挂着竞态写之后的
指纹，fresh() 对比的是同一次写之后的盘面，永久命中，陈旧内容与陈旧
Checksum 无限期返回、无信号。指纹改为紧贴读前的 stat 快照：任何在指纹
之后完成的写都会让盘面与条目失配，下一次复验必然重载；stat 失败的候选
不读（无指纹可挂），#36 的错误记录口径不变。附 stat/read 测试缝供竞态
时序确定性编排（skill_test.go 注入，生产路径不触碰）。
skillkit#78：RewriteBody 此前用裸 HasPrefix(content, "---") 判文档头——
parseFrontmatter 与 RewriteMode 经 trimDocHead 接受的 BOM / 前导空白
SKILL.md（Windows 编辑器常见产物，stripBOM 存在的缘由）被编辑路径以
「无 frontmatter」误拒。入口对齐 RewriteMode：trimDocHead 后再判围栏，
围栏保留与 #66 坐标切分语义不变。

- 杂务(deps)：依赖地板追到最新——github.com/oekit/pkg v0.42.0

| 依赖 | 现地板 | 抬到 |
|---|---|---|
| `github.com/oekit/pkg` | `v0.41.1` | `v0.42.0` |
## 为什么
矩阵的**发版时刻地板时效**门禁（`verify_tag_floor --ref HEAD`）要求：各仓 HEAD
（= 下次发版会发出去的内容）的依赖地板，相对那一刻的依赖最新版本，不能已落后。
oekit 基座与矩阵模块都在高频发版 ⇒ 地板是**跑步机**，落后一格就会判红。
## 兼容性实测
本工具在提交前跑完 `go build` / `go vet` / `go test -race -count=1 ./...`（全绿才提交）
与 `go mod tidy -diff`（无漂移）⇒ 改动**只影响依赖版本**，零行为差异。
只改 `go.mod` / `go.sum` ⇒ 不需要 CHANGELOG 条目（版本段由下次发版的 `prewrite` 生成）。
Co-authored-by: Murphy-Yi <zhangyi@Murphy-YideMac-Mini.local>

- 杂务(deps)：依赖地板追到最新——github.com/oagentkit/textx v0.10.7

| 依赖 | 现地板 | 抬到 |
|---|---|---|
| `github.com/oagentkit/textx` | `v0.10.6` | `v0.10.7` |
## 为什么
矩阵的**发版时刻地板时效**门禁（`verify_tag_floor --ref HEAD`）要求：各仓 HEAD
（= 下次发版会发出去的内容）的依赖地板，相对那一刻的依赖最新版本，不能已落后。
oekit 基座与矩阵模块都在高频发版 ⇒ 地板是**跑步机**，落后一格就会判红。
## 兼容性实测
本工具在提交前跑完 `go build` / `go vet` / `go test -race -count=1 ./...`（全绿才提交）
与 `go mod tidy -diff`（无漂移）⇒ 改动**只影响依赖版本**，零行为差异。
只改 `go.mod` / `go.sum` ⇒ 不需要 CHANGELOG 条目（版本段由下次发版的 `prewrite` 生成）。

- 杂务(deps)：依赖地板追到最新——github.com/oagentkit/guardx v0.24.1

| 依赖 | 现地板 | 抬到 |
|---|---|---|
| `github.com/oagentkit/guardx` | `v0.24.0` | `v0.24.1` |
## 为什么
矩阵的**发版时刻地板时效**门禁（`verify_tag_floor --ref HEAD`）要求：各仓 HEAD
（= 下次发版会发出去的内容）的依赖地板，相对那一刻的依赖最新版本，不能已落后。
oekit 基座与矩阵模块都在高频发版 ⇒ 地板是**跑步机**，落后一格就会判红。
## 兼容性实测
本工具在提交前跑完 `go build` / `go vet` / `go test -race -count=1 ./...`（全绿才提交）
与 `go mod tidy -diff`（无漂移）⇒ 改动**只影响依赖版本**，零行为差异。
只改 `go.mod` / `go.sum` ⇒ 不需要 CHANGELOG 条目（版本段由下次发版的 `prewrite` 生成）。

- 杂务(deps)：依赖地板追到最新——github.com/oekit/pkg v0.41.1

| 依赖 | 现地板 | 抬到 |
|---|---|---|
| `github.com/oekit/pkg` | `v0.40.0` | `v0.41.1` |
## 为什么
矩阵的**发版时刻地板时效**门禁（`verify_tag_floor --ref HEAD`）要求：各仓 HEAD
（= 下次发版会发出去的内容）的依赖地板，相对那一刻的依赖最新版本，不能已落后。
oekit 基座与矩阵模块都在高频发版 ⇒ 地板是**跑步机**，落后一格就会判红。
## 兼容性实测
本工具在提交前跑完 `go build` / `go vet` / `go test -race -count=1 ./...`（全绿才提交）
与 `go mod tidy -diff`（无漂移）⇒ 改动**只影响依赖版本**，零行为差异。
只改 `go.mod` / `go.sum` ⇒ 不需要 CHANGELOG 条目（版本段由下次发版的 `prewrite` 生成）。


## [v0.10.4] — 2026-11-09

- 杂务(deps)：依赖地板追到最新——guardx v0.24.0 / textx v0.10.6

## [v0.10.3] — 2026-11-09

### 修复：skillkit 第三轮 issue 清零（#66 #67）

* fix(skill): RewriteBody 用真实头偏移切原串——混排 EOL（CRLF 起始围栏+LF 内行）不再坐标错位腐蚀 frontmatter（#66）
* fix(skill): use_skill 截断后 Checksum 对截断产物重算——自证锚点绑定实际返回内容，合法截断不再报假阳性（#67）
---------
Co-authored-by: Murphy-Yi <zhangyi@Murphy-YideMac-Mini.local>

- 杂务(deps)：依赖地板追到最新——github.com/oekit/pkg v0.40.0

| 依赖 | 现地板 | 抬到 |
|---|---|---|
| `github.com/oekit/pkg` | `v0.39.0` | `v0.40.0` |
## 为什么
矩阵的**发版时刻地板时效**门禁（`verify_tag_floor --ref HEAD`）要求：各仓 HEAD
（= 下次发版会发出去的内容）的依赖地板，相对那一刻的依赖最新版本，不能已落后。
oekit 基座与矩阵模块都在高频发版 ⇒ 地板是**跑步机**，落后一格就会判红。
## 兼容性实测
本工具在提交前跑完 `go build` / `go vet` / `go test -race -count=1 ./...`（全绿才提交）
与 `go mod tidy -diff`（无漂移）⇒ 改动**只影响依赖版本**，零行为差异。
只改 `go.mod` / `go.sum` ⇒ 不需要 CHANGELOG 条目（版本段由下次发版的 `prewrite` 生成）。
Co-authored-by: Murphy-Yi <zhangyi@Murphy-YideMac-Mini.local>

- 杂务(deps)：依赖地板追到最新——github.com/oagentkit/textx v0.10.5

| 依赖 | 现地板 | 抬到 |
|---|---|---|
| `github.com/oagentkit/textx` | `v0.10.4` | `v0.10.5` |
## 为什么
矩阵的**发版时刻地板时效**门禁（`verify_tag_floor --ref HEAD`）要求：各仓 HEAD
（= 下次发版会发出去的内容）的依赖地板，相对那一刻的依赖最新版本，不能已落后。
oekit 基座与矩阵模块都在高频发版 ⇒ 地板是**跑步机**，落后一格就会判红。
## 兼容性实测
本工具在提交前跑完 `go build` / `go vet` / `go test -race -count=1 ./...`（全绿才提交）
与 `go mod tidy -diff`（无漂移）⇒ 改动**只影响依赖版本**，零行为差异。
只改 `go.mod` / `go.sum` ⇒ 不需要 CHANGELOG 条目（版本段由下次发版的 `prewrite` 生成）。
Co-authored-by: Murphy-Yi <zhangyi@Murphy-YideMac-Mini.local>

- 杂务(deps)：依赖地板追到最新——github.com/oagentkit/guardx v0.23.4

| 依赖 | 现地板 | 抬到 |
|---|---|---|
| `github.com/oagentkit/guardx` | `v0.23.3` | `v0.23.4` |
| `github.com/oagentkit/textx` | `v0.10.3` | `v0.10.4` |
| `github.com/oekit/pkg` | `v0.38.0` | `v0.39.0` |
## 为什么
矩阵的**发版时刻地板时效**门禁（`verify_tag_floor --ref HEAD`）要求：各仓 HEAD
（= 下次发版会发出去的内容）的依赖地板，相对那一刻的依赖最新版本，不能已落后。
oekit 基座与矩阵模块都在高频发版 ⇒ 地板是**跑步机**，落后一格就会判红。
## 兼容性实测
本工具在提交前跑完 `go build` / `go vet` / `go test -race -count=1 ./...`（全绿才提交）
与 `go mod tidy -diff`（无漂移）⇒ 改动**只影响依赖版本**，零行为差异。
只改 `go.mod` / `go.sum` ⇒ 不需要 CHANGELOG 条目（版本段由下次发版的 `prewrite` 生成）。
Co-authored-by: Murphy-Yi <zhangyi@Murphy-YideMac-Mini.local>


## [v0.10.2] — 2026-11-08

### 修复(skill)：RewriteMode 未知 frontmatter 键原样保留 + FileProvider 缓存失效路径（#57 #56）

skillkit#57 RewriteMode 曾用 LibMeta 结构体 round-trip 切换 mode——
yaml.Unmarshal 忽略 schema 外键 + re-marshal 只写已知字段，SKILL.md
开放标准允许的 license/owner 等键被静默删除，与文档「其余元数据字段
原样保留」相悖。改为 yaml.Node 局部手术：只对 mode 标量原位改值（无则
追加），其余键连同键序/注释原样存活；文档头判定改用 trimDocHead 与解析
器同口径（BOM/前导空白不再被「无 frontmatter」误拒）。
skillkit#56 FileProvider.Resolve 缓存永久 sticky——盘上编辑（含宿主经
RewriteMode/RewriteBody 写回的编辑流）永不生效，陈旧 Checksum 毒化守卫。
缓存条目改为携带源文件 mtime+size 指纹，命中前锁外 stat 复验（一次 stat
换编辑-重载循环无进程级陈旧；粗粒度 FS 的等长同 mtime 改写由新增
Invalidate(name) 显式失效兜底）；条目不可变、并发无单飞——Resolve 不引入
全局串行点。

- 杂务(deps)：依赖地板追到最新——基座 v0.38.0/v0.29.0/v0.8.11 + 矩阵 tier0 新 tag（按本仓依赖取齐）

发版时刻地板时效门禁要求追平。
兼容性实测：go build / go vet / go test -race -count=1 ./... 全绿，go mod tidy -diff 无漂移。


## [v0.10.1] — 2026-11-08

- 杂务(deps)：依赖地板追到最新——github.com/oekit/pkg v0.37.4

pkg v0.37.4（httpjson NormalizeBaseURL scheme 大小写归一 #239 + refresh_versions）落地，地板同步追平（matrix-tools#188 发版链）。
兼容性实测：go build / go vet / go test -race -count=1 ./... 全绿，go mod tidy -diff 无漂移。

- 杂务(deps)：依赖地板追到最新——concurrency v0.28.3 / logx v0.8.8 / pkg v0.37.3 / textx v0.10.1 / guardx v0.23.1 / runtime v0.12.1（按本仓依赖取齐）

发版时刻地板时效门禁（verify_tag_floor --ref HEAD）要求追平（matrix-tools#188 发版链）。
兼容性实测：go build / go vet / go test -race -count=1 ./... 全绿，go mod tidy -diff 无漂移。

- 杂务(deps)：依赖地板追到最新——github.com/oagentkit/guardx v0.23.0

| 依赖 | 现地板 | 抬到 |
|---|---|---|
| `github.com/oagentkit/guardx` | `v0.20.0` | `v0.23.0` |
## 为什么
矩阵的**发版时刻地板时效**门禁（`verify_tag_floor --ref HEAD`）要求：各仓 HEAD
（= 下次发版会发出去的内容）的依赖地板，相对那一刻的依赖最新版本，不能已落后。
oekit 基座与矩阵模块都在高频发版 ⇒ 地板是**跑步机**，落后一格就会判红。
## 兼容性实测
本工具在提交前跑完 `go build` / `go vet` / `go test -race -count=1 ./...`（全绿才提交）
与 `go mod tidy -diff`（无漂移）⇒ 改动**只影响依赖版本**，零行为差异。
只改 `go.mod` / `go.sum` ⇒ 不需要 CHANGELOG 条目（版本段由下次发版的 `prewrite` 生成）。
Co-authored-by: Murphy-Yi <zhangyi@Murphy-YideMac-Mini.local>

- 杂务(build)：补统一 Makefile——本地 ci 门禁与共享 CI 同口径（agentkit#138）

Co-authored-by: Murphy-Yi <zhangyi@Murphy-YideMac-Mini.local>

- 杂务(deps)：依赖地板追到最新——github.com/oekit/pkg v0.37.1

| 依赖 | 现地板 | 抬到 |
|---|---|---|
| `github.com/oekit/pkg` | `v0.37.0` | `v0.37.1` |
## 为什么
矩阵的**发版时刻地板时效**门禁（`verify_tag_floor --ref HEAD`）要求：各仓 HEAD
（= 下次发版会发出去的内容）的依赖地板，相对那一刻的依赖最新版本，不能已落后。
oekit 基座与矩阵模块都在高频发版 ⇒ 地板是**跑步机**，落后一格就会判红。
## 兼容性实测
本工具在提交前跑完 `go build` / `go vet` / `go test -race -count=1 ./...`（全绿才提交）
与 `go mod tidy -diff`（无漂移）⇒ 改动**只影响依赖版本**，零行为差异。
只改 `go.mod` / `go.sum` ⇒ 不需要 CHANGELOG 条目（版本段由下次发版的 `prewrite` 生成）。
Co-authored-by: Murphy-Yi <zhangyi@Murphy-YideMac-Mini.local>


## [v0.10.0] — 2026-11-07

### 修复(skill)：remove_after 补历法校验 fail-closed；frontmatter 宽松解析改走 yaml.v3 解码标准值形态

## 概要
一个 PR 修两个 issue：#45（remove_after 历法校验）与 #46（frontmatter 宽松解析器的标准 YAML 值形态）。
close #45
close #46
## #45 remove_after 只验形态不验历法
**根因**：`skill/library.go` 的 `validateDeprecated` 只用 regex 验 `\d{4}-\d{2}-\d{2}` 形态，「2026-13-45」「2027-02-30」照常通过 Validate；下游 `DeprecationExpired` 对 `time.Parse` 失败恒返回 false——非法日期 = 弃用窗口永不过期（静默无限期弃用窗口），lint/reload 的过期判定永远不触发。
**修法**：`Validate` 在形态检查之后追加 `time.Parse("2006-01-02")` 真实历法日校验，非法历法日（13 月 45 日、2 月 30 日、非闰年 2 月 29 日等）加载期即报错，fail-closed 不静默忽略；真实闰日（2028-02-29）不受误伤。`DeprecationExpired` 的解析失败分支保留为防御路径并注明上游闸门（Library 加载路径到不了）。
**测试**（`skill/lifecycle_test.go::TestValidateRemoveAfterCalendar`）：三类非法历法日报「非法日期」、闰年 2 月 29 日放行、非法历法日经 `LoadFromFS` 端到端 fail-fast。
## #46 FileProvider 宽松解析器把标准 YAML 值形态解析成垃圾
**根因**：`skill/frontmatter.go` 的 `parseFrontmatter` 是手写逐行 `strings.Cut(":")` + `trimQuoted` 的宽松解析器，只认单行标量——块标量 `description: |` 的值是字面 `"|"`（多行内容静默全丢）、`>-` 折叠同理、引号值带行尾注释则引号+注释整串泄漏进发现清单（`ListSkills` → `Meta` → `ListPrompt`，模型按描述决策 use_skill 的依据被静默摧毁）；且解析「成功」故 OnCorrupt（#36 观测面）不触发，零信号。
**修法**：围栏切分保持不变（宽松策略保留：无/未闭合围栏仍整份当正文，FileProvider 单文件损坏不阻塞发现），围栏内的值解码改走仓内既有 `gopkg.in/yaml.v3`（与 `ParseRichFrontmatter` 同源，不新增依赖）——新增 `fmLite` 小结构以 `yaml.Node` 承接 name/description/version 三键（其余键仍忽略），块标量（| / > 及缩进体）、单双引号（含转义）、行尾注释、空值、整数 tag 等标准 YAML 形态均正确解出，与严格面同一 YAML 语义；非标量节点返回空串，不产出 `"{}"`/`"[]"` 类垃圾值。块整体非合法 YAML（如未闭合引号、tab 缩进、重复键）时回落旧逐行宽松解析（发现不阻塞），并返回 warn，由 `loadSkill` 经 `OnCorrupt` 上报——降级产出可疑值不再零信号。
**测试**（`skill/frontmatter_test.go`）：
- `TestParseFrontmatterStandardYAMLValueForms`：8 类标准值形态表驱动断言（含宽松/严格两面 description 一致性——块标量含缩进 `---` 分隔线的完整口径）；
- `TestFileProviderListSkillsStandardYAMLFrontmatter`：发现清单端到端，块标量描述完整进 `Meta`，`ListPrompt` 不再渲染字面 `|`；
- `TestFileProviderFrontmatterFallbackWarn`：非合法 YAML 触发 OnCorrupt 降级警告（带定位与原因），清单不缺项、逐行可解的键仍保留。
## 契约遵守
- `LoadVocab` 的 yaml KnownFields 严格解码未动；`ParseRichFrontmatter` 语义未动。
- 围栏切分（`splitFrontmatter`/`findClosingFence`，#33 唯一事实源）未动。
- 未动 CHANGELOG.md / go.mod / AGENTS.md / README / matrix-tools；未打 tag。
## 本地验证
`gofmt -l .` 为空、`go vet ./...`、`go build ./...`、`go test ./...` 全绿。

- 杂务(deps)：依赖地板追到最新——github.com/oekit/pkg v0.37.0

| 依赖 | 现地板 | 抬到 |
|---|---|---|
| `github.com/oekit/pkg` | `v0.36.0` | `v0.37.0` |
## 为什么
矩阵的**发版时刻地板时效**门禁（`verify_tag_floor --ref HEAD`）要求：各仓 HEAD
（= 下次发版会发出去的内容）的依赖地板，相对那一刻的依赖最新版本，不能已落后。
oekit 基座与矩阵模块都在高频发版 ⇒ 地板是**跑步机**，落后一格就会判红。
## 兼容性实测
本工具在提交前跑完 `go build` / `go vet` / `go test -race -count=1 ./...`（全绿才提交）
与 `go mod tidy -diff`（无漂移）⇒ 改动**只影响依赖版本**，零行为差异。
只改 `go.mod` / `go.sum` ⇒ 不需要 CHANGELOG 条目（版本段由下次发版的 `prewrite` 生成）。

- 杂务(deps)：依赖地板追到最新——github.com/oagentkit/guardx v0.20.0

| 依赖 | 现地板 | 抬到 |
|---|---|---|
| `github.com/oagentkit/guardx` | `v0.19.0` | `v0.20.0` |
| `github.com/oagentkit/textx` | `v0.9.0` | `v0.10.0` |
| `github.com/oekit/pkg` | `v0.34.1` | `v0.36.0` |
## 为什么
矩阵的**发版时刻地板时效**门禁（`verify_tag_floor --ref HEAD`）要求：各仓 HEAD
（= 下次发版会发出去的内容）的依赖地板，相对那一刻的依赖最新版本，不能已落后。
oekit 基座与矩阵模块都在高频发版 ⇒ 地板是**跑步机**，落后一格就会判红。
## 兼容性实测
本工具在提交前跑完 `go build` / `go vet` / `go test -race -count=1 ./...`（全绿才提交）
与 `go mod tidy -diff`（无漂移）⇒ 改动**只影响依赖版本**，零行为差异。
只改 `go.mod` / `go.sum` ⇒ 不需要 CHANGELOG 条目（版本段由下次发版的 `prewrite` 生成）。


## [v0.9.0] — 2026-11-06

### 修复(skill,lineage)：Provider 语义五修——版本约束/错误区分/别名缓存/确定性/时区口径（#35-#39）

skillkit#35——scanAliases 失败不缓存 + 首扫并发收敛：一次瞬时目录错误此前
把空别名表缓存进 p.aliases（展示名归一进程内永久失效）；失败态下次调用
重试，成功态快路径读缓存。
skillkit#36——loadSkill 区分 fs.ErrNotExist 与其余 ReadFile 错误：权限拒绝
不再误诊为「未找到」（全候选失败且有非NotExist错误时原样上抛）；
ListSkills 损坏条目经新回调 OnCorrupt 留痕（本仓 tier 纪律不引日志依赖，
宿主接线告警面；静默丢弃让清单缺一条无从归因）。
skillkit#37——Ref.Version 请求约束兑现（两 Provider 都校验）：此前
FileProvider 只当缓存键、Library 完全忽略，请求 2.0.0 静默返回 1.0.0；
doc 从「声明无实现」改为「声明+实现一致」。精确匹配；区间约束走 Decl 面。
skillkit#38——sortFocus 边比较器补 To.Kind/Relation 全键：同名异类节点间
序此前不稳定（sort.Slice 不保证稳定序），跨 run 不可复现违反确定性契约。
32 边 ×16 轮逐边相等钉子。
skillkit#39——DeprecationExpired 改 UTC 解析：remove_after 裸日期此前按
time.Local，过期边界随宿主时区漂移最多一天且与 agent-memory 的 UTC 口径
跨包两套哲学。
测试五组钉子；既有 checksum 测试的 9.9.9 请求改为不约束（版本约束生效后
该请求按新语义本就该报错）。全仓 -race 绿。
关联 oagentkit/skillkit#35 oagentkit/skillkit#36 oagentkit/skillkit#37 oagentkit/skillkit#38 oagentkit/skillkit#39
Co-authored-by: Murphy-Yi <zhangyi@Murphy-YideMac-Mini.local>

### 修复(skill)：ListPrompt Title 消毒 + RewriteBody 围栏语义归一

* fix(skill): ListPrompt Title 消毒 + RewriteBody 围栏语义归一（#33 #34）
skillkit#34——ListPrompt 的 doc 声称 Name/Title/Description 同属半可信、
「消毒在出口统一兜底」，但只有 desc 过 EscapeUntrusted：Title（frontmatter
name 引号标量可含 \n 转义）裸拼进行首括注——description 里的注入被中和、
title 里的原样越出条目成为独立标题/伪列表项，同一威胁模型开了旁路。
Title 先折叠单行再消毒（折叠后的行首结构标记由 EscapeUntrusted 中和）。
skillkit#33——RewriteBody 按 isFenceLine（TrimSpace）逐行判闭合，与注释
宣称的 findClosingFence 同一约定相悖：frontmatter 内缩进的 ---（YAML 块
标量分隔线）被误当闭合，编辑路径截断元数据（Section B/version 全丢）且
输出不可再解析；同输入 ParseRichFrontmatter 正常。RewriteBody 改走
findClosingFence（围栏语义唯一事实源），isFenceLine 的「唯一事实源」注释
随之修正为「起始围栏语义」。
TestRewriteBodyKeepsIndentedFenceInBlockScalar（截断回归+输出可再解析）/
TestRewriteBodyClosesAtColumnZeroFence（列 0 常规路径不回归）/
TestListPromptTitleSanitized（注入越出条目钉子+中和不删内容+正常 Title
不受影响）。全仓 -race 绿。
关联 oagentkit/skillkit#33 oagentkit/skillkit#34
* style: gofmt 对齐追加测试
* chore: 删除已无调用方的 isFenceLine（RewriteBody 归一 findClosingFence 后退役；围栏语义唯一事实源收口）
* docs: findClosingFence 注释自足（isFenceLine 退役后不再互相引用）
---------
Co-authored-by: Murphy-Yi <zhangyi@Murphy-YideMac-Mini.local>

- 杂务(deps)：依赖地板追到最新——github.com/oagentkit/guardx v0.19.0

| 依赖 | 现地板 | 抬到 |
|---|---|---|
| `github.com/oagentkit/guardx` | `v0.18.0` | `v0.19.0` |
| `github.com/oagentkit/textx` | `v0.8.0` | `v0.9.0` |
## 为什么
矩阵的**发版时刻地板时效**门禁（`verify_tag_floor --ref HEAD`）要求：各仓 HEAD
（= 下次发版会发出去的内容）的依赖地板，相对那一刻的依赖最新版本，不能已落后。
oekit 基座与矩阵模块都在高频发版 ⇒ 地板是**跑步机**，落后一格就会判红。
## 兼容性实测
本工具在提交前跑完 `go build` / `go vet` / `go test -race -count=1 ./...`（全绿才提交）
与 `go mod tidy -diff`（无漂移）⇒ 改动**只影响依赖版本**，零行为差异。
只改 `go.mod` / `go.sum` ⇒ 不需要 CHANGELOG 条目（版本段由下次发版的 `prewrite` 生成）。
Co-authored-by: Murphy-Yi <zhangyi@Murphy-YideMac-Mini.local>

- 杂务(deps)：依赖地板追到最新——本轮新 tag（首轮 tidy 撞上新发版 proxy 延迟，重跑补完）

Co-authored-by: Murphy-Yi <zhangyi@Murphy-YideMac-Mini.local>

- 杂务(deps)：依赖地板追到最新——github.com/oekit/pkg v0.34.1

| 依赖 | 现地板 | 抬到 |
|---|---|---|
| `github.com/oekit/pkg` | `v0.34.0` | `v0.34.1` |
## 为什么
矩阵的**发版时刻地板时效**门禁（`verify_tag_floor --ref HEAD`）要求：各仓 HEAD
（= 下次发版会发出去的内容）的依赖地板，相对那一刻的依赖最新版本，不能已落后。
oekit 基座与矩阵模块都在高频发版 ⇒ 地板是**跑步机**，落后一格就会判红。
## 兼容性实测
本工具在提交前跑完 `go build` / `go vet` / `go test -race -count=1 ./...`（全绿才提交）
与 `go mod tidy -diff`（无漂移）⇒ 改动**只影响依赖版本**，零行为差异。
只改 `go.mod` / `go.sum` ⇒ 不需要 CHANGELOG 条目（版本段由下次发版的 `prewrite` 生成）。
Co-authored-by: Murphy-Yi <zhangyi@Murphy-YideMac-Mini.local>

- 杂务(deps)：依赖地板追到最新——github.com/oagentkit/guardx v0.17.1

| 依赖 | 现地板 | 抬到 |
|---|---|---|
| `github.com/oagentkit/guardx` | `v0.17.0` | `v0.17.1` |
| `github.com/oagentkit/textx` | `v0.7.1` | `v0.7.2` |
| `github.com/oekit/pkg` | `v0.31.0` | `v0.34.0` |
## 为什么
矩阵的**发版时刻地板时效**门禁（`verify_tag_floor --ref HEAD`）要求：各仓 HEAD
（= 下次发版会发出去的内容）的依赖地板，相对那一刻的依赖最新版本，不能已落后。
oekit 基座与矩阵模块都在高频发版 ⇒ 地板是**跑步机**，落后一格就会判红。
## 兼容性实测
本工具在提交前跑完 `go build` / `go vet` / `go test -race -count=1 ./...`（全绿才提交）
与 `go mod tidy -diff`（无漂移）⇒ 改动**只影响依赖版本**，零行为差异。
只改 `go.mod` / `go.sum` ⇒ 不需要 CHANGELOG 条目（版本段由下次发版的 `prewrite` 生成）。
Co-authored-by: Murphy-Yi <zhangyi@Murphy-YideMac-Mini.local>

- 杂务(deps)：依赖地板追到最新——github.com/oagentkit/guardx v0.17.0

guardx v0.17.0（2026-10-28 发布，含 logredact 载荷裁剪档与 invoke.Outcome 第三格）⇒ 本仓的地板落后 ⇒ 抬上来。
矩阵的**发版时刻地板时效**门禁（`verify_tag_floor --ref HEAD`，宽口径）要求：**各仓 HEAD**
（= 下次发版会发出去的内容）的依赖地板，相对那一刻的依赖最新版本，不能已落后。
兼容性实测：本仓只用这些依赖的少量 API ⇒ go build / go vet / go test -race 全绿，
**零行为差异**；go 指令未被顶起。
只改 go.mod / go.sum ⇒ 不需要 CHANGELOG 条目（版本段由下次发版的 `prewrite` 生成）。

- 杂务(deps)：依赖地板追到最新——github.com/oekit/pkg v0.31.0

oekit 基座发了新版本（pkg v0.31.0 / concurrency v0.23.0 / logx v0.8.0）⇒ 本仓的
依赖地板落后 ⇒ 抬到最新，让**下一次发版**不再带出落后地板（打 tag 时刻的地板
必须是当时的最新——`verify_tag_floor` / `release.py doctor` 都在查这条）。
兼容性实测：本仓只用这些基座包的少量 API ⇒ go build / go vet / go test -race
全绿，**零行为差异**；go 指令未被顶起（本仓的 go 地板不动）。

- 杂务(deps)：依赖地板追平上游本轮新版本——textx v0.7.1

本仓 HEAD 的依赖地板已落后于 textx 的最新版本（`verify_tag_floor.py --org oagentkit --ref HEAD` 判红）。
本仓一旦发版，就会把落后地板发出去——而 tag 受 protect-tags ruleset 保护、打完不能重打，
只能靠后续版本补齐。
本版只抬依赖地板，不改任何行为：
- github.com/oagentkit/textx v0.6.2 → v0.7.1
对应 issue matrix-tools#143。

- 杂务(deps)：依赖地板追平上游本轮新版本——github.com/oagentkit/guardx@v0.15.0、github.com/oekit/pkg@v0.28.0

本仓 HEAD 的依赖地板已落后于各依赖的最新版本（`verify_tag_floor.py --org oagentkit --ref HEAD` 判红）。
本仓一旦发版，就会把落后地板发出去——而 tag 受 protect-tags ruleset 保护、打完不能重打，
只能靠后续版本补齐。
本版只抬依赖地板，不改任何行为：
- github.com/oagentkit/guardx v0.15.0 → v0.15.0
- github.com/oekit/pkg v0.28.0 → v0.28.0
对应 issue matrix-tools#143。


- 文档(skillkit)：刷正 README 的外部依赖清单 —— 与 go.mod 实测对账

本仓 README 的「外部依赖」清单与**各仓 go.mod 的事实**对不上。
矩阵的单一数据源 `matrix-tools/modules.json` 也有同样漂移，已在
[matrix-tools#126](https://github.com/oagentkit/matrix-tools/pull/126) 同步刷正。
## 现状核实
```bash
$ grep -A8 "^require (" skillkit/go.mod | grep -v indirect
	github.com/Masterminds/semver/v3 v3.3.1
	github.com/cloudwego/eino v0.9.21
	github.com/oagentkit/guardx v0.14.0
	github.com/oagentkit/textx v0.6.2
	github.com/oekit/pkg v0.22.0
	gopkg.in/yaml.v3 v3.0.1
```
⇒ 第三方直接依赖是 **3 个**（semver/v3、**eino**、yaml.v3），本仓 README 只列了 2 个。
## 影响
`eino` 是重量级地板（拉 sonic 等一整套）。README 的「外部依赖地板」小节漏了它 ⇒ 读者按文档判断「这仓只吃 yaml + semver」会**低估**依赖面。
## 改动
**只改清单那一行（`- `gopkg.in/yaml.v3`` / `- `github.com/Masterminds/semver/v3`` ⇒ 多一行 `- `github.com/cloudwego/eino``）**，其余段落逐字未动。
## 验收
- [x] 清单与 `go.mod` 实测一致
- [x] 与 `matrix-tools/modules.json` 的登记一致（同期在 matrix-tools#126 刷正）
- [x] **纯 .md 改动** ⇒ 不需要 CHANGELOG 条目、不打 tag、不发版（无发版步骤）
关联 oagentkit/matrix-tools#125

- 文档(skillkit)：刷正 README.md 的层位声明——tier 1 → tier 2

## 现状
本仓 `README.md` 的两处层位声明写的是**层位重定位之前**的旧层位（tier 1），
而矩阵单一数据源 `matrix-tools/modules.json` 说本仓是 **tier 2（集成层）**。
层位重定位的依据见 matrix-tools 提交 `6d800c4`「层位 = 依赖地板」——
tier 2 的定义是「依赖 tier0/tier1 + 重第三方地板」（`modules.json` 的 `tiers` 映射）。
| 位置 | 改前 | 改后 |
|---|---|---|
| `README.md` 第 3 行（定位声明行） | `**tier 1**` | `**tier 2**` |
| `README.md` 第 17 行（`## 定位` 自述行） | `**tier 1**` | `**tier 2**` |
- `modules.json`：`skillkit` 的 `tier` = **2**，`tiers["2"]` = 集成层
- 本仓 `AGENTS.md` 第 5 行：**已是** `> 矩阵分层：**tier 2**（集成层，依赖 tier0/tier1 + 重第三方地板）`
  ⇒ 改之前，本仓的**两个第一入口互相矛盾**（`AGENTS.md` 说 tier 2、`README.md` 说 tier 1）
## 影响
`README.md` 是仓级的第一入口（人读 + agent 读）。它写 tier 1 而数据源是 tier 2 ⇒
读者按 README 判断依赖方向就会算错，而分层是矩阵的**硬不变量**（tier N 不得依赖 tier > N）。
## 改动
**只改层位数字与层名，其余段落逐字不动**：
- 第 3 行：`**tier 1**` → `**tier 2**`，层名 `（能力层，依赖 tier0）` → `（集成层，依赖 tier0/tier1 + 重第三方地板）`
- `## 定位`：`**tier 1**` → `**tier 2**`，层名 `能力层` → `集成层`
层名与整段括号说明对齐到 tier 2 的权威表述后，与本仓 `AGENTS.md`
及 `llm` / `toolgate` / `tools` / `agent` 四个 tier 2 仓的 README 逐字一致。
diff 只含层位行，`README.md` 的手写段落逐字未动。
## 验收
- [x] `python3 matrix-tools/scripts/verify_repo_tiers.py --root .` ⇒ 本仓 README **红 0**（本地实跑，26 处声明全一致）
- [x] 层位数字与层名与 `modules.json` 一致，且与本仓 `AGENTS.md` 一致
- [x] **纯 .md 改动** ⇒ 不需要 CHANGELOG 条目、不打 tag、不发版（无发版步骤）
关联 oagentkit/matrix-tools#123

- 文档(skillkit)：刷正 AGENTS.md 层位声明——tier 1 → tier 2

AGENTS.md 的层位声明与矩阵单一数据源 modules.json 不一致。
- 矩阵分层：**tier 1**（能力层，依赖 tier0）→ **tier 2**（集成层，依赖 tier0/tier1 + 重第三方地板）
- 「本仓是 tier 1，只允许依赖 tier ≤ 1 的仓」→ tier 2
modules.json 里 skillkit 的 tier = 2（层名「集成层」；层位 t1 → t2 时带 eino 地板，
skill 的工具决策用 eino tool schema）。层位写错会让人按 tier 1 的口径判依赖方向，
而分层是矩阵的硬不变量。
属 matrix-tools#109 的仓级层位漂移；同批修 queue / obs / mcpkit / rag。
diff 只含层位行，AGENTS.md 的手写段落逐字未动。


## [v0.8.5] — 2026-10-24

- 杂务(deps)：依赖地板跟到矩阵本轮新版本——guardx v0.14.0 / oekit/pkg v0.22.0

## 下游要做什么
**`go mod tidy`**。地板抬高对下游是**隐式破坏性变更**：不跑就报
`updates to go.mod needed`，且报错不提真正原因（**不是本仓的 API 变了**）。
本仓**自身代码零改动**——只看本仓的 diff 看不出这是一次地板抬高。
## 为什么抬
上一版 tag（v0.8.4）打出去时，它依赖的模块的**新版已经先发出**了——
即「打 tag 那一刻地板就落后」。矩阵的「发版时刻地板时效」门禁
（`verify_tag_floor.py --org oagentkit --latest-only`）对下一次发版会判红：
- github.com/oagentkit/guardx@v0.14.0
- github.com/oekit/pkg@v0.22.0
门禁给的修法是「抬 go.mod 地板 → go mod tidy → 重新发 tag」。
本 PR 就是那条修法。
## 为什么无行为差异
本仓只 import 这两个依赖的两个包：
`github.com/oagentkit/guardx/fence`（skill/decision.go）与
`github.com/oekit/pkg/encoding`（skill/skill.go）。
而上游两版的差异分别落在 `permgate/`（guardx）与 `pathx/` + `stringx/`（oekit/pkg）
——都不在本仓的 import 面内 ⇒ 用到的转发面与符号集一字未动。
## 门禁
```
go build / vet / test -race   ✅
go mod tidy -diff             ✅ 空
gofmt -l .                    ✅ 空
```


## [v0.8.4] — 2026-10-21

- 杂务(deps)：依赖地板跟到矩阵本轮新版本（第二轮：消除打 tag 时刻的地板落后）

* chore(deps): 依赖地板跟到矩阵本轮新版本——textx v0.6.2
## 下游要做什么
**`go mod tidy`**。地板抬高对下游是**隐式破坏性变更**：不跑就报
`updates to go.mod needed`，且报错不提真正原因（**不是本仓的 API 变了**）。
本仓**自身代码零改动**——只看本仓的 diff 看不出这是一次地板抬高。
## 为什么抬
上一版 tag 打出去时，它依赖的模块的**新版已经先发出**了——
即「打 tag 那一刻地板就落后」。矩阵的「发版时刻地板时效」门禁
（`verify_tag_floor.py --org oagentkit --latest-only`）当场报红：
- github.com/oagentkit/textx@v0.6.2
门禁给的修法是「抬 go.mod 地板 → go mod tidy → 重新发 tag」。
本 PR 就是那条修法。
## 为什么无行为差异
被抬的这几个版本**自身也都是纯依赖地板追平**（各仓代码零改动，只抬
`oekit/pkg` / `oekit/concurrency` / `guardx` / `llm` 四类地板之一），
本仓用到的转发面与符号集一字未动。
## 门禁
```
go build / vet / test -race   ✅
gofmt -l .                    ✅ 空
```
* docs(changelog): 重生成 CHANGELOG.md——补回浅克隆期间漏掉的条目
上一版 tag（v0.8.3）的 CHANGELOG 段**漏了 2 条提交**：本地克隆是浅克隆
（depth 限制），生成器遍历 tag 区间时看不到更早的提交，于是那两条被漏掉。
本次重生成（本地已补齐完整历史）把 v0.8.3 段补回：
- 杂务(deps)：oagentkit/guardx 依赖地板抬到 v0.9.0
- 文档(AGENTS)：补「发版顺序」小节，指向 org CONTRIBUTING 单一副本
并预写 v0.8.4 待发版段（依赖地板跟到 textx v0.6.2）。
**为什么必须修**：下游 `go get` 拿到的是 tag 快照——快照里的 CHANGELOG 若
缺条目，发布说明就是残缺的（矩阵的 CHANGELOG 覆盖门禁正是查这个）。


## [v0.8.3] — 2026-10-21

- 杂务(deps)：依赖地板追平上游本轮新版本——oagentkit/guardx v0.13.1 oekit/pkg v0.20.0

## 下游要做什么
**`go mod tidy`**。地板抬高对下游是隐式 breaking：不跑就报
`updates to go.mod needed`，且报错不提真正原因（**不是本仓的 API 变了**）。
本仓**自身代码零改动**——只看本仓的 diff 看不出这是一次地板抬高。
## 抬了哪几块、为什么无行为差异
**`github.com/oagentkit/guardx` v0.13.0 → v0.13.1**
上游 diff 只有 `go.mod` / `go.sum` / `CHANGELOG.md`——**没有 Go 源码变更**
（内容是它自己把 `oekit/pkg` 地板抬到 v0.20.0）。本仓用到的包与转发面一字未动。
**`github.com/oekit/pkg` v0.18.0 → v0.20.0**
上游 diff 只动了 5 个包：
- `errors`：**纯增量**——新增 `Errorf`（格式化 + `%w` 包装，nil-safe）。
  `New`/`As`/`Is`/`Unwrap`/`Wrap` 是同型透传，一字未动。
- `httpjson`：**退避溢出修复**——旧写法 `base << (attempt-1)` 在 attempt 较大时
  int64 溢出成负数，负数绕过 `MaxDelay` 封顶，`time.After(负数)` 立刻返回
  ⇒ **该轮退避彻底消失**。改为乘法 + 饱和封顶。正常 attempt 数（1..20）下
  结果**逐字相同**，差别只在溢出的病态路径。
- `urlguard`：**纯增量**——新增 `WithExactHosts`，未设置时行为与既有完全一致。
- `stringx`：**仅注释**（补 `TruncEllipsis` 的 n 语义说明）。
- `pathx`：新增包（`SanitizeRel` 等）。
**本仓实际 import 的 `oekit/pkg` 子包：encoding **
与上游改动相交的：**（无）**
## 门禁
```
go build / vet / test -race   ✅
gofmt -l .                    ✅ 空
```

- 杂务(deps)：guardx 依赖地板抬到 v0.13.0

本仓只用 guardx 的既有包（fence），而它在 v0.11.0 → v0.13.0 之间**未变**
（那几版只加了 sandbox 的 Capable 与 wire）——故这是纯地板追平，无行为差异。
矩阵的依赖地板时效门禁（verify_dep_floors.py）会把「地板落后远端最新」报出来。
门禁自检：go build / vet / test ✅，gofmt ✅，go mod tidy ✅

- 杂务(deps)：oagentkit/guardx 依赖地板抬到 v0.11.0

guardx v0.11.0 新增 sandbox/wire 跨进程调用协议，是**新增能力**而非 API 变更——
本仓无需改代码，只抬地板。
为什么现在补：guardx v0.11.0 是本轮刚发布的，本仓上一个 tag 早于它，
于是这一格欠账一直挂着。依赖地板报告器把「子仓发版让下游落后」显式化，
但报告器恒不失败——真正的硬约束是「打 tag 时刻地板必须是当时的最新」。
本地门禁：go mod tidy 无差异 / go build / go vet 通过。

- 杂务(deps)：oagentkit/guardx 依赖地板抬到 v0.9.0

guardx v0.9.0 新增了 toolfail 与 sandbox 两个 tier0 原语，
是**新增能力**而非 API 变更——本仓无需改代码，只抬地板。
为什么现在补：guardx v0.9.0 是 2026-10-21 发布的，本仓上一个 tag 早于它，
于是这一格欠账一直挂着（依赖地板报告器把「子仓发版让下游落后」显式化，
但报告器恒不失败——真正的硬约束是「打 tag 时刻地板必须是当时的最新」）。
本地门禁：go mod tidy 无差异 / go build / go vet 通过。


- 文档(AGENTS)：补「发版顺序」小节，指向 org CONTRIBUTING 单一副本

13 个业务仓的 AGENTS.md 此前对发版顺序零覆盖（只有 textx 有个
「### 版本与发版」标题，但里面只讲 SemVer 与 module 路径），
发版顺序只写在 org 规则仓的 .github/CONTRIBUTING.md 里，单仓读者看不到。
本次只加**指针 + 本仓相关提醒**，runbook 本体保持**单一副本**
（CONTRIBUTING §发版），避免 13 份副本各自漂移。
覆盖四条：
- CHANGELOG.md 由 matrix-tools/scripts/gen_changelogs.py 从 git 历史生成，不要手改
- tag 必须落在 CHANGELOG 提交之后（下游 go get 拿到的是 tag 快照）
- 警告：无参重生成会吞掉待发版段，--next-version 是唯一能产出待发版段的开关
- 本仓相关提醒（见正文）


## [v0.8.2] — 2026-10-20

- 杂务(deps)：依赖地板抬到最新

地板对齐到发版时刻的各依赖最新 tag：
- oekit/pkg、oekit/concurrency 抬到 v0.18.0（oekit/logx 已是最新 v0.7.1）
- 矩阵内跨仓依赖抬到各仓最新 tag
依据 matrix-tools#44：最新 tag 的依赖地板在「打 tag 那一刻」就已落后一格，
只能靠后续版本补齐（tag 受 protect-tags ruleset 保护不能重打）。


## [v0.8.1] — 2026-10-20

### 修复(go)：Go 地板降到 go 1.24.0，CI go-version 同步到 1.24

承接 matrix-tools#22（tier0 三仓的地板被无谓抬高）——本轮把 tier1+ 里
**依赖链全兼容**的仓也降下来（matrix-tools#34 的排查结论）。
## 依据
- **依赖链上没有要求更高版本的包**：本仓的每个依赖（直接与间接）的 go 指令
  都 ≤ 1.24.x，本仓声明 1.26.0 没有任何依赖需要它。
- **代码本身兼容 1.24**：真实 go1.24.0 工具链（`GOTOOLCHAIN=go1.24.0`）
  build + test 全绿——实跑，不是推断。
- **门槛是 go 指令不是 toolchain**（编译器实证）：`go 1.28.0` 被拒，
  `go 1.24.0` + `toolchain go1.28.0` 通过。
## toolchain 行的取舍：删掉
toolchain **不构成门槛**，但 `toolchain go1.27.2` 会让 `go` 命令**自动下载**
1.27.2，与「最低支持版本是 1.24」相悖。删掉，也与基座 oekit 六个基础仓一致。
## 依赖地板同步重钉
降 go 指令后，`require` 必须指向**已发新 tag 的 tier0 仓**（它们的 tag 快照
里才是 go 1.24.0）——否则 MVS 会拉到旧 tag 的 1.26 地板，降级形同虚设。
本仓的内部 require 已随 `go mod tidy` 一并抬到对应的新 tag。
## 改动
- `go.mod`：`go 1.26.0` → `go 1.24.0`；删 `toolchain go1.27.2`；
  内部 require 抬到 tier0 新 tag
- `.github/workflows/ci.yml`：`go-version: '1.26'` → `'1.24'`
CI 的 go-version 与 go.mod 的 go 指令**成对**——这是跨层通用纪律，不因本次
降级而放松。CI 用高版本构建、go.mod 承诺低版本时，CI 全绿但用户按最低版本
编译失败，门禁绿了而承诺没兑现。
## 验证
- `GOTOOLCHAIN=go1.24.0 go build ./...` → exit 0
- `GOTOOLCHAIN=go1.24.0 go test ./...` → 全绿
- `go mod tidy` → 零差异
门禁见 oagentkit/matrix-tools：`scripts/verify_go_floor.py`（判据 4 覆盖
tier1+ 相对地板，判据独立不复用任何生成器）。

## [v0.8.0] — 2026-10-19

- 杂务(deps)：依赖地板对齐远端最新（guardx v0.5.0 / textx v0.5.0） (13)

matrix-tools 的 verify_dep_floors.py 首次全矩阵扫描查出的地板漂移，本仓是受影响仓之一。
本轮按 tier 拓扑顺序推进：tier0（textx / runtime / guardx / queue）先发，
本仓再把地板跟到它们的新 tag——否则就是「发了新的、旧的立刻变落后」，永远收敛不了。
所升版本与原钉版本相比源码零改动，逐版实测确认：
- textx v0.4.0→v0.5.0 / runtime v0.6.0→v0.7.0 / guardx v0.4.1→v0.5.0
  全树逐字节相同（diff -rq 退出码 0）
- oekit/pkg v0.14.0→v0.15.1  仅新增 urlguard 包（纯标准库，pkg 的 go.mod
  连 require 块都没有，零新增三方依赖）
- oekit/concurrency v0.17.0→v0.17.1  仅三个文档文件
- oekit/logx v0.7.0→v0.7.1  仅 CI 配置与文档
无 API 变更、无兼容风险。
本地门禁：go mod tidy -diff 无输出 / go build / go vet / gofmt / go test -race 全通过。
关联 oagentkit/matrix-tools#13


## [v0.7.0] — 2026-10-19

- 杂务(deps)：依赖地板对齐远端最新（pkg v0.15.1）

matrix-tools 的 verify_dep_floors.py 首次全矩阵扫描查出：43 条第一方依赖里
28 条落后，横跨 11 个仓；本仓是其中之一。
所升版本与本仓原钉版本相比源码零改动，逐版实测确认：
- textx v0.3.0→v0.4.0 / obs v0.4.1→v0.5.0 / runtime v0.5.0→v0.6.0 /
  mcpkit v0.5.1→v0.6.0  全树逐字节相同（diff -rq 退出码 0）
- oekit/pkg v0.14.0→v0.15.1  仅新增 urlguard 包（纯标准库，pkg 的 go.mod
  连 require 块都没有，零新增三方依赖）
- oekit/concurrency v0.17.0→v0.17.1  仅三个文档文件
- oekit/logx v0.7.0→v0.7.1  仅 CI 配置与文档
无 API 变更、无兼容风险。
本地门禁：go mod tidy -diff 无输出 / go build / go vet / gofmt / go test -race 全通过。


## [v0.6.1] — 2026-10-19

- 杂务(deps)：依赖地板跟进 guardx v0.4.1 / textx v0.4.0

guardx v0.3.0 → v0.4.1：拿到 egress 补 0.0.0.0/8 的 SSRF 修复
（oagentkit/guardx#2）。本仓只用 guardx/fence，而 v0.3.0 与 v0.4.0 指向
同一 commit、fence 与 logredact 零变更，所以这次升级**不含任何行为变化**，
是纯粹的地板滞后清理。
textx v0.3.0 → v0.4.0：同样落后的地板，顺带跟进。
go.sum 仅这两个模块的版本替换，三方依赖一条未动。


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

- 杂务(deps)：依赖地板升到「历史最新 + 0.1.0」

本轮做了版本线重置：清空全部历史 tag，按「历史最新 + 0.1.0」重打新版本
（如 llm v0.6.0 → v0.7.0、toolgate v0.6.1 → v0.7.0、agentkit v1.3.0 → v1.4.0）。
各仓 go.mod 的矩阵内依赖地板同步跟进到新版本号。
**go.sum 的 oagentkit 行已整段删除**，由 CI 的 `go mod tidy` 重算：
新 tag 的内容与任何历史版本都不同，`h1` 必须现算；本地没有 Go 工具链，
且 `/go.mod` 的 dirhash 无法自行复现（zip hash 可复现，gomod hash 不可复现）。
重算依赖 `.github` 的 go-ci 配了 `GOPRIVATE=github.com/oagentkit/*` 直连 VCS：
goproxy.cn 缓存着旧版本，若走 proxy 会拉回**旧代码**并算出旧 hash，
门禁全绿但依赖是错的。
（第二个提交会补上重算好的 go.sum。）
* chore(deps): 依赖地板升到「历史最新 + 0.1.0」
chore(deps): 依赖地板升到「历史最新 + 0.1.0」
本轮做了版本线重置：清空全部历史 tag，按「历史最新 + 0.1.0」重打新版本
（如 llm v0.6.0 → v0.7.0、toolgate v0.6.1 → v0.7.0、agentkit v1.3.0 → v1.4.0）。
各仓 go.mod 的矩阵内依赖地板同步跟进到新版本号。
**go.sum 的 oagentkit 行已整段删除**，由 CI 的 `go mod tidy` 重算：
新 tag 的内容与任何历史版本都不同，`h1` 必须现算；本地没有 Go 工具链，
且 `/go.mod` 的 dirhash 无法自行复现（zip hash 可复现，gomod hash 不可复现）。
重算依赖 `.github` 的 go-ci 配了 `GOPRIVATE=github.com/oagentkit/*` 直连 VCS：
goproxy.cn 缓存着旧版本，若走 proxy 会拉回**旧代码**并算出旧 hash，
门禁全绿但依赖是错的。
（第二个提交会补上重算好的 go.sum。）
* chore(deps): 回填 CI 重算的 go.sum
`h1` 由 CI 里真实的 `go mod tidy` 算出后回填——本地没有 Go 工具链，
且 `/go.mod` 的 dirhash 无法自行复现（zip hash 可复现，gomod hash 不可复现），
所以这是唯一权威来源。
重算依赖 `.github` go-ci 里的 `GOPRIVATE=github.com/oagentkit/*` 直连 VCS：
goproxy.cn 仍缓存着版本线重置前的旧内容，走 proxy 会拉回**旧代码**并算出旧 hash，
门禁全绿但依赖是错的。
验证方式就是本 PR 的 `go mod tidy -diff`：若回填有任何偏差，它会照样报差异。
* chore(deps): 回填 CI 重算的 go.sum
`h1` 由 CI 里真实的 `go mod tidy` 算出后回填——本地没有 Go 工具链，
且 `/go.mod` 的 dirhash 无法自行复现（zip hash 可复现，gomod hash 不可复现），
所以这是唯一权威来源。
重算依赖 `.github` go-ci 里的 `GOPRIVATE=github.com/oagentkit/*` 直连 VCS：
goproxy.cn 仍缓存着版本线重置前的旧内容，走 proxy 会拉回**旧代码**并算出旧 hash，
门禁全绿但依赖是错的。
验证方式就是本 PR 的 `go mod tidy -diff`：若回填有任何偏差，它会照样报差异。
* chore(deps): 回填 CI 重算的 go.sum
`h1` 由 CI 里真实的 `go mod tidy` 算出后回填——本地没有 Go 工具链，
且 `/go.mod` 的 dirhash 无法自行复现（zip hash 可复现，gomod hash 不可复现），
所以这是唯一权威来源。
重算依赖 `.github` go-ci 里的 `GOPRIVATE=github.com/oagentkit/*` 直连 VCS：
goproxy.cn 仍缓存着版本线重置前的旧内容，走 proxy 会拉回**旧代码**并算出旧 hash，
门禁全绿但依赖是错的。
验证方式就是本 PR 的 `go mod tidy -diff`：若回填有任何偏差，它会照样报差异。

- 杂务(deps)：跟进最新 tag

oagentkit/guardx v0.1.2→v0.2.0、oagentkit/textx v0.1.0→v0.2.0
go.sum 取 sumdb 权威 h1，仅就地替换目标 mod 两行。

- 杂务(deps)：oekit 依赖升级

pkg v0.13.0→v0.14.0
姊妹矩阵 oekit 的依赖地板此前被漏掉（上一轮只扫了 oagentkit 内部依赖），本次一并对齐到最新。
升级前已确认 oekit/concurrency v0.13.1..v0.16.0 的 8 个提交全是 feat/docs/chore，
无 breaking/remove 标记。其中 #28 订正了 `WithCooldown(0)` 的语义——
即 llms.txt 里警告过的那个陷阱，升上去正好拿到修正后的行为。
go.sum 的 h1 取自 sumdb，仅就地替换目标 mod 的两行、其余行不动。

- 杂务(deps)：guardx v0.1.1 → v0.1.2

guardx v0.1.2 已发布（含 injectband 新包），本仓跟进。
go.sum 的 h1 取自 sumdb，仅就地替换该 mod 的两行。

- 杂务(deps)：跨仓依赖 guardx v0.1.0 → v0.1.1

全 org 跨仓依赖地板统一升级（第一批）。
go.sum 的 h1 取自 sumdb（经 goproxy.cn 代理），即 go mod tidy 会写入的同一个权威值。
只就地替换该 mod 的两行、其余行一个字节不动——不重排全文件。go.sum 行序由 cmd/go
自己维护（/go.mod 行按版本字符串序在前，有 zip h1: 的普通行统一排在同 mod 末尾），
手写重排反而会引入漂移。

- 杂务(deps)：oekit 三模块升至最新（pkg v0.13.0 / concurrency v0.12.0 / logx v0.7.0）

拿到上游新增能力：
- concurrency #22 RetryableWithHeaders：403 头感知判定（次级限流 / 主限流耗尽）
- concurrency #23 FanInLimit：受限并发扇出
- concurrency #24 WithTokenClock、#21 Lockout（keyed lockout/penalty）
- logx #12 TraceHandler：给裸 slog.Handler 注入 trace_id
破坏性变更核对（均无影响）：
- logx v0.7.0 ParseLogLevel 签名改为 (Level, error) —— 我们未使用该函数
- concurrency v0.9.0 的 Slot.Panicked 废弃 —— rag 已迁到 Slot.Err + errors.As
门禁：go build / go vet / go test -race 全绿。


- CI：接入 org 共享门禁（规则本体在 oagentkit/.github）
