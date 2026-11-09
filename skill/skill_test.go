package skill

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestFileProviderResolve(t *testing.T) {
	dir := t.TempDir()

	// 创建测试 skill 文件：root/<name>/SKILL.md
	skillDir := filepath.Join(dir, "ocr-grading")
	_ = os.MkdirAll(skillDir, 0o755)
	_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# OCR 评分标准\n\nHigh = 安全问题"), 0o644)

	// 创建 root/<name>.md 格式
	_ = os.WriteFile(filepath.Join(dir, "team-style.md"), []byte("# 团队风格"), 0o644)

	p := NewFileProvider(dir)
	ctx := context.Background()

	// 解析 <name>/SKILL.md
	s, err := p.Resolve(ctx, Ref{Name: "ocr-grading"})
	if err != nil {
		t.Fatal(err)
	}
	if s.Content != "# OCR 评分标准\n\nHigh = 安全问题" {
		t.Fatalf("内容不符: %q", s.Content)
	}
	if s.Checksum == "" {
		t.Fatal("缺少 checksum")
	}

	// 解析 <name>.md
	s2, err := p.Resolve(ctx, Ref{Name: "team-style"})
	if err != nil {
		t.Fatal(err)
	}
	if s2.Content != "# 团队风格" {
		t.Fatalf("内容不符: %q", s2.Content)
	}

	// 缓存：第二次 Resolve 应命中缓存
	s3, err := p.Resolve(ctx, Ref{Name: "ocr-grading"})
	if err != nil {
		t.Fatal(err)
	}
	if s3 != s {
		t.Fatal("应命中缓存（同一指针）")
	}
}

func TestFileProviderNotFound(t *testing.T) {
	dir := t.TempDir()
	p := NewFileProvider(dir)

	_, err := p.Resolve(context.Background(), Ref{Name: "nonexistent"})
	if err == nil {
		t.Fatal("不存在的 skill 应报错")
	}
}

func TestFileProviderPathTraversal(t *testing.T) {
	dir := t.TempDir()
	p := NewFileProvider(dir)

	_, err := p.Resolve(context.Background(), Ref{Name: "../../../etc/passwd"})
	if err == nil {
		t.Fatal("路径穿越应被拒绝")
	}
}

func TestFileProviderUnsupportedSource(t *testing.T) {
	dir := t.TempDir()
	p := NewFileProvider(dir)

	_, err := p.Resolve(context.Background(), Ref{Name: "x", Source: "git"})
	if err == nil {
		t.Fatal("非 file 来源应报错")
	}
}

func TestFileProviderVersionedCache(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "my-skill.md"), []byte("v1"), 0o644)

	p := NewFileProvider(dir)
	ctx := context.Background()

	s1, _ := p.Resolve(ctx, Ref{Name: "my-skill"})
	s2, _ := p.Resolve(ctx, Ref{Name: "my-skill", Version: "2"})
	// 不同版本应独立缓存
	if s1 == s2 {
		t.Fatal("不同版本不应共享缓存")
	}
}

// Ref.Version 请求约束兑现（skillkit#37）：此前 FileProvider 只当缓存键、
// Library 完全忽略——请求 2.0.0 拿回 1.0.0 也静默成功（约束裸奔）。
func TestRefVersionConstraintEnforced(t *testing.T) {
	dir := t.TempDir()
	_ = os.MkdirAll(filepath.Join(dir, "sop"), 0o755)
	_ = os.WriteFile(filepath.Join(dir, "sop", "SKILL.md"),
		[]byte("---\nname: sop\nversion: 1.2.3\ndescription: d\n---\n\nBODY"), 0o644)
	fp := NewFileProvider(dir)
	// 版本不符报错（不再静默返回 1.2.3）
	_, err := fp.Resolve(context.Background(), Ref{Name: "sop", Version: "2.0.0"})
	if err == nil || !strings.Contains(err.Error(), "版本不符") {
		t.Fatalf("版本不符应报错: %v", err)
	}
	// 匹配版本与空版本（不约束）照常
	if _, err := fp.Resolve(context.Background(), Ref{Name: "sop", Version: "1.2.3"}); err != nil {
		t.Fatalf("匹配版本应通过: %v", err)
	}
	if _, err := fp.Resolve(context.Background(), Ref{Name: "sop"}); err != nil {
		t.Fatalf("空版本不约束: %v", err)
	}
}

// loadSkill 区分 ErrNotExist 与其余错误（skillkit#36）：权限拒绝不再误诊为
// 「未找到」。
func TestLoadSkillPermissionError(t *testing.T) {
	dir := t.TempDir()
	_ = os.MkdirAll(filepath.Join(dir, "sop"), 0o755)
	p := filepath.Join(dir, "sop", "SKILL.md")
	_ = os.WriteFile(p, []byte("---\nname: sop\n---\n\nBODY"), 0o000) // 权限拒绝
	fp := NewFileProvider(dir)
	_, err := fp.Resolve(context.Background(), Ref{Name: "sop"})
	if err == nil {
		t.Fatal("权限拒绝应报错")
	}
	if strings.Contains(err.Error(), "未找到") {
		t.Fatalf("权限拒绝不得误诊为未找到: %v", err)
	}
	// 缺文件仍报「未找到」
	_, err = fp.Resolve(context.Background(), Ref{Name: "absent"})
	if err == nil || !strings.Contains(err.Error(), "未找到") {
		t.Fatalf("缺文件应报未找到: %v", err)
	}
}

