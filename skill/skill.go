// Package skill 技能内容与元数据：SKILL.md 解析（frontmatter 围栏 + canonical
// checksum）、进程内技能库（多根扫描、热替换）、版本区间约束（decl/version）、
// 决策式按需加载（decision.go 的 use_skill 渐进披露）。
// 内容文件解析带路径遍历保护（`..` 段拒绝；注意 symlink 不在防护范围——root
// 应为可信目录）；文件解析部分零外部依赖，决策使用依赖 eino。
package skill

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/oekit/pkg/encoding"
)

// Ref 内容引用。
//
// Version 是**请求约束**（oagentkit/skillkit#37）：非空时两侧 Provider 都会
// 校验——解析出的 skill 版本 ≠ 请求版本即报错（此前 FileProvider 只拿它当
// 缓存键、Library 完全忽略，约束静默裸奔：请求 2.0.0 拿回 1.0.0 也成功）。
// 校验用精确匹配（decl/version 的区间约束走 Decl 面构建 ref）。
type Ref struct {
	Name    string
	Version string
	Source  string // file（v3.0 仅实现 file）
}

// Skill 解析后的内容（带 checksum 防内容漂移难排查）。
// SKILL.md 支持 Agent Skills 标准的 frontmatter（--- 围栏内 name/description 行）：
// description 服务于发现与决策使用（渐进披露），正文才是消费主体。
type Skill struct {
	// Name 技能名。⚠️ **两 Provider 口径不同**（oagentkit/skillkit#83，既有
	// 契约、双 Provider 对照测试钉住）：FileProvider 取 frontmatter 展示名
	// （未声明时回落目录名）；Library 恒取规范引用名（目录名）。做身份比对/
	// 去重的消费方请以 Ref.Name（请求引用名）为准，勿用本字段跨 Provider
	// 比对——同一份 SKILL.md（目录名 sop、frontmatter name: 展示名）两侧
	// 分别得「展示名」与「sop」。统一口径是破坏式变更（FileProvider 侧行为
	// 被 decision_test 的 TestFrontmatterParsed 钉住），需维护者裁决后另开
	// breaking 变更，届时展示名应有独立字段。
	Name        string
	Version     string // frontmatter 声明（未声明为空串）；ref.Version 是请求约束/缓存键，不回显
	Description string // frontmatter description（无则空）
	Content     string // frontmatter 之后的正文
	// Checksum 内容校验和，canonical 口径：sha256(正文) 前 16 位 hex——绑定
	// 实际注入提示词的内容（frontmatter 是元数据，不进 prompt）。
	// FileProvider 与 Library 两个 Provider 一致，观测守卫的比对基准不随 Provider 漂移。
	Checksum string
}

// contentChecksum canonical 校验和（sha256 hex 前 16 位），两个 Provider 共用单源。
// 哈希经 oekit pkg/encoding 单源（ekit v0.31 起：弱哈希入口已在该包治理移除，本函数
// 只取前 16 hex 的截断口径；ekit 单体冻结后随 oekit 拆分仓继续单源）。
func contentChecksum(body string) string {
	return encoding.Sha256(body)[:16]
}

// Meta skill 元数据（不含正文——渐进披露：提示词只注入这一层）。
type Meta struct {
	// Name 规范引用名（目录名/文件名）——use_skill 加载与 allowed 清单的唯一依据。
	Name string
	// Title 展示名（frontmatter name；与 Name 相同或目录缺失时为空）。
	// 仅用于清单展示，加载仍按 Name；模型若用 Title 调用，AsSkillTool 会归一化。
	Title       string
	Version     string
	Description string
}

// Provider 内容解析接口。
type Provider interface {
	Resolve(ctx context.Context, ref Ref) (*Skill, error)
}

