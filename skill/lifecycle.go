package skill

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// RewriteMode 结构化切换 mode：解析 frontmatter → 改 mode → yaml 序列化写回，
// 正文与其余元数据字段（嵌套 requires_mcp/provides 等）原样保留。
// 无 frontmatter 或未闭合 → 报错（不猜格式，与解析 fail-fast 同哲学）。
func RewriteMode(content, mode string) (string, error) {
	if mode != ModeStatic && mode != ModeOnDemand {
		return "", fmt.Errorf("skill: mode 非法 %q（static|on_demand）", mode)
	}
	if !strings.HasPrefix(content, "---") {
		return "", fmt.Errorf("skill: SKILL.md 无 frontmatter，无法切换 mode")
	}
	meta, body, err := ParseRichFrontmatter("tmp", content)
	if err != nil {
		return "", err
	}
	meta.Mode = mode
	fm, err := marshalFrontmatter(meta)
	if err != nil {
		return "", err
	}
	if body == "" {
		return fm + "\n", nil
	}
	return fm + "\n\n" + body + "\n", nil
}

// RewriteBody 保留 frontmatter 围栏（原文不动），仅替换正文——编辑器路径。
// 无 frontmatter 或未闭合 → 报错（围栏是元数据的唯一保护壳，不整段重建）。
func RewriteBody(content, body string) (string, error) {
	if !strings.HasPrefix(content, "---") {
		return "", fmt.Errorf("skill: SKILL.md 无 frontmatter，无法编辑正文")
	}
	rest := strings.TrimPrefix(content[3:], "\r")
	if !strings.HasPrefix(rest, "\n") {
		return "", fmt.Errorf("skill: SKILL.md 无 frontmatter，无法编辑正文")
	}
	// 围栏语义唯一事实源 = findClosingFence（列 0 闭合；oagentkit/skillkit#33）：
	// 此前按 isFenceLine 逐行 TrimSpace 判闭合——frontmatter 内缩进的 `---`
	// （YAML 块标量里的分隔线）被误当闭合，编辑路径把元数据截断（Section B
	// 与 version 全丢），且输出不可再解析；同输入 ParseRichFrontmatter 正常。
	blockEnd, _, closed := findClosingFence(rest)
	if !closed {
		return "", fmt.Errorf("skill: frontmatter 未闭合，拒绝编辑")
	}
	// kept = 起始围栏到闭合围栏行（含）：blockEnd 指向闭合 "---" 前的 \n
	kept := content[:3+blockEnd] + "\n---"
	return kept + "\n\n" + strings.TrimRight(body, "\n") + "\n", nil
}

// marshalFrontmatter 序列化回 SKILL.md frontmatter（只写已知字段；Title 还原为
// frontmatter name，缺省 version 0.0.0 不写回——避免文件被无意义膨胀）。
// schema 唯一事实源是 LibMeta（omitempty 标签即写回约定）。
func marshalFrontmatter(m LibMeta) (string, error) {
	out := m
	out.Name = m.Title // Title 还原为 frontmatter name
	if out.Version == DefaultVersion {
		out.Version = ""
	}
	b, err := yaml.Marshal(&out)
	if err != nil {
		return "", err
	}
	return "---\n" + string(b) + "---", nil
}
