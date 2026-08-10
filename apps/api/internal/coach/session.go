package coach

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/zhangyongjie/job-assistant/internal/db"
	"github.com/zhangyongjie/job-assistant/internal/llm"
)

const CrisisHelpText = `如果你正在经历强烈的自我伤害念头，或担心自己可能伤害他人，请立刻寻求专业或紧急帮助：
· 当地紧急求助电话（中国大陆可拨打 120 / 当地心理援助热线）
· 可联系身边可信的人陪同就医
本产品是职场心理教练，不能替代持证心理咨询或精神科诊疗，也不会提供危机陪护。`

var sceneLabel = map[string]string{
	"job_search":    "求职 / 跳槽",
	"promotion":     "晋升 / 述职",
	"communication": "职场沟通 / 冲突",
}

const systemPrompt = `你是「职场心理教练」，服务 3–10 年 ToB / 安全 / 合规方向职场人。
你的任务：在求职、晋升、沟通高压节点，帮用户拆开「事实 vs 自我评判」，稳住状态，给出 24 小时内可执行的下一步；需要过关训练时建议跳转人事/业务/谈薪工具。

硬性边界：
1. 不做临床诊断，不说「你有抑郁症/焦虑症」等病名，不开药。
2. 不替用户做重大人生决定（离职/接受 offer/撕破脸），只澄清标准与选项。
3. 共情后必须落到可验证的下一步，禁止纯鸡汤收尾。
4. 若用户流露自伤/伤人等危机信号：crisisFlag=true，reply 中明确转介专业帮助，停止常规深度陪聊。
5. 只输出 JSON。`

func SceneLabel(scene string) string {
	if v, ok := sceneLabel[scene]; ok {
		return v
	}
	return scene
}

func Start(client *llm.Client, scene, relatedEvent, taskHint string) (*db.CoachSession, error) {
	sess := &db.CoachSession{
		Scene:       scene,
		Title:       SceneLabel(scene) + " · 教练会话",
		RelatedEvent: relatedEvent,
		Messages:    []db.CoachMessage{},
		ActionItems: []string{},
		Scripts:     []string{},
		Status:      "active",
	}

	opening := heuristicOpening(scene, relatedEvent)
	if client != nil && client.Enabled() {
		user := fmt.Sprintf(`场景：%s（%s）
关联事件：%s
关联任务摘要：%s

请给出开场：先确认场景与当前节点，邀请用户用 2–3 句描述「发生了什么」和「此刻最卡住的感受」。
输出 JSON：
{
  "reply":"教练开场",
  "actionItems":[],
  "scripts":[],
  "crisisFlag":false,
  "suggestGate":""
}`, scene, SceneLabel(scene), relatedEvent, truncate(taskHint, 800))
		raw, err := client.ChatJSON(systemPrompt, user)
		if err == nil {
			var out struct {
				Reply string `json:"reply"`
			}
			if decode(raw, &out) == nil && strings.TrimSpace(out.Reply) != "" {
				opening = out.Reply
			}
		}
	}
	sess.Messages = append(sess.Messages, db.CoachMessage{Role: "coach", Content: opening})
	return sess, nil
}

