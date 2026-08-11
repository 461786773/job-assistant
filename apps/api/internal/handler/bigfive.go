package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/zhangyongjie/job-assistant/internal/bigfive"
	"github.com/zhangyongjie/job-assistant/internal/db"
)

func (h *Handler) ListBigFive(w http.ResponseWriter, r *http.Request) {
	claims := requireUser(w, r)
	if claims == nil {
		return
	}
	items, err := h.Store.ListBigFiveProfiles(claims.UserID, 10)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	for i := range items {
		sanitizeBigFiveForClient(&items[i])
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) LatestBigFive(w http.ResponseWriter, r *http.Request) {
	claims := requireUser(w, r)
	if claims == nil {
		return
	}
	item, err := h.Store.LatestBigFiveProfile(claims.UserID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if item == nil {
		writeErr(w, http.StatusNotFound, "尚未完成职场画像")
		return
	}
	sanitizeBigFiveForClient(item)
	writeJSON(w, http.StatusOK, item)
}

func (h *Handler) GetBigFive(w http.ResponseWriter, r *http.Request) {
	claims := requireUser(w, r)
	if claims == nil {
		return
	}
	id := chi.URLParam(r, "id")
	item, err := h.Store.GetBigFiveProfile(id, claims.UserID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if item == nil {
		writeErr(w, http.StatusNotFound, "画像不存在")
		return
	}
	sanitizeBigFiveForClient(item)
	writeJSON(w, http.StatusOK, item)
}

func (h *Handler) CreateBigFive(w http.ResponseWriter, r *http.Request) {
	claims := requireUser(w, r)
	if claims == nil {
		return
	}
	var body struct {
		Answers bigfive.Answers `json:"answers"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "无效 JSON")
		return
	}
	result, err := bigfive.Build(body.Answers)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	answersJSON, _ := json.Marshal(body.Answers)
	scoresJSON, _ := json.Marshal(result.Scores)
	now := db.Now()
	rec := &db.WorkplaceBigFiveProfile{
		ID:              newID(),
		UserID:          claims.UserID,
		Version:         result.Version,
		RawAnswers:      answersJSON,
		Scores:          scoresJSON,
		PersonaID:       result.PersonaID,
		PersonaTitle:    result.PersonaTitle,
		PersonaBlurb:    result.PersonaBlurb,
		PersonaBody:     result.PersonaBody,
		Tags:            result.Tags,
		CoachHints:      result.CoachHints,
		SummaryForCoach: result.SummaryCoach,
		CompletedAt:     now,
		CreatedAt:       now,
	}
	if err := h.Store.CreateBigFiveProfile(rec); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	sanitizeBigFiveForClient(rec)
	writeJSON(w, http.StatusCreated, rec)
}

// sanitizeBigFiveForClient 去掉教练侧提示，并刷新用户可见描写。
func sanitizeBigFiveForClient(p *db.WorkplaceBigFiveProfile) {
	if p == nil {
		return
	}
	p.CoachHints = nil
	p.SummaryForCoach = ""
	if body := bigfive.BodyForPersona(p.PersonaID); body != "" {
		p.PersonaBody = body
	}
}
