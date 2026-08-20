package bigfive

import (
	"strings"
	"testing"
)

func TestBuildScoringAndPersona(t *testing.T) {
	// 高 C、低 E → quiet_architect；反向题也测一下（每维 3 题）
	a := Answers{
		Q1: 4, Q2: 4, Q3: 2, // O mid-high (s3 reverse → 4)
		Q4: 5, Q5: 5, Q6: 1, // C high (s6 reverse → 5)
		Q7: 2, Q8: 2, Q9: 5, // E low (s9 reverse → 1)
		Q10: 3, Q11: 3, Q12: 3,
		Q13: 3, Q14: 3, Q15: 3,
	}
	p, err := Build(a)
	if err != nil {
		t.Fatal(err)
	}
	if p.Scores.Conscientiousness.Band != "high" {
		t.Fatalf("C band=%s mean=%.1f", p.Scores.Conscientiousness.Band, p.Scores.Conscientiousness.Mean)
	}
	if p.Scores.Extraversion.Band != "low" {
		t.Fatalf("E band=%s mean=%.1f", p.Scores.Extraversion.Band, p.Scores.Extraversion.Mean)
	}
	if p.PersonaID != "quiet_architect" {
		t.Fatalf("persona=%s want quiet_architect", p.PersonaID)
	}
	if p.PersonaBody == "" || strings.Contains(p.PersonaBody, "硬伤：") || strings.Contains(p.PersonaBody, "双刃：") {
		t.Fatalf("body should be one paragraph without 双刃/硬伤 labels, got %q", p.PersonaBody)
	}
	if len(p.Tags) == 0 || len(p.CoachHints) == 0 {
		t.Fatalf("tags/hints empty: %v %v", p.Tags, p.CoachHints)
	}
}

func TestValidate(t *testing.T) {
	a := Answers{Q1: 0}
	if err := Validate(a); err == nil {
		t.Fatal("expected validate error")
	}
}

func TestSoftComboAvoidsGenericSingle(t *testing.T) {
	// 多数题偏中，但 A/C 相对更高 → 应落到组合/双维，而非空泛单维
	a := Answers{
		Q1: 3, Q2: 3, Q3: 3,
		Q4: 4, Q5: 4, Q6: 2,
		Q7: 3, Q8: 3, Q9: 3,
		Q10: 4, Q11: 4, Q12: 2,
		Q13: 3, Q14: 3, Q15: 3,
	}
	p, err := Build(a)
	if err != nil {
		t.Fatal(err)
	}
	generic := map[string]bool{
		"open_explorer": true, "reliable_closer": true, "social_spark": true,
		"harmony_keeper": true, "deep_feeler": true,
	}
	if generic[p.PersonaID] {
		t.Fatalf("persona too generic: %s (%s)", p.PersonaID, p.PersonaTitle)
	}
	if p.PersonaBody == "" || len(p.PersonaBody) < 40 {
		t.Fatalf("body too thin: %q", p.PersonaBody)
	}
}

func TestEachDimThreeItemsMaxDisplay(t *testing.T) {
	// 五维全拉满 → 每维 display=100
	a := Answers{
		Q1: 5, Q2: 5, Q3: 1, // O: 5,5,5
		Q4: 5, Q5: 5, Q6: 1, // C
		Q7: 5, Q8: 5, Q9: 1, // E
		Q10: 5, Q11: 5, Q12: 1, // A
		Q13: 5, Q14: 5, Q15: 1, // N
	}
	p, err := Build(a)
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range []struct {
		name string
		got  int
	}{
		{"O", p.Scores.Openness.Display},
		{"C", p.Scores.Conscientiousness.Display},
		{"E", p.Scores.Extraversion.Display},
		{"A", p.Scores.Agreeableness.Display},
		{"N", p.Scores.Neuroticism.Display},
	} {
		if d.got != 100 {
			t.Fatalf("%s display=%d want 100", d.name, d.got)
		}
	}
}
