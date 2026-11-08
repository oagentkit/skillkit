package skill

import (
	"strings"
	"testing"
	"testing/fstest"
	"time"
)

// mf 构造 MapFS 文件项。
func mf(s string) *fstest.MapFile { return &fstest.MapFile{Data: []byte(s)} }

// lifecycleFS 校验/写回夹具：_shared 技能（frozen 缺省）+ 包内技能（契约字段齐全）。
func lifecycleFS() fstest.MapFS {
	return fstest.MapFS{
		"_shared/skills/sop/SKILL.md":          mf("---\nname: SOP\ndescription: 通用诊断流程\n---\n\n流程正文"),
		"os-basics/skills/net-detect/SKILL.md": mf("---\nname: 网络检测\ndescription: 网络攻击检测方法\nmode: on_demand\nversion: 1.2.0\nmaturity: stable\nrequires_mcp:\n  - server: ask-ops\n    tools: [get_logs]\n---\n\n检测方法论正文"),
	}
}

// TestLoadValidateFailFast mode/maturity 枚举、SemVer、弃用窗口、requires_mcp 非空——全部加载期拦截。
func TestLoadValidateFailFast(t *testing.T) {
	bad := []fstest.MapFS{
		{"p/skills/s1/SKILL.md": mf("---\nname: x\nmode: magic\n---\n正文")},
		{"p/skills/s1/SKILL.md": mf("---\nname: x\nmaturity: semi\n---\n正文")},
		{"p/skills/s1/SKILL.md": mf("---\nname: x\nversion: 1.x\n---\n正文")},
		{"p/skills/s1/SKILL.md": mf("---\nname: x\nmaturity: deprecated\n---\n正文")},
		{"p/skills/s1/SKILL.md": mf("---\nname: x\nmaturity: deprecated\ndeprecated:\n  remove_after: tomorrow\n---\n正文")},
		{"p/skills/s1/SKILL.md": mf("---\nname: x\nrequires_mcp:\n  - tools: [a]\n---\n正文")},
		{"p/skills/s1/SKILL.md": mf("---\nname: [unclosed\n---\n正文")},
	}
	for i, fs := range bad {
		if _, err := LoadFromFS(fs); err == nil {
			t.Fatalf("夹具 %d 应 fail-fast", i)
		}
	}
	// 错误带文件定位。
	_, err := LoadFromFS(bad[0])
	if err != nil && !strings.Contains(err.Error(), "p/skills/s1/SKILL.md") {
		t.Fatalf("错误应带路径定位: %v", err)
	}
}

// TestValidateRemoveAfterCalendar 历法校验回归（skillkit#45）：remove_after 此前
// 只验 \d{4}-\d{2}-\d{2} 形态——「2026-13-45」「2027-02-29（非闰年）」通过 Validate，
// 而 DeprecationExpired 对解析失败恒 false：非法日期=弃用窗口永不过期（静默
// 无限期）。现 Validate 用 time.Parse 验真实历法日，加载期 fail-closed。
func TestValidateRemoveAfterCalendar(t *testing.T) {
	bad := []string{"2026-13-45", "2027-02-30", "2027-02-29"} // 13 月 45 日 / 2 月 30 日 / 非闰年 2 月 29 日
	for _, d := range bad {
		err := validateDeprecated(&Deprecated{RemoveAfter: d})
		if err == nil || !strings.Contains(err.Error(), "非法日期") {
			t.Fatalf("%s 应报非法历法日: %v", d, err)
		}
	}
	// 真实历法日不受误伤：闰年 2 月 29 日合法。
	if err := validateDeprecated(&Deprecated{RemoveAfter: "2028-02-29"}); err != nil {
		t.Fatalf("闰年 2 月 29 日不应误伤: %v", err)
	}
	// 加载期端到端：非法历法日整库 fail-fast。
	fsys := fstest.MapFS{"p/skills/s1/SKILL.md": mf("---\nname: x\nmaturity: deprecated\ndeprecated:\n  remove_after: 2026-13-45\n---\n正文")}
	if _, err := LoadFromFS(fsys); err == nil {
		t.Fatal("非法历法 remove_after 应在加载期拦截")
	}
}

