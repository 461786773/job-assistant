package hr

import "testing"

func TestReplaceExact(t *testing.T) {
	got, ok, method := replaceOnce("hello world", "world", "地球")
	if !ok || method != "exact" || got != "hello 地球" {
		t.Fatalf("got=%q ok=%v method=%s", got, ok, method)
	}
}

func TestReplaceWhitespace(t *testing.T) {
	resume := "负责产品规划\n\n与研发协作"
	before := "负责产品规划 与研发协作"
	after := "负责 ToB 产品规划，并与研发协作落地"
	got, ok, method := replaceOnce(resume, before, after)
	if !ok || method != "whitespace" {
		t.Fatalf("ok=%v method=%s got=%q", ok, method, got)
	}
	if got != after {
		t.Fatalf("unexpected: %q", got)
	}
}

func TestApplyRewritesSequential(t *testing.T) {
	resume := "AAA BBB CCC"
	items := []RewriteItem{
		{Target: "1", Before: "AAA", After: "A1"},
		{Target: "2", Before: "BBB", After: "B2"},
	}
	out, results := ApplyRewrites(resume, items, []int{0, 1})
	if out != "A1 B2 CCC" {
		t.Fatalf("out=%q", out)
	}
	if !results[0].OK || !results[1].OK {
		t.Fatalf("results=%+v", results)
	}
}

func TestApplyMissing(t *testing.T) {
	out, results := ApplyRewrites("only this", []RewriteItem{
		{Before: "missing", After: "x"},
	}, []int{0})
	if out != "only this" || results[0].OK {
		t.Fatalf("out=%q results=%+v", out, results)
	}
}
