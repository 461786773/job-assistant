package interview

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/zhangyongjie/job-assistant/internal/llm"
)

type ChatMsg struct {
	Role    string `json:"role"` // interviewer | candidate | coach
	Content string `json:"content"`
}

type Diagnosis struct {
	DepthScore   int      `json:"depthScore"`
	Summary      string   `json:"summary"`
	WeakPoints   []string `json:"weakPoints"`
	STARMissing  []string `json:"starMissing"`
	CoachScripts []string `json:"coachScripts"`
}

type Session struct {
	Messages  []ChatMsg  `json:"messages"`
	Round     int        `json:"round"`
	MaxRounds int        `json:"maxRounds"`
	Status    string     `json:"status"` // active | done
	Diagnosis *Diagnosis `json:"diagnosis,omitempty"`
	UpdatedAt string     `json:"updatedAt,omitempty"`
}

const systemPrompt = `你是主机厂/安全合规方向的业务面试官，面试对象是 3–10 年 ToB 产品经理。
目标：识别「讲述表层化」——只有背景职责、缺少取舍/个人贡献/结果/复盘。
规则：
1. 问法要具体，一次只问一个问题。
2. 追问优先：决策取舍、冲突推动、计划拆解、迁移边界、个人贡献。
3. 不要编造简历没有的经历。
4. 只输出 JSON。`

func Start(client *llm.Client, resume, jd, company, role string) (*Session, error) {
	sess := &Session{
		Messages:  []ChatMsg{},
		Round:     1,
		MaxRounds: 5,
		Status:    "active",
	}
	if client == nil || !client.Enabled() {
		q := "先用 1 分钟做自我介绍，并说明你和这个岗位最匹配的 2 个点。请尽量落到具体项目，而不是罗列公司名。"
		sess.Messages = append(sess.Messages, ChatMsg{Role: "interviewer", Content: q})
		return sess, nil
	}

	user := fmt.Sprintf(`公司：%s
岗位：%s

【JD】
%s

【简历摘要】
%s

请开场。输出 JSON：
{"question":"面试官第一问（自我介绍/匹配点）","intent":"intro"}`, company, role, truncate(jd, 2500), truncate(resume, 3500))

	raw, err := client.ChatJSON(systemPrompt, user)
	if err != nil {
		return nil, err
	}
	var out struct {
		Question string `json:"question"`
	}
	if err := decodeJSON(raw, &out); err != nil || strings.TrimSpace(out.Question) == "" {
		out.Question = "先用 1 分钟自我介绍，并说明你与本岗位最匹配的两点（请落到项目）。"
	}
	sess.Messages = append(sess.Messages, ChatMsg{Role: "interviewer", Content: out.Question})
	return sess, nil
}

func Reply(client *llm.Client, sess *Session, answer, resume, jd, company, role string) (*Session, error) {
	if sess == nil {
		return nil, fmt.Errorf("会话不存在，请先开始面试模拟")
	}
	if sess.Status == "done" {
		return sess, nil
	}
	answer = strings.TrimSpace(answer)
	if answer == "" {
		return nil, fmt.Errorf("请先输入回答")
	}
	sess.Messages = append(sess.Messages, ChatMsg{Role: "candidate", Content: answer})

	shouldFinish := sess.Round >= sess.MaxRounds
	if client == nil || !client.Enabled() {
		return heuristicReply(sess, answer, shouldFinish), nil
	}

	hist := formatHistory(sess.Messages)
	if shouldFinish {
		user := fmt.Sprintf(`面试结束，请诊断候选人是否「表层化」。

公司：%s 岗位：%s
【JD】
%s
【简历摘要】
%s
【对话】
%s

输出 JSON：
{
  "diagnosis": {
    "depthScore": 0-100,
    "summary": "一句话结论",
    "weakPoints": ["..."],
    "starMissing": ["S|T|A|R 缺什么"],
    "coachScripts": ["60-90秒补强口述稿1", "稿2"]
  }
}`, company, role, truncate(jd, 2000), truncate(resume, 2500), hist)

		raw, err := client.ChatJSON(systemPrompt+"\n现在进入总结诊断阶段。", user)
		if err != nil {
			return nil, err
		}
		var out struct {
			Diagnosis Diagnosis `json:"diagnosis"`
		}
		if err := decodeJSON(raw, &out); err != nil {
			return nil, fmt.Errorf("诊断解析失败: %w", err)
		}
		sess.Diagnosis = &out.Diagnosis
		sess.Status = "done"
		sess.Messages = append(sess.Messages, ChatMsg{
			Role:    "coach",
			Content: formatCoach(out.Diagnosis),
		})
		return sess, nil
	}

	user := fmt.Sprintf(`根据候选人上一答，决定下一问（反表层化追问）。当前第 %d/%d 轮。

公司：%s 岗位：%s
【JD】
%s
【简历摘要】
%s
【对话】
%s

输出 JSON：
{
  "surfaceLevel": true/false,
  "briefFeedback": "一句话点评（可空）",
  "question": "下一追问",
  "intent": "deep_dive|pressure|plan_split|tradeoff"
}`, sess.Round+1, sess.MaxRounds, company, role, truncate(jd, 2000), truncate(resume, 2500), hist)

	raw, err := client.ChatJSON(systemPrompt, user)
	if err != nil {
		return nil, err
	}
	var out struct {
		BriefFeedback string `json:"briefFeedback"`
		Question      string `json:"question"`
	}
	if err := decodeJSON(raw, &out); err != nil || strings.TrimSpace(out.Question) == "" {
		out.Question = "刚才这件事情里，你做了哪些取舍？如果只保一个目标，你放弃了什么？为什么？"
	}
	if strings.TrimSpace(out.BriefFeedback) != "" {
		sess.Messages = append(sess.Messages, ChatMsg{Role: "coach", Content: out.BriefFeedback})
	}
	sess.Messages = append(sess.Messages, ChatMsg{Role: "interviewer", Content: out.Question})
	sess.Round++
	return sess, nil
}