// TestLoadContractDefaults 缺省值（_shared=frozen / 包内=experimental / 0.0.0 / static）
// 与契约字段（requires_mcp.tools / provides）解析。
func TestLoadContractDefaults(t *testing.T) {
	lib, err := LoadFromFS(lifecycleFS())
	if err != nil {
		t.Fatal(err)
	}
	if m := lib.Describe("sop"); m.Maturity != MaturityFrozen || m.Mode != ModeStatic || m.Version != DefaultVersion {
		t.Fatalf("sop 缺省异常: %+v", m)
	}
	m := lib.Describe("net-detect")
	if m.Maturity != MaturityStable || m.Version != "1.2.0" || m.Mode != ModeOnDemand {
		t.Fatalf("net-detect 元数据异常: %+v", m)
	}
	if len(m.RequiresMCP) != 1 || m.RequiresMCP[0].Server != "ask-ops" ||
		len(m.RequiresMCP[0].Tools) != 1 || m.RequiresMCP[0].Tools[0] != "get_logs" {
		t.Fatalf("requires_mcp 解析异常: %+v", m.RequiresMCP)
	}
	if got := lib.Describe("sop").Source; got != "_shared" {
		t.Fatalf("source=%q", got)
	}
}

// TestDeprecatedExpiredInUse 过期且被引用 → 失败清单；未过期/无引用不进。
func TestDeprecatedExpiredInUse(t *testing.T) {
	fsys := fstest.MapFS{"p/skills/s1/SKILL.md": mf("---\nname: s1\nmaturity: deprecated\ndeprecated:\n  replaced_by: s2\n  remove_after: 2099-01-01\n---\n正文")}
	lib, err := LoadFromFS(fsys)
	if err != nil {
		t.Fatal(err)
	}
	if m := lib.Describe("s1"); m.DeprecationExpired(time.Now()) || m.Deprecated.ReplacedBy != "s2" {
		t.Fatalf("2099 窗口不应过期/replaced_by 丢失: %+v", m)
	}
	if got := lib.DeprecatedExpiredInUse(map[string][]string{"s1": {"a"}}, time.Now()); len(got) != 0 {
		t.Fatalf("未过期不应进失败清单: %v", got)
	}
	past := fstest.MapFS{"p/skills/s1/SKILL.md": mf("---\nname: s1\nmaturity: deprecated\ndeprecated:\n  remove_after: 2020-01-01\n---\n正文")}
	lib2, _ := LoadFromFS(past)
	if got := lib2.DeprecatedExpiredInUse(map[string][]string{"s1": {"a"}}, time.Now()); len(got) != 1 {
		t.Fatalf("过期且被引用应进失败清单: %v", got)
	}
	if got := lib2.DeprecatedExpiredInUse(map[string][]string{}, time.Now()); len(got) != 0 {
		t.Fatalf("无引用不应失败: %v", got)
	}
}

// TestRewriteMode 结构化 mode 写回保留嵌套字段（requires_mcp.tools/min_version、provides）与正文。
func TestRewriteMode(t *testing.T) {
	src := strings.Join([]string{
		"---", "name: ssh-check", "description: 检测", "mode: static",
		"version: 1.1.0", "maturity: stable", "provides: [ssh-attack-detection, login-audit]",
		"requires_mcp:", "  - server: security-assistant", "    min_version: \"1.2.0\"", "    tools: [collect_security_logs]",
		"---", "", "正文段",
	}, "\n")
	out, err := RewriteMode(src, ModeOnDemand)
	if err != nil {
		t.Fatal(err)
	}
	lib, err := LoadFromFS(fstest.MapFS{"p/skills/ssh-check/SKILL.md": mf(out)})
	if err != nil {
		t.Fatalf("写回后应仍可加载: %v\n%s", err, out)
	}
	m := lib.Describe("ssh-check")
	if m.Mode != ModeOnDemand || m.Maturity != MaturityStable || m.Version != "1.1.0" {
		t.Fatalf("写回后元数据丢失: %+v", m)
	}
	if len(m.Provides) != 2 || len(m.RequiresMCP) != 1 || m.RequiresMCP[0].MinVersion != "1.2.0" || m.RequiresMCP[0].Server != "security-assistant" {
		t.Fatalf("写回丢契约字段: provides=%v mcp=%+v", m.Provides, m.RequiresMCP)
	}
	if body, _ := lib.Body("ssh-check"); body != "正文段" {
		t.Fatalf("正文被破坏: %q", body)
	}
	if _, err := RewriteMode("无围栏", ModeStatic); err == nil {
		t.Fatal("无 frontmatter 应报错")
	}
	if _, err := RewriteMode(src, "magic"); err == nil {
		t.Fatal("非法 mode 应报错")
	}
}