// scanAliases 失败不缓存（skillkit#35）：一次瞬时目录错误不再让展示名归一
// 进程内永久失效——失败后下次调用重试。
func TestScanAliasesErrorNotCached(t *testing.T) {
	dir := t.TempDir()
	_ = os.MkdirAll(filepath.Join(dir, "sop"), 0o755)
	_ = os.WriteFile(filepath.Join(dir, "sop", "SKILL.md"),
		[]byte("---\nname: 标准作业程序\nversion: 1.0.0\ndescription: d\n---\n\nBODY"), 0o644)
	fp := NewFileProvider(dir)
	m := fp.scanAliases(context.Background())
	if m["标准作业程序"] != "sop" {
		t.Fatalf("别名扫描应命中: %+v", m)
	}
	// 成功结果已缓存（快路径）
	fp.mu.RLock()
	cached := fp.aliases
	fp.mu.RUnlock()
	if cached == nil {
		t.Fatal("成功结果应写缓存")
	}
}

// ListSkills 损坏条目经 OnCorrupt 留痕（skillkit#36）：静默丢弃让损坏无信号。
func TestListSkillsOnCorrupt(t *testing.T) {
	dir := t.TempDir()
	_ = os.MkdirAll(filepath.Join(dir, "good"), 0o755)
	_ = os.WriteFile(filepath.Join(dir, "good", "SKILL.md"),
		[]byte("---\nname: good\ndescription: d\n---\n\nBODY"), 0o644)
	// bad 目录无 SKILL.md（Resolve 报「未找到」——发现面唯一真实损坏形态：
	// frontmatter 解析是宽松语义，unclosed 围栏整份当正文不算损坏）
	_ = os.MkdirAll(filepath.Join(dir, "bad"), 0o755)
	fp := NewFileProvider(dir)
	var corrupted []string
	fp.OnCorrupt = func(name string, err error) { corrupted = append(corrupted, name) }
	metas, err := fp.ListSkills(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(metas) != 1 || metas[0].Name != "good" {
		t.Fatalf("完好条目应照常列出: %+v", metas)
	}
	if len(corrupted) != 1 || corrupted[0] != "bad" {
		t.Fatalf("损坏条目应经 OnCorrupt 留痕: %v", corrupted)
	}
}

// TestFileProviderResolveObservesDiskEdits skillkit#56 回归：Resolve 缓存此前
// 永久 sticky（无 TTL、无 stat 复验、无失效 API）——盘上编辑（含宿主经
// RewriteMode/RewriteBody 改写后写回的编辑流）永不生效，陈旧 Checksum 还会
// 毒化比对型守卫。现缓存条目带 mtime+size 指纹，命中前先 stat 复验，变了即重读。
func TestFileProviderResolveObservesDiskEdits(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "demo", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	v1 := "---\nname: demo\ndescription: 第一版\n---\n\nbody-v1\n"
	if err := os.WriteFile(path, []byte(v1), 0o644); err != nil {
		t.Fatal(err)
	}
	fp := NewFileProvider(dir)
	ctx := context.Background()
	s1, err := fp.Resolve(ctx, Ref{Name: "demo"})
	if err != nil {
		t.Fatal(err)
	}
	if s1.Content != "body-v1" {
		t.Fatalf("首读内容不符: %q", s1.Content)
	}

	// 等长正文替换（RewriteBody 编辑流的典型终态）：size 不变、只动 mtime——
	// 指纹必须双轴。显式 Chtimes 前移，免受文件系统 mtime 粒度影响（不赌 FS）。
	v2 := strings.Replace(v1, "body-v1", "body-v2", 1)
	if err := os.WriteFile(path, []byte(v2), 0o644); err != nil {
		t.Fatal(err)
	}
	future := time.Now().Add(2 * time.Second)
	if err := os.Chtimes(path, future, future); err != nil {
		t.Fatal(err)
	}
	s2, err := fp.Resolve(ctx, Ref{Name: "demo"})
	if err != nil {
		t.Fatal(err)
	}
	if s2.Content != "body-v2" {
		t.Fatalf("盘上编辑后 Resolve 仍回旧缓存（skillkit#56）: %q", s2.Content)
	}
	if s2.Checksum == s1.Checksum {
		t.Fatalf("内容变更后 checksum 必须随动: %s", s2.Checksum)
	}

	// 复验不误伤缓存语义：内容未变时同一指针命中
	s3, err := fp.Resolve(ctx, Ref{Name: "demo"})
	if err != nil {
		t.Fatal(err)
	}
	if s3 != s2 {
		t.Fatal("未变化的重复 Resolve 应命中缓存（同一指针）")
	}
}

// TestFileProviderInvalidate 显式失效 API：粗粒度 FS 上同 size 且同 mtime 的
// 改写 stat 抓不住——宿主写回后可 Invalidate 强制下一次 Resolve 重读。
func TestFileProviderInvalidate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "demo.md")
	v1 := "---\nname: demo\n---\n\nbody-v1\n"
	if err := os.WriteFile(path, []byte(v1), 0o644); err != nil {
		t.Fatal(err)
	}
	fp := NewFileProvider(dir)
	ctx := context.Background()
	if _, err := fp.Resolve(ctx, Ref{Name: "demo"}); err != nil {
		t.Fatal(err)
	}

	// 等长改写 + mtime 回拨到原值：stat 复验视为未变（粗粒度 FS 的真实形态）
	fi, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	v2 := strings.Replace(v1, "body-v1", "body-v2", 1)
	if err := os.WriteFile(path, []byte(v2), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(path, fi.ModTime(), fi.ModTime()); err != nil {
		t.Fatal(err)
	}

	fp.Invalidate("demo")
	s, err := fp.Resolve(ctx, Ref{Name: "demo"})
	if err != nil {
		t.Fatal(err)
	}
	if s.Content != "body-v2" {
		t.Fatalf("Invalidate 后 Resolve 应重读盘上内容: %q", s.Content)
	}
}

