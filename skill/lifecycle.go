package skill

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// RewriteMode 结构化切换 mode：frontmatter 块整体按 yaml.Node 承接，只对 mode
// 标量做局部手术（已有则原位改值，没有则追加），其余键——含 LibMeta schema 外
// 的未知键（SKILL.md 开放标准允许的 license/owner/工具注解等）——连同键序与
// 注释一并原样保留，兑现「其余元数据字段原样保留」的文档承诺。
// （oagentkit/skillkit#57：此前的 LibMeta 结构体 round-trip——yaml.Unmarshal
// 忽略未知键 + re-marshal 只写已知字段——会把 schema 外的键静默删除。）
// 无 frontmatter 或未闭合 → 报错（不猜格式，与解析 fail-fast 同哲学）；文档头
// 判定与解析器同口径（trimDocHead 剥 BOM/前导空白，带 BOM 的文档不再被拒）。
func RewriteMode(content, mode string) (string, error) {
	if mode != ModeStatic && mode != ModeOnDemand {
		return "", fmt.Errorf("skill: mode 非法 %q（static|on_demand）", mode)
	}
	trimmed := trimDocHead(content)
	if !strings.HasPrefix(trimmed, "---") {
		return "", fmt.Errorf("skill: SKILL.md 无 frontmatter，无法切换 mode")
	}
	block, body, ok := splitFrontmatter(trimmed)
	if !ok {
		return "", fmt.Errorf("skill: frontmatter 未闭合")
	}
	mapping, err := frontmatterMapping(block)
	if err != nil {
		return "", err
	}
	if err := setMappingScalar(mapping, "mode", mode); err != nil {
		return "", err
	}
	b, err := yaml.Marshal(mapping)
	if err != nil {
		return "", err
	}
	fm := "---\n" + string(b) + "---"
	if body == "" {
		return fm + "\n", nil
	}
	return fm + "\n\n" + body + "\n", nil
}

// frontmatterMapping 把 frontmatter 块承接为键值映射节点（node 手术的手术台）。
// 空块等价空映射（`---\n---` 退化形态）；块整体非法 YAML fail-fast；顶层不是
// 映射（标量/序列）同样报错——与旧结构体承接的失败语义同口径。
func frontmatterMapping(block string) (*yaml.Node, error) {
	if strings.TrimSpace(block) == "" {
		return &yaml.Node{Kind: yaml.MappingNode}, nil
	}
	var root yaml.Node
	if err := yaml.Unmarshal([]byte(block), &root); err != nil {
		return nil, fmt.Errorf("skill: frontmatter YAML 非法: %w", err)
	}
	if root.Kind != yaml.DocumentNode || len(root.Content) != 1 || root.Content[0].Kind != yaml.MappingNode {
		return nil, fmt.Errorf("skill: frontmatter 须为 YAML 键值映射，无法切换 mode")
	}
	return root.Content[0], nil
}

// setMappingScalar 在键值映射节点上原位设置标量键：已有同键则只改值节点
// （键节点与其注释/位置不动；非标量值视为非法 frontmatter，fail-fast 不静默
// 覆盖）；没有则把新键值对追加到末尾。
func setMappingScalar(m *yaml.Node, key, val string) error {
	for i := 0; i+1 < len(m.Content); i += 2 {
		k, v := m.Content[i], m.Content[i+1]
		if k.Kind == yaml.ScalarNode && k.Value == key {
			if v.Kind != yaml.ScalarNode {
				return fmt.Errorf("skill: frontmatter %s 须为标量，无法切换", key)
			}
			v.Tag = "!!str"
			v.Value = val
			return nil
		}
	}
	m.Content = append(m.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key},
		&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: val},
	)
	return nil
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
