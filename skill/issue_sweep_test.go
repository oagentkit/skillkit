package skill

// 2026-11-11 issue 清零战回归测试（#83 #84）：钉住 Name 口径现状与 miss 口径。

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

// TestNameSemanticsAcrossProviders_Issue83 Skill.Name 两 Provider 的口径**现状**
// 钉住（#83）：FileProvider=frontmatter 展示名（未声明回落目录名）、Library=
// 规范引用名。字段文档已写明差异与「身份比对请用 Ref.Name」的守则——本测试
// 让任何一侧的无意漂移立即可见；统一口径是破坏式变更，届时改此测试+字段文档。
func TestNameSemanticsAcrossProviders_Issue83(t *testing.T) {
	content := "---\nname: 展示名\n---\n\n正文。\n"
	fp := NewFileProvider(t.TempDir())
	if err := os.MkdirAll(filepath.Join(fp.Root, "sop"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fp.Root, "sop", "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	s1, err := fp.Resolve(context.Background(), Ref{Name: "sop"})
	if err != nil {
		t.Fatal(err)
	}
	lib, err := LoadFromFS(fstest.MapFS{"_shared/skills/sop/SKILL.md": &fstest.MapFile{Data: []byte(content)}})
	if err != nil {
		t.Fatal(err)
	}
	s2, err := lib.Resolve(context.Background(), Ref{Name: "sop"})
	if err != nil {
		t.Fatal(err)
	}
	if s1.Name != "展示名" {
		t.Fatalf("FileProvider.Name 应为 frontmatter 展示名，实得 %q", s1.Name)
	}
	if s2.Name != "sop" {
		t.Fatalf("Library.Name 应为规范引用名，实得 %q", s2.Name)
	}
}

// TestDescribeMissReturnsNameOnly_Issue84 未知名字（含 nil 接收者）一律返回
// 仅含名字的零值——同一函数不再有两种 miss 语义。
func TestDescribeMissReturnsNameOnly_Issue84(t *testing.T) {
	lib := &Library{} // 非 nil、无条目
	m := lib.Describe("不存在")
	if m.Name != "不存在" {
		t.Fatalf("miss 应返回仅含名字的零值，实得 %+v", m)
	}
	var nilLib *Library
	if m2 := nilLib.Describe("任意"); m2.Name != "任意" {
		t.Fatalf("nil 接收者口径漂移: %+v", m2)
	}
}