// FileProvider file 来源实现：root/<name>/SKILL.md 或 root/<name>.md。
//
// 缓存与新鲜度（oagentkit/skillkit#56）：缓存条目带源文件 stat 指纹
// （mtime+size），Resolve 命中前先复验，盘上编辑（含宿主经 RewriteMode/
// RewriteBody 改写后写回的编辑流）下一次 Resolve 即可见——每次命中付出一次
// os.Stat 的代价，换取「编辑-重载」循环无进程级陈旧。指纹取**读取前**的
// stat 快照（skillkit#77）：读改窗内的竞态写不会被吸收进指纹让旧内容永久
// 免检。粗粒度 FS 上同 size 且同 mtime 的改写指纹抓不住，可经 Invalidate
// 显式强制重读。复验发现文件消失按陈旧处理走重读——删除以「未找到」错误
// 如实上抛，不再被旧缓存掩护。
type FileProvider struct {
	// OnCorrupt 单个 skill 解析失败回调（nil 安全，oagentkit/skillkit#36）：
	// ListSkills 对损坏条目 continue 跳过（发现不阻塞），损坏经此观测——
	// 此前静默丢弃，清单缺一条无从归因。本仓不引入日志依赖（tier 纪律），
	// 宿主经回调接告警面。
	OnCorrupt func(name string, err error)

	Root string

	// 测试缝（skillkit#77 竞态复现用）：nil = os.Stat / os.ReadFile。生产
	// 路径不触碰（构造后由测试在首次 Resolve 前注入一次），注入后 stat/读
	// 的时序可在测试里确定性编排。
	statFn     func(string) (os.FileInfo, error)
	readFileFn func(string) ([]byte, error)

	mu      sync.RWMutex
	cache   map[string]*fileCacheEntry // key = name@version（与 cacheKey 同键空间）
	aliases map[string]string          // 展示名(frontmatter name) → 规范引用名（懒扫描）
}

// fileCacheEntry 缓存条目：解析产物 + 源文件指纹。指纹两轴缺一不可——粗粒度
// FS 上 size 抓不住等长替换、mtime 抓不住同窗双写。条目入库后不可变（并发
// 持有只读）；内容变更即换新条目，不做原地改写。
type fileCacheEntry struct {
	skill *Skill
	path  string
	mtime time.Time
	size  int64
}

// NewFileProvider 创建文件来源的内容解析器。
func NewFileProvider(root string) *FileProvider {
	return &FileProvider{Root: root, cache: map[string]*fileCacheEntry{}}
}

func cacheKey(ref Ref) string {
	if ref.Version != "" {
		return ref.Name + "@" + ref.Version
	}
	return ref.Name
}

// statFile / readFile 源文件访问单源（测试缝注入点，nil 回落标准库）。
func (p *FileProvider) statFile(path string) (os.FileInfo, error) {
	if p.statFn != nil {
		return p.statFn(path)
	}
	return os.Stat(path)
}

func (p *FileProvider) readFile(path string) ([]byte, error) {
	if p.readFileFn != nil {
		return p.readFileFn(path)
	}
	return os.ReadFile(path)
}

// Invalidate 显式失效缓存：清掉 name 的全部缓存键（含 name@version 变体，
// 归一化口径同 safeSkillName）。供宿主在 stat 指纹抓不住的写路径后调用
// （粗粒度 FS 上的等长同 mtime 改写、NFS 等不可靠 mtime 来源），也可用于
// 写后强制一致。未知名字安全（无操作）。
func (p *FileProvider) Invalidate(name string) {
	name = strings.TrimSpace(name)
	clean := filepath.Clean("/" + name)[1:]
	p.mu.Lock()
	defer p.mu.Unlock()
	for key := range p.cache {
		if key == name || key == clean ||
			strings.HasPrefix(key, name+"@") || strings.HasPrefix(key, clean+"@") {
			delete(p.cache, key)
		}
	}
}

