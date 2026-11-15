package lineage

// 2026-11-15 第四轮审计清零回归（#120）：Focus depth=2 的无向边去重。

import "testing"

// TestFocusEdgesDeduped_Issue120 节点有 seen 去重、边此前没有：depth=2 时
// 第一跳邻居的邻接含焦点自身（UsedBy/GrantedBy 本就是反向边），同一关系以
// A→B 与 B→A 各输出一次。现按「两端有序对 + Relation」去重，每条无向关系
// 恰出现一次，节点集合不变。
func TestFocusEdgesDeduped_Issue120(t *testing.T) {
	lin, _ := fixture()
	res, ok := lin.Focus("sec", 2)
	if !ok {
		t.Fatal("sec 焦点应存在")
	}
	// 节点集不受边去重影响：server sec + expert specialists/sec + skill sop/pod
	if len(res.Nodes) != 4 {
		t.Fatalf("depth=2 节点集合错误（去重不得误删节点）: %+v", res.Nodes)
	}
	// 无向关系恰出现一次
	seen := map[string]int{}
	for _, e := range res.Edges {
		a := e.From.Kind + "/" + e.From.Name
		b := e.To.Kind + "/" + e.To.Name
		if a > b {
			a, b = b, a
		}
		seen[a+"|"+b+"|"+e.Relation]++
	}
	for k, n := range seen {
		if n != 1 {
			t.Fatalf("无向关系 %s 出现 %d 次（skillkit#120）: %+v", k, n, res.Edges)
		}
	}
	// 反向验证：不得同时存在 A→B 与 B→A（同 relation）
	dirs := map[string]bool{}
	for _, e := range res.Edges {
		from := e.From.Kind + "/" + e.From.Name
		to := e.To.Kind + "/" + e.To.Name
		if dirs[to+"->"+from+"|"+e.Relation] {
			t.Fatalf("同一关系双向重复输出: %+v", res.Edges)
		}
		dirs[from+"->"+to+"|"+e.Relation] = true
	}
	if len(res.Edges) != len(seen) {
		t.Fatalf("边集合含重复: edges=%d unique=%d", len(res.Edges), len(seen))
	}
	// 夹具推演：grants（sec↔expert）、requires_mcp（sec↔sop）、skill（expert↔sop、
	// expert↔pod）共 4 条无向关系
	if len(res.Edges) != 4 {
		t.Fatalf("depth=2 边集合应含 4 条无向关系，实得 %d: %+v", len(res.Edges), res.Edges)
	}

	// depth=1 焦点星形边不误伤（sec 的两条一跳邻接）
	res1, _ := lin.Focus("sec", 1)
	if len(res1.Edges) != 2 {
		t.Fatalf("depth=1 边集合应 2 条: %+v", res1.Edges)
	}
}
