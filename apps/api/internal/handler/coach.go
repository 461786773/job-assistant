package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/zhangyongjie/job-assistant/internal/bigfive"
	"github.com/zhangyongjie/job-assistant/internal/coach"
	"github.com/zhangyongjie/job-assistant/internal/db"
)

type createCoachBody struct {
	Scene               string `json:"scene"`
	RelatedTaskID       string `json:"relatedTaskId"`
	RelatedEvent        string `json:"relatedEvent"`
	RelatedQuickCheckID string `json:"relatedQuickCheckId"`
	Mode                string `json:"mode"` // trial | formal；默认 formal
	SkipQuickGate       bool   `json:"skipQuickGate"`
}

func startOfLocalDay(now time.Time) time.Time {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		loc = time.Local
	}
	t := now.In(loc)
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, loc)
}

func (h *Handler) ListCoachSessions(w http.ResponseWriter, r *http.Request) {
	claims := requireUser(w, r)
	if claims == nil {
		return
	}
	items, err := h.Store.ListCoachSessions(claims.UserID, 50)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) GetCoachSession(w http.ResponseWriter, r *http.Request) {
	claims := requireUser(w, r)
	if claims == nil {
		return
	}
	id := chi.URLParam(r, "id")
	sess, err := h.Store.GetCoachSession(id, claims.UserID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if sess == nil {
		writeErr(w, http.StatusNotFound, "会话不存在")
		return
	}
	writeJSON(w, http.StatusOK, sess)
}

func (h *Handler) CreateCoachSession(w http.ResponseWriter, r *http.Request) {
	claims := requireUser(w, r)
	if claims == nil {
		return
	}
	var body createCoachBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "无效 JSON")
		return
	}
	scene := strings.TrimSpace(body.Scene)
	if scene == "" {
		scene = "job_search"
	}
	if scene != "job_search" && scene != "promotion" && scene != "communication" {
		writeErr(w, http.StatusBadRequest, "scene 需为 job_search / promotion / communication")
		return
	}

	taskHint := ""
	relatedTaskID := strings.TrimSpace(body.RelatedTaskID)
	if relatedTaskID != "" {
		task, err := h.Store.GetTask(relatedTaskID, claims.UserID)
		if err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		if task == nil {
			writeErr(w, http.StatusBadRequest, "关联任务不存在")
			return
		}
		taskHint = task.Title + "\n" + task.Company + " · " + task.TargetRole + "\n" + truncateRunes(task.JDText, 600)
	}

	quickHint := ""
	quickID := strings.TrimSpace(body.RelatedQuickCheckID)
	mode := strings.ToLower(strings.TrimSpace(body.Mode))
	if mode == "" {
		if body.SkipQuickGate {
			mode = "trial"
		} else {
			mode = "formal"
		}
	}
	if mode != "trial" && mode != "formal" {
		writeErr(w, http.StatusBadRequest, "mode 需为 trial 或 formal")
		return
	}

	var quick *db.QuickSelfCheck
	if quickID != "" {
		item, err := h.Store.GetQuickSelfCheck(quickID, claims.UserID)
		if err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		if item == nil {
			writeErr(w, http.StatusBadRequest, "关联的自评不存在")
			return
		}
		quick = item
		quickHint = coach.FormatQuickCheck(item)
	} else if mode == "formal" {
		// 认真聊：当日已有快照可沿用（日历日，非滚动 24h）
		recent, err := h.Store.LatestQuickSelfCheckSince(claims.UserID, startOfLocalDay(time.Now()))
		if err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		if recent == nil {
			writeJSON(w, http.StatusConflict, map[string]any{
				"error":    "认真聊之前，想先邀你花三分钟看看自己",
				"code":     "quick_check_required",
				"redirect": "/wellbeing/quick",
			})
			return
		}
		quick = recent
		quickID = recent.ID
		quickHint = coach.FormatQuickCheck(recent)
	} else if mode == "trial" {
		// 轻松聊：无门禁，但有快照则注入最新（§0.8）
		if latestQuick, _ := h.Store.LatestQuickSelfCheck(claims.UserID); latestQuick != nil {
			quick = latestQuick
			quickID = latestQuick.ID
			quickHint = coach.FormatQuickCheck(latestQuick)
		}
	}

	assessmentHint := ""
	bigFiveHint := ""
	primaryNeed := ""
	if user, _ := h.Store.GetUserByID(claims.UserID); user != nil {
		primaryNeed = user.PrimaryNeed
	}
	if latest, _ := h.Store.LatestInitialAssessment(claims.UserID); latest != nil {
		assessmentHint = latest.SummaryForCoach
	}
	if bf, _ := h.Store.LatestBigFiveProfile(claims.UserID); bf != nil {
		bigFiveHint = bigfive.SummaryForCoachLive(bf.PersonaID, bf.Scores)
		if bigFiveHint == "" {
			bigFiveHint = bf.SummaryForCoach
		}
	}

	sess, err := coach.Start(h.LLM, scene, strings.TrimSpace(body.RelatedEvent), taskHint, quickHint, assessmentHint, bigFiveHint, primaryNeed)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}
	now := db.Now()
	sess.ID = newID()
	sess.UserID = claims.UserID
	sess.RelatedTaskID = relatedTaskID
	if mode == "trial" {
		sess.Title = coach.SceneLabel(scene) + " · 轻松聊聊"
	}
	sess.CreatedAt = now
	sess.UpdatedAt = now
	if err := h.Store.CreateCoachSession(sess); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if quick != nil && quick.RelatedCoachSessionID == "" && mode == "formal" {
		quick.RelatedCoachSessionID = sess.ID
		_ = h.Store.UpdateQuickSelfCheck(quick)
	}
	_ = quickID
	writeJSON(w, http.StatusCreated, sess)
}

