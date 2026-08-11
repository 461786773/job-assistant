// Package bigfive 实现职场大五画像（OCEAN 简版）计分、人设与标签。
package bigfive

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

const Version = "v1"

// Answers q1..q12，取值 1–5。
type Answers struct {
	Q1  int `json:"q1"`
	Q2  int `json:"q2"`
	Q3  int `json:"q3"`
	Q4  int `json:"q4"`
	Q5  int `json:"q5"`
	Q6  int `json:"q6"`
	Q7  int `json:"q7"`
	Q8  int `json:"q8"`
	Q9  int `json:"q9"`
	Q10 int `json:"q10"`
	Q11 int `json:"q11"`
	Q12 int `json:"q12"`
}

type DimScore struct {
	Mean    float64 `json:"mean"`
	Display int     `json:"display"`
	Band    string  `json:"band"` // low | mid | high
}

type Scores struct {
	Openness          DimScore `json:"openness"`
	Conscientiousness DimScore `json:"conscientiousness"`
	Extraversion      DimScore `json:"extraversion"`
	Agreeableness     DimScore `json:"agreeableness"`
	Neuroticism       DimScore `json:"neuroticism"`
	TopDims           []string `json:"topDims"`
	LowDims           []string `json:"lowDims"`
}

type ProfileResult struct {
	Version       string   `json:"version"`
	Scores        Scores   `json:"scores"`
	PersonaID     string   `json:"personaId"`
	PersonaTitle  string   `json:"personaTitle"`
	PersonaBlurb  string   `json:"personaBlurb"`
	PersonaBody   string   `json:"personaBody"`
	Tags          []string `json:"tags"`
	CoachHints    []string `json:"coachHints"`
	SummaryCoach  string   `json:"summaryForCoach"`
}

var dimOrder = []string{"O", "C", "E", "A", "N"}

var dimLabel = map[string]string{
	"O": "开放性", "C": "尽责性", "E": "外向性", "A": "宜人性", "N": "情绪波动性",
}

func (a Answers) rawSlice() []int {
	return []int{a.Q1, a.Q2, a.Q3, a.Q4, a.Q5, a.Q6, a.Q7, a.Q8, a.Q9, a.Q10, a.Q11, a.Q12}
}

func Validate(a Answers) error {
	for i, v := range a.rawSlice() {
		if v < 1 || v > 5 {
			return fmt.Errorf("第 %d 题须为 1–5", i+1)
		}
	}
	return nil
}

func scoreItem(raw int, reverse bool) float64 {
	if reverse {
		return float64(6 - raw)
	}
	return float64(raw)
}

func bandOf(mean float64) string {
	if mean < 2.5 {
		return "low"
	}
	if mean < 3.5 {
		return "mid"
	}
	return "high"
}

func displayOf(mean float64) int {
	return int(math.Round((mean - 1) / 4 * 100))
}

func dim(mean float64) DimScore {
	return DimScore{Mean: round1(mean), Display: displayOf(mean), Band: bandOf(mean)}
}

func round1(v float64) float64 {
	return math.Round(v*10) / 10
}

