package lineage

import (
	"slices"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/oagentkit/skillkit/pack"
	"github.com/oagentkit/skillkit/skill"
)

// fixture 专家 sec（挂摘覆盖引用 sop+pod，授予 sec server）；技能 sop requires_mcp sec；
// 契约清单 sec v2（collect/inspect）。
func fixture() (*Lineage, map[string]*pack.ToolManifest) {
	experts := []Expert{{
		Slug: "specialists/sec", Overridden: true,
		Skills: []string{"sop", "pod"},
		Grants: []Grant{{Server: "sec", Allow: []string{"collect", "undeclared-x"}}},
	}}
	skills := []Skill{
		{Name: "sop", Source: "_shared", Version: "1.1.0",
			RequiresMCP: []skill.MCPDep{{Server: "sec"}}},
		{Name: "pod", Source: "k8s", Version: "0.9.0"},
	}
	manifests := map[string]*pack.ToolManifest{
		"sec": {Server: "sec", Name: "sec", Version: "2.0.0",
			Tools: []pack.ManifestTool{{Name: "collect"}, {Name: "inspect"}}, From: "k8s"},
	}
	return Build(experts, skills, manifests), manifests
}

// TestBuild 边组装：used_by 覆盖感知、grants、requires_mcp、manifest 挂载。
func TestBuild(t *testing.T) {
	lin, _ := fixture()
	e := lin.Experts["specialists/sec"]
	if !e.Overridden || !slices.Equal(e.Skills, []string{"pod", "sop"}) ||
		!slices.Equal(e.Grants, []string{"sec"}) {
		t.Fatalf("专家边错误: %+v", e)
	}
	se := lin.Skills["sop"]
	if !slices.Equal(se.UsedBy, []string{"specialists/sec"}) || se.Source != "_shared" || se.Version != "1.1.0" {
		t.Fatalf("技能边错误: %+v", se)
	}
	m := lin.MCPS["sec"]
	if !slices.Equal(m.GrantedBy, []string{"specialists/sec"}) ||
		!slices.Equal(m.Tools, []string{"collect", "undeclared-x"}) ||
		m.Manifest == nil || m.Manifest.Version != "2.0.0" {
		t.Fatalf("MCP 边错误: %+v", m)
	}
	if !slices.Equal(m.UsedBySkills, []string{"sop"}) {
		t.Fatalf("UsedBySkills 错误: %+v", m)
	}
	// 未引用技能也登记（焦点查询可及）。
	if _, ok := lin.Skills["pod"]; !ok {
		t.Fatal("未引用技能应登记")
	}
	// nil 安全：nil 图 Focus 不炸。
	if _, ok := (*Lineage)(nil).Focus("x", 1); ok {
		t.Fatal("nil 图应返回 false")
	}
}

// TestFocus 焦点邻接子图：技能焦点含专家与 server；未知焦点 false；深度裁剪。
func TestFocus(t *testing.T) {
	lin, _ := fixture()
	res, ok := lin.Focus("sop", 1)
	if !ok {
		t.Fatal("sop 焦点应存在")
	}
	kinds := map[string]bool{}
	for _, n := range res.Nodes {
		kinds[n.Kind+"_"+n.Name] = true
	}
	if !kinds["expert_specialists/sec"] || !kinds["skill_sop"] || !kinds["server_sec"] {
		t.Fatalf("邻接节点不全: %+v", res.Nodes)
	}
	// 反向：server 焦点能回到专家与依赖技能。
	res2, _ := lin.Focus("sec", 2)
	names := map[string]bool{}
	for _, n := range res2.Nodes {
		names[n.Kind+"_"+n.Name] = true
	}
	if !names["expert_specialists/sec"] || !names["skill_sop"] {
		t.Fatalf("深度 2 邻接不全: %+v", res2.Nodes)
	}
	if _, ok := lin.Focus("nosuch", 1); ok {
		t.Fatal("未知焦点应 false")
	}
}