func (h *Handler) ReplyCoachSession(w http.ResponseWriter, r *http.Request) {
	claims := requireUser(w, r)
	if claims == nil {
		return
	}
	id := chi.URLParam(r, "id")
	sess, err := h.Store.GetCoachSession(id, claims.UserID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if sess == nil {
		writeErr(w, http.StatusNotFound, "会话不存在")
		return
	}
	var body struct {
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "无效 JSON")
		return
	}

	taskHint := ""
	if sess.RelatedTaskID != "" {
		if task, _ := h.Store.GetTask(sess.RelatedTaskID, claims.UserID); task != nil {
			taskHint = task.Title + "\n" + task.Company + " · " + task.TargetRole
		}
	}

	assessmentHint := ""
	bigFiveHint := ""
	quickHint := ""
	if latest, _ := h.Store.LatestInitialAssessment(claims.UserID); latest != nil {
		assessmentHint = latest.SummaryForCoach
	}
	if bf, _ := h.Store.LatestBigFiveProfile(claims.UserID); bf != nil {
		bigFiveHint = bigfive.SummaryForCoachLive(bf.PersonaID, bf.Scores)
		if bigFiveHint == "" {
			bigFiveHint = bf.SummaryForCoach
		}
	}
	// §0.8：本会话关联快照优先，否则取最新快照
	if linked, _ := h.Store.QuickSelfCheckByCoachSession(claims.UserID, sess.ID); linked != nil {
		quickHint = coach.FormatQuickCheck(linked)
	} else if latestQuick, _ := h.Store.LatestQuickSelfCheck(claims.UserID); latestQuick != nil {
		quickHint = coach.FormatQuickCheck(latestQuick)
	}
	profileCtx := coach.MergeUserProfile(bigFiveHint, assessmentHint, quickHint)

	updated, err := coach.Reply(h.LLM, sess, body.Message, taskHint, profileCtx)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	updated.UpdatedAt = db.Now()
	if err := h.Store.UpdateCoachSession(updated); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (h *Handler) DeleteCoachSession(w http.ResponseWriter, r *http.Request) {
	claims := requireUser(w, r)
	if claims == nil {
		return
	}
	id := chi.URLParam(r, "id")
	if err := h.Store.DeleteCoachSession(id, claims.UserID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeErr(w, http.StatusNotFound, "会话不存在")
			return
		}
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func truncateRunes(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}
