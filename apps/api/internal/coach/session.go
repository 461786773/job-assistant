package coach

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/zhangyongjie/job-assistant/internal/db"
	"github.com/zhangyongjie/job-assistant/internal/llm"
	"github.com/zhangyongjie/job-assistant/internal/prompts"
)

// CrisisHelpText 兼容旧引用；正文见 prompts.CrisisHelp。
const CrisisHelpText = prompts.CrisisHelp

var sceneLabel = map[string]string{
	"job_search":    "求职 / 跳槽",
	"promotion":     "晋升 / 述职",
	"communication": "职场沟通 / 冲突",
}

func SceneLabel(scene string) string {
	if v, ok := sceneLabel[scene]; ok {
		return v
	}
	return scene
}

// MergeUserProfile 把职场画像、心理评估（基线）、快照合成一段「合读」上下文。
func MergeUserProfile(bigFiveHint, assessmentHint, quickHint string) string {
	var parts []string
	if s := strings.TrimSpace(bigFiveHint); s != "" {
		parts = append(parts, "【职场风格画像·较稳定】\n"+s)
	}
	if s := strings.TrimSpace(assessmentHint); s != "" {
		parts = append(parts, "【心理评估·基线·最新】\n"+s)
	}
	if s := strings.TrimSpace(quickHint); s != "" {
		parts = append(parts, "【快照·此刻状态】\n"+s)
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, "\n\n") + "\n\n合读说明：画像看协作与压力习惯，心理评估（基线）看近两周压力与卡住点，快照看此刻状态。三者描述同一个人，须合在一起用；禁止诊断病名，禁止只引用其中一块。"
}

