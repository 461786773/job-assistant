package hr

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/zhangyongjie/job-assistant/internal/llm"
)

type DimensionScore struct {
	Key     string `json:"key"`
	Label   string `json:"label"`
	Score   int    `json:"score"`
	Comment string `json:"comment"`
}

type Issue struct {
	Severity string `json:"severity"` // critical | warn | info
	Title    string `json:"title"`
	Detail   string `json:"detail"`
}

type RewriteItem struct {
	Target  string `json:"target"`  // 段落/经历定位
	Action  string `json:"action"`  // keep | compress | rewrite | quantify
	Before  string `json:"before"`
	After   string `json:"after"`
	Reason  string `json:"reason"`
}

type Report struct {
	TotalScore   int              `json:"totalScore"`
	Summary      string           `json:"summary"`
	Dimensions   []DimensionScore `json:"dimensions"`
	Issues       []Issue          `json:"issues"`
	Rewrites     []RewriteItem    `json:"rewrites"`
	Source       string           `json:"source"` // llm | heuristic
	GeneratedAt  string           `json:"generatedAt"`
}

const systemPrompt = `你是一位严谨的社招人事（HR），专门筛选 3–10 年 ToB / 安全 / 合规 / 车联网相关产品经理简历。
你的任务是：对照目标 JD，从「人事 6 秒筛简历」视角给分、找硬伤、给可执行改写清单。

硬性规则：
1. 不要编造候选人没有的经历；缺信息就在 issues 里标「待补充」。
2. 分数必须有扣分理由，禁止空泛夸奖。
3. 关注：与 JD 匹配度、职业主线是否清晰、量化与客户场景可信度、时间断层/过短经历/术语硬伤/夸大风险。
4. 改写建议要具体到可改的句子或段落，action 只能是 keep/compress/rewrite/quantify。
5. 只输出 JSON，不要 Markdown。`

func Analyze(client *llm.Client, resume, jd, company, role string) (*Report, error) {
	if client != nil && client.Enabled() {
		rep, err := analyzeWithLLM(client, resume, jd, company, role)
		if err == nil {
			return rep, nil
		}
		// fall through to heuristic with error note? Better return LLM error so user fixes config.
		return nil, err
	}
	return analyzeHeuristic(resume, jd, company, role), nil
}

func analyzeWithLLM(client *llm.Client, resume, jd, company, role string) (*Report, error) {
	user := fmt.Sprintf(`公司：%s
目标岗位：%s

【目标 JD】
%s

【候选人简历】
%s

请输出 JSON，字段如下：
{
  "totalScore": 0-100 整数,
  "summary": "一句话人事结论",
  "dimensions": [
    {"key":"match","label":"JD匹配度","score":0-100,"comment":"..."},
    {"key":"story","label":"主线清晰度","score":0-100,"comment":"..."},
    {"key":"credibility","label":"可信度/量化","score":0-100,"comment":"..."},
    {"key":"risk","label":"风险控制（越高越好）","score":0-100,"comment":"..."}
  ],
  "issues": [
    {"severity":"critical|warn|info","title":"...","detail":"..."}
  ],
  "rewrites": [
    {"target":"定位","action":"keep|compress|rewrite|quantify","before":"原句摘录","after":"改写示例","reason":"..."}
  ]
}`, company, role, jd, resume)

	content, err := client.ChatJSON(systemPrompt, user)
	if err != nil {
		return nil, err
	}
	content = stripCodeFence(content)
	var rep Report
	if err := json.Unmarshal([]byte(content), &rep); err != nil {
		return nil, fmt.Errorf("模型返回非预期 JSON: %w; 片段: %s", err, truncate(content, 200))
	}
	rep.Source = "llm"
	normalize(&rep)
	return &rep, nil
}

