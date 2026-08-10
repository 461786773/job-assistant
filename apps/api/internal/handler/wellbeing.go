package handler

import (
	"encoding/json"
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
				"at":         c.At,
				"stress":     c.StressScore,
				"eventType":  c.EventType,
				"note":       c.Note,
			})
			if len(highEvents) >= 5 {
				break
			}
		}
	}
	return map[string]any{
		"avgStress7":      avg7,
		"checkInCount7":   n7,
		"avgStress30":     avg30,
		"checkInCount30":  n30,
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