func heuristicReply(sess *Session, answer string, finish bool) *Session {
	shallow := len([]rune(answer)) < 80 || (!strings.Contains(answer, "我") && strings.Contains(answer, "我们"))
	if finish || sess.Round >= sess.MaxRounds {
		d := Diagnosis{
			DepthScore:  55,
			Summary:     "启发式诊断：回答偏短或个人贡献不够清晰（未接模型时的占位结论）。",
			WeakPoints:  []string{"建议补齐 Action/Result", "少用「我们」，说清你做了什么"},
			STARMissing: []string{"A", "R"},
			CoachScripts: []string{
				"用 STAR：情境一句话 → 你的任务 → 你具体动作（含取舍）→ 结果与复盘。",
			},
		}
		if shallow {
			d.DepthScore = 42
			d.Summary = "启发式诊断：回答偏表层，缺少取舍与结果。"
		}
		sess.Diagnosis = &d
		sess.Status = "done"
		sess.Messages = append(sess.Messages, ChatMsg{Role: "coach", Content: formatCoach(d)})
		return sess
	}
	q := "再展开一点：这件事里你个人拍板或推动的关键一步是什么？有没有备选方案被你否掉？"
	if shallow {
		q = "你的回答信息偏少。请用一个具体例子：当时有哪些选项、你选了哪个、为什么、结果如何？"
	}
	sess.Messages = append(sess.Messages, ChatMsg{Role: "interviewer", Content: q})
	sess.Round++
	return sess
}

func formatCoach(d Diagnosis) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("【面试模拟诊断】深度分 %d\n%s\n", d.DepthScore, d.Summary))
	if len(d.WeakPoints) > 0 {
		b.WriteString("薄弱点：\n- " + strings.Join(d.WeakPoints, "\n- ") + "\n")
	}
	if len(d.CoachScripts) > 0 {
		b.WriteString("补强口述：\n")
		for i, s := range d.CoachScripts {
			b.WriteString(fmt.Sprintf("%d) %s\n", i+1, s))
		}
	}
	return strings.TrimSpace(b.String())
}

func formatHistory(msgs []ChatMsg) string {
	var b strings.Builder
	for _, m := range msgs {
		b.WriteString(m.Role + ": " + m.Content + "\n")
	}
	return b.String()
}

func decodeJSON(raw string, dest any) error {
	raw = strings.TrimSpace(raw)
	raw = stripFence(raw)
	if err := json.Unmarshal([]byte(raw), dest); err == nil {
		return nil
	}
	// extract first JSON object from reasoning dumps
	re := regexp.MustCompile(`(?s)\{.*\}`)
	m := re.FindString(raw)
	if m == "" {
		return fmt.Errorf("no json object")
	}
	return json.Unmarshal([]byte(m), dest)
}

func stripFence(s string) string {
	if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```json")
		s = strings.TrimPrefix(s, "```JSON")
		s = strings.TrimPrefix(s, "```")
		if i := strings.LastIndex(s, "```"); i >= 0 {
			s = s[:i]
		}
	}
	return strings.TrimSpace(s)
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "..."
}
