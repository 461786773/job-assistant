package prompts

import "testing"

func TestPublicCopyHasUserFacingTemplates(t *testing.T) {
	c := PublicCopy()
	for _, key := range []string{"crisisHelp", "assessmentBoundaryNote", "assessmentCrisisHeadline"} {
		s, _ := c[key].(string)
		if s == "" {
			t.Fatalf("%s 为空，用户侧文案须集中在 prompts 包", key)
		}
	}
	steps, _ := c["assessmentCrisisSteps"].([]string)
	if len(steps) == 0 {
		t.Fatal("assessmentCrisisSteps 为空")
	}
	if CrisisHelp != c["crisisHelp"] {
		t.Fatal("PublicCopy.crisisHelp 必须与 CrisisHelp 同一份")
	}
}