// Resolve 解析引用，返回内容（带缓存，命中前 stat 指纹复验——见 FileProvider
// 新鲜度契约）。磁盘读取前检查 ctx 取消。
func (p *FileProvider) Resolve(ctx context.Context, ref Ref) (*Skill, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("skill: %w", err)
	}
	if ref.Source == "" {
		ref.Source = "file"
	}
	if ref.Source != "file" {
		return nil, fmt.Errorf("skill: 来源 %q 暂未实现（仅支持 file），ref=%s", ref.Source, ref.Name)
	}

	name, err := safeSkillName(ref.Name)
	if err != nil {
		return nil, err
	}
	// 缓存键用**归一化后**的引用名（skillkit#85）：此前先算键后归一——
	// "./foo" 与 "foo" 落不同键（键空间随拼写分裂），而 Invalidate 只删
	// 归一化键 ⇒ 非规范拼写装出的条目**永不过期**，等长同 mtime 改写后旧
	// 内容持续外发。键空间归一后同一 skill 单条目，Invalidate 的归一化
	// 删除口径与键空间一致。
	key := cacheKey(Ref{Name: name, Version: ref.Version})
	p.mu.RLock()
	cached := p.cache[key]
	p.mu.RUnlock()
	if cached != nil && p.fresh(cached) {
		return cached.skill, nil
	}
	s, path, fi, err := p.loadSkill(ref, name)
	if err != nil {
		// 陈旧条目的重读失败（如源文件已被删）：清掉旧条目再如实上抛，
		// 不留半失效状态
		p.mu.Lock()
		delete(p.cache, key)
		p.mu.Unlock()
		return nil, err
	}
	// 指纹即读取前的 stat 快照（skillkit#77）：读取前的指纹保证任何在其后
	// 完成的写都让盘面与条目失配，下一次复验必然重载——读窗内被编辑时本次
	// 返回的可能是旧内容（竞态语义既定），但旧内容不会再被竞态写吸收进
	// 指纹而永久免检（「读后取指纹」的缺陷形态）。
	p.mu.Lock()
	p.cache[key] = &fileCacheEntry{skill: s, path: path, mtime: fi.ModTime(), size: fi.Size()}
	p.mu.Unlock()
	return s, nil
}

// fresh 缓存条目新鲜度：对源文件重新 stat，mtime+size 双轴一致才算新鲜。
// stat 在锁外执行（只读不可变字段 Root 与条目自身，不触碰共享 map）；
// stat 失败（文件被删等）按陈旧处理——具体错误交由重读路径如实上抛。
// 并发取舍：不做单飞，两个 goroutine 同看到一个陈旧键会各自重读一遍
// （重复 I/O、后写覆盖，读到的都是同一份盘上内容）——Resolve 不引入
// 全局串行点，锁内只做 map 读写。
func (p *FileProvider) fresh(e *fileCacheEntry) bool {
	fi, err := p.statFile(e.path)
	if err != nil {
		return false
	}
	return fi.ModTime().Equal(e.mtime) && fi.Size() == e.size
}

// safeSkillName 规范化引用名并拒绝路径穿越（前缀 / 使 .. 无法越出 root）。
// 仅拒绝路径段级的 ..，放行 foo..bar 这类合法名。
func safeSkillName(refName string) (string, error) {
	name := filepath.Clean("/" + strings.TrimSpace(refName))[1:]
	if name == "" {
		return "", fmt.Errorf("skill: 非法引用 %q", refName)
	}
	for _, seg := range strings.Split(name, "/") {
		if seg == ".." {
			return "", fmt.Errorf("skill: 非法引用 %q", refName)
		}
	}
	return name, nil
}