// TestDiff 版本/成熟度/弃用跃迁、工具面增减、引用边增减；nil 基线=首帧无 diff。
func TestDiff(t *testing.T) {
	prev, _ := fixture()
	next := Build([]Expert{{
		Slug:   "specialists/sec",
		Skills: []string{"pod"},
		Grants: []Grant{{Server: "sec", Allow: []string{"collect"}}},
	}}, []Skill{
		{Name: "sop", Source: "_shared", Version: "1.2.0", Maturity: skill.MaturityDeprecated,
			Deprecated: &skill.Deprecated{RemoveAfter: "2026-01-01"}},
		{Name: "pod", Source: "k8s", Version: "0.9.0"},
	}, map[string]*pack.ToolManifest{
		"sec": {Server: "sec", Name: "sec", Version: "2.0.0",
			Tools: []pack.ManifestTool{{Name: "collect"}, {Name: "probe"}}, From: "k8s"},
	})

	impacts := Diff(prev, next)
	byType := map[string][]Impact{}
	for _, im := range impacts {
		byType[im.Type] = append(byType[im.Type], im)
	}
	if len(byType["skill_version_changed"]) != 1 || byType["skill_version_changed"][0].From != "1.1.0" || byType["skill_version_changed"][0].To != "1.2.0" {
		t.Fatalf("版本跃迁 impact 错误: %+v", byType["skill_version_changed"])
	}
	if len(byType["skill_maturity_changed"]) != 1 {
		t.Fatalf("成熟度跃迁 impact 缺失: %+v", byType)
	}
	if len(byType["deprecation"]) != 1 || byType["deprecation"][0].To != "2026-01-01" {
		t.Fatalf("弃用跃迁 impact 错误: %+v", byType["deprecation"])
	}
	tc := byType["tools_changed"]
	// 并集语义：prev 显式授予 undeclared-x 也计入（manifest∪授予），cur 侧它消失 → removed。
	if len(tc) != 1 || !slices.Equal(tc[0].ToolsAdded, []string{"probe"}) ||
		!slices.Equal(tc[0].ToolsRemoved, []string{"inspect", "undeclared-x"}) {
		t.Fatalf("工具面 impact 错误: %+v", tc)
	}
	rc := byType["refs_changed"]
	// v0.10.14：refs_changed 只填 Skills*（旧 tools_* 同步填充已按无兼容层纪律删除）。
	if len(rc) != 1 || rc[0].Expert != "specialists/sec" ||
		len(rc[0].ToolsAdded) != 0 || len(rc[0].ToolsRemoved) != 0 ||
		!slices.Equal(rc[0].SkillsRemoved, []string{"sop"}) {
		t.Fatalf("引用边 impact 错误: %+v", rc)
	}
	if got := Diff(prev, prev); len(got) != 0 {
		t.Fatalf("同图 diff 应为空: %+v", got)
	}
	if got := Diff(nil, next); len(got) != 0 {
		t.Fatalf("首帧应无 impact: %+v", got)
	}
}

// TestHub Set/Resync/Get/Manifests 语义与 nil 安全。
func TestHub(t *testing.T) {
	var nilHub *Hub
	nilHub.Resync(nil) // no-op
	if nilHub.Get() != nil || nilHub.Manifests() != nil {
		t.Fatal("nil hub 应安全 no-op")
	}
	lin, manifests := fixture()
	h := NewHub()
	h.Set(lin, manifests)
	if h.Get() != lin {
		t.Fatal("Set 应存图并返回")
	}
	// Resync 只换图，清单沿用。
	next := Build(nil, nil, nil)
	h.Resync(next)
	if h.Get() != next || h.Manifests() == nil {
		t.Fatal("Resync 应换图留清单")
	}
}

func TestSkillFromMeta(t *testing.T) {
	// LibMeta → Skill 投影的字段映射（映射唯一事实源的回归钉）。
	dep := &skill.Deprecated{Reason: "r", RemoveAfter: "2026-12-31"}
	meta := skill.LibMeta{
		Name: "sop", Version: "1.2.0", Maturity: skill.MaturityStable, Source: "k8s",
		RequiresMCP: []skill.MCPDep{{Server: "sec", Tools: []string{"collect"}}},
		Provides:    []string{"triage"},
		Deprecated:  dep,
		// 以下字段不应进入血缘输入
		Title: "展示名", Path: "k8s/skills/sop/SKILL.md", Description: "desc",
	}
	got := SkillFromMeta(meta)
	// Deprecated 按值比对（skillkit#112 起投影期解引用复制，指针同一性不再是
	// 契约——共享指针正是被修掉的别名面）。
	if got.Name != "sop" || got.Version != "1.2.0" || got.Maturity != "stable" ||
		got.Source != "k8s" || got.Deprecated == nil || *got.Deprecated != *dep ||
		!slices.Equal(got.Provides, []string{"triage"}) ||
		len(got.RequiresMCP) != 1 || got.RequiresMCP[0].Server != "sec" {
		t.Fatalf("SkillFromMeta 映射不符: %+v", got)
	}
}