// TestRewriteModePreservesUnknownKeys skillkit#57 回归：frontmatter 块按
// yaml.Node 局部手术（只动 mode 标量）后，LibMeta schema 外的键——SKILL.md
// 开放标准允许的 license/owner 及任意工具注解——连同键序与注释原样存活。
// 此前的 LibMeta 结构体 round-trip 会把它们静默删除（yaml.Unmarshal 忽略
// 未知键，re-marshal 只写已知字段），与文档「其余元数据字段原样保留」相悖。
func TestRewriteModePreservesUnknownKeys(t *testing.T) {
	src := strings.Join([]string{
		"---",
		"name: demo",
		"description: 演示",
		"mode: static",
		"# 保留注释行",
		"license: mit", // schema 外键（开放标准允许）
		"owner: platform-team",
		"x-tool-annotation: {a: 1}",
		"---",
		"",
		"Body.",
	}, "\n")
	out, err := RewriteMode(src, ModeOnDemand)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"license: mit", "owner: platform-team", "x-tool-annotation", "# 保留注释行", "a: 1"} {
		if !strings.Contains(out, want) {
			t.Fatalf("未知键 %q 被静默丢弃（skillkit#57）:\n%s", want, out)
		}
	}
	if !strings.Contains(out, "mode: on_demand") {
		t.Fatalf("mode 未切换:\n%s", out)
	}
	// round-trip 幂等：再次切换未知键继续存活
	out2, err := RewriteMode(out, ModeStatic)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out2, "license: mit") {
		t.Fatalf("二次切换仍须保留未知键:\n%s", out2)
	}
}

// TestRewriteModeToleratesDocHead 文档头判定与解析器同口径（skillkit#57 次要项）：
// 解析入口 trimDocHead 剥 BOM 与前导空白——能被 ParseRichFrontmatter 正常解析
// 的文档（BOM/前导空行）不应被 RewriteMode 以「无 frontmatter」拒绝。
func TestRewriteModeToleratesDocHead(t *testing.T) {
	for i, src := range []string{
		"\ufeff---\nmode: static\n---\n正文", // Windows 编辑器 BOM
		"\n\t---\nmode: static\n---\n正文",   // 前导空白
	} {
		out, err := RewriteMode(src, ModeOnDemand)
		if err != nil {
			t.Fatalf("夹具 %d：解析器可读的文档头不应被拒: %v", i, err)
		}
		if !strings.Contains(out, "mode: on_demand") {
			t.Fatalf("夹具 %d：mode 未切换: %q", i, out)
		}
	}
}

// TestRewriteModeNoTitle 无 frontmatter name 的技能写回不得引入伪 name。
func TestRewriteModeNoTitle(t *testing.T) {
	src := "---\nmode: static\n---\n正文"
	out, err := RewriteMode(src, ModeOnDemand)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "name:") {
		t.Fatalf("无展示名不应写回 name 键: %q", out)
	}
}

// TestRewriteBody 保留 frontmatter 原文，仅替换正文。
func TestRewriteBody(t *testing.T) {
	src := "---\nname: s1\nmode: static\n---\n\n旧正文"
	out, err := RewriteBody(src, "新正文")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "name: s1") || !strings.Contains(out, "新正文") || strings.Contains(out, "旧正文") {
		t.Fatalf("RewriteBody 异常: %q", out)
	}
	if _, err := RewriteBody("无围栏", "x"); err == nil {
		t.Fatal("无 frontmatter 应报错")
	}
}

