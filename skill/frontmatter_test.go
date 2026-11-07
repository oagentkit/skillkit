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

// TestParseFrontmatterStandardYAMLValueForms 标准 YAML 值形态回归（skillkit#46）：
// 手写逐行宽松解析器只认单行标量——块标量 `description: |` 的值是字面 "|"
// （多行内容静默全丢）、>- 折叠同理、引号值带行尾注释则引号+注释整串泄漏。
// 现围栏内的值解码走 yaml.v3（与 ParseRichFrontmatter 同源），标准形态正确解出
// 且不产生降级信号。
func TestParseFrontmatterStandardYAMLValueForms(t *testing.T) {
	cases := []struct {
		label string
		fm    string
		name  string
		desc  string
		ver   string
	}{
		{"块标量 literal", "name: demo\ndescription: |\n  第一步：准备凭据。\n  第二步：轮换。\n", "demo", "第一步：准备凭据。\n第二步：轮换。", ""},
		{"折叠标量 strip", "name: demo\ndescription: >-\n  折叠成一行\n  的描述\n", "demo", "折叠成一行 的描述", ""},
		{"双引号转义", "name: \"ocr\"\ndescription: \"带\\t转义与「引号」\"\n", "ocr", "带\t转义与「引号」", ""},
		{"单引号双写转义", "name: demo\ndescription: 'It''s a skill'\n", "demo", "It's a skill", ""},
		{"行尾注释（带引号值）", "name: demo\ndescription: \"真实描述\" # yaml 行尾注释\n", "demo", "真实描述", ""},
		{"行尾注释（裸值）", "name: demo\ndescription: 真实描述 # 注释\n", "demo", "真实描述", ""},
		{"空值", "name: demo\ndescription:\n", "demo", "", ""},
		{"整数形态 version", "name: demo\ndescription: d\nversion: 2\n", "demo", "d", "2"},
	}
	for _, c := range cases {
		name, desc, ver, body, warn := parseFrontmatter("---\n" + c.fm + "---\n\n正文内容\n")
		if warn != nil {
			t.Fatalf("%s: 合法 YAML 不应产生降级信号: %v", c.label, warn)
		}
		if name != c.name {
			t.Fatalf("%s: name=%q, want %q", c.label, name, c.name)
		}
		if desc != c.desc {
			t.Fatalf("%s: desc=%q, want %q", c.label, desc, c.desc)
		}
		if ver != c.ver {
			t.Fatalf("%s: ver=%q, want %q", c.label, ver, c.ver)
		}
		if body != "正文内容" {
			t.Fatalf("%s: 正文被破坏: %q", c.label, body)
		}
	}
	// 两套解析器同一 YAML 语义：同输入下宽松面与 yaml 严格面给出一致 description
	// （块标量含缩进 `---` 分隔线的完整口径）。
	src := "---\nname: demo\ndescription: |\n  Section A\n  ---\n  Section B\nversion: 1.0.0\n---\n\n正文内容\n"
	_, looseDesc, _, _, warn := parseFrontmatter(src)
	if warn != nil {
		t.Fatalf("块标量+缩进围栏不应降级: %v", warn)
	}
	rich, _, err := ParseRichFrontmatter("demo", src)
	if err != nil {
		t.Fatal(err)
	}
	if looseDesc != rich.Description {
		t.Fatalf("宽松/严格两面的 description 应一致: %q vs %q", looseDesc, rich.Description)
	}
}

// TestFileProviderListSkillsStandardYAMLFrontmatter 发现清单端到端（skillkit#46）：
// 块标量 description 曾以字面 "|" 进 Meta/ListPrompt（模型按描述决策 use_skill
// 的依据被静默摧毁）——现 ListSkills 的描述即真实多行内容。
func TestFileProviderListSkillsStandardYAMLFrontmatter(t *testing.T) {
	dir := t.TempDir()
	skillDir := dir + "/probe-skill"
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	src := "---\nname: probe-skill\ndescription: |\n  轮换凭据的分步方法论。\n  含前置检查。\n---\n\n正文内容\n"
	if err := os.WriteFile(skillDir+"/SKILL.md", []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	p := NewFileProvider(dir)
	metas, err := p.ListSkills(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(metas) != 1 || metas[0].Description != "轮换凭据的分步方法论。\n含前置检查。" {
		t.Fatalf("块标量描述应完整进清单: %+v", metas)
	}
	out := ListPrompt(metas)
	if !strings.Contains(out, "轮换凭据的分步方法论") || strings.Contains(out, "**：|") {
		t.Fatalf("ListPrompt 不得再渲染字面 |: %s", out)
	}
}

// TestFileProviderFrontmatterFallbackWarn 降级信号（skillkit#46）：块整体非合法
// YAML（如未闭合引号）时仍回落逐行宽松解析（发现不阻塞），但降级必须可观测——
// OnCorrupt 收到警告，不再静默产出可疑值。
func TestFileProviderFrontmatterFallbackWarn(t *testing.T) {
	dir := t.TempDir()
	skillDir := dir + "/broken-yaml"
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	src := "---\nname: \"unclosed\ndescription: 单行描述仍可用\nversion: 1.0.0\n---\n\n正文内容\n"
	if err := os.WriteFile(skillDir+"/SKILL.md", []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	var warned bool
	p := NewFileProvider(dir)
	p.OnCorrupt = func(name string, err error) {
		warned = true
		if name != "broken-yaml" || err == nil || !strings.Contains(err.Error(), "降级") {
			t.Errorf("降级警告应带定位与原因: name=%q err=%v", name, err)
		}
	}
	metas, err := p.ListSkills(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	// 宽松策略保留：降级不阻塞发现，逐行可解的键仍进清单。
	if len(metas) != 1 || metas[0].Description != "单行描述仍可用" || metas[0].Version != "1.0.0" {
		t.Fatalf("降级后逐行解析值应保留: %+v", metas)
	}
	if !warned {
		t.Fatal("frontmatter 非合法 YAML 应触发 OnCorrupt 降级警告（零信号回归）")
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
