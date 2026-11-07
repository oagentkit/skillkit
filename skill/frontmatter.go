package skill

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// findClosingFence 在起始围栏之后的内容里定位闭合 ---（闭合围栏唯一事实源：
// 列 0 ---，允许行尾空白/CRLF；「---x」不是围栏）。
// 返回块终点（闭合行前的 \n 处）与正文起点（闭合行之后）；未闭合 ok=false。
func findClosingFence(rest string) (blockEnd, bodyStart int, ok bool) {
	for off := 0; ; {
		j := strings.Index(rest[off:], "\n---")
		if j < 0 {
			return 0, 0, false
		}
		pos := off + j
		after := rest[pos+4:] // "---" 之后到行尾
		lineEnd := strings.IndexByte(after, '\n')
		tail := after
		if lineEnd >= 0 {
			tail = after[:lineEnd]
		}
		if strings.TrimSpace(tail) == "" { // 行尾只有空白/CRLF：合法闭合
			if lineEnd < 0 {
				return pos, len(rest), true
			}
			return pos, pos + 4 + lineEnd + 1, true
		}
		off = pos + 1 // 「---x」伪围栏：跳过继续找
	}
}

// splitFrontmatter 从已 trimDocHead 且以 --- 起头的内容里切出 frontmatter 块与
// canonical 正文（TrimSpace 后）。两解析器唯一围栏切分入口——起始围栏必须独立成行
// （`---` 后立刻换行，CRLF 先吃 \r），闭合围栏语义见 findClosingFence。
// ok=false：起始围栏不成立 / 闭合围栏缺失；调用方按各自策略（宽松回退 vs 严格报错）。
func splitFrontmatter(trimmed string) (block, body string, ok bool) {
	if !strings.HasPrefix(trimmed, "---") {
		return "", "", false
	}
	rest := strings.TrimPrefix(trimmed[3:], "\r")
	if !strings.HasPrefix(rest, "\n") {
		return "", "", false // `---foo` 不是围栏（起始围栏须独立成行）
	}
	blockEnd, bodyStart, closed := findClosingFence(rest)
	if !closed {
		return "", "", false
	}
	// canonical 正文：剥壳后去首尾空白——两解析器同口径，Provider 的
	// Content（及其 Checksum）才逐字节一致。
	return rest[:blockEnd], strings.TrimSpace(rest[bodyStart:]), true
}

// fmLite 宽松 frontmatter 的 yaml 承接结构：只取 name/description/version 三键，
// 其余键忽略（与旧逐行解析同一宽松语义）。字段以 yaml.Node 承接——标量的
// 值形态（块标量 | / >、单双引号含转义、行尾注释、空值、整数等非字符串 tag）
// 由 yaml.v3 解析层给出最终内容，与 ParseRichFrontmatter 同一 YAML 语义。
type fmLite struct {
	Name        yaml.Node `yaml:"name"`
	Description yaml.Node `yaml:"description"`
	Version     yaml.Node `yaml:"version"`
}

// nodeScalarText 提取标量节点文本：!!null 与非标量节点（映射/序列——宽松面
// 无此形态的合法输入）返回空串，不产出 "{}"/"[]" 类垃圾值。
func nodeScalarText(n *yaml.Node) string {
	if n.Kind != yaml.ScalarNode || n.Tag == "!!null" {
		return ""
	}
	return n.Value
}

