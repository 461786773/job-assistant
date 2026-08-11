package hr

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/zhangyongjie/job-assistant/internal/llm"
	"github.com/zhangyongjie/job-assistant/internal/prompts"
)

type aiApplyResponse struct {
	UpdatedResume string `json:"updatedResume"`
	Changes       []struct {
		Index  int    `json:"index"`
		OK     bool   `json:"ok"`
		Note   string `json:"note"`
		Target string `json:"target"`
	} `json:"changes"`
	Summary string `json:"summary"`
}

// ApplyRewritesWithAI asks the model to surgically edit resume for selected rewrite items.
func ApplyRewritesWithAI(client *llm.Client, resume, jd, company, role string, items []RewriteItem, indexes []int) (string, []ApplyResult, error) {
	if client == nil || !client.Enabled() {
		return "", nil, fmt.Errorf("LLM 未配置")
	}
	if strings.TrimSpace(resume) == "" {
		return "", nil, fmt.Errorf("简历为空")
	}

	type itemPayload struct {
		Index  int    `json:"index"`
		Target string `json:"target"`
		Action string `json:"action"`
		Before string `json:"before"`
		After  string `json:"after"`
		Reason string `json:"reason"`
	}
	selected := make([]itemPayload, 0, len(indexes))
	for _, idx := range indexes {
		if idx < 0 || idx >= len(items) {
			continue
		}
		it := items[idx]
		selected = append(selected, itemPayload{
			Index: idx, Target: it.Target, Action: it.Action,
			Before: it.Before, After: it.After, Reason: it.Reason,
		})
	}
	if len(selected) == 0 {
		return resume, nil, fmt.Errorf("没有有效的改写项")
	}

	rawItems, _ := json.MarshalIndent(selected, "", "  ")
	user := fmt.Sprintf(`公司：%s
目标岗位：%s

【目标 JD】
%s

【当前简历全文】
%s

【需要应用的改写项】
%s

请输出 JSON：
{
  "updatedResume": "应用改写后的完整简历正文",
  "changes": [
    {"index": 0, "ok": true, "target": "...", "note": "改了什么 / 为何跳过"}
  ],
  "summary": "一句话说明本次改动"
}
要求：changes 必须覆盖上面每一个 index；若某条无法安全落地则 ok=false 并说明原因，且不要为该条捏造事实。`,
		company, role, jd, resume, string(rawItems))

	content, err := client.ChatJSON(prompts.ResumeRewrite, user)
	if err != nil {
		return "", nil, err
	}
	content = stripCodeFence(content)
	var parsed aiApplyResponse
	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		return "", nil, fmt.Errorf("模型返回非预期 JSON: %w; 片段: %s", err, truncate(content, 200))
	}
	updated := strings.TrimSpace(parsed.UpdatedResume)
	if updated == "" {
		return "", nil, fmt.Errorf("模型未返回 updatedResume")
	}

	byIdx := map[int]ApplyResult{}
	for _, ch := range parsed.Changes {
		byIdx[ch.Index] = ApplyResult{
			Index:   ch.Index,
			Target:  ch.Target,
			OK:      ch.OK,
			Method:  "ai",
			Message: ch.Note,
		}
	}
	results := make([]ApplyResult, 0, len(indexes))
	for _, idx := range indexes {
		if r, ok := byIdx[idx]; ok {
			if r.Target == "" && idx >= 0 && idx < len(items) {
				r.Target = items[idx].Target
			}
			results = append(results, r)
			continue
		}
		// model omitted this index — infer from whether resume changed
		res := ApplyResult{Index: idx, Method: "ai", Message: "模型未回报该条"}
		if idx >= 0 && idx < len(items) {
			res.Target = items[idx].Target
		}
		if updated != strings.TrimSpace(resume) {
			res.OK = true
			res.Message = "已包含在整体改写中"
		}
		results = append(results, res)
	}

	anyOK := false
	for _, r := range results {
		if r.OK {
			anyOK = true
			break
		}
	}
	if !anyOK && updated == strings.TrimSpace(resume) {
		return resume, results, nil
	}
	return updated, results, nil
}
