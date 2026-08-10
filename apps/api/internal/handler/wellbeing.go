package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/zhangyongjie/job-assistant/internal/db"
)

type createCheckInBody struct {
	At                    string   `json:"at"`
	StressScore           int      `json:"stressScore"`
	MoodTags              []string `json:"moodTags"`
	EnergyScore           int      `json:"energyScore"`
	EventType             string   `json:"eventType"`
	Note                  string   `json:"note"`
	RelatedTaskID         string   `json:"relatedTaskId"`
	RelatedCoachSessionID string   `json:"relatedCoachSessionId"`
}

func (h *Handler) ListCheckIns(w http.ResponseWriter, r *http.Request) {
	claims := requireUser(w, r)
	if claims == nil {
		return
	}
	items, err := h.Store.ListCheckIns(claims.UserID, 90)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	summary := wellbeingSummary(items)
	writeJSON(w, http.StatusOK, map[string]any{
		"items":   items,
		"summary": summary,
	})
}

func (h *Handler) CreateCheckIn(w http.ResponseWriter, r *http.Request) {
	claims := requireUser(w, r)
	if claims == nil {
		return
	}
	var body createCheckInBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "无效 JSON")
		return
	}
	if body.StressScore < 1 || body.StressScore > 10 {
		writeErr(w, http.StatusBadRequest, "压力分需为 1–10")
		return
	}
	at := strings.TrimSpace(body.At)
	if at == "" {
		at = db.Now()
	}
	c := &db.WellbeingCheckIn{
		ID:                    newID(),
		UserID:                claims.UserID,
		At:                    at,
		StressScore:           body.StressScore,
		MoodTags:              body.MoodTags,
		EnergyScore:           body.EnergyScore,
		EventType:             strings.TrimSpace(body.EventType),
		Note:                  strings.TrimSpace(body.Note),
		RelatedTaskID:         strings.TrimSpace(body.RelatedTaskID),
		RelatedCoachSessionID: strings.TrimSpace(body.RelatedCoachSessionID),
		CreatedAt:             db.Now(),
	}
	if c.MoodTags == nil {
		c.MoodTags = []string{}
	}
	if err := h.Store.CreateCheckIn(c); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, c)
}