func Build(a Answers) (ProfileResult, error) {
	if err := Validate(a); err != nil {
		return ProfileResult{}, err
	}
	s1 := scoreItem(a.Q1, false)
	s2 := scoreItem(a.Q2, true)
	s3 := scoreItem(a.Q3, false)
	s4 := scoreItem(a.Q4, false)
	s5 := scoreItem(a.Q5, true)
	s6 := scoreItem(a.Q6, false)
	s7 := scoreItem(a.Q7, true)
	s8 := scoreItem(a.Q8, false)
	s9 := scoreItem(a.Q9, true)
	s10 := scoreItem(a.Q10, false)
	s11 := scoreItem(a.Q11, false)
	s12 := scoreItem(a.Q12, true)

	scores := Scores{
		Openness:          dim((s1 + s2) / 2),
		Conscientiousness: dim((s3 + s4 + s5) / 3),
		Extraversion:      dim((s6 + s7) / 2),
		Agreeableness:     dim((s8 + s9) / 2),
		Neuroticism:       dim((s10 + s11 + s12) / 3),
	}
	means := map[string]float64{
		"O": scores.Openness.Mean,
		"C": scores.Conscientiousness.Mean,
		"E": scores.Extraversion.Mean,
		"A": scores.Agreeableness.Mean,
		"N": scores.Neuroticism.Mean,
	}
	scores.TopDims = pickExtreme(means, true, 2)
	scores.LowDims = pickExtreme(means, false, 2)

	personaID, title, blurb, body := pickPersona(scores, means)
	tags := pickTags(scores)
	hints := pickHints(scores)
	summary := FormatForCoach(title, blurb, scores, tags, hints)

	return ProfileResult{
		Version:      Version,
		Scores:       scores,
		PersonaID:    personaID,
		PersonaTitle: title,
		PersonaBlurb: blurb,
		PersonaBody:  body,
		Tags:         tags,
		CoachHints:   hints,
		SummaryCoach: summary,
	}, nil
}

func pickExtreme(means map[string]float64, high bool, n int) []string {
	type pair struct {
		k string
		v float64
	}
	arr := make([]pair, 0, 5)
	for _, k := range dimOrder {
		arr = append(arr, pair{k, means[k]})
	}
	sort.SliceStable(arr, func(i, j int) bool {
		if high {
			if arr[i].v == arr[j].v {
				return indexOf(dimOrder, arr[i].k) < indexOf(dimOrder, arr[j].k)
			}
			return arr[i].v > arr[j].v
		}
		if arr[i].v == arr[j].v {
			return indexOf(dimOrder, arr[i].k) < indexOf(dimOrder, arr[j].k)
		}
		return arr[i].v < arr[j].v
	})
	out := make([]string, 0, n)
	for i := 0; i < n && i < len(arr); i++ {
		out = append(out, arr[i].k)
	}
	return out
}

func indexOf(xs []string, v string) int {
	for i, x := range xs {
		if x == v {
			return i
		}
	}
	return 99
}

func band(scores Scores, code string) string {
	switch code {
	case "O":
		return scores.Openness.Band
	case "C":
		return scores.Conscientiousness.Band
	case "E":
		return scores.Extraversion.Band
	case "A":
		return scores.Agreeableness.Band
	case "N":
		return scores.Neuroticism.Band
	}
	return "mid"
}

func isHigh(b string) bool { return b == "high" }
func isLow(b string) bool  { return b == "low" }
func isMidOrHigh(b string) bool {
	return b == "mid" || b == "high"
}

