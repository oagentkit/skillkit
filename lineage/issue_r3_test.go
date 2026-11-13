package lineage

// 2026-11-13 第三轮清零回归（#112）：#108 的指针字段残留面——Deprecated 与
// Manifest 指针不得别名入图，Build 后调用方改指针目标不得穿透已发布快照。

import (
	"testing"

	"github.com/oagentkit/skillkit/pack"
	"github.com/oagentkit/skillkit/skill"
)

// TestBuildCopiesPointerFields_Issue112 Build 后改调用方的 *Deprecated /
// *Manifest（含 Tools 切片），已发布快照必须原样——与 #108 的切片口径对齐。
func TestBuildCopiesPointerFields_Issue112(t *testing.T) {
	dep := &skill.Deprecated{Reason: "old", ReplacedBy: "new-skill", RemoveAfter: "2026-12-01"}
	man := &pack.ToolManifest{
		Server: "srv", Name: "srv", Version: "1.0.0",
		Tools: []pack.ManifestTool{{Name: "t1", Desc: "d1"}},
	}
	lin := Build(
		nil,
		[]Skill{{Name: "s1", Deprecated: dep}},
		map[string]*pack.ToolManifest{"srv": man},
	)

	// 指针目标是活数据：装配方后续读改（library 生命周期路径同款写法）。
	dep.RemoveAfter = "2099-01-01"
	dep.Reason = "MUT"
	man.Version = "9.9.9"
	man.Tools[0].Name = "MUT"
	man.Tools = append(man.Tools, pack.ManifestTool{Name: "evil"})

	se := lin.Skills["s1"]
	if se.Deprecated == nil {
		t.Fatal("Deprecated 应入图")
	}
	if se.Deprecated.RemoveAfter != "2026-12-01" || se.Deprecated.Reason != "old" ||
		se.Deprecated.ReplacedBy != "new-skill" {
		t.Fatalf("Deprecated 指针别名穿透进图（skillkit#112）: %+v", se.Deprecated)
	}
	mc := lin.MCPS["srv"].Manifest
	if mc == nil {
		t.Fatal("Manifest 应入图")
	}
	if mc.Version != "1.0.0" || len(mc.Tools) != 1 || mc.Tools[0].Name != "t1" {
		t.Fatalf("Manifest 指针别名穿透进图（skillkit#112）: %+v", mc)
	}

	// Hub 发布后再写：并发读者可见的快照不变（#108 的验收面延伸到指针字段）。
	h := NewHub()
	h.Set(lin, nil)
	dep.RemoveAfter = "2100-01-01"
	man.Version = "0.0.1"
	got := h.Get()
	if got.Skills["s1"].Deprecated.RemoveAfter != "2026-12-01" {
		t.Fatalf("已发布快照 Deprecated 被写穿: %+v", got.Skills["s1"].Deprecated)
	}
	if got.MCPS["srv"].Manifest.Version != "1.0.0" {
		t.Fatalf("已发布快照 Manifest 被写穿: %+v", got.MCPS["srv"].Manifest)
	}
}

// TestSkillFromMetaCopiesDeprecatedPointer_Issue112 投影期即断开指针（issue
// 列的 :59 面）：SkillFromMeta 返回的输入结构不与 LibMeta 共享 Deprecated。
func TestSkillFromMetaCopiesDeprecatedPointer_Issue112(t *testing.T) {
	meta := skill.LibMeta{
		Name:       "s2",
		Deprecated: &skill.Deprecated{RemoveAfter: "2026-12-01"},
	}
	s := SkillFromMeta(meta)
	meta.Deprecated.RemoveAfter = "2099-01-01"
	if s.Deprecated == nil || s.Deprecated.RemoveAfter != "2026-12-01" {
		t.Fatalf("SkillFromMeta 投影与 LibMeta 共享 Deprecated 指针（skillkit#112）: %+v", s.Deprecated)
	}
	// nil 形态保持 nil（不造空结构）。
	if got := SkillFromMeta(skill.LibMeta{Name: "s3"}); got.Deprecated != nil {
		t.Fatalf("无 Deprecated 应保持 nil: %+v", got.Deprecated)
	}
}