func Start(client *llm.Client, scene, relatedEvent, taskHint, quickHint, assessmentHint, bigFiveHint, primaryNeed string) (*db.CoachSession, error) {
	sess := &db.CoachSession{
		Scene:        scene,
		Title:        SceneLabel(scene) + " · 教练会话",
		RelatedEvent: relatedEvent,
		Messages:     []db.CoachMessage{},
		ActionItems:  []string{},
		Scripts:      []string{},
		Status:       "active",
	}

	profileCtx := MergeUserProfile(bigFiveHint, assessmentHint, quickHint)
	opening := heuristicOpening(scene, relatedEvent, profileCtx)
	if client != nil && client.Enabled() {
		user := fmt.Sprintf(`场景：%s（%s）
用户诉求：%s
关联事件：%s
关联任务摘要：%s

【用户综合档案】（画像 + 评估 + 自评，必须合读）
%s

请给出开场：合读档案后自然接住用户（禁止诊断病名；不要分别背诵两套材料），邀请用户用 2–3 句补充「发生了什么」。
若困扰分≥8 或评测危机偏高，语气放缓并轻提示专业支持 / 预约通道。
输出 JSON：
{
  "reply":"教练开场",
  "actionItems":[],
  "scripts":[],
  "crisisFlag":false,
  "suggestGate":""
}`, scene, SceneLabel(scene), primaryNeed, relatedEvent, truncate(taskHint, 800), truncate(profileCtx, 2200))
		raw, err := client.ChatJSON(prompts.CoachSystem, user)
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

func Reply(client *llm.Client, sess *db.CoachSession, userText, taskHint, profileCtx string) (*db.CoachSession, error) {
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

【用户综合档案】（画像 + 评估 + 自评，必须合读，每轮都要参考）
%s

【对话】
%s

请按此顺序写 reply：
1) 若用户本轮有情绪/疲惫/压力：第一句必须共情（具体点出心理层面的感受，勿空泛「辛苦了」）；
2) 结合综合档案调整接话方式与侧重点（画像定节奏，评估定压力与目标），勿把两套资料割裂；
3) 再澄清事实 vs 自我评判，或命名情绪与需求；
4) 再给 1–3 个选项，并收束到一个 24h 动作（可选附「下一句怎么说」）。
若需要过关训练，suggestGate 填 hr|interview|salary，否则空字符串。
输出 JSON：
{
  "reply":"教练回复（有情绪时必须以共情句开头）",
  "actionItems":["24h 内可做的动作"],
  "scripts":["可选话术"],
  "crisisFlag":false,
  "suggestGate":"",
  "done":false
}`, sess.Scene, sess.RelatedEvent, truncate(taskHint, 800), truncate(profileCtx, 2200), hist)

	raw, err := client.ChatJSON(prompts.CoachSystem, user)
	if err != nil {
		return nil, err
	}
	var out struct {
		Reply       string   `json:"reply"`
		ActionItems []string `json:"actionItems"`
		Scripts     []string `json:"scripts"`
		SuggestGate string   `json:"suggestGate"`
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
	if g := normalizeSuggestGate(out.SuggestGate); g != "" {
		sess.SuggestGate = g
	}
	if out.Done {
		sess.Status = "done"
	}
	return sess, nil
}

func normalizeSuggestGate(g string) string {
	switch strings.ToLower(strings.TrimSpace(g)) {
	case "hr", "interview", "salary":
		return strings.ToLower(strings.TrimSpace(g))
	default:
		return ""
	}
}

func heuristicOpening(scene, event, profileCtx string) string {
	base := fmt.Sprintf("我们先从「%s」这个节点聊。", SceneLabel(scene))
	if strings.TrimSpace(event) != "" {
		base += fmt.Sprintf("你提到和「%s」有关。", event)
	}
	if strings.TrimSpace(profileCtx) != "" {
		base += "\n\n我这边已经把你的职场画像和评估放在一起看了（风格参考，不是诊断）。"
		if strings.Contains(profileCtx, "困扰分 8") || strings.Contains(profileCtx, "困扰分 9") || strings.Contains(profileCtx, "困扰分 10") {
			base += "你现在的困扰分偏高。我们可以先把步子放慢；若持续很难受，也可预约私人心理辅导通道——我这边是职场教练，不能替代诊疗。"
		}
		base += "\n\n想先补充一两句：最近具体发生了什么？还是直接从「今天想带走的那一点」聊起？"
		return base
	}
	return base + "可以先用两三句话告诉我：刚才实际发生了什么？以及你此刻最卡住的一个感受是什么？（我会帮你把事实和自我评判拆开。）"
}

var feelingLabel = map[string]string{
	"tired": "累", "irritable": "烦", "numb": "空",
	"afraid": "怕", "stuck": "堵", "indifferent": "无所谓",
}
var durationLabel = map[string]string{
	"few_days": "就这两三天", "one_two_weeks": "一两周了",
	"over_month": "一个月以上", "unclear_chronic": "说不清，好像一直这样",
}
var impactLabel = map[string]string{
	"sleep": "睡眠", "appetite": "胃口", "focus": "集中力",
	"temper": "脾气", "body": "身体", "mood_only": "主要是心里不爽",
}
var takeawayLabel = map[string]string{
	"clarity": "想通一件事", "strength": "找回一点力量", "tiny_tool": "有一个能用的小办法",
	"just_talk": "只是想找个人说说话", "unsure_but_here": "我也不知道想带走什么，但我来了",
}

// FormatQuickCheck turns a saved self-check into coach context (no diagnosis).
func FormatQuickCheck(c *db.QuickSelfCheck) string {
	if c == nil {
		return ""
	}
	feelings := make([]string, 0, len(c.Feelings))
	for _, f := range c.Feelings {
		if v, ok := feelingLabel[f]; ok {
			feelings = append(feelings, v)
		} else {
			feelings = append(feelings, f)
		}
	}
	impacts := make([]string, 0, len(c.Impacts))
	for _, i := range c.Impacts {
		if v, ok := impactLabel[i]; ok {
			impacts = append(impacts, v)
		} else {
			impacts = append(impacts, i)
		}
	}
	lines := []string{
		fmt.Sprintf("- 感觉：%s", strings.Join(feelings, "、")),
		fmt.Sprintf("- 持续：%s", orDash(durationLabel[c.Duration], c.Duration)),
		fmt.Sprintf("- 影响：%s", strings.Join(impacts, "、")),
		fmt.Sprintf("- 困扰分 %d/10", c.DistressScore),
		fmt.Sprintf("- 今天想带走：%s", orDash(takeawayLabel[c.Takeaway], c.Takeaway)),
	}
	if strings.TrimSpace(c.TriggerNote) != "" {
		lines = append(lines, "- 触发事件："+strings.TrimSpace(c.TriggerNote))
	}
	return strings.Join(lines, "\n")
}

func orDash(a, b string) string {
	if strings.TrimSpace(a) != "" {
		return a
	}
	if strings.TrimSpace(b) != "" {
		return b
	}
	return "—"
}

func heuristicReply(sess *db.CoachSession, userText string) *db.CoachSession {
	reply := "我听到了。我们先拆一下：哪一句是可核对的事实，哪一句是你对自己的评判？然后选一个 24 小时内最小动作——哪怕只是写清「今晚只改简历主线一段」也可以。"
	if looksEmotionHeavy(userText) {
		reply = "我了解这种疲累——来自心里的，往往比身体上的疲惫更重。你可以先不用急着「振作」。我们慢慢来：这份累，更像撑太久了，还是最近有一件事特别耗你？"
	}
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

func looksEmotionHeavy(text string) bool {
	keys := []string{"太累", "好累", "累了", "疲惫", "撑不住", "喘不过气", "心累", "崩溃", "难受", "焦虑", "委屈", "想哭", "空耗", "燃尽"}
	for _, k := range keys {
		if strings.Contains(text, k) {
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
