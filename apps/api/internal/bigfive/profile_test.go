package bigfive

import (
	"strings"
	"testing"
)

func TestBuildScoringAndPersona(t *testing.T) {
	// 高 C、低 E → quiet_architect；反向题也测一下
	a := Answers{
		Q1: 4, Q2: 2, // O mid-high
		Q3: 5, Q4: 5, Q5: 1, // C high (s5 reverse: 5)
		Q6: 2, Q7: 5, // E low (s7 reverse: 1)
		Q8: 3, Q9: 3,
		Q10: 3, Q11: 3, Q12: 3,
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
	if p.PersonaBody == "" || strings.Contains(p.PersonaBody, "开放性") {
		t.Fatalf("body should be warm prose, got %q", p.PersonaBody)
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