func pickPersona(scores Scores, means map[string]float64) (id, title, blurb, body string) {
	// 组合人设优先（更具体）；body 为给用户看的温暖描写，不做五维点名清单
	combos := []struct {
		id, title, blurb, body string
		ok                     func() bool
	}{
		{"bridge_builder", "会议室润滑剂", "你一进场，火药味都像被调小了一格。",
			"你习惯先把人接住，再把事往前推。会场里有你在，大家更容易听见彼此，而不是只听见立场。",
			func() bool { return isHigh(band(scores, "A")) && isHigh(band(scores, "E")) }},
		{"quiet_architect", "幕后架构师", "话不多，但版本计划和风险清单常常是你默默补齐的。",
			"你不太抢话筒，却常在关键处把缝补上。独处时想得更清楚，交付前那份踏实，多半来自你提前铺好的路。",
			func() bool { return isHigh(band(scores, "C")) && isLow(band(scores, "E")) }},
		{"storm_pilot", "高压领航员", "你会紧张，但紧张时常被你翻译成检查清单。",
			"压力会先抵达你的身体，但你很少停在原地发慌——更常把它拆成下一步、再下一步。紧的时候，反而更清醒。",
			func() bool { return isHigh(band(scores, "N")) && isHigh(band(scores, "C")) }},
		{"bold_challenger", "当场拆弹手", "场面冷了你敢开口，问题脏了你敢点名。",
			"你不怕把难听的话说清楚。场面僵住时，往往是你先把真正的问题拎到桌面上——锋利，但常能推进。",
			func() bool { return isLow(band(scores, "A")) && isHigh(band(scores, "E")) }},
		{"curious_migrator", "跨界翻译官", "你总能把旧经验讲成新岗位听得懂的语言。",
			"换赛道、换岗位时，你不太会死磕「我从前怎么做」，更愿意问「这边听得懂哪一句」。旧履历在你嘴里，常常能长出新接口。",
			func() bool { return isHigh(band(scores, "O")) && isMidOrHigh(band(scores, "C")) }},
		{"steady_anchor", "定海神针", "别人还在晃，你已经在问「下一步谁做、何时完」。",
			"热闹里你不太容易跟着晃。别人还在交换情绪时，你已经在找可落地的那一步——谁做、何时完、怎样算完。",
			func() bool { return isLow(band(scores, "N")) && isHigh(band(scores, "C")) }},
		{"empath_radar", "情绪雷达站", "你对氛围很敏感——既是超能力，也容易过载。",
			"屋里刚变沉默，你往往比别人先察觉。这让你很会照顾关系，也容易把别人的天气当成自己的天气。记得留一点回血的空隙。",
			func() bool { return isHigh(band(scores, "A")) && isHigh(band(scores, "N")) }},
		{"solo_craftsman", "独处工匠", "最好的方案，往往在你戴上耳机之后出现。",
			"灵感不太爱在嘈杂里露脸。给你一段安静，你更能把散落的线索织成能用的方案——不是不合群，是充电方式不同。",
			func() bool { return isLow(band(scores, "E")) && isHigh(band(scores, "O")) }},
		{"flexible_firefighter", "机动救火员", "计划改了三次你也能上场，现场感拉满。",
			"计划翻车时，你反而容易进状态。临场反应快，现场感强；要是再给自己留一点「先喘口气再冲」的余地，会更持久。",
			func() bool { return isLow(band(scores, "C")) && isHigh(band(scores, "E")) }},
		{"principled_editor", "严格审稿人", "你要的是对齐标准，不是和稀泥的「大家都好」。",
			"你对「差不多就行」过敏。对齐标准、把关质量，是你安心的方式；偶尔也练习一下——观点可以硬，关系不必撕破。",
			func() bool { return isLow(band(scores, "A")) && isHigh(band(scores, "C")) }},
	}
	for _, c := range combos {
		if c.ok() {
			return c.id, c.title, c.blurb, c.body
		}
	}

	top := "O"
	if len(scores.TopDims) > 0 {
		top = scores.TopDims[0]
	}
	_ = means
	singles := map[string][4]string{
		"O": {"open_explorer", "新地图测绘员", "未知区域对你来说更像副本入口，不像障碍。",
			"新事物不太吓到你，反而像多了一扇可推的门。你愿意先摸一摸，再决定要不要走——好奇心是你的燃料。"},
		"C": {"reliable_closer", "收尾必胜组", "「做完」比「做过」更让你有安全感。",
			"你在意的不是热闹开场，是能不能体面收尾。清单、节点、跟进——这些小事叠起来，就是别人说的「靠谱」。"},
		"E": {"social_spark", "场上点火器", "冷场三秒，你已经在找那个能接话的人。",
			"你容易把场子暖起来。对齐、推进、把冷空气变成对话——你习惯站在「先动起来」的那一边。"},
		"A": {"harmony_keeper", "共识收纳盒", "你习惯先把人接住，再把事放进去。",
			"冲突来时，你常先找彼此还站得住的共同点。关系被你看得很重；也记得，照顾别人之前，给自己留一句真心话。"},
		"N": {"deep_feeler", "内心气象台", "你比别人更早收到「要变天」的信号。",
			"你对压力和气氛更敏感，信号来得早。这不是脆弱，是天线更长——学会区分「事实」和「脑子里的电影」，会轻松许多。"},
	}
	s := singles[top]
	return s[0], s[1], s[2], s[3]
}

