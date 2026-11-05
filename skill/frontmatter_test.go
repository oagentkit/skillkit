package skill

import (
	"context"
	"os"
	"strings"
	"testing"
)

// TestFileProviderBOMRegression 回归（v0.10.30 第四轮审计）：parseFrontmatter
// 也必须剥 BOM——FileProvider.Resolve 主加载路径曾整份当正文、元数据静默丢失。
func TestFileProviderBOMRegression(t *testing.T) {
	dir := t.TempDir()
	skillDir := dir + "/bom-skill"
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := "\xef\xbb\xbf---\nname: bom-skill\ndescription: 带 BOM 的技能\n---\n\n正文内容\n"
	if err := os.WriteFile(skillDir+"/SKILL.md", []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	p := NewFileProvider(dir)
	metas, err := p.ListSkills(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(metas) != 1 || metas[0].Name != "bom-skill" || metas[0].Description != "带 BOM 的技能" {
		t.Fatalf("BOM 路径元数据丢失: %+v", metas)
	}
	sk, err := p.Resolve(context.Background(), Ref{Name: "bom-skill"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(sk.Content, "---") || !strings.Contains(sk.Content, "正文内容") {
		t.Fatalf("正文不应含 frontmatter 围栏: %q", sk.Content)
	}
}

// TestFileProviderCRLFRegression 回归（第五轮审计）：CRLF 行尾的起始围栏——
// parseFrontmatter 曾要求围栏后紧跟 \n 而 \r\n 直接 bail，整份当正文、元数据
// 丢失且与 Library 路径的 Content/Checksum 分歧。
func TestFileProviderCRLFRegression(t *testing.T) {
	dir := t.TempDir()
	skillDir := dir + "/crlf-skill"
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := "---\r\nname: crlf-skill\r\ndescription: CRLF 技能\r\n---\r\n\r\n正文内容\r\n"
	if err := os.WriteFile(skillDir+"/SKILL.md", []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	p := NewFileProvider(dir)
	metas, err := p.ListSkills(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(metas) != 1 || metas[0].Name != "crlf-skill" || metas[0].Description != "CRLF 技能" {
		t.Fatalf("CRLF 路径元数据丢失: %+v", metas)
	}
	sk, err := p.Resolve(context.Background(), Ref{Name: "crlf-skill"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(sk.Content, "---") || !strings.Contains(sk.Content, "正文内容") {
		t.Fatalf("正文不应含围栏: %q", sk.Content)
	}
}

// RewriteBody 围栏语义唯一事实源（skillkit#33）：frontmatter 内缩进的 `---`
// （YAML 块标量里的分隔线）不是闭合——此前按 isFenceLine TrimSpace 判闭合，
// 编辑路径把元数据截断且输出不可再解析（同输入 ParseRichFrontmatter 正常）。
func TestRewriteBodyKeepsIndentedFenceInBlockScalar(t *testing.T) {
	src := "---\nname: demo\ndescription: |\n  Section A\n  ---\n  Section B\nversion: 1.0.0\n---\n\nOLD BODY\n"
	out, err := RewriteBody(src, "NEW-BODY")
	if err != nil {
		t.Fatalf("RewriteBody: %v", err)
	}
	for _, want := range []string{"Section B", "version: 1.0.0", "NEW-BODY"} {
		if !strings.Contains(out, want) {
			t.Fatalf("块标量内容丢失 %q: %q", want, out)
		}
	}
	if strings.Contains(out, "OLD BODY") {
		t.Fatalf("旧正文应被替换: %q", out)
	}
	// 输出可再解析（ParseRichFrontmatter 走 findClosingFence 同语义）
	m, _, err := ParseRichFrontmatter("demo", out)
	if err != nil {
		t.Fatalf("重写输出应可再解析: %v\nout=%q", err, out)
	}
	if m.Name != "demo" {
		t.Fatalf("name 应保留: %+v", m)
	}
}

// 列 0 闭合围栏仍被识别（唯一事实源的两语义一致面）。
func TestRewriteBodyClosesAtColumnZeroFence(t *testing.T) {
	src := "---\nname: demo\ndescription: d\n---\n\nOLD\n"
	out, err := RewriteBody(src, "NEW")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "NEW") || strings.Contains(out, "OLD") {
		t.Fatalf("常规围栏路径回归: %q", out)
	}
}