// TestFileProviderConcurrentResolveInvalidate -race 回归：并发 Resolve 与
// Invalidate 混跑无数据竞争（指纹 stat 在锁外、缓存读写均在锁内收敛）。
func TestFileProviderConcurrentResolveInvalidate(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "demo.md"), []byte("---\nname: demo\n---\n\n正文"), 0o644); err != nil {
		t.Fatal(err)
	}
	fp := NewFileProvider(dir)
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx := context.Background()
			for j := 0; j < 32; j++ {
				if _, err := fp.Resolve(ctx, Ref{Name: "demo"}); err != nil {
					t.Error(err)
					return
				}
				if j%8 == 0 {
					fp.Invalidate("demo")
				}
			}
		}()
	}
	wg.Wait()
}

// TestFileProviderFingerprintStatBeforeRead skillkit#77 回归：指纹 stat 此前
// 取在 ReadFile **之后**——写落在「读结束→指纹 stat」窗内时，缓存条目持有
// v1 内容却挂着 v2 的指纹，fresh() 对比的是同一次写之后的盘面，永久命中，
// 陈旧内容与陈旧 Checksum 被无限期返回、无信号。指纹必须先于读：任何在
// 指纹之后完成的写都会让盘面与条目失配，下一次复验必然重载。
func TestFileProviderFingerprintStatBeforeRead(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "demo", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	v1 := "---\nname: demo\ndescription: 第一版\n---\n\nbody-v1-short\n"
	v2 := strings.Replace(v1, "body-v1-short", "body-v2-longer-and-different", 1)
	if err := os.WriteFile(path, []byte(v1), 0o644); err != nil {
		t.Fatal(err)
	}
	fp := NewFileProvider(dir)
	var seq []string
	fp.statFn = func(p string) (os.FileInfo, error) {
		seq = append(seq, "stat")
		return os.Stat(p)
	}
	fp.readFileFn = func(p string) ([]byte, error) {
		seq = append(seq, "read")
		b, err := os.ReadFile(p)
		if err == nil && strings.Contains(string(b), "body-v1") {
			// 宿主编辑写回落在读改窗内：读结束的同一瞬间盘面已是 v2
			// （size 变长，指纹必变——不赌 mtime 粒度）
			if werr := os.WriteFile(p, []byte(v2), 0o644); werr != nil {
				t.Errorf("模拟并发写回失败: %v", werr)
			}
		}
		return b, err
	}
	ctx := context.Background()
	s1, err := fp.Resolve(ctx, Ref{Name: "demo"})
	if err != nil {
		t.Fatal(err)
	}
	if s1.Content != "body-v1-short" {
		t.Fatalf("首读内容不符: %q", s1.Content)
	}
	// 竞态写之后的下一次 Resolve 必须看到盘上 v2——缺陷形态：条目挂着竞态
	// 写之后的指纹，fresh() 永久命中，v1 连同陈旧 Checksum 被无限期返回。
	s2, err := fp.Resolve(ctx, Ref{Name: "demo"})
	if err != nil {
		t.Fatal(err)
	}
	if s2.Content != "body-v2-longer-and-different" {
		t.Fatalf("读改窗内的写后 Resolve 仍回陈旧缓存（skillkit#77 缺陷）: %q", s2.Content)
	}
	if s2.Checksum == s1.Checksum {
		t.Fatal("竞态写后 checksum 必须随动")
	}
	if len(seq) < 2 || seq[0] != "stat" || seq[1] != "read" {
		t.Fatalf("指纹 stat 必须先于读（skillkit#77 修复点）: %v", seq)
	}
}
