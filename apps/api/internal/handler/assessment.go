package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/zhangyongjie/job-assistant/internal/assessment"
	"github.com/zhangyongjie/job-assistant/internal/db"
)

func (h *Handler) ListAssessments(w http.ResponseWriter, r *http.Request) {
	claims := requireUser(w, r)
	if claims == nil {
		return
	}
	items, err := h.Store.ListInitialAssessments(claims.UserID, 20)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) LatestAssessment(w http.ResponseWriter, r *http.Request) {
	claims := requireUser(w, r)
	if claims == nil {
		return
	}
	item, err := h.Store.LatestInitialAssessment(claims.UserID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if item == nil {
		writeErr(w, http.StatusNotFound, "尚无评测")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *Handler) GetAssessment(w http.ResponseWriter, r *http.Request) {
	claims := requireUser(w, r)
	if claims == nil {
		return
	}
	id := chi.URLParam(r, "id")
	item, err := h.Store.GetInitialAssessment(id, claims.UserID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if item == nil {
		writeErr(w, http.StatusNotFound, "评测不存在")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *Handler) CreateAssessment(w http.ResponseWriter, r *http.Request) {
	claims := requireUser(w, r)
	if claims == nil {
		return
	}
	var ans assessment.Answers
	if err := json.NewDecoder(r.Body).Decode(&ans); err != nil {
		writeErr(w, http.StatusBadRequest, "无效 JSON")
		return
	}
	ans.CrisisLevel = strings.TrimSpace(ans.CrisisLevel)
	if ans.CrisisLevel != "none" && ans.CrisisLevel != "fleeting" && ans.CrisisLevel != "elevated" {
		writeErr(w, http.StatusBadRequest, "请完成安全筛查")
		return
	}
	if len(ans.KeyEvents) > 2 {
		writeErr(w, http.StatusBadRequest, "关键事件最多选 2 项")
		return
	}
	if len(ans.MoodTags) > 3 {
		writeErr(w, http.StatusBadRequest, "情绪标签最多选 3 项")
		return
	}
	if len(ans.Goals) > 3 {
		writeErr(w, http.StatusBadRequest, "期望最多选 3 项")
		return
	}

	metrics, analysis, summary, err := assessment.Analyze(h.LLM, ans)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	rec, err := assessment.BuildRecord(claims.UserID, ans, metrics, analysis, summary)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	rec.ID = newID()
	if err := h.Store.CreateInitialAssessment(rec); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, rec)
}

type setNeedBody struct {
	Need string `json:"need"`
}

func (h *Handler) SetPrimaryNeed(w http.ResponseWriter, r *http.Request) {
	claims := requireUser(w, r)
	if claims == nil {
		return
	}
	var body setNeedBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "无效 JSON")
		return
	}
	need := strings.TrimSpace(body.Need)
	allowed := map[string]bool{
		"job_search": true, "promotion": true, "communication": true,
		"counsel_first": true, "unsure": true,
	}
	if !allowed[need] {
		writeErr(w, http.StatusBadRequest, "诉求需为 job_search / promotion / communication / counsel_first / unsure")
		return
	}
	if err := h.Store.UpdateUserPrimaryNeed(claims.UserID, need); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"primaryNeed": need})
}

func (h *Handler) ListBookings(w http.ResponseWriter, r *http.Request) {
	claims := requireUser(w, r)
	if claims == nil {
		return
	}
	items, err := h.Store.ListCounselingBookings(claims.UserID, 30)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

type createBookingBody struct {
	PreferredSlots string `json:"preferredSlots"`
	Note           string `json:"note"`
	ContactChannel string `json:"contactChannel"`
}

func (h *Handler) CreateBooking(w http.ResponseWriter, r *http.Request) {
	claims := requireUser(w, r)
	if claims == nil {
		return
	}
	var body createBookingBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "无效 JSON")
		return
	}
	slots := strings.TrimSpace(body.PreferredSlots)
	if slots == "" {
		writeErr(w, http.StatusBadRequest, "请填写时段偏好")
		return
	}
	now := db.Now()
	b := &db.CounselingBooking{
		ID:             newID(),
		UserID:         claims.UserID,
		PreferredSlots: slots,
		Note:           strings.TrimSpace(body.Note),
		ContactChannel: strings.TrimSpace(body.ContactChannel),
		Status:         "requested",
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := h.Store.CreateCounselingBooking(b); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, b)
}

type patchBookingBody struct {
	Status string `json:"status"`
}

func (h *Handler) PatchBooking(w http.ResponseWriter, r *http.Request) {
	claims := requireUser(w, r)
	if claims == nil {
		return
	}
	id := chi.URLParam(r, "id")
	b, err := h.Store.GetCounselingBooking(id, claims.UserID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if b == nil {
		writeErr(w, http.StatusNotFound, "预约不存在")
		return
	}
	var body patchBookingBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "无效 JSON")
		return
	}
	status := strings.TrimSpace(body.Status)
	// 用户侧仅允许 requested / cancelled；confirmed / done 由人工侧处理（P1-4）
	userAllowed := map[string]bool{"requested": true, "cancelled": true}
	if !userAllowed[status] {
		writeErr(w, http.StatusBadRequest, "用户只能设为「待确认」或「已取消」；确认与完成需人工处理")
		return
	}
	// 已完成的预约不可再改
	if b.Status == "done" {
		writeErr(w, http.StatusBadRequest, "已完成的预约不能再改状态")
		return
	}
	b.Status = status
	b.UpdatedAt = db.Now()
	if err := h.Store.UpdateCounselingBooking(b); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, b)
}