func (h *Handler) DeleteCheckIn(w http.ResponseWriter, r *http.Request) {
	claims := requireUser(w, r)
	if claims == nil {
		return
	}
	id := chi.URLParam(r, "id")
	if err := h.Store.DeleteCheckIn(id, claims.UserID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeErr(w, http.StatusNotFound, "打卡不存在")
			return
		}
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func wellbeingSummary(items []db.WellbeingCheckIn) map[string]any {
	now := time.Now().UTC()
	avg7, n7 := avgStressSince(items, now.AddDate(0, 0, -7))
	avg30, n30 := avgStressSince(items, now.AddDate(0, 0, -30))
	moods := map[string]int{}
	highEvents := []map[string]any{}
	for _, c := range items {
		for _, m := range c.MoodTags {
			moods[m]++
		}
		if c.StressScore >= 7 {
			highEvents = append(highEvents, map[string]any{
				"at":        c.At,
				"stress":    c.StressScore,
				"eventType": c.EventType,
				"note":      c.Note,
			})
			if len(highEvents) >= 5 {
				break
			}
		}
	}
	return map[string]any{
		"avgStress7":       avg7,
		"checkInCount7":    n7,
		"avgStress30":      avg30,
		"checkInCount30":   n30,
		"moodDistribution": moods,
		"recentHighStress": highEvents,
	}
}

func avgStressSince(items []db.WellbeingCheckIn, since time.Time) (float64, int) {
	sum, n := 0, 0
	for _, c := range items {
		t, err := time.Parse(time.RFC3339, c.At)
		if err != nil {
			continue
		}
		if t.Before(since) {
			continue
		}
		sum += c.StressScore
		n++
	}
	if n == 0 {
		return 0, 0
	}
	return float64(sum) / float64(n), n
}

type createQuickSelfCheckBody struct {
	At                    string   `json:"at"`
	Feelings              []string `json:"feelings"`
	Duration              string   `json:"duration"`
	Impacts               []string `json:"impacts"`
	DistressScore         int      `json:"distressScore"`
	TriggerNote           string   `json:"triggerNote"`
	Takeaway              string   `json:"takeaway"`
	RelatedCoachSessionID string   `json:"relatedCoachSessionId"`
	AlsoCheckIn           *bool    `json:"alsoCheckIn"` // default true: mirror into wellbeing timeline
}

var allowedFeelings = map[string]string{
	"tired": "累", "irritable": "烦", "numb": "空",
	"afraid": "怕", "stuck": "堵", "indifferent": "无所谓",
}
var allowedDurations = map[string]bool{
	"few_days": true, "one_two_weeks": true, "over_month": true, "unclear_chronic": true,
}
var allowedImpacts = map[string]bool{
	"sleep": true, "appetite": true, "focus": true, "temper": true, "body": true, "mood_only": true,
}
var allowedTakeaways = map[string]bool{
	"clarity": true, "strength": true, "tiny_tool": true, "just_talk": true, "unsure_but_here": true,
}

func (h *Handler) ListQuickSelfChecks(w http.ResponseWriter, r *http.Request) {
	claims := requireUser(w, r)
	if claims == nil {
		return
	}
	items, err := h.Store.ListQuickSelfChecks(claims.UserID, 30)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) GetQuickSelfCheck(w http.ResponseWriter, r *http.Request) {
	claims := requireUser(w, r)
	if claims == nil {
		return
	}
	id := chi.URLParam(r, "id")
	item, err := h.Store.GetQuickSelfCheck(id, claims.UserID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if item == nil {
		writeErr(w, http.StatusNotFound, "自评不存在")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *Handler) CreateQuickSelfCheck(w http.ResponseWriter, r *http.Request) {
	claims := requireUser(w, r)
	if claims == nil {
		return
	}
	var body createQuickSelfCheckBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "无效 JSON")
		return
	}
	feelings, err := normalizeFeelings(body.Feelings)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	duration := strings.TrimSpace(body.Duration)
	if !allowedDurations[duration] {
		writeErr(w, http.StatusBadRequest, "请选择感觉持续时长")
		return
	}
	impacts, err := normalizeImpacts(body.Impacts)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	if body.DistressScore < 0 || body.DistressScore > 10 {
		writeErr(w, http.StatusBadRequest, "状态分需为 0–10")
		return
	}
	takeaway := strings.TrimSpace(body.Takeaway)
	if !allowedTakeaways[takeaway] {
		writeErr(w, http.StatusBadRequest, "请选择今天最想带走什么")
		return
	}

	at := strings.TrimSpace(body.At)
	if at == "" {
		at = db.Now()
	}
	now := db.Now()
	c := &db.QuickSelfCheck{
		ID:                    newID(),
		UserID:                claims.UserID,
		At:                    at,
		Version:               "v1",
		Feelings:              feelings,
		Duration:              duration,
		Impacts:               impacts,
		DistressScore:         body.DistressScore,
		TriggerNote:           strings.TrimSpace(body.TriggerNote),
		Takeaway:              takeaway,
		RelatedCoachSessionID: strings.TrimSpace(body.RelatedCoachSessionID),
		CreatedAt:             now,
	}
	if err := h.Store.CreateQuickSelfCheck(c); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	alsoCheckIn := true
	if body.AlsoCheckIn != nil {
		alsoCheckIn = *body.AlsoCheckIn
	}
	if alsoCheckIn {
		stress := body.DistressScore
		if stress < 1 {
			stress = 1
		}
		moods := make([]string, 0, len(feelings))
		for _, f := range feelings {
			if label, ok := allowedFeelings[f]; ok {
				moods = append(moods, label)
			}
		}
		note := "三分钟自评"
		if c.TriggerNote != "" {
			note += " · " + c.TriggerNote
		}
		_ = h.Store.CreateCheckIn(&db.WellbeingCheckIn{
			ID:                    newID(),
			UserID:                claims.UserID,
			At:                    at,
			StressScore:           stress,
			MoodTags:              moods,
			EnergyScore:           0,
			EventType:             "other",
			Note:                  note,
			RelatedCoachSessionID: c.RelatedCoachSessionID,
			CreatedAt:             now,
		})
	}

	writeJSON(w, http.StatusCreated, c)
}

func (h *Handler) DeleteQuickSelfCheck(w http.ResponseWriter, r *http.Request) {
	claims := requireUser(w, r)
	if claims == nil {
		return
	}
	id := chi.URLParam(r, "id")
	if err := h.Store.DeleteQuickSelfCheck(id, claims.UserID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeErr(w, http.StatusNotFound, "自评不存在")
			return
		}
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func normalizeFeelings(in []string) ([]string, error) {
	if len(in) == 0 {
		return nil, fmt.Errorf("请至少勾选一种感觉（最多 2 项）")
	}
	if len(in) > 2 {
		return nil, fmt.Errorf("感觉最多勾选 2 项")
	}
	out := make([]string, 0, len(in))
	seen := map[string]bool{}
	for _, raw := range in {
		v := strings.TrimSpace(raw)
		if _, ok := allowedFeelings[v]; !ok {
			return nil, fmt.Errorf("无效的感觉选项")
		}
		if seen[v] {
			continue
		}
		seen[v] = true
		out = append(out, v)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("请至少勾选一种感觉（最多 2 项）")
	}
	return out, nil
}

func normalizeImpacts(in []string) ([]string, error) {
	if len(in) == 0 {
		return nil, fmt.Errorf("请至少勾选一项影响面")
	}
	out := make([]string, 0, len(in))
	seen := map[string]bool{}
	hasMoodOnly := false
	for _, raw := range in {
		v := strings.TrimSpace(raw)
		if !allowedImpacts[v] {
			return nil, fmt.Errorf("无效的影响面选项")
		}
		if seen[v] {
			continue
		}
		seen[v] = true
		if v == "mood_only" {
			hasMoodOnly = true
		}
		out = append(out, v)
	}
	// 「都没影响」与其他项互斥：若同时勾选，只保留 mood_only
	if hasMoodOnly && len(out) > 1 {
		out = []string{"mood_only"}
	}
	return out, nil
}
