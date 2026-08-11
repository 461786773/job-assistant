package salary

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/zhangyongjie/job-assistant/internal/llm"
	"github.com/zhangyongjie/job-assistant/internal/prompts"
)

type Comp struct {
	MonthlyBase      float64 `json:"monthlyBase"`
	MonthlyAllowance float64 `json:"monthlyAllowance"`
	Months           float64 `json:"months"` // e.g. 13 or 12
	YearEndMonths    float64 `json:"yearEndMonths"`
	YearEndGuaranteed bool   `json:"yearEndGuaranteed"`
	BonusYearly      float64 `json:"bonusYearly"`
	HousingFundBase  float64 `json:"housingFundBase"`
	HousingFundRate  float64 `json:"housingFundRate"` // 0.08
	Notes            string  `json:"notes"`
}

type Case struct {
	Current   Comp     `json:"current"`
	Offer     Comp     `json:"offer"`
	TargetPkg float64  `json:"targetPkg"`
	FloorPkg  float64  `json:"floorPkg"`
	Analysis  *Analysis `json:"analysis,omitempty"`
}

type Analysis struct {
	CurrentConservative float64  `json:"currentConservative"`
	CurrentTarget       float64  `json:"currentTarget"`
	OfferConservative   float64  `json:"offerConservative"`
	OfferTarget         float64  `json:"offerTarget"`
	Gaps                []string `json:"gaps"`
	AskPoints           []string `json:"askPoints"`
	Scripts             []string `json:"scripts"`
	Summary             string   `json:"summary"`
	Source              string   `json:"source"`
}

func Analyze(client *llm.Client, c *Case) (*Analysis, error) {
	if c == nil {
		return nil, fmt.Errorf("缺少薪酬数据")
	}
	curC, curT := packageRange(c.Current)
	offC, offT := packageRange(c.Offer)
	a := &Analysis{
		CurrentConservative: curC,
		CurrentTarget:       curT,
		OfferConservative:   offC,
		OfferTarget:         offT,
		Source:              "calc",
	}
	a.Gaps = []string{}
	a.AskPoints = []string{}
	a.Scripts = []string{}

	if !c.Offer.YearEndGuaranteed && c.Offer.YearEndMonths > 0 {
		a.Gaps = append(a.Gaps, "年终未书面保底，应按区间看，不要按满额单点谈")
		a.AskPoints = append(a.AskPoints, "年终几个月、是否保底、绩效系数口径能否写进 offer")
	}
	if c.Offer.MonthlyBase > 0 && c.Offer.MonthlyBase < 5000 {
		a.Gaps = append(a.Gaps, "基本工资偏低，可能影响加班费等基数")
		a.AskPoints = append(a.AskPoints, "总额不变前提下，能否提高基本/岗位、略降绩效占比")
	}
	if c.Current.HousingFundBase > 0 && c.Offer.HousingFundBase > 0 && c.Offer.HousingFundBase < c.Current.HousingFundBase {
		a.Gaps = append(a.Gaps, "公积金基数低于当前，年包隐性变少")
		a.AskPoints = append(a.AskPoints, "公积金基数与公司缴纳比例确认")
	}
	if c.FloorPkg > 0 && offT < c.FloorPkg {
		a.Gaps = append(a.Gaps, fmt.Sprintf("目标年包乐观值 %.0f 低于你的底线 %.0f", offT, c.FloorPkg))
	}
	a.AskPoints = append(a.AskPoints, "试用期是否打折、项目奖金规则、调薪频率")
	a.Scripts = append(a.Scripts,
		"我想确认一下：定薪时是否可以按我当前「应收结构」评估，而不只看近几个月流水？我这边有结构说明和截图。",
		"年终部分希望能明确：几个月、是否保底、绩效系数如何算；如果暂时不能写满额，我们按保守口径对齐预期。",
	)
	a.Summary = fmt.Sprintf("当前保守/目标约 %.1f / %.1f 万；Offer 保守/目标约 %.1f / %.1f 万。", curC/10000, curT/10000, offC/10000, offT/10000)

	if client != nil && client.Enabled() {
		raw, err := client.ChatJSON(
			prompts.SalaryNegotiate,
			fmt.Sprintf(`当前包：保守%.0f 目标%.0f
Offer包：保守%.0f 目标%.0f
底线%.0f 目标%.0f
已识别缺口：%s

输出 JSON：
{"summary":"...","askPoints":["..."],"scripts":["话术1","话术2"]}`,
				curC, curT, offC, offT, c.FloorPkg, c.TargetPkg, strings.Join(a.Gaps, "；")),
		)
		if err == nil {
			var extra struct {
				Summary   string   `json:"summary"`
				AskPoints []string `json:"askPoints"`
				Scripts   []string `json:"scripts"`
			}
			if json.Unmarshal([]byte(extractObj(raw)), &extra) == nil {
				if extra.Summary != "" {
					a.Summary = extra.Summary
				}
				if len(extra.AskPoints) > 0 {
					a.AskPoints = extra.AskPoints
				}
				if len(extra.Scripts) > 0 {
					a.Scripts = extra.Scripts
				}
				a.Source = "llm"
			}
		}
	}
	return a, nil
}

func packageRange(c Comp) (conservative, target float64) {
	months := c.Months
	if months <= 0 {
		months = 12
	}
	monthly := c.MonthlyBase + c.MonthlyAllowance
	baseYear := monthly * months
	ye := c.MonthlyBase * c.YearEndMonths
	if c.YearEndMonths <= 0 {
		ye = 0
	}
	bonus := c.BonusYearly
	hf := 0.0
	if c.HousingFundBase > 0 && c.HousingFundRate > 0 {
		hf = c.HousingFundBase * c.HousingFundRate * 12 // company side rough
	}
	conservative = baseYear + bonus*0.5 + hf
	if c.YearEndGuaranteed {
		conservative += ye
	} else {
		conservative += ye * 0.5
	}
	target = baseYear + ye + bonus + hf
	return conservative, target
}

func extractObj(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "{") {
		return s
	}
	i := strings.Index(s, "{")
	j := strings.LastIndex(s, "}")
	if i >= 0 && j > i {
		return s[i : j+1]
	}
	return s
}
