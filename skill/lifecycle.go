package skill

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// RewriteMode 结构化切换 mode：frontmatter 块整体按 yaml.Node 承接，只对 mode
// 标量做**字节级局部手术**（已有则原位替换该标量的原始字节，没有则追加一行），
// 其余内容——含 schema 外的未知键（SKILL.md 开放标准允许的 license/owner/工具
// 注解等）、键序、注释、缩进、空行与行尾注释——逐字原样保留。
// （oagentkit/skillkit#57：此前的 LibMeta 结构体 round-trip 会把 schema 外的键
// 静默删除；oagentkit/skillkit#124：node 手术后整块 yaml.Marshal 仍会按 yaml.v3
// 规则重排——默认 4 空格缩进、空行丢失——「只改 mode」的承诺在格式面不成立。
// 无法安全定位原始标量（标签/锚点/块标量等非常规形态）时回退 node 手术 +
// 序列化：语义正确优先，缩进按原块探测，其余格式尽力保留。）
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
	// block 以起始围栏后的换行开头（splitFrontmatter 的坐标系）；剥掉它再
	// 做行/列定位——yaml.Node 的 Line 从这行空行起算会整体偏移一格。
	block = strings.TrimPrefix(block, "\n")
	mapping, err := frontmatterMapping(block)
	if err != nil {
		return "", err
	}
	rewritten, done, err := rewriteModeInPlace(block, mapping, mode)
	if err != nil {
		return "", err
	}
	if !done {
		// 回退路径：node 手术改写语义，序列化按原块缩进（默认 2，不再用
		// yaml.v3 的 4 空格默认值重排）。
		if err := setMappingScalar(mapping, "mode", mode); err != nil {
			return "", err
		}
		rewritten, err = marshalMount(mapping, block)
		if err != nil {
			return "", err
		}
	}
	fm := "---\n" + rewritten + "---"
	if body == "" {
		return fm + "\n", nil
	}
	return fm + "\n\n" + body + "\n", nil
}

// rewriteModeInPlace 在原始 frontmatter 块里定位 mode 标量并只替换它的字节
// （或没有 mode 键时追加一行）。ok=false 表示无法安全定位（调用方回退序列化）。
// 返回的块以换行结尾（与 yaml.Marshal 的输出形态一致，拼接闭合围栏用）。
func rewriteModeInPlace(block string, mapping *yaml.Node, mode string) (string, bool, error) {
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		k, v := mapping.Content[i], mapping.Content[i+1]
		if k.Kind != yaml.ScalarNode || k.Value != "mode" {
			continue
		}
		if v.Kind != yaml.ScalarNode {
			// 非标量 mode 由 setMappingScalar 报同款错误（回退路径）
			return "", false, nil
		}
		raw, ok := yamlScalarRaw(v)
		if !ok || raw == "" {
			return "", false, nil
		}
		start, ok := lineColumnOffset(block, v.Line, v.Column)
		if !ok || !strings.HasPrefix(block[start:], raw) {
			return "", false, nil
		}
		newRaw := modeScalarRaw(mode, v.Style)
		out := block[:start] + newRaw + block[start+len(raw):]
		return ensureTrailingNewline(out), true, nil
	}
	return appendModeLine(block, mode), true, nil
}

// yamlScalarRaw 还原标量节点的原始单行文本（用于定位替换区间）。只接受能
// 精确还原的形态：plain（无锚点）与单/双引号——含标签/锚点/块标量时返回
// false，交回退路径，不猜。
func yamlScalarRaw(v *yaml.Node) (string, bool) {
	if v.Kind != yaml.ScalarNode || v.Anchor != "" {
		return "", false
	}
	switch v.Style {
	case 0: // plain（未加引号、无标签）
		return v.Value, true
	case yaml.DoubleQuotedStyle:
		return strconv.Quote(v.Value), true
	case yaml.SingleQuotedStyle:
		return "'" + strings.ReplaceAll(v.Value, "'", "''") + "'", true
	default:
		return "", false
	}
}

// modeScalarRaw 按原标量风格产出新值的原始文本（值都是 ASCII 标识符，
// plain/单/双引号三种风格均无需转义）。
func modeScalarRaw(mode string, style yaml.Style) string {
	switch style {
	case yaml.DoubleQuotedStyle:
		return `"` + mode + `"`
	case yaml.SingleQuotedStyle:
		return "'" + mode + "'"
	default:
		return mode
	}
}