// sortFocus 边全键比较（skillkit#38）：同名异类节点（skill x / server x）
// 间序此前不稳定，输出跨 run 不可复现违反确定性契约——同输入两次排序结果
// 必须逐边相等（多次重复以暴露不稳定）。
func TestSortFocusDeterministicSameNameDifferentKind(t *testing.T) {
	build := func() *LineageFocus {
		res := &LineageFocus{}
		for i := 0; i < 32; i++ {
			res.Edges = append(res.Edges,
				LineageEdge{From: LineageNode{Kind: "skill", Name: "x"}, To: LineageNode{Kind: "server", Name: "x"}, Relation: "calls"},
				LineageEdge{From: LineageNode{Kind: "skill", Name: "x"}, To: LineageNode{Kind: "tool", Name: "x"}, Relation: "lists"},
			)
		}
		sortFocus(res)
		return res
	}
	first := build()
	for round := 0; round < 16; round++ {
		next := build()
		for i := range first.Edges {
			if first.Edges[i] != next.Edges[i] {
				t.Fatalf("第 %d 轮排序不稳定（同名异类节点间序漂移）: 边 %d %+v vs %+v", round, i, first.Edges[i], next.Edges[i])
			}
		}
	}
}

// TestBuildCopiesCallerInputs（skillkit#108）：Build 是「只读图机制」入口——
// 不得原地改写装配方输入（stabilize 的排序不得写穿），图内集合不得与装配方
// 底层数组同址（否则装配方复用切片会穿透改写 Hub 已发布的并发读快照）。
func TestBuildCopiesCallerInputs(t *testing.T) {
	skills := []string{"zeta", "alpha"}
	provides := []string{"p1", "p2"}
	deps := []skill.MCPDep{{Server: "sec", Tools: []string{"collect"}}}
	lin := Build(
		[]Expert{{Slug: "e1", Skills: skills}},
		[]Skill{{Name: "s1", RequiresMCP: deps, Provides: provides}},
		nil,
	)

	// 写穿：Build 不得重排装配方给的顺序。
	if !slices.Equal(skills, []string{"zeta", "alpha"}) {
		t.Fatalf("Build 原地重排了调用方 Expert.Skills：%v", skills)
	}
	// 别名穿透矩阵：改写装配方底层数组，图内三项 + dep.Tools 均不得跟随。
	skills[0] = "MUT"
	provides[0] = "MUT"
	deps[0].Server = "MUT"
	deps[0].Tools[0] = "MUT"
	if got := lin.Experts["e1"].Skills; !slices.Equal(got, []string{"alpha", "zeta"}) {
		t.Fatalf("ExpertEdges.Skills 与调用方共享底层数组（排序也写穿）：%v", got)
	}
	se := lin.Skills["s1"]
	if !slices.Equal(se.Provides, []string{"p1", "p2"}) {
		t.Fatalf("SkillEdges.Provides 与调用方共享底层数组：%v", se.Provides)
	}
	if len(se.RequiresMCP) != 1 || se.RequiresMCP[0].Server != "sec" ||
		!slices.Equal(se.RequiresMCP[0].Tools, []string{"collect"}) {
		t.Fatalf("SkillEdges.RequiresMCP（含 Tools 明细）与调用方共享底层数组：%+v", se.RequiresMCP)
	}

	// 已发布快照：Hub.Set 之后装配方再写，快照不得变（并发读者看到的图不可变）。
	h := NewHub()
	h.Set(lin, nil)
	skills[1] = "MUT"
	provides[1] = "MUT"
	if got := h.Get().Experts["e1"].Skills; !slices.Equal(got, []string{"alpha", "zeta"}) {
		t.Fatalf("已发布快照被装配方写穿：%v", got)
	}
	if got := h.Get().Skills["s1"].Provides; !slices.Equal(got, []string{"p1", "p2"}) {
		t.Fatalf("已发布快照 Provides 被装配方写穿：%v", got)
	}

	// 别名链源头：SkillFromMeta 按引用投影 LibMeta，Build 后改写 LibMeta 不得进图。
	meta := skill.LibMeta{
		Name:        "s2",
		Provides:    []string{"x"},
		RequiresMCP: []skill.MCPDep{{Server: "srv2", Tools: []string{"t2"}}},
	}
	lin2 := Build(nil, []Skill{SkillFromMeta(meta)}, nil)
	meta.Provides[0] = "MUT"
	meta.RequiresMCP[0].Server = "MUT"
	meta.RequiresMCP[0].Tools[0] = "MUT"
	s2 := lin2.Skills["s2"]
	if !slices.Equal(s2.Provides, []string{"x"}) ||
		s2.RequiresMCP[0].Server != "srv2" || !slices.Equal(s2.RequiresMCP[0].Tools, []string{"t2"}) {
		t.Fatalf("LibMeta 别名链穿透进图：%+v", s2)
	}
}