// loadSkill 读盘并剥 frontmatter 组装 Skill（checksum 对正文计算）。
// 返回实际命中的候选路径与**读取前**的 stat 快照——缓存指纹（mtime+size）
// 指向它（#56；指纹先于内容是 skillkit#77 的修复点：读后取指纹会把读改窗
// 内的竞态写吸收进指纹，旧内容从此免检）。
func (p *FileProvider) loadSkill(ref Ref, name string) (*Skill, string, os.FileInfo, error) {
	candidates := []string{
		filepath.Join(p.Root, name, "SKILL.md"),
		filepath.Join(p.Root, name+".md"),
	}
	var data []byte
	var loaded string
	var fi os.FileInfo
	var readErr error
	for _, c := range candidates {
		// 指纹先于内容：stat 紧贴读前取快照——任何在指纹之后完成的写都让
		// 盘面与条目失配、下一次复验必然重载；stat 失败的候选不读（无指纹
		// 可挂），错误同 #36 口径记录：非「未找到」在全候选失败时如实上抛
		si, err := p.statFile(c)
		if err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				readErr = err
			}
			continue
		}
		b, err := p.readFile(c)
		if err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				// 权限拒绝/I/O 错误不是「未找到」——原样记录并在全候选失败时
				// 上抛（oagentkit/skillkit#36：此前任意 ReadFile 错误都被当
				// 不存在吞掉，权限故障被误诊为文件缺失，排查方向完全反了）
				readErr = err
			}
			continue
		}
		data, loaded, fi = b, c, si
		break
	}
	if data == nil {
		if readErr != nil {
			return nil, "", nil, fmt.Errorf("skill: 读取 %s 失败: %w", ref.Name, readErr)
		}
		return nil, "", nil, fmt.Errorf("skill: 未找到 %s（尝试 %v）", ref.Name, candidates)
	}
	// frontmatter 元数据剥离；checksum 对正文计算（canonical 口径，见 Skill.Checksum）
	fmName, fmDesc, fmVer, body, fmWarn := parseFrontmatter(string(data))
	if fmWarn != nil && p.OnCorrupt != nil {
		// 降级警告（oagentkit/skillkit#46）：frontmatter 块非合法 YAML 时逐行
		// 宽松解析仍产出元数据（发现不阻塞），但「值可能是垃圾」必须可观测——
		// OnCorrupt 亦承载降级信号（本仓不引日志依赖，tier 纪律），宿主接告警面。
		p.OnCorrupt(ref.Name, fmWarn)
	}
	// Version 请求约束兑现（oagentkit/skillkit#37）：请求版本非空且与
	// frontmatter 声明不符即报错——此前只当缓存键，约束静默裸奔
	if ref.Version != "" && fmVer != ref.Version {
		return nil, "", nil, fmt.Errorf("skill: %s 版本不符：请求 %s，库内 %s", ref.Name, ref.Version, fmVer)
	}
	displayName := ref.Name
	if fmName != "" {
		displayName = fmName
	}
	return &Skill{
		Name:        displayName,
		Version:     fmVer,
		Description: fmDesc,
		Content:     body,
		Checksum:    contentChecksum(body),
	}, loaded, fi, nil
}

// ---------- 发现与决策使用（渐进披露） ----------

// ListSkills 枚举目录下全部 skill 的元数据（不含正文）。
// 供提示词注入"可用 skill 清单"——agent 据此决策加载哪个（经 use_skill 工具）。
// 实现 Lister 接口。
func (p *FileProvider) ListSkills(ctx context.Context) ([]Meta, error) {
	entries, err := os.ReadDir(p.Root)
	if err != nil {
		return nil, fmt.Errorf("skill: 目录读取失败 %s: %w", p.Root, err)
	}
	var out []Meta
	for _, e := range entries {
		name := e.Name()
		var refName string
		switch {
		case e.IsDir():
			refName = name // root/<name>/SKILL.md
		case strings.HasSuffix(name, ".md"):
			refName = strings.TrimSuffix(name, ".md") // root/<name>.md
		default:
			continue
		}
		s, err := p.Resolve(ctx, Ref{Name: refName})
		if err != nil {
			// 单个 skill 损坏不阻塞发现，但要**留痕**（rag#33 同款教训：
			// 静默丢弃让损坏无信号，清单缺一条无从归因）。本仓不引入日志
			// 依赖（tier 纪律）——观测走 OnCorrupt 回调，宿主接线告警面。
			if p.OnCorrupt != nil {
				p.OnCorrupt(refName, err)
			}
			continue
		}
		title := ""
		if s.Name != refName {
			title = s.Name // frontmatter name 与目录名不同：作为展示别名
		}
		out = append(out, Meta{Name: refName, Title: title, Version: s.Version, Description: s.Description})
	}
	return out, nil
}