// lineColumnOffset 把 yaml.Node 的 1-based (line, column) 换算成 block 的
// 字节偏移。列按字节计——mode 键与值前缀都是 ASCII（顶层键），不存在
// rune/byte 歧义。越界返回 false。
func lineColumnOffset(block string, line, col int) (int, bool) {
	if line < 1 || col < 1 {
		return 0, false
	}
	off := 0
	for i := 1; i < line; i++ {
		j := strings.IndexByte(block[off:], '\n')
		if j < 0 {
			return 0, false
		}
		off += j + 1
	}
	lineText := block[off:]
	if j := strings.IndexByte(lineText, '\n'); j >= 0 {
		lineText = lineText[:j]
	}
	if col-1 > len(lineText) {
		return 0, false
	}
	return off + col - 1, true
}

// ensureTrailingNewline 保证块以行终止结尾；块末若是 CRLF 的 \r 半边，
// 补 \n 即完整 CRLF。
func ensureTrailingNewline(block string) string {
	if strings.HasSuffix(block, "\n") {
		return block
	}
	return block + "\n"
}

// appendModeLine 在块末尾追加 mode 键（结尾无换行）；空块直接成行。
// 行终止跟随原块的 CRLF/LF 形态，不额外引入混排。
func appendModeLine(block, mode string) string {
	if block == "" {
		return "mode: " + mode + "\n"
	}
	eol := "\n"
	if strings.Contains(block, "\r\n") {
		eol = "\r\n"
	}
	if !strings.HasSuffix(block, "\n") {
		if strings.HasSuffix(block, "\r") {
			block += "\n" // 补齐原块末行的 CRLF
		} else {
			block += eol
		}
	}
	return block + "mode: " + mode + eol
}

// marshalMount 回退序列化：yaml.Encoder + 探测原块缩进（不再用 yaml.v3
// 默认的 4 空格整块重排，尽可能贴近手写排版）。
func marshalMount(mapping *yaml.Node, originalBlock string) (string, error) {
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(frontmatterIndent(originalBlock))
	if err := enc.Encode(mapping); err != nil {
		return "", fmt.Errorf("skill: frontmatter 序列化失败: %w", err)
	}
	if err := enc.Close(); err != nil {
		return "", fmt.Errorf("skill: frontmatter 序列化失败: %w", err)
	}
	return buf.String(), nil
}

// frontmatterIndent 探测原块的首个多空格缩进（2..8），探测不到用 2。
// 仅服务回退序列化路径。
func frontmatterIndent(block string) int {
	for _, line := range strings.Split(block, "\n") {
		trimmed := strings.TrimLeft(line, " ")
		if trimmed == "" || len(trimmed) == len(line) {
			continue
		}
		if n := len(line) - len(trimmed); n >= 2 {
			if n > 8 {
				n = 8
			}
			return n
		}
	}
	return 2
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
// 文档头判定与解析器同口径（skillkit#78）：trimDocHead 剥 BOM/前导空白——
// 带 BOM（Windows 编辑器常见产物）或前导空行的文档 parseFrontmatter 与
// RewriteMode 都接受，编辑路径此前裸 HasPrefix(content, "---") 以「无
// frontmatter」误拒（文件有 frontmatter，解析器也认）；围栏块之后的围栏
// 保留与坐标切分语义不变（#66），坐标全部基于剥离后的 trimmed。
func RewriteBody(content, body string) (string, error) {
	trimmed := trimDocHead(content)
	if !strings.HasPrefix(trimmed, "---") {
		return "", fmt.Errorf("skill: SKILL.md 无 frontmatter，无法编辑正文")
	}
	rest := strings.TrimPrefix(trimmed[3:], "\r")
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
	// kept = 起始围栏到闭合围栏行（含）：blockEnd 指向闭合 "---" 前的 \n。
	// 切串必须用 rest 的**真实头偏移**（skillkit#66）：起始围栏 CRLF 时
	// 上面的 TrimPrefix 已吃掉 \r，rest 坐标系相对 trimmed 左移 1 字节——
	// 按写死的 3 切，混排 EOL（CRLF 起始围栏 + LF 内行，Windows 编辑器
	// 保存的单文件常态形态）下切点落在真实围栏 \n 的前一个字节，恰好裁掉
	// frontmatter 末行内容的最后一个字节（version: 1.2.3 → 1.2.，回读解析
	// 成功零报错的静默递进腐蚀；全 CRLF 文档碰巧丢的是 \r 才漏网）。
	head := len(trimmed) - len(rest) // 3（LF 起始围栏）或 4（CRLF）
	kept := trimmed[:head+blockEnd] + "\n---"
	return kept + "\n\n" + strings.TrimRight(body, "\n") + "\n", nil
}
