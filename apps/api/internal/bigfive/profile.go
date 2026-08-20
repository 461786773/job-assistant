// Package bigfive 实现职场大五画像（OCEAN 简版）计分、人设与标签。
package bigfive

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
)

const Version = "v1.4"

// Answers q1..q15，取值 1–5；每维 3 题（O1–3, C4–6, E7–9, A10–12, N13–15）。
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
	Q13 int `json:"q13"`
	Q14 int `json:"q14"`
	Q15 int `json:"q15"`
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
	return []int{
		a.Q1, a.Q2, a.Q3, a.Q4, a.Q5, a.Q6, a.Q7, a.Q8, a.Q9,
		a.Q10, a.Q11, a.Q12, a.Q13, a.Q14, a.Q15,
	}
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
	s2 := scoreItem(a.Q2, false)
	s3 := scoreItem(a.Q3, true)
	s4 := scoreItem(a.Q4, false)
	s5 := scoreItem(a.Q5, false)
	s6 := scoreItem(a.Q6, true)
	s7 := scoreItem(a.Q7, false)
	s8 := scoreItem(a.Q8, false)
	s9 := scoreItem(a.Q9, true)
	s10 := scoreItem(a.Q10, false)
	s11 := scoreItem(a.Q11, false)
	s12 := scoreItem(a.Q12, true)
	s13 := scoreItem(a.Q13, false)
	s14 := scoreItem(a.Q14, false)
	s15 := scoreItem(a.Q15, true)

	scores := Scores{
		Openness:          dim((s1 + s2 + s3) / 3),
		Conscientiousness: dim((s4 + s5 + s6) / 3),
		Extraversion:      dim((s7 + s8 + s9) / 3),
		Agreeableness:     dim((s10 + s11 + s12) / 3),
		Neuroticism:       dim((s13 + s14 + s15) / 3),
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
	summary := FormatForCoach(title, blurb, body, scores, tags, hints)

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

// softHigh / softLow：相对极值也算「突出」，避免多数题选 3 时人人落到平庸单维人设。
func softHigh(means map[string]float64, code string) bool {
	mx := maxMean(means)
	v := means[code]
	return v >= 3.2 && v >= mx-0.45
}

func softLow(means map[string]float64, code string) bool {
	mn := minMean(means)
	v := means[code]
	return v <= 2.9 && v <= mn+0.45
}

func maxMean(means map[string]float64) float64 {
	mx := -1.0
	for _, k := range dimOrder {
		if means[k] > mx {
			mx = means[k]
		}
	}
	return mx
}

func minMean(means map[string]float64) float64 {
	mn := 99.0
	for _, k := range dimOrder {
		if means[k] < mn {
			mn = means[k]
		}
	}
	return mn
}

func pickPersona(scores Scores, means map[string]float64) (id, title, blurb, body string) {
	hi := func(code string) bool {
		return isHigh(band(scores, code)) || softHigh(means, code)
	}
	lo := func(code string) bool {
		return isLow(band(scores, code)) || softLow(means, code)
	}
	// 组合人设优先：职场冲突感更强，避免「开朗/靠谱」式空标签
	combos := []struct {
		id string
		ok func() bool
	}{
		{"bridge_builder", func() bool { return hi("A") && hi("E") }},
		{"quiet_architect", func() bool { return hi("C") && lo("E") }},
		{"storm_pilot", func() bool { return hi("N") && hi("C") }},
		{"bold_challenger", func() bool { return lo("A") && hi("E") }},
		{"curious_migrator", func() bool { return hi("O") && (hi("C") || isMidOrHigh(band(scores, "C"))) }},
		{"steady_anchor", func() bool { return lo("N") && hi("C") }},
		{"empath_radar", func() bool { return hi("A") && hi("N") }},
		{"solo_craftsman", func() bool { return lo("E") && hi("O") }},
		{"flexible_firefighter", func() bool { return lo("C") && hi("E") }},
		{"principled_editor", func() bool { return lo("A") && hi("C") }},
		{"guarded_analyst", func() bool { return lo("E") && hi("N") }},
		{"warm_pragmatist", func() bool { return hi("A") && hi("C") }},
		{"restless_scout", func() bool { return hi("O") && hi("E") }},
		{"iron_gatekeeper", func() bool { return hi("C") && lo("O") }},
	}
	for _, c := range combos {
		if c.ok() {
			return lookupPersona(c.id)
		}
	}

	top1, top2 := "C", "O"
	if len(scores.TopDims) >= 1 {
		top1 = scores.TopDims[0]
	}
	if len(scores.TopDims) >= 2 {
		top2 = scores.TopDims[1]
	}
	if id, ok := dualPersonaID(top1, top2); ok {
		return lookupPersona(id)
	}

	singles := map[string]string{
		"O": "open_explorer", "C": "reliable_closer", "E": "social_spark",
		"A": "harmony_keeper", "N": "deep_feeler",
	}
	return lookupPersona(singles[top1])
}

func dualPersonaID(a, b string) (string, bool) {
	key := a + b
	if a > b {
		key = b + a
	}
	table := map[string]string{
		"AC": "warm_pragmatist",
		"AN": "empath_radar",
		"AE": "bridge_builder",
		"AO": "gentle_inventor",
		"CN": "storm_pilot",
		"CE": "stage_pm",
		"CO": "curious_migrator",
		"EN": "visible_voltage",
		"EO": "restless_scout",
		"NO": "night_mapper",
	}
	id, ok := table[key]
	return id, ok
}

type personaCopy struct {
	title, blurb, edge, shadow string
}

func lookupPersona(id string) (pid, title, blurb, body string) {
	p, ok := personaCatalog[id]
	if !ok {
		p = personaCatalog["reliable_closer"]
		id = "reliable_closer"
	}
	return id, p.title, p.blurb, composePersonaBody(p)
}

func composePersonaBody(p personaCopy) string {
	edge := strings.TrimSpace(p.edge)
	shadow := strings.TrimSpace(p.shadow)
	if edge == "" {
		return shadow
	}
	if shadow == "" {
		return edge
	}
	// 一段合写：锋利面 + 代价，禁止「双刃：/硬伤：」分栏标题（问卷 §5.3）
	return edge + " " + shadow
}

// 人设文案库：禁止纯表扬；标题抓眼；短句带刺；正文 = 一段合写（锋利面与代价）。
var personaCatalog = map[string]personaCopy{
	"bridge_builder": {
		"人形灭火器",
		"你会场一坐，火药味降两度——账单却记在你情绪账户上。",
		"你把对立翻译成「还能谈」的议题，场面要破局时别人看你。短板是：你常把别人的火先吞进自己肚子，散会了还在复盘「我是不是说重了」。",
		"润滑剂的耗材是你自己。冲突散了你未必散得了；讨好式灭火练久了，真心话会越来越晚出口。",
	},
	"quiet_architect": {
		"隐形总控",
		"话筒不在你手里，坑却常是你默默填的——然后功劳漂走。",
		"你不抢舞台，关键缝里把路铺好、风险表写上。代价是贡献太干净：容易被当成背景板，邀功不会时，成果归嗓门更大的人。",
		"「默默靠谱」换不来曝光。不学会把贡献说出口，你的价值只会在别人简历里显影。",
	},
	"storm_pilot": {
		"焦虑发动机",
		"你不是不怕，是把紧张硬焊成清单——再拿加班证明「我还能行」。",
		"高压下你拆步骤、盯节点、先做最脏的活，团队靠这股绷着的推进力过关。问题是你分不清「必要紧张」和「自我鞭打」。",
		"松下来身体才讨债。预警被你当燃料烧，烧穿前你还以为自己很强。",
	},
	"bold_challenger": {
		"当场拆弹手",
		"场面冷了你敢开口，问题脏了你敢点名——也敢把人得罪干净。",
		"你能把真问题拎上桌，锋利、短、有效；僵局常因你一句难听话松动。同时，直球会在人心里存档成「刺头」。",
		"观点赢了、关系裂了，是最贵的账单。缺缓冲话术时，你付双倍：事情成了，同盟没了。",
	},
	"curious_migrator": {
		"履历变形金刚",
		"旧经验到你嘴里总能接新接口——也容易被怀疑只会讲故事。",
		"换赛道你先问「这边听哪一句」，迁移叙事是武器。日常则容易被贴「想太多 / 不安分」；点子多、锚点少时，可信度掉得很快。",
		"故事讲得漂亮，落地跟不上，人设就从「能迁移」滑成「会包装」。收口比开脑洞更考验你。",
	},
	"steady_anchor": {
		"情绪绝缘体",
		"别人还在晃，你已经在问谁做、何时完——然后被默认你永远不累。",
		"热闹里你不跟着抖，能把混乱压成下一步。短板是稳会被误读成「不懂人 / 容量无限」；你少说需求，别人就加码。",
		"绝缘太久，连自己的信号也收不到。不会喊疼时，定盘星会先裂开。",
	},
	"empath_radar": {
		"人形气压计",
		"屋里刚变沉默你就察觉——接着过载，睡眠来结账。",
		"你读空气准，会接尴尬、顾关系。超能力的电费从你精力里扣：别人的天气变成你的天气，边界一松就全场内耗代偿。",
		"敏感≠深度。不练边界，你只是团队的免费情绪污水处理厂。",
	},
	"solo_craftsman": {
		"耳机里的军师",
		"好方案爱等你戴上耳机——群聊里你却像失踪人口。",
		"嘈杂里你掉帧，安静里能出深水活。短板刺眼：沉默被读成不配合，不曝光进度时，存在感会被活活饿死。",
		"深水输出救不了「没人知道你在干嘛」。不主动同步，你再强也像黑箱。",
	},
	"flexible_firefighter": {
		"救火成瘾者",
		"计划改三次你也能冲——可怕的是你开始离不开火。",
		"翻车现场你状态拉满：临场快、能把崩盘拽回来。靠救火证明价值久了，会慢慢不会「无着做」；没火发慌，有火透支。",
		"英雄时刻很爽，可持续性很差。火一灭，你的价值叙事也会空一截。",
	},
	"principled_editor": {
		"质量原教旨",
		"你要标准对齐，不要和稀泥——也因此常被写成「难搞」。",
		"你对「差不多」过敏，砍水分、把关口，是交付站得住的原因。同时标准一硬、面子一薄，对事严厉会滑成对人刺痛。",
		"正确但孤立，是最惨的胜利。不会把标准翻译成人听得懂的「为什么」，你就只剩杠点。",
	},
	"guarded_analyst": {
		"会前三剧本",
		"你开口前已演完翻车版——慢热常被骂成犹豫。",
		"你不爱即兴，推演完再说话，踩坑少。被催时若硬装流畅反而露馅；你缺的是「我需要思考窗」的声明，不是能力。",
		"过度预演会吃掉行动窗口。剧本写到第三版还不出手，分析就成了拖延的高定外套。",
	},
	"warm_pragmatist": {
		"双肩挑死人",
		"人要留住、事要做成——两头都想赢，两头都在抽你的血。",
		"你很少只讲情怀或只讲 KPI，安慰人时已在想下一步。代价是不敢排优先级时，温柔变成自我剥削，最后被两头压扁。",
		"「我都可以」是最高级的陷阱。不会说不，你就不是桥梁，是桥面裂缝。",
	},
	"restless_scout": {
		"情报饥渴症",
		"新东西你都想先摸——侦察没有收队，就叫分心。",
		"你开地图快、前线感强，探路值钱。热情一散焦点，别人只看见到处点火、少见收官；新鲜感撤退后剩半成品。",
		"好奇心若不配截止日期，就是组织里的干扰源。先定「探到什么算收工」。",
	},
	"iron_gatekeeper": {
		"红线人形立牌",
		"你先问风险再问机会——也常被嫌挡路。",
		"合规、不可回滚的坑，你竖警戒线早，少踩雷是沉默功劳。只说「不行」不说「保大家走更远」，你会变成阻力符号，创新场里先被开火。",
		"守门若不带替代路径，就只是挡路。红线要配「怎么绕、怎么降级」，否则没人听。",
	},
	"gentle_inventor": {
		"温水创新派",
		"你推新前先摸别人怕不怕——很少砸场，也很少炸穿。",
		"创新像推窗：阻力小，因为你把「怕」接住了。太顾感受时点子磨成无棱；缺成功标准，温柔创新 = 温吞无结果。",
		"怕得罪人就会得罪进度。创新需要可验的一刀，不是永远的「我们再看看」。",
	},
	"stage_pm": {
		"台上项目官",
		"场上带着走、会后钉节点——表演和交付你都想拿，电费也双倍。",
		"对外撑场、对内收口，你像会走路的看板。「表演推进」耗电高；分不清场上人设和真实容量时，会后崩盘只是时间问题。",
		"掌声不是续航。下场不回血，项目官会先变成项目事故。",
	},
	"visible_voltage": {
		"明线高电压",
		"场上有电，私下跳闸——硬撑「我很好」时烧毁是真的。",
		"推进力外显、能点着冷场，别人靠你带节奏。会后回放脑更吵；电压装出来的时候，账单从睡眠和脾气里扣。",
		"亮着不等于健康。不会关灯休息，你就只是一根即将熔断的保险丝。",
	},
	"night_mapper": {
		"凌晨制图员",
		"压力一大你就重画地图——常常画到睡不着，还以为自己在努力。",
		"焦虑能推你找新解法，危机感变重构动力。若不拆「事实 / 担心」，地图全是鬼影；用思考麻痹行动，越想越空转。",
		"脑子加班≠工作推进。凌晨制图若不落地成白天三步，就只是昂贵的失眠。",
	},
	"open_explorer": {
		"新坑试吃员",
		"未知像副本——也像永远吃不完、从不买单的试吃盘。",
		"你敢先摸再决定，破冰试错时值钱。探索无收口就是点子很多、落地很少；新鲜感一撤，半成品和白眼留下。",
		"好奇心要配「试完怎么收」。否则你不是探索者，是半成品制造商。",
	},
	"reliable_closer": {
		"收尾强迫症",
		"「做过」安慰不了你，只有「做完」能——然后没人问你累不累。",
		"清单、节点、跟进叠出靠谱，别人睡得着常因默认有你收尾。你把一切扛完，换来「怎么不早说累」；不会喊停，靠谱会累死你。",
		"收尾型人格最容易变成组织的免费保险。保险没有保费声明，理赔的是你的健康。",
	},
	"social_spark": {
		"冷场灭霸",
		"冷场三秒你就去找接话人——会后才发现自己账户透支了。",
		"你能把空气变成对话，场上推进快。热闹是透支账户；不坦白回血需求，你会用下一场热闹掩盖这一场的空。",
		"社交高光很上瘾。戒不掉「我来救场」，你就练不会真正的休息。",
	},
	"harmony_keeper": {
		"共识讨好债",
		"你先接人再放事——共识成了，委屈进账了。",
		"冲突里找共同点，关系场当缓冲垫，少撕破脸常有你的份。代价是真心话排不到队，变成好好先生/女士：表面和谐，内里一笔讨好债。",
		"和谐若靠你单方面吞刺，那不叫团队健康，叫你个人负债。",
	},
	"deep_feeler": {
		"全天候警报器",
		"你更早收到「要变天」——也更难关掉误报。",
		"天线长，压力信号来得早，准的时候能帮团队躲坑。事实与「脑子里的电影」一糊，敏感就从雷达变成全天警报，内耗刷屏。",
		"感受多不是深度本身。不练「标注哪些是事实」，你只是一台关不掉的报警器。",
	},
}

// BodyForPersona 按人设返回用户可见描写（用于回看页刷新旧版生硬文案）。
func BodyForPersona(personaID string) string {
	p, ok := personaCatalog[personaID]
	if !ok {
		return ""
	}
	return composePersonaBody(p)
}

// TitleBlurbForPersona 回看页刷新标题/短句（旧档也可升级文案）。
func TitleBlurbForPersona(personaID string) (title, blurb string) {
	p, ok := personaCatalog[personaID]
	if !ok {
		return "", ""
	}
	return p.title, p.blurb
}

// ShadowForPersona 硬伤短句（结果页 / 教练上下文可单独引用）。
func ShadowForPersona(personaID string) string {
	p, ok := personaCatalog[personaID]
	if !ok {
		return ""
	}
	return p.shadow
}

// TagsForScores 按五维档位刷新标签（回看旧档可升级代价向文案）。
func TagsForScores(scores Scores) []string {
	return pickTags(scores)
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
		add("#新坑试吃", "#收口困难户")
	case "low":
		add("#路径依赖户", "#创新过敏体质")
	}
	switch scores.Conscientiousness.Band {
	case "high":
		add("#收尾强迫", "#不喊停会死")
	case "low":
		add("#救火成瘾", "#无火就慌")
	}
	switch scores.Extraversion.Band {
	case "high":
		add("#场上带电", "#会后空账")
	case "low":
		add("#深水黑箱", "#存在感易饿死")
	}
	switch scores.Agreeableness.Band {
	case "high":
		add("#讨好债户", "#边界漏风")
	case "low":
		add("#直球伤人", "#正确但孤立")
	}
	switch scores.Neuroticism.Band {
	case "high":
		add("#会后复盘脑", "#误报刷屏")
	case "low":
		add("#情绪绝缘", "#被当永动机")
	}

	if isHigh(scores.Conscientiousness.Band) && isHigh(scores.Neuroticism.Band) {
		add("#焦虑当燃料")
	}
	if isHigh(scores.Conscientiousness.Band) && isLow(scores.Neuroticism.Band) {
		add("#稳到被透支")
	}
	if isHigh(scores.Extraversion.Band) && isHigh(scores.Agreeableness.Band) {
		add("#灭火耗材本人")
	}
	if isLow(scores.Extraversion.Band) && isHigh(scores.Agreeableness.Band) {
		add("#小圈讨好仓")
	}
	if isHigh(scores.Openness.Band) && isLow(scores.Extraversion.Band) {
		add("#耳机方案、群聊失踪")
	}
	if isLow(scores.Agreeableness.Band) && isHigh(scores.Conscientiousness.Band) {
		add("#质量原教旨")
	}
	if isHigh(scores.Openness.Band) && isHigh(scores.Agreeableness.Band) {
		add("#温吞创新风险")
	}
	if scores.Openness.Band == "mid" && scores.Conscientiousness.Band == "mid" &&
		scores.Extraversion.Band == "mid" && scores.Agreeableness.Band == "mid" && scores.Neuroticism.Band == "mid" {
		add("#尚未显影", "#别用中庸骗自己")
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
func FormatForCoach(title, blurb, body string, scores Scores, tags, hints []string) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("人设：%s——%s\n", title, blurb))
	if strings.TrimSpace(body) != "" {
		b.WriteString(strings.TrimSpace(body) + "\n")
	}
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
	b.WriteString("边界：风格参考（锋利面与代价揉在一段描写里），禁止诊断病名，禁止当作录用判断。")
	return b.String()
}

// SummaryForCoachLive 按当前文案库重算教练摘要（旧档注入时不依赖过期 SummaryForCoach）。
func SummaryForCoachLive(personaID string, scoresJSON json.RawMessage) string {
	title, blurb := TitleBlurbForPersona(personaID)
	body := BodyForPersona(personaID)
	if title == "" && body == "" {
		return ""
	}
	var scores Scores
	if len(scoresJSON) > 0 {
		_ = json.Unmarshal(scoresJSON, &scores)
	}
	tags := TagsForScores(scores)
	hints := pickHints(scores)
	return FormatForCoach(title, blurb, body, scores, tags, hints)
}