// parseFrontmatter 解析 SKILL.md 的 YAML frontmatter（Agent Skills 开放标准的
// 最小子集：--- 围栏内的 name/description/version 行），不认识的键忽略——
// skill 正文才是消费主体，元数据只服务发现与决策。
// 宽松策略：无/未闭合围栏一律整份当正文（FileProvider 单文件损坏不阻塞发现）。
//
// 围栏内的值解码走 yaml.v3（oagentkit/skillkit#46）：此前手写逐行 Cut 只认
// 单行标量——`description: |` 的值是字面 "|"（块标量多行内容静默全丢）、
// 引号值带行尾注释则引号+注释整串泄漏进发现清单，且解析「成功」零信号。
// 块整体不是合法 YAML 时回落旧逐行宽松解析（宽松策略保留），并返回非 nil
// warn 交调用方观测——降级产出垃圾值的风险不再静默。
func parseFrontmatter(content string) (name, desc, ver, body string, warn error) {
	trimmed := trimDocHead(content)
	block, body, ok := splitFrontmatter(trimmed)
	if !ok {
		return "", "", "", content, nil // 宽松策略既定语义（围栏不成立整份当正文），非降级
	}

	var fm fmLite
	if err := yaml.Unmarshal([]byte(block), &fm); err != nil {
		warn = fmt.Errorf("skill: frontmatter 块非合法 YAML，降级逐行宽松解析: %w", err)
	} else {
		return nodeScalarText(&fm.Name), nodeScalarText(&fm.Description), nodeScalarText(&fm.Version), body, nil
	}

	// 降级回落：旧逐行宽松解析（只认单行标量；块标量等形态此时已无解，
	// warn 是其产出垃圾值的唯一信号面）。
	for _, line := range strings.Split(block, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		val = trimQuoted(strings.TrimSpace(val))
		switch strings.TrimSpace(key) {
		case "name":
			name = val
		case "description":
			desc = val
		case "version":
			ver = val
		}
	}
	return name, desc, ver, body, warn
}

// trimQuoted 仅当值整体被成对引号包裹时剥除——值内部以引号结尾的合法内容
// （如：他说 "hello"）不受影响。
func trimQuoted(v string) string {
	for _, q := range []string{`"`, `'`} {
		if len(v) >= 2 && strings.HasPrefix(v, q) && strings.HasSuffix(v, q) {
			return v[1 : len(v)-1]
		}
	}
	return v
}

// stripBOM 剥 UTF-8 BOM（Windows 编辑器常见产物；带 BOM 的 SKILL.md 若不剥，
// "---" 前缀判定失败、整份被当正文、元数据静默丢失）。
func stripBOM(content string) string {
	return strings.TrimPrefix(content, "\ufeff")
}

// trimDocHead 剥 BOM 与前导空白（两解析器入口同一口径；围栏后语义各自保持）。
func trimDocHead(content string) string {
	return strings.TrimLeft(stripBOM(content), " \t\r\n")
}

// ParseRichFrontmatter 用 yaml 解析 SKILL.md frontmatter（name/description/mode/
// version/maturity/requires_mcp/deprecated/compatibility）。返回元数据与正文。
// frontmatter name 写入 Title（展示名）；目录名仍作规范 Name。
// 严格策略：起始围栏出现但未闭合 → 报错（Library fail-fast，与 FileProvider 宽松回退相对）。
func ParseRichFrontmatter(dirName, content string) (LibMeta, string, error) {
	meta := LibMeta{Name: dirName, Mode: ModeStatic}
	body := content
	trimmed := trimDocHead(content)
	if !strings.HasPrefix(trimmed, "---") {
		return meta, body, nil
	}
	fmRaw, body, ok := splitFrontmatter(trimmed)
	if !ok {
		return meta, body, fmt.Errorf("skill: frontmatter 未闭合")
	}

	// 直接解进 LibMeta（schema 唯一事实源）；Name/Title 语义 fixup 在下方。
	var fm LibMeta
	if err := yaml.Unmarshal([]byte(fmRaw), &fm); err != nil {
		return meta, body, fmt.Errorf("skill: frontmatter YAML 非法: %w", err)
	}
	meta = fm
	meta.Name = dirName  // 目录名作规范名
	meta.Title = fm.Name // frontmatter name 作展示名
	if meta.Mode == "" {
		meta.Mode = ModeStatic
	}
	return meta, body, nil
}
