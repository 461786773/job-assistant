package handler

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/zhangyongjie/job-assistant/internal/config"
	"github.com/zhangyongjie/job-assistant/internal/db"
	"github.com/zhangyongjie/job-assistant/internal/hr"
	"github.com/zhangyongjie/job-assistant/internal/interview"
	"github.com/zhangyongjie/job-assistant/internal/llm"
	"github.com/zhangyongjie/job-assistant/internal/resume"
	"github.com/zhangyongjie/job-assistant/internal/salary"
)

type Handler struct {
	Store *db.Store
	Cfg   config.Config
	LLM   *llm.Client
}

func New(store *db.Store, cfg config.Config, llmClient *llm.Client) *Handler {
	return &Handler{Store: store, Cfg: cfg, LLM: llmClient}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func newID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":      true,
		"service": "job-assistant",
		"llm":     h.LLM != nil && h.LLM.Enabled(),
	})
}

func (h *Handler) ListTasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.Store.ListTasks()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": tasks})
}

func (h *Handler) GetTask(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	task, err := h.Store.GetTask(id)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if task == nil {
		writeErr(w, http.StatusNotFound, "任务不存在")
		return
	}
	writeJSON(w, http.StatusOK, task)
}

type createTaskBody struct {
	Title      string `json:"title"`
	Company    string `json:"company"`
	TargetRole string `json:"targetRole"`
	JDText     string `json:"jdText"`
	ResumeText string `json:"resumeText"`
	Notes      string `json:"notes"`
}

func (h *Handler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var body createTaskBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "无效的 JSON")
		return
	}
	now := db.Now()
	title := strings.TrimSpace(body.Title)
	if title == "" {
		if body.Company != "" || body.TargetRole != "" {
			title = strings.TrimSpace(body.Company + " · " + body.TargetRole)
		} else {
			title = "未命名任务"
		}
	}
	task := &db.Task{
		ID:         newID(),
		Title:      title,
		Company:    strings.TrimSpace(body.Company),
		TargetRole: strings.TrimSpace(body.TargetRole),
		JDText:     strings.TrimSpace(body.JDText),
		ResumeText: strings.TrimSpace(body.ResumeText),
		Status:     "draft",
		Notes:      strings.TrimSpace(body.Notes),
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if task.ResumeText != "" {
		task.ResumeFormat = "txt"
	}
	if err := h.Store.CreateTask(task); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, task)
}

type updateTaskBody struct {
	Title      *string `json:"title"`
	Company    *string `json:"company"`
	TargetRole *string `json:"targetRole"`
	JDText     *string `json:"jdText"`
	ResumeText *string `json:"resumeText"`
	Status     *string `json:"status"`
	Notes      *string `json:"notes"`
}

func (h *Handler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	task, err := h.Store.GetTask(id)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if task == nil {
		writeErr(w, http.StatusNotFound, "任务不存在")
		return
	}
	var body updateTaskBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "无效的 JSON")
		return
	}
	if body.Title != nil {
		task.Title = strings.TrimSpace(*body.Title)
	}
	if body.Company != nil {
		task.Company = strings.TrimSpace(*body.Company)
	}
	if body.TargetRole != nil {
		task.TargetRole = strings.TrimSpace(*body.TargetRole)
	}
	if body.JDText != nil {
		task.JDText = strings.TrimSpace(*body.JDText)
	}
	if body.ResumeText != nil {
		task.ResumeText = strings.TrimSpace(*body.ResumeText)
		if task.ResumeFilename == "" {
			task.ResumeFormat = "txt"
		}
	}
	if body.Status != nil {
		task.Status = strings.TrimSpace(*body.Status)
	}
	if body.Notes != nil {
		task.Notes = strings.TrimSpace(*body.Notes)
	}
	task.UpdatedAt = db.Now()
	if err := h.Store.UpdateTask(task); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, task)
}

func (h *Handler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.Store.DeleteTask(id); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) UploadResume(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	task, err := h.Store.GetTask(id)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if task == nil {
		writeErr(w, http.StatusNotFound, "任务不存在")
		return
	}

	if err := r.ParseMultipartForm(20 << 20); err != nil {
		writeErr(w, http.StatusBadRequest, "无法解析上传文件（限制 20MB）")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeErr(w, http.StatusBadRequest, "请使用字段名 file 上传简历")
		return
	}
	defer file.Close()

	raw, err := io.ReadAll(file)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "读取上传文件失败")
		return
	}

	parsed, err := resume.Parse(header.Filename, bytes.NewReader(raw))
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}

	safeName := fmt.Sprintf("%s_%d_%s", id, time.Now().Unix(), filepath.Base(header.Filename))
	dest := filepath.Join(h.Cfg.UploadDir, safeName)
	if err := os.WriteFile(dest, raw, 0o644); err != nil {
		writeErr(w, http.StatusInternalServerError, "保存上传文件失败")
		return
	}

	task.ResumeText = parsed.Text
	task.ResumeFilename = header.Filename
	task.ResumeFormat = parsed.Format
	task.UpdatedAt = db.Now()
	if err := h.Store.UpdateTask(task); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"task":   task,
		"chars":  len([]rune(parsed.Text)),
		"format": parsed.Format,
	})
}

