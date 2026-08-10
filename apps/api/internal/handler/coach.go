package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/zhangyongjie/job-assistant/internal/coach"
	"github.com/zhangyongjie/job-assistant/internal/db"
)

type createCoachBody struct {
	Scene         string `json:"scene"`
	RelatedTaskID string `json:"relatedTaskId"`
	RelatedEvent  string `json:"relatedEvent"`
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

	sess, err := coach.Start(h.LLM, scene, strings.TrimSpace(body.RelatedEvent), taskHint)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}
	now := db.Now()
	sess.ID = newID()
	sess.UserID = claims.UserID
	sess.RelatedTaskID = relatedTaskID
	sess.CreatedAt = now
	sess.UpdatedAt = now
	if err := h.Store.CreateCoachSession(sess); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
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

	updated, err := coach.Reply(h.LLM, sess, body.Message, taskHint)
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
