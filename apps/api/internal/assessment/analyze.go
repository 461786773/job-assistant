package assessment

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/zhangyongjie/job-assistant/internal/db"
	"github.com/zhangyongjie/job-assistant/internal/llm"
	"github.com/zhangyongjie/job-assistant/internal/prompts"
)

type Answers struct {
	Consent            bool     `json:"consent"`
	PrimaryScene       string   `json:"primaryScene"`
	WorkStatus         string   `json:"workStatus"`
	KeyEvents          []string `json:"keyEvents"`
	KeyEventsOther     string   `json:"keyEventsOther"`
	TenureBand         string   `json:"tenureBand"`
	B1                 int      `json:"b1"`
	B2                 int      `json:"b2"`
	B3                 int      `json:"b3"`
	B4                 int      `json:"b4"`
	B5                 int      `json:"b5"`
	B6                 int      `json:"b6"`
	MoodTags           []string `json:"moodTags"`
	MoodOther          string   `json:"moodOther"`
	Stressors          []string `json:"stressors"`
	StressorsOther     string   `json:"stressorsOther"`
	C2                 int      `json:"c2"`
	C3a                int      `json:"c3a"`
	C3b                int      `json:"c3b"`
	C4a                int      `json:"c4a"`
	C4b                int      `json:"c4b"`
	C5a                int      `json:"c5a"`
	C5b                int      `json:"c5b"`
	Coping             []string `json:"coping"`
	SupportLevel       string   `json:"supportLevel"`
	CheckinWillingness string   `json:"checkinWillingness"`
	CrisisLevel        string   `json:"crisisLevel"`
	Goals              []string `json:"goals"`
	FreeTextBlockers   string   `json:"freeTextBlockers"`
	FreeTextOther      string   `json:"freeTextOther"`
}

type Metrics struct {
	StressBaseline   int     `json:"stressBaseline"`
	EmotionLoad      float64 `json:"emotionLoad"`
	EnergyScore      int     `json:"energyScore"`
	SleepImpact      int     `json:"sleepImpact"`
	PainIntensity    int     `json:"painIntensity"`
	SuggestedScene   string  `json:"suggestedScene"`
	SuggestedNeed    string  `json:"suggestedNeed"`
	RecommendCounsel bool    `json:"recommendCounsel"`
}

type AIAnalysis struct {
	Headline       string   `json:"headline"`
	SuggestedScene string   `json:"suggestedScene"`
	NextSteps      []string `json:"nextSteps"`
	BoundaryNote   string   `json:"boundaryNote"`
	Crisis         bool     `json:"crisis"`
}

var sceneLabel = map[string]string{
	"job_search":    "求职 / 跳槽",
	"promotion":     "晋升 / 述职",
	"communication": "职场沟通 / 冲突",
	"mixed":         "多场景交织",
	"other":         "其他职场压力",
}

func Analyze(client *llm.Client, ans Answers) (Metrics, AIAnalysis, string, error) {
	m := computeMetrics(ans)
	analysis := heuristicAnalysis(ans, m)
	summary := FormatSummary(ans, m, analysis)

	if client != nil && client.Enabled() {
		payload, _ := json.Marshal(map[string]any{
			"answers": ans,
			"metrics": m,
		})
		user := fmt.Sprintf(`问卷与内部指标（勿对用户复述算法名）：
%s

输出 JSON：
{
  "headline":"一句描述性状态摘要（无病名）",
  "suggestedScene":"job_search|promotion|communication|mixed|other",
  "nextSteps":["建议下一步1","建议下一步2"],
  "boundaryNote":"非诊疗边界一句",
  "crisis":false
}`, string(payload))
		raw, err := client.ChatJSON(prompts.AssessmentAnalyze, user)
		if err == nil {
			var out AIAnalysis
			if decode(raw, &out) == nil && strings.TrimSpace(out.Headline) != "" {
				if out.SuggestedScene == "" {
					out.SuggestedScene = m.SuggestedScene
				}
				if out.Crisis {
					out.Crisis = ans.CrisisLevel == "elevated"
				}
				if ans.CrisisLevel == "elevated" {
					out.Crisis = true
				}
				analysis = out
				summary = FormatSummary(ans, m, analysis)
			}
		}
	}
	return m, analysis, summary, nil
}

func computeMetrics(ans Answers) Metrics {
	emo := mean(ans.B2, ans.B3, ans.B4)
	scene := ans.PrimaryScene
	if scene == "" {
		scene = "mixed"
	}
	need := mapSceneToNeed(scene)
	recommend := emo >= 3.5 || ans.B1 >= 4 || ans.CrisisLevel == "elevated" || ans.CrisisLevel == "fleeting"
	return Metrics{
		StressBaseline:   ans.B1,
		EmotionLoad:      emo,
		EnergyScore:      ans.B5,
		SleepImpact:      ans.B6,
		PainIntensity:    ans.C2,
		SuggestedScene:   scene,
		SuggestedNeed:    need,
		RecommendCounsel: recommend,
	}
}

func mapSceneToNeed(scene string) string {
	switch scene {
	case "job_search":
		return "job_search"
	case "promotion":
		return "promotion"
	case "communication":
		return "communication"
	case "mixed":
		return "unsure"
	default:
		return "counsel_first"
	}
}