func (h *Handler) ParseResume(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(20 << 20); err != nil {
		writeErr(w, http.StatusBadRequest, "无法解析上传文件（限制 20MB）")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeErr(w, http.StatusBadRequest, "请使用字段名 file 上传简历")
		return
	}
	defer file.Close()

	raw, err := io.ReadAll(file)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "读取上传文件失败")
		return
	}

	parsed, err := resume.Parse(header.Filename, bytes.NewReader(raw))
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"text":     parsed.Text,
		"format":   parsed.Format,
		"filename": header.Filename,
		"chars":    len([]rune(parsed.Text)),
	})
}

func (h *Handler) AnalyzeHR(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	task, err := h.Store.GetTask(id)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if task == nil {
		writeErr(w, http.StatusNotFound, "任务不存在")
		return
	}
	if strings.TrimSpace(task.ResumeText) == "" {
		writeErr(w, http.StatusBadRequest, "请先填写或上传简历")
		return
	}
	if strings.TrimSpace(task.JDText) == "" {
		writeErr(w, http.StatusBadRequest, "请先填写目标 JD")
		return
	}

	report, err := hr.Analyze(h.LLM, task.ResumeText, task.JDText, task.Company, task.TargetRole)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}
	report.GeneratedAt = db.Now()
	raw, err := json.Marshal(report)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	task.HrReport = raw
	if task.Status == "draft" {
		task.Status = "hr_done"
	}
	task.UpdatedAt = db.Now()
	if err := h.Store.UpdateTask(task); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"task":   task,
		"report": report,
	})
}

func (h *Handler) InterviewStart(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	task, err := h.Store.GetTask(id)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if task == nil {
		writeErr(w, http.StatusNotFound, "任务不存在")
		return
	}
	if strings.TrimSpace(task.ResumeText) == "" || strings.TrimSpace(task.JDText) == "" {
		writeErr(w, http.StatusBadRequest, "请先完善简历与 JD")
		return
	}
	sess, err := interview.Start(h.LLM, task.ResumeText, task.JDText, task.Company, task.TargetRole)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}
	sess.UpdatedAt = db.Now()
	raw, _ := json.Marshal(sess)
	task.Interview = raw
	task.UpdatedAt = db.Now()
	if err := h.Store.UpdateTask(task); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"task": task, "interview": sess})
}

func (h *Handler) InterviewReply(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	task, err := h.Store.GetTask(id)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if task == nil {
		writeErr(w, http.StatusNotFound, "任务不存在")
		return
	}
	var body struct {
		Answer string `json:"answer"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "无效 JSON")
		return
	}
	var sess interview.Session
	if len(task.Interview) == 0 {
		writeErr(w, http.StatusBadRequest, "请先开始业务关")
		return
	}
	if err := json.Unmarshal(task.Interview, &sess); err != nil {
		writeErr(w, http.StatusInternalServerError, "会话损坏")
		return
	}
	updated, err := interview.Reply(h.LLM, &sess, body.Answer, task.ResumeText, task.JDText, task.Company, task.TargetRole)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}
	updated.UpdatedAt = db.Now()
	raw, _ := json.Marshal(updated)
	task.Interview = raw
	if updated.Status == "done" {
		task.Status = "interview_done"
	}
	task.UpdatedAt = db.Now()
	if err := h.Store.UpdateTask(task); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"task": task, "interview": updated})
}

func (h *Handler) SalaryAnalyze(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	task, err := h.Store.GetTask(id)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if task == nil {
		writeErr(w, http.StatusNotFound, "任务不存在")
		return
	}
	var body salary.Case
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "无效 JSON")
		return
	}
	analysis, err := salary.Analyze(h.LLM, &body)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	body.Analysis = analysis
	raw, _ := json.Marshal(body)
	task.Salary = raw
	task.Status = "salary_done"
	task.UpdatedAt = db.Now()
	if err := h.Store.UpdateTask(task); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"task": task, "salary": body})
}
