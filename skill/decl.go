package skill

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// Decl 技能的结构化声明引用（agent.yaml skills / pack.yaml depends_on.skills 等
// 装配声明的列表元素）。兼容两种 yaml 形态：裸串（=仅名字，无约束）与
// {name, version, optional} 映射。version 是 semver 区间表达式
// （">=1.0.0 <2.0.0"，空格=AND）；空=不约束。Optional=true 表示引用缺失时
// 允许警告降级（是否降级由装配方决定，本包不做存在性检查）。
type Decl struct {
	Name     string `yaml:"name" json:"name"`
	Version  string `yaml:"version,omitempty" json:"version,omitempty"`
	Optional bool   `yaml:"optional,omitempty" json:"optional,omitempty"`
}

// UnmarshalYAML 兼容裸串与映射两种形态（yaml.v3）。
//
// 两种形态同一校验口径（oagentkit/skillkit#123）：name 先 TrimSpace 再判空
// ——此前映射分支只判 `v.Name == ""`，`{name: " sop"}` 与全空白 `{name: " "}`
// 原样透传，与裸串 `- " sop"` 的行为分裂。映射分支还做键白名单校验：yaml.v3
// 的宽松 Decode 会静默丢弃未知键，`verison:` 类拼写让 Version 留空 = 版本
// 约束静默消失（正是 #37 在 Ref 面修掉的「约束裸奔」搬到声明面）——未知键
// fail-loud。
func (d *Decl) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind == yaml.ScalarNode {
		var s string
		if err := value.Decode(&s); err != nil {
			return err
		}
		d.Name = strings.TrimSpace(s)
		if d.Name == "" {
			return fmt.Errorf("skill: 技能引用为空")
		}
		return nil
	}
	if value.Kind == yaml.MappingNode {
		// 未知键 fail-loud（宽松 Decode 会静默忽略，拼写错误无信号）
		for i := 0; i+1 < len(value.Content); i += 2 {
			k := value.Content[i]
			if k.Kind != yaml.ScalarNode {
				return fmt.Errorf("skill: 技能引用含非标量键")
			}
			switch k.Value {
			case "name", "version", "optional":
			default:
				hint := ""
				if near := nearestDeclKey(k.Value); near != "" {
					hint = fmt.Sprintf("（是否想写 %q？）", near)
				}
				return fmt.Errorf("skill: 技能引用含未知键 %q%s（支持 name/version/optional）", k.Value, hint)
			}
		}
	}
	type plain Decl
	var v plain
	if err := value.Decode(&v); err != nil {
		return err
	}
	v.Name = strings.TrimSpace(v.Name)
	if v.Name == "" {
		return fmt.Errorf("skill: 技能引用缺 name")
	}
	*d = Decl(v)
	return nil
}

// nearestDeclKey 未知键与白名单键的编辑距离 ≤2 时返回最近者——给
// `verison:`/`verson:` 类拼写一个「是否想写」提示，而不只是「未知键」。
func nearestDeclKey(key string) string {
	best, bestDist := "", 3 // 距离 ≥3 不给提示
	for _, want := range []string{"name", "version", "optional"} {
		if d := editDistance(key, want); d < bestDist {
			best, bestDist = want, d
		}
	}
	return best
}

// editDistance Levenshtein 距离（两行滚动，O(min(m,n)) 空间）。
func editDistance(a, b string) int {
	ar, br := []rune(a), []rune(b)
	if len(ar) > len(br) {
		ar, br = br, ar
	}
	prev := make([]int, len(ar)+1)
	cur := make([]int, len(ar)+1)
	for i := range prev {
		prev[i] = i
	}
	for j := 1; j <= len(br); j++ {
		cur[0] = j
		for i := 1; i <= len(ar); i++ {
			cost := 1
			if ar[i-1] == br[j-1] {
				cost = 0
			}
			cur[i] = min(prev[i]+1, cur[i-1]+1, prev[i-1]+cost)
		}
		prev, cur = cur, prev
	}
	return prev[len(ar)]
}