func heuristicAnalysis(ans Answers, m Metrics) AIAnalysis {
	scene := sceneLabel[m.SuggestedScene]
	if scene == "" {
		scene = "职场高压节点"
	}
	headline := fmt.Sprintf("近两周压力基线约 %d/5，情绪负荷偏%s，当前更贴近「%s」场景。",
		m.StressBaseline, loadWord(m.EmotionLoad), scene)
	if m.EmotionLoad >= 3.5 && ans.B4 >= 4 {
		headline = fmt.Sprintf("近两周压力与自我否定感偏高，当前更适合先从「%s」的教练疏导开始，把评判和可改进点拆开。", scene)
	}
	steps := []string{"先开一次 AI 心理疏导，澄清事实与自我评判", "用三分钟自评或打卡看见状态变化"}
	if m.SuggestedNeed == "job_search" {
		steps = append(steps, "需要过关时再进入人事 / 业务 / 谈薪训练")
	} else if m.SuggestedNeed == "promotion" || m.SuggestedNeed == "communication" {
		steps = append(steps, "在对应场景疏导里定一个 24 小时可执行动作")
	}
	if m.RecommendCounsel {
		steps = append(steps, "若持续很难受，可预约私人心理辅导通道")
	}
	crisis := ans.CrisisLevel == "elevated"
	if crisis {
		headline = "你标出了需要被认真对待的痛苦信号。建议优先寻求专业或紧急支持；产品内职场教练不适合处理危机时刻。"
		steps = []string{"查看危机求助资源并联系身边可信的人", "确认已知晓资源后，再考虑非危机向的打卡与疏导"}
	}
	return AIAnalysis{
		Headline:       headline,
		SuggestedScene: m.SuggestedScene,
		NextSteps:      steps,
		BoundaryNote:   "本评估用于自我觉察与教练个性化，不是心理诊断，也不能替代持证咨询或精神科诊疗。",
		Crisis:         crisis,
	}
}

func FormatSummary(ans Answers, m Metrics, analysis AIAnalysis) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("主场景：%s\n", sceneLabelOr(ans.PrimaryScene)))
	b.WriteString(fmt.Sprintf("压力基线 %d/5；情绪负荷 %.1f/5；精力 %d/5；睡眠影响 %d/5\n",
		m.StressBaseline, m.EmotionLoad, m.EnergyScore, m.SleepImpact))
	if len(ans.Stressors) > 0 {
		b.WriteString("压力源：" + strings.Join(ans.Stressors, "、") + "\n")
	}
	if len(ans.Goals) > 0 {
		b.WriteString("期望：" + strings.Join(ans.Goals, "、") + "\n")
	}
	if strings.TrimSpace(ans.FreeTextBlockers) != "" {
		b.WriteString("最卡住：" + strings.TrimSpace(ans.FreeTextBlockers) + "\n")
	}
	b.WriteString("危机标记：" + ans.CrisisLevel + "\n")
	b.WriteString("AI摘要：" + analysis.Headline)
	return b.String()
}

func sceneLabelOr(s string) string {
	if v, ok := sceneLabel[s]; ok {
		return v
	}
	return s
}

func loadWord(v float64) string {
	if v >= 3.5 {
		return "高"
	}
	if v >= 2.5 {
		return "中"
	}
	return "低"
}

func mean(vals ...int) float64 {
	if len(vals) == 0 {
		return 0
	}
	sum := 0
	for _, v := range vals {
		sum += v
	}
	return float64(sum) / float64(len(vals))
}

func decode(raw string, dest any) error {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)
	return json.Unmarshal([]byte(raw), dest)
}

// BuildRecord fills a db.InitialAssessment from answers + analysis.
func BuildRecord(userID string, ans Answers, m Metrics, analysis AIAnalysis, summary string) (*db.InitialAssessment, error) {
	if !ans.Consent {
		return nil, fmt.Errorf("请先勾选知情同意")
	}
	if ans.PrimaryScene == "" {
		return nil, fmt.Errorf("请选择当前主场景")
	}
	if ans.CrisisLevel == "" {
		return nil, fmt.Errorf("请完成安全筛查")
	}
	answersRaw, err := json.Marshal(ans)
	if err != nil {
		return nil, err
	}
	scores := map[string]int{
		"b1": ans.B1, "b2": ans.B2, "b3": ans.B3, "b4": ans.B4, "b5": ans.B5, "b6": ans.B6, "c2": ans.C2,
		"c3a": ans.C3a, "c3b": ans.C3b, "c4a": ans.C4a, "c4b": ans.C4b, "c5a": ans.C5a, "c5b": ans.C5b,
	}
	scoresRaw, _ := json.Marshal(scores)
	metricsRaw, _ := json.Marshal(m)
	aiRaw, _ := json.Marshal(analysis)
	now := db.Now()
	return &db.InitialAssessment{
		UserID:             userID,
		Version:            "v1",
		Answers:            answersRaw,
		PrimaryScene:       ans.PrimaryScene,
		WorkStatus:         ans.WorkStatus,
		KeyEvents:          ans.KeyEvents,
		TenureBand:         ans.TenureBand,
		Scores:             scoresRaw,
		MoodTags:           ans.MoodTags,
		Stressors:          ans.Stressors,
		Coping:             ans.Coping,
		SupportLevel:       ans.SupportLevel,
		CheckinWillingness: ans.CheckinWillingness,
		CrisisLevel:        ans.CrisisLevel,
		Goals:              ans.Goals,
		FreeTextBlockers:   strings.TrimSpace(ans.FreeTextBlockers),
		FreeTextOther:      strings.TrimSpace(ans.FreeTextOther),
		Metrics:            metricsRaw,
		AIAnalysis:         aiRaw,
		SummaryForCoach:    summary,
		CompletedAt:        now,
		CreatedAt:          now,
	}, nil
}