// TestBuildDedupsGrantAndDepReverseFields（skillkit#109）：同 server 多 grant /
// 多依赖声明时，grants/granted_by/used_by_skills 去重——与同循环内 Tools 的
// 并集口径一致（反查字段是单一事实源，重复项会流进全部 list API 与 Diff impact）。
func TestBuildDedupsGrantAndDepReverseFields(t *testing.T) {
	lin := Build(
		[]Expert{{
			Slug:   "e1",
			Skills: []string{"s1"},
			Grants: []Grant{
				{Server: "sec", Allow: []string{"a"}},
				{Server: "sec", Allow: []string{"b"}},
			},
		}},
		[]Skill{{
			Name:        "s1",
			RequiresMCP: []skill.MCPDep{{Server: "sec"}, {Server: "sec", Tools: []string{"b"}}},
		}},
		nil,
	)
	m := lin.MCPS["sec"]
	if !slices.Equal(m.GrantedBy, []string{"e1"}) {
		t.Fatalf("同 server 多 grant 未去重 GrantedBy：%v", m.GrantedBy)
	}
	if !slices.Equal(m.UsedBySkills, []string{"s1"}) {
		t.Fatalf("同 server 多依赖声明未去重 UsedBySkills：%v", m.UsedBySkills)
	}
	if !slices.Equal(lin.Experts["e1"].Grants, []string{"sec"}) {
		t.Fatalf("同 server 多 grant 未去重 Grants：%v", lin.Experts["e1"].Grants)
	}
	// 并集不回归：两笔 grant 的 allow 仍是并集去重。
	if !slices.Equal(m.Tools, []string{"a", "b"}) {
		t.Fatalf("Tools 并集口径回归：%v", m.Tools)
	}

	// 跨专家：不同专家各记一次，同专家重复 grant 不叠记。
	lin2 := Build([]Expert{
		{Slug: "e1", Grants: []Grant{{Server: "sec"}, {Server: "sec"}}},
		{Slug: "e2", Grants: []Grant{{Server: "sec"}}},
	}, nil, nil)
	if !slices.Equal(lin2.MCPS["sec"].GrantedBy, []string{"e1", "e2"}) {
		t.Fatalf("跨专家 GrantedBy 错误：%v", lin2.MCPS["sec"].GrantedBy)
	}
}

// TestHubResyncConcurrentSetNoLostUpdate（skillkit#110）：Resync 的「沿用上次 Set
// 的清单」是 atomic.Pointer 上的读-改-写——若非原子（Load 与 Store 之间被并发
// Set 插入），会把更新的契约清单静默回滚成旧版。两条不变量钉住根因：
//  1. 终态：全部写方落定后，清单必须是本轮 Set 写入的 2.0.0（Resync 只换图）；
//  2. 过程：观察方（watcher）看到的清单版本只前进不后退（Resync 恒不改清单值）。
//
// 正确实现（CAS 循环）下两条恒成立；非原子 RMW 下随并发度放大命中。
func TestHubResyncConcurrentSetNoLostUpdate(t *testing.T) {
	h := NewHub()
	lin := Build(nil, nil, nil)
	manifest := func(v string) map[string]*pack.ToolManifest {
		return map[string]*pack.ToolManifest{"srv": {Server: "srv", Version: v}}
	}
	const rounds = 20000
	for r := 0; r < rounds; r++ {
		h.Set(lin, manifest("1.0.0"))
		var writes, watch sync.WaitGroup
		var done, sawNew, regressed atomic.Bool

		// 观察方：清单版本只前进不后退。
		watch.Add(1)
		go func() {
			defer watch.Done()
			for !done.Load() {
				m := h.Manifests()["srv"]
				if m == nil {
					continue
				}
				if m.Version == "2.0.0" {
					sawNew.Store(true)
				} else if sawNew.Load() {
					regressed.Store(true)
					return
				}
			}
		}()

		for i := 0; i < 4; i++ {
			writes.Add(2)
			go func() {
				defer writes.Done()
				for j := 0; j < 4; j++ {
					h.Set(lin, manifest("2.0.0"))
				}
			}()
			go func() {
				defer writes.Done()
				for j := 0; j < 4; j++ {
					h.Resync(lin)
				}
			}()
		}
		writes.Wait()
		done.Store(true)
		watch.Wait()

		if regressed.Load() {
			t.Fatalf("第 %d 轮：观察方看到清单版本回退——Resync 用旧清单覆盖并发 Set", r)
		}
		if m := h.Manifests()["srv"]; m == nil || m.Version != "2.0.0" {
			t.Fatalf("第 %d 轮：Resync 把并发 Set 的新清单静默回滚，终态清单=%+v", r, m)
		}
	}
}
