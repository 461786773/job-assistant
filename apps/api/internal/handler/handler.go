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
	"github.com/zhangyongjie/job-assistant/internal/auth"
	"github.com/zhangyongjie/job-assistant/internal/config"
	"github.com/zhangyongjie/job-assistant/internal/db"
	"github.com/zhangyongjie/job-assistant/internal/hr"
	"github.com/zhangyongjie/job-assistant/internal/interview"
	"github.com/zhangyongjie/job-assistant/internal/llm"
	"github.com/zhangyongjie/job-assistant/internal/resume"
	"github.com/zhangyongjie/job-assistant/internal/salary"
)

type Handler struct {
	Store  *db.Store
	Cfg    config.Config
	LLM    *llm.Client
	Tokens *auth.TokenManager
}

func New(store *db.Store, cfg config.Config, llmClient *llm.Client, tokens *auth.TokenManager) *Handler {
	return &Handler{Store: store, Cfg: cfg, LLM: llmClient, Tokens: tokens}
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
		"role":    "workplace-coach",
		"llm":     h.LLM != nil && h.LLM.Enabled(),
	})
}

func (h *Handler) ListTasks(w http.ResponseWriter, r *http.Request) {
	claims := requireUser(w, r)
	if claims == nil {
		return
	}
	tasks, err := h.Store.ListTasks(claims.UserID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": tasks})
}

func (h *Handler) GetTask(w http.ResponseWriter, r *http.Request) {
	claims := requireUser(w, r)
	if claims == nil {
		return
	}
	id := chi.URLParam(r, "id")
	task, err := h.Store.GetTask(id, claims.UserID)
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
	claims := requireUser(w, r)
	if claims == nil {
		return
	}
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
		UserID:     claims.UserID,
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
	claims := requireUser(w, r)
	if claims == nil {
		return
	}
	id := chi.URLParam(r, "id")
	task, err := h.Store.GetTask(id, claims.UserID)
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
	claims := requireUser(w, r)
	if claims == nil {
		return
	}
	id := chi.URLParam(r, "id")
	if err := h.Store.DeleteTask(id, claims.UserID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeErr(w, http.StatusNotFound, "任务不存在")
			return
		}
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) UploadResume(w http.ResponseWriter, r *http.Request) {
	claims := requireUser(w, r)
	if claims == nil {
		return
	}
	id := chi.URLParam(r, "id")
	task, err := h.Store.GetTask(id, claims.UserID)
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
	if requireUser(w, r) == nil {
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
	writeJSON(w, http.StatusOK, map[string]any{
		"text":     parsed.Text,
		"format":   parsed.Format,
		"filename": header.Filename,
		"chars":    len([]rune(parsed.Text)),
	})
}

func (h *Handler) AnalyzeHR(w http.ResponseWriter, r *http.Request) {
	claims := requireUser(w, r)
	if claims == nil {
		return
	}
	id := chi.URLParam(r, "id")
	task, err := h.Store.GetTask(id, claims.UserID)
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

type applyHRBody struct {
	Indexes []int  `json:"indexes"`
	All     bool   `json:"all"`
	Mode    string `json:"mode"` // ai | direct；默认 ai（无 LLM 时回退 direct）
}

func (h *Handler) ApplyHRRewrites(w http.ResponseWriter, r *http.Request) {
	claims := requireUser(w, r)
	if claims == nil {
		return
	}
	id := chi.URLParam(r, "id")
	task, err := h.Store.GetTask(id, claims.UserID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if task == nil {
		writeErr(w, http.StatusNotFound, "任务不存在")
		return
	}
	if len(task.HrReport) == 0 {
		writeErr(w, http.StatusBadRequest, "请先生成简历评分")
		return
	}
	var report hr.Report
	if err := json.Unmarshal(task.HrReport, &report); err != nil {
		writeErr(w, http.StatusBadRequest, "简历优化报告损坏，请重新分析")
		return
	}
	if len(report.Rewrites) == 0 {
		writeErr(w, http.StatusBadRequest, "没有可应用的改写项")
		return
	}

	var body applyHRBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "无效 JSON")
		return
	}
	indexes := body.Indexes
	if body.All {
		indexes = make([]int, len(report.Rewrites))
		for i := range report.Rewrites {
			indexes[i] = i
		}
	} else if len(indexes) == 0 {
		writeErr(w, http.StatusBadRequest, "请指定 indexes 或 all=true")
		return
	}

	mode := strings.ToLower(strings.TrimSpace(body.Mode))
	if mode == "" {
		mode = "ai"
	}
	if mode != "ai" && mode != "direct" {
		writeErr(w, http.StatusBadRequest, "mode 只能是 ai 或 direct")
		return
	}

	prev := task.ResumeText
	var (
		next    string
		results []hr.ApplyResult
	)

	if mode == "ai" {
		if h.LLM == nil || !h.LLM.Enabled() {
			// 无模型时自动降级为直接替换
			mode = "direct"
		} else {
			var aiErr error
			next, results, aiErr = hr.ApplyRewritesWithAI(
				h.LLM, task.ResumeText, task.JDText, task.Company, task.TargetRole,
				report.Rewrites, indexes,
			)
			if aiErr != nil {
				writeErr(w, http.StatusBadGateway, "AI 改写失败: "+aiErr.Error())
				return
			}
		}
	}

	if mode == "direct" {
		next, results = hr.ApplyRewrites(task.ResumeText, report.Rewrites, indexes)
	}

	applied := 0
	for _, r := range results {
		if r.OK {
			applied++
		}
	}
	changed := next != prev && (applied > 0 || (mode == "ai" && strings.TrimSpace(next) != "" && next != prev))
	// AI 可能整体改写成功但 changes 标记不全
	if mode == "ai" && next != prev {
		changed = true
		if applied == 0 {
			applied = 1
			for i := range results {
				results[i].OK = true
				if results[i].Method == "" {
					results[i].Method = "ai"
				}
			}
		}
	}

	if !changed {
		writeJSON(w, http.StatusOK, map[string]any{
			"task":               task,
			"results":            results,
			"applied":            0,
			"previousResumeText": prev,
			"changed":            false,
			"mode":               mode,
		})
		return
	}

	task.ResumeText = next
	task.UpdatedAt = db.Now()
	if err := h.Store.UpdateTask(task); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"task":               task,
		"results":            results,
		"applied":            applied,
		"previousResumeText": prev,
		"changed":            true,
		"mode":               mode,
	})
}

func (h *Handler) InterviewStart(w http.ResponseWriter, r *http.Request) {
	claims := requireUser(w, r)
	if claims == nil {
		return
	}
	id := chi.URLParam(r, "id")
	task, err := h.Store.GetTask(id, claims.UserID)
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
	claims := requireUser(w, r)
	if claims == nil {
		return
	}
	id := chi.URLParam(r, "id")
	task, err := h.Store.GetTask(id, claims.UserID)
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
		writeErr(w, http.StatusBadRequest, "请先开始面试模拟")
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
	claims := requireUser(w, r)
	if claims == nil {
		return
	}
	id := chi.URLParam(r, "id")
	task, err := h.Store.GetTask(id, claims.UserID)
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