func analyzeHeuristic(resume, jd, company, role string) *Report {
	resumeLower := strings.ToLower(resume)
	jdTokens := extractKeywords(jd)
	hit := 0
	var missing []string
	for _, kw := range jdTokens {
		if strings.Contains(resumeLower, strings.ToLower(kw)) {
			hit++
		} else if len(missing) < 8 {
			missing = append(missing, kw)
		}
	}
	matchScore := 55
	if len(jdTokens) > 0 {
		matchScore = 40 + int(float64(hit)/float64(len(jdTokens))*55)
	}

	digitCount := len(regexp.MustCompile(`[0-9]+`).FindAllString(resume, -1))
	credScore := 45
	if digitCount >= 8 {
		credScore = 75
	} else if digitCount >= 3 {
		credScore = 60
	}

	storyScore := 58
	if strings.Contains(resume, "产品经理") || strings.Contains(resume, "产品") {
		storyScore += 8
	}
	if company != "" && strings.Contains(resume, company) {
		storyScore += 5
	}
	if role != "" {
		for _, p := range strings.Fields(role) {
			if len([]rune(p)) >= 2 && strings.Contains(resume, p) {
				storyScore += 4
				break
			}
		}
	}
	if storyScore > 90 {
		storyScore = 90
	}

	riskScore := 70
	issues := []Issue{}
	if utf8.RuneCountInString(resume) < 400 {
		riskScore -= 15
		issues = append(issues, Issue{Severity: "warn", Title: "简历过短", Detail: "正文信息偏少，人事难以判断交付深度，建议补客户/职责/结果。"})
	}
	if !regexp.MustCompile(`20\d{2}`).MatchString(resume) {
		riskScore -= 10
		issues = append(issues, Issue{Severity: "warn", Title: "缺少时间线", Detail: "未见明显年份，建议补齐各段工作起止时间。"})
	}
	if digitCount < 3 {
		riskScore -= 10
		issues = append(issues, Issue{Severity: "warn", Title: "缺少量化", Detail: "几乎没有数字结果（客户数、周期、覆盖范围等），可信度偏弱。"})
	}
	if len(missing) > 0 {
		issues = append(issues, Issue{
			Severity: "critical",
			Title:    "JD 关键词覆盖不足",
			Detail:   "简历中较少出现：" + strings.Join(missing, "、") + "。请按目标岗补主线，而不是堆无关经历。",
		})
	}
	if len(issues) == 0 {
		issues = append(issues, Issue{Severity: "info", Title: "启发式初评", Detail: "当前未配置中转站，结果为规则初评。配置 JA_LLM_* 后可获得人事视角深度报告。"})
	} else {
		issues = append(issues, Issue{Severity: "info", Title: "启发式初评", Detail: "当前未配置 LLM 中转站，以上为规则引擎结果；配置后将替换为模型人事评分。"})
	}

	rewrites := []RewriteItem{}
	if digitCount < 3 {
		rewrites = append(rewrites, RewriteItem{
			Target: "工作/项目结果句",
			Action: "quantify",
			Before: "负责平台能力建设并推动落地",
			After:  "负责××平台能力建设，推动 N 个主机厂/客户落地，覆盖从××到××的关键阀点，按期完成中期交付",
			Reason: "用可验证数字和场景替换空泛职责",
		})
	}
	if len(missing) > 0 {
		rewrites = append(rewrites, RewriteItem{
			Target: "个人优势 / 最近一段经历首句",
			Action: "rewrite",
			Before: "（当前首句若未点题目标岗）",
			After:  fmt.Sprintf("面向%s，突出与 JD 相关的关键词：%s", nonempty(role, "目标岗位"), strings.Join(missing[:min(3, len(missing))], "、")),
			Reason: "让人事 6 秒内看到岗位匹配信号",
		})
	}
	rewrites = append(rewrites, RewriteItem{
		Target: "过短或弱相关经历",
		Action: "compress",
		Before: "多段经历平均用力",
		After:  "最近目标相关经历写细，其余压缩为 2–3 条结果句",
		Reason: "避免主线发散",
	})

	total := (matchScore + storyScore + credScore + riskScore) / 4
	rep := &Report{
		TotalScore: total,
		Summary:    fmt.Sprintf("启发式初评 %d 分：JD 关键词命中 %d/%d。配置中转站后可生成完整人事报告。", total, hit, len(jdTokens)),
		Dimensions: []DimensionScore{
			{Key: "match", Label: "JD匹配度", Score: matchScore, Comment: fmt.Sprintf("关键词命中 %d/%d", hit, len(jdTokens))},
			{Key: "story", Label: "主线清晰度", Score: storyScore, Comment: "基于标题/岗位词的粗判"},
			{Key: "credibility", Label: "可信度/量化", Score: credScore, Comment: fmt.Sprintf("检测到约 %d 处数字", digitCount)},
			{Key: "risk", Label: "风险控制", Score: riskScore, Comment: "时间线、篇幅、硬伤粗检"},
		},
		Issues:   issues,
		Rewrites: rewrites,
		Source:   "heuristic",
	}
	normalize(rep)
	return rep
}

func extractKeywords(jd string) []string {
	// pull CJK/tech tokens of length>=2 from JD; keep distinctive ones
	re := regexp.MustCompile(`[\p{Han}]{2,8}|[A-Za-z][A-Za-z0-9+/#.-]{1,20}`)
	raw := re.FindAllString(jd, -1)
	stop := map[string]bool{
		"负责": true, "参与": true, "完成": true, "工作": true, "经验": true, "优先": true, "以上": true,
		"具备": true, "熟悉": true, "了解": true, "能力": true, "岗位": true, "职责": true, "要求": true,
		"以及": true, "进行": true, "相关": true, "我们": true, "公司": true, "项目": true, "团队": true,
		"and": true, "the": true, "for": true, "with": true, "to": true,
	}
	seen := map[string]bool{}
	var out []string
	for _, w := range raw {
		k := strings.ToLower(w)
		if stop[k] || stop[w] || seen[k] {
			continue
		}
		seen[k] = true
		out = append(out, w)
		if len(out) >= 18 {
			break
		}
	}
	return out
}

func normalize(rep *Report) {
	if rep.TotalScore < 0 {
		rep.TotalScore = 0
	}
	if rep.TotalScore > 100 {
		rep.TotalScore = 100
	}
	if rep.Dimensions == nil {
		rep.Dimensions = []DimensionScore{}
	}
	if rep.Issues == nil {
		rep.Issues = []Issue{}
	}
	if rep.Rewrites == nil {
		rep.Rewrites = []RewriteItem{}
	}
	for i := range rep.Dimensions {
		if rep.Dimensions[i].Score < 0 {
			rep.Dimensions[i].Score = 0
		}
		if rep.Dimensions[i].Score > 100 {
			rep.Dimensions[i].Score = 100
		}
	}
}

func stripCodeFence(s string) string {
	s = strings.TrimSpace(s)
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

func nonempty(a, b string) string {
	if strings.TrimSpace(a) != "" {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