// TestRewriteBodyMixedEOL 混排 EOL 回归（skillkit#66）：CRLF 起始围栏 + LF
// 内行（Windows 编辑器保存的单文件常态形态）此前坐标错位一字节——TrimPrefix
// 吃掉起始围栏 \r 后 rest 坐标系左移，kept 却按无 \r 的 content 坐标系切，
// 每次编辑递进裁掉 frontmatter 末行内容的最后一个字节（version: 1.2.3 →
// 1.2.），回读解析成功零报错。断言口径：frontmatter 围栏块逐字节保留、
// 回读解析元数据无损。全 CRLF 形态此前「碰巧不损」（丢的恰是 \r，表现为
// 换行归一）——正是它漏网的原因，一并钉住。
func TestRewriteBodyMixedEOL(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want string
	}{
		{"CRLF 起始围栏+LF 内行", "---\r\nname: probe-skill\nversion: 1.2.3\n---\n\nBODY\n",
			"---\r\nname: probe-skill\nversion: 1.2.3\n---\n\nNEW\n"},
		{"全 LF 基线", "---\nname: probe-skill\nversion: 1.2.3\n---\n\nBODY\n",
			"---\nname: probe-skill\nversion: 1.2.3\n---\n\nNEW\n"},
		{"全 CRLF", "---\r\nname: probe-skill\r\nversion: 1.2.3\r\n---\r\n\r\nBODY\r\n",
			"---\r\nname: probe-skill\r\nversion: 1.2.3\r\n---\n\nNEW\n"},
		{"LF 起始围栏+CRLF 内行", "---\nname: probe-skill\r\nversion: 1.2.3\r\n---\n\nBODY\n",
			"---\nname: probe-skill\r\nversion: 1.2.3\r\n---\n\nNEW\n"},
	}
	for _, tc := range cases {
		out, err := RewriteBody(tc.src, "NEW")
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		if out != tc.want {
			t.Fatalf("%s:\ngot  %q\nwant %q", tc.name, out, tc.want)
		}
		// 回读解析零损：末行值不被截断（腐蚀形态回读是成功的，单靠不报错抓不住）。
		meta, body, err := ParseRichFrontmatter("probe-skill", out)
		if err != nil {
			t.Fatalf("%s: 回读解析失败: %v", tc.name, err)
		}
		if meta.Title != "probe-skill" || meta.Version != "1.2.3" {
			t.Fatalf("%s: 元数据被腐蚀: title=%q version=%q", tc.name, meta.Title, meta.Version)
		}
		if body != "NEW" {
			t.Fatalf("%s: 正文应为 NEW, got %q", tc.name, body)
		}
	}
}

// TestVersionInRange 区间求解：空区间恒真；空格=AND；非法区间/版本报错。
func TestVersionInRange(t *testing.T) {
	cases := []struct {
		version, expr string
		want          bool
		wantErr       bool
	}{
		{"1.2.0", "", true, false},
		{"1.2.0", ">=1.0.0 <2.0.0", true, false},
		{"2.0.0", ">=1.0.0 <2.0.0", false, false},
		{"1.0.0", ">=1.0.0", true, false},
		{"0.0.0", "^0.1.0", false, false},
		{"1.2.0", "非法区间", false, true},
		{"not-a-version", ">=1.0.0", false, true},
	}
	for _, c := range cases {
		got, err := VersionInRange(c.version, c.expr)
		if c.wantErr {
			if err == nil {
				t.Fatalf("VersionInRange(%q,%q) 应报错", c.version, c.expr)
			}
			continue
		}
		if err != nil {
			t.Fatalf("VersionInRange(%q,%q): %v", c.version, c.expr, err)
		}
		if got != c.want {
			t.Fatalf("VersionInRange(%q,%q)=%v, want %v", c.version, c.expr, got, c.want)
		}
	}
}