func Reply(client *llm.Client, sess *db.CoachSession, userText, taskHint string) (*db.CoachSession, error) {
	if sess == nil {
		return nil, fmt.Errorf("会话不存在")
	}
	if sess.Status == "done" || sess.CrisisFlag {
		return sess, nil
	}
	userText = strings.TrimSpace(userText)
	if userText == "" {
		return nil, fmt.Errorf("请先输入内容")
	}
	sess.Messages = append(sess.Messages, db.CoachMessage{Role: "user", Content: userText})

	if detectCrisis(userText) {
		sess.CrisisFlag = true
		sess.Status = "done"
		sess.Messages = append(sess.Messages, db.CoachMessage{Role: "coach", Content: CrisisHelpText})
		return sess, nil
	}

	if client == nil || !client.Enabled() {
		return heuristicReply(sess, userText), nil
	}

	hist := formatHistory(sess.Messages)
	user := fmt.Sprintf(`场景：%s
关联事件：%s
关联任务摘要：%s

【对话】
%s

请按教练骨架推进（澄清事实/评判 → 命名情绪与需求 → 1–3 个选项 → 一个 24h 动作 + 可选「下一句怎么说」）。
若需要过关训练，suggestGate 填 hr|interview|salary，否则空字符串。
输出 JSON：
{
  "reply":"教练回复",
  "actionItems":["24h 内可做的动作"],
  "scripts":["可选话术"],
  "crisisFlag":false,
  "suggestGate":"",
  "done":false
}`, sess.Scene, sess.RelatedEvent, truncate(taskHint, 800), hist)

	raw, err := client.ChatJSON(systemPrompt, user)
	if err != nil {
		return nil, err
	}
	var out struct {
		Reply       string   `json:"reply"`
		ActionItems []string `json:"actionItems"`
		Scripts     []string `json:"scripts"`
		CrisisFlag  bool     `json:"crisisFlag"`
		Done        bool     `json:"done"`
	}
	if err := decode(raw, &out); err != nil || strings.TrimSpace(out.Reply) == "" {
		return heuristicReply(sess, userText), nil
	}
	if out.CrisisFlag {
		sess.CrisisFlag = true
		sess.Status = "done"
		sess.Messages = append(sess.Messages, db.CoachMessage{Role: "coach", Content: CrisisHelpText})
		return sess, nil
	}
	sess.Messages = append(sess.Messages, db.CoachMessage{Role: "coach", Content: out.Reply})
	if len(out.ActionItems) > 0 {
		sess.ActionItems = out.ActionItems
	}
	if len(out.Scripts) > 0 {
		sess.Scripts = out.Scripts
	}
	if out.Done {
		sess.Status = "done"
	}
	return sess, nil
}

func heuristicOpening(scene, event string) string {
	base := fmt.Sprintf("我们先从「%s」这个节点聊。", SceneLabel(scene))
	if strings.TrimSpace(event) != "" {
		base += fmt.Sprintf("你提到和「%s」有关。", event)
	}
	return base + "可以先用两三句话告诉我：刚才实际发生了什么？以及你此刻最卡住的一个感受是什么？（我会帮你把事实和自我评判拆开。）"
}

func heuristicReply(sess *db.CoachSession, userText string) *db.CoachSession {
	reply := "我听到了。我们先拆一下：哪一句是可核对的事实，哪一句是你对自己的评判？然后选一个 24 小时内最小动作——哪怕只是写清「今晚只改简历主线一段」也可以。"
	if len(sess.Messages) >= 5 {
		reply = "这一轮我们可以先收个口：请定一个今天就能做完的小动作（例如：列出挂面里可改进的 2 点；或给上级发一句澄清预期的草稿）。需要练表达时，再去过关训练里跑人事/业务/谈薪。"
		sess.ActionItems = []string{"写下今天唯一要完成的一件小事，并设一个结束时间"}
		sess.Status = "done"
	}
	sess.Messages = append(sess.Messages, db.CoachMessage{Role: "coach", Content: reply})
	return sess
}

func detectCrisis(text string) bool {
	keys := []string{"自杀", "不想活", "结束生命", "弄死自己", "伤害自己", "割腕", "跳楼", "杀死"}
	t := strings.ToLower(text)
	for _, k := range keys {
		if strings.Contains(t, k) {
			return true
		}
	}
	return false
}

func formatHistory(msgs []db.CoachMessage) string {
	var b strings.Builder
	for _, m := range msgs {
		role := "教练"
		if m.Role == "user" {
			role = "用户"
		}
		b.WriteString(role)
		b.WriteString("：")
		b.WriteString(m.Content)
		b.WriteString("\n")
	}
	return b.String()
}

func decode(raw string, dest any) error {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)
	return json.Unmarshal([]byte(raw), dest)
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}