// CanonicalName 把展示名（frontmatter name）归一化为规范引用名（目录名/文件名）。
// 未命中返回 ("", false)。别名表懒扫描构建（进程内 skill 目录通常不变）；
// 首次扫描为全目录磁盘 I/O，ctx 取消随链透传。
func (p *FileProvider) CanonicalName(ctx context.Context, name string) (string, bool) {
	// 精确名优先（与 Library.CanonicalName 的 byName 契约对齐）：技能 A 的
	// frontmatter name 恰与技能 B 的目录名相同时，别名表不得把 B 重定向到 A。
	// 同 loadName 的防穿越归一（前缀 / 使 .. 无法越出 root）
	name = filepath.Clean("/" + strings.TrimSpace(name))[1:]
	if name != "" {
		if _, err := os.Stat(filepath.Join(p.Root, name, "SKILL.md")); err == nil {
			return name, true
		}
		if _, err := os.Stat(filepath.Join(p.Root, name+".md")); err == nil {
			return name, true
		}
	}
	p.mu.RLock()
	aliases := p.aliases
	p.mu.RUnlock()
	if aliases == nil {
		aliases = p.scanAliases(ctx)
	}
	n, ok := aliases[name]
	return n, ok
}

func (p *FileProvider) scanAliases(ctx context.Context) map[string]string {
	// 并发收敛 + 失败不缓存（oagentkit/skillkit#35）：此前吞掉 ListSkills
	// 错误并把**空别名表**缓存进 p.aliases——一次瞬时目录错误（权限抖动/
	// 挂载闪断）让展示名归一进程内永久失效（恒 not-found）；并发首扫各自
	// 全目录 I/O 重复执行。
	p.mu.Lock()
	if p.aliases != nil {
		m := p.aliases
		p.mu.Unlock()
		return m
	}
	p.mu.Unlock()

	metas, err := p.ListSkills(ctx)
	if err != nil {
		// 失败不写缓存：下次调用重试（ListSkills 内部已对单 skill 损坏
		// continue 容忍——走到这里的 err 是目录级故障，值得重试）
		return map[string]string{}
	}
	m := map[string]string{}
	for _, meta := range metas {
		if meta.Title != "" && meta.Title != meta.Name {
			// 同 Title 取名字典序最小者——与 Library.byTitle 同一确定性规则
			//（此前字典序扫描+后者覆盖=取最大者，与 Library 先到先得不同构）
			if prev, exists := m[meta.Title]; !exists || meta.Name < prev {
				m[meta.Title] = meta.Name
			}
		}
	}
	p.mu.Lock()
	p.aliases = m
	p.mu.Unlock()
	return m
}

// Lister 可选能力：支持"发现 + 决策使用"的 Provider（渐进披露模式需要）。
type Lister interface {
	ListSkills(ctx context.Context) ([]Meta, error)
}

// AliasResolver 可选能力：把展示别名归一化为规范引用名（渐进披露链路闭环：
// 清单可能展示 frontmatter name，模型会原样回填给 use_skill）。
type AliasResolver interface {
	CanonicalName(ctx context.Context, name string) (string, bool)
}

// 编译期断言：FileProvider 支持发现与别名归一化。
var (
	_ Lister        = (*FileProvider)(nil)
	_ AliasResolver = (*FileProvider)(nil)
)
