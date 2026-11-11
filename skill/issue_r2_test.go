package skill

// 2026-11-11 第二轮清零回归（#85 #86）。

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

// TestInvalidateCoversNonCanonicalSpellings_Issue85 非规范拼写（./foo）装载
// 的条目必须能被 Invalidate("foo") 失效——缓存键空间归一。
func TestInvalidateCoversNonCanonicalSpellings_Issue85(t *testing.T) {
	content := "---\nname: sop\n---\n\n正文 v1。\n"
	fp := NewFileProvider(t.TempDir())
	if err := os.MkdirAll(filepath.Join(fp.Root, "sop"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fp.Root, "sop", "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	s1, err := fp.Resolve(context.Background(), Ref{Name: "./sop"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(s1.Content, "v1") {
		t.Fatalf("首装应读到 v1: %q", s1.Content)
	}
	// 等长同 mtime 改写（stat 指纹抓不住的写）。
	content2 := "---\nname: sop\n---\n\n正文 v2。\n"
	if err := os.WriteFile(filepath.Join(fp.Root, "sop", "SKILL.md"), []byte(content2), 0o644); err != nil {
		t.Fatal(err)
	}
	fp.Invalidate("sop")
	s2, err := fp.Resolve(context.Background(), Ref{Name: "./sop"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(s2.Content, "v2") {
		t.Fatalf("Invalidate 后应读到 v2——非规范拼写条目失效不掉（skillkit#85）: %q", s2.Content)
	}
}

// TestLibraryTitleNormalizedLikeFileProvider_Issue86 展示名与引用名相同时
// Meta.Title 置空——两 Provider 一个口径。
func TestLibraryTitleNormalizedLikeFileProvider_Issue86(t *testing.T) {
	content := "---\nname: foo\n---\n\n正文。\n"
	lib, err := LoadFromFS(fstest.MapFS{"_shared/skills/foo/SKILL.md": &fstest.MapFile{Data: []byte(content)}})
	if err != nil {
		t.Fatal(err)
	}
	metas, err := lib.ListSkills(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(metas) != 1 {
		t.Fatalf("应 1 条: %d", len(metas))
	}
	if metas[0].Title != "" {
		t.Fatalf("Title 与 Name 相同应置空（两 Provider 同口径，skillkit#86）: %q", metas[0].Title)
	}
	// 展示名与引用名不同：Title 保留。
	content2 := "---\nname: 展示名\n---\n\n正文。\n"
	lib2, err := LoadFromFS(fstest.MapFS{"_shared/skills/bar/SKILL.md": &fstest.MapFile{Data: []byte(content2)}})
	if err != nil {
		t.Fatal(err)
	}
	metas2, err := lib2.ListSkills(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(metas2) != 1 || metas2[0].Title != "展示名" {
		t.Fatalf("不同展示名应保留: %+v", metas2)
	}
}