// BodyForPersona 按人设返回用户可见描写（用于回看页刷新旧版生硬文案）。
func BodyForPersona(personaID string) string {
	bodies := map[string]string{
		"bridge_builder":        "你习惯先把人接住，再把事往前推。会场里有你在，大家更容易听见彼此，而不是只听见立场。",
		"quiet_architect":       "你不太抢话筒，却常在关键处把缝补上。独处时想得更清楚，交付前那份踏实，多半来自你提前铺好的路。",
		"storm_pilot":           "压力会先抵达你的身体，但你很少停在原地发慌——更常把它拆成下一步、再下一步。紧的时候，反而更清醒。",
		"bold_challenger":       "你不怕把难听的话说清楚。场面僵住时，往往是你先把真正的问题拎到桌面上——锋利，但常能推进。",
		"curious_migrator":      "换赛道、换岗位时，你不太会死磕「我从前怎么做」，更愿意问「这边听得懂哪一句」。旧履历在你嘴里，常常能长出新接口。",
		"steady_anchor":         "热闹里你不太容易跟着晃。别人还在交换情绪时，你已经在找可落地的那一步——谁做、何时完、怎样算完。",
		"empath_radar":          "屋里刚变沉默，你往往比别人先察觉。这让你很会照顾关系，也容易把别人的天气当成自己的天气。记得留一点回血的空隙。",
		"solo_craftsman":        "灵感不太爱在嘈杂里露脸。给你一段安静，你更能把散落的线索织成能用的方案——不是不合群，是充电方式不同。",
		"flexible_firefighter":  "计划翻车时，你反而容易进状态。临场反应快，现场感强；要是再给自己留一点「先喘口气再冲」的余地，会更持久。",
		"principled_editor":     "你对「差不多就行」过敏。对齐标准、把关质量，是你安心的方式；偶尔也练习一下——观点可以硬，关系不必撕破。",
		"open_explorer":         "新事物不太吓到你，反而像多了一扇可推的门。你愿意先摸一摸，再决定要不要走——好奇心是你的燃料。",
		"reliable_closer":       "你在意的不是热闹开场，是能不能体面收尾。清单、节点、跟进——这些小事叠起来，就是别人说的「靠谱」。",
		"social_spark":          "你容易把场子暖起来。对齐、推进、把冷空气变成对话——你习惯站在「先动起来」的那一边。",
		"harmony_keeper":        "冲突来时，你常先找彼此还站得住的共同点。关系被你看得很重；也记得，照顾别人之前，给自己留一句真心话。",
		"deep_feeler":           "你对压力和气氛更敏感，信号来得早。这不是脆弱，是天线更长——学会区分「事实」和「脑子里的电影」，会轻松许多。",
	}
	return bodies[personaID]
}

func pickTags(scores Scores) []string {
	tags := make([]string, 0, 8)
	add := func(xs ...string) {
		for _, t := range xs {
			if len(tags) >= 6 {
				return
			}
			dup := false
			for _, e := range tags {
				if e == t {
					dup = true
					break
				}
			}
			if dup {
				continue
			}
			tags = append(tags, t)
		}
	}

	switch scores.Openness.Band {
	case "high":
		add("#灵感捕手", "#愿意改方案", "#迁移叙事有素材")
	case "low":
		add("#稳扎稳打", "#少折腾派")
	}
	switch scores.Conscientiousness.Band {
	case "high":
		add("#靠谱收尾", "#清单体质", "#简历好改型")
	case "low":
		add("#即兴选手", "#反流程侠")
	}
	switch scores.Extraversion.Band {
	case "high":
		add("#场上发光", "#对齐推进器")
	case "low":
		add("#深水静音", "#会后充电", "#一对一比群聊更在状态")
	}
	switch scores.Agreeableness.Band {
	case "high":
		add("#关系优先", "#温和说服")
	case "low":
		add("#直球表达", "#标准对齐", "#冲突场需要话术缓冲")
	}
	switch scores.Neuroticism.Band {
	case "high":
		add("#敏感天线", "#会后复盘脑", "#述职前要热身")
	case "low":
		add("#情绪稳盘", "#抗压缓冲")
	}

	if isHigh(scores.Conscientiousness.Band) && isHigh(scores.Neuroticism.Band) {
		add("#焦虑转化为行动")
	}
	if isHigh(scores.Conscientiousness.Band) && isLow(scores.Neuroticism.Band) {
		add("#稳如交付日")
	}
	if isHigh(scores.Extraversion.Band) && isHigh(scores.Agreeableness.Band) {
		add("#空气清新剂")
	}
	if isLow(scores.Extraversion.Band) && isHigh(scores.Agreeableness.Band) {
		add("#小范围深同盟")
	}
	if isHigh(scores.Openness.Band) && isLow(scores.Extraversion.Band) {
		add("#独自开地图")
	}
	if isLow(scores.Agreeableness.Band) && isHigh(scores.Conscientiousness.Band) {
		add("#质量守门员")
	}
	if isHigh(scores.Openness.Band) && isHigh(scores.Agreeableness.Band) {
		add("#温和创新者")
	}
	if scores.Openness.Band == "mid" && scores.Conscientiousness.Band == "mid" &&
		scores.Extraversion.Band == "mid" && scores.Agreeableness.Band == "mid" && scores.Neuroticism.Band == "mid" {
		add("#弹性适应体", "#尚未显影")
	}

	if len(tags) > 6 {
		tags = tags[:6]
	}
	return tags
}

func pickHints(scores Scores) []string {
	hints := make([]string, 0, 3)
	add := func(s string) {
		if len(hints) >= 3 {
			return
		}
		hints = append(hints, s)
	}
	if isHigh(scores.Neuroticism.Band) {
		add("开场先帮用户区分「事实」和「脑子里的电影」，再谈行动。")
	}
	if isHigh(scores.Conscientiousness.Band) {
		add("对方吃「可执行下一步」；少空泛鼓励，多一起列最小步骤。")
	}
	if isLow(scores.Extraversion.Band) {
		add("允许短回复与沉默；少用逼问式连续追问。")
	}
	if isHigh(scores.Agreeableness.Band) {
		add("注意 TA 可能过度自责或先顾别人；帮 TA 练习边界句。")
	}
	if isLow(scores.Agreeableness.Band) {
		add("认可直接风格，同时练习「观点硬、关系不撕破」的说法。")
	}
	if isHigh(scores.Openness.Band) {
		add("用迁移与类比挖深度，避免只停留在新点子清单。")
	}
	if len(hints) == 0 {
		add("先确认当下最卡的一件事，再一起拆成可验证的下一步。")
	}
	return hints
}

func bandCN(b string) string {
	switch b {
	case "high":
		return "偏高"
	case "low":
		return "偏低"
	default:
		return "中等"
	}
}

// FormatForCoach 生成注入教练会话的摘要。
func FormatForCoach(title, blurb string, scores Scores, tags, hints []string) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("职场画像：%s——%s\n", title, blurb))
	b.WriteString(fmt.Sprintf("五维：O%.1f(%s) C%.1f(%s) E%.1f(%s) A%.1f(%s) N%.1f(%s)\n",
		scores.Openness.Mean, bandCN(scores.Openness.Band),
		scores.Conscientiousness.Mean, bandCN(scores.Conscientiousness.Band),
		scores.Extraversion.Mean, bandCN(scores.Extraversion.Band),
		scores.Agreeableness.Mean, bandCN(scores.Agreeableness.Band),
		scores.Neuroticism.Mean, bandCN(scores.Neuroticism.Band),
	))
	if len(tags) > 0 {
		b.WriteString("标签：" + strings.Join(tags, " ") + "\n")
	}
	if len(hints) > 0 {
		b.WriteString("配合提示：\n")
		for _, h := range hints {
			b.WriteString("- " + h + "\n")
		}
	}
	b.WriteString("边界：风格参考，禁止诊断病名，禁止当作录用判断。")
	return b.String()
}
