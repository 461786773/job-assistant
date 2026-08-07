package server

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/zhangyongjie/job-assistant/internal/config"
	"github.com/zhangyongjie/job-assistant/internal/handler"
)

func NewRouter(h *handler.Handler, cfg config.Config) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Get("/api/health", h.Health)
	r.Route("/api/tasks", func(r chi.Router) {
		r.Get("/", h.ListTasks)
		r.Post("/", h.CreateTask)
		r.Get("/{id}", h.GetTask)
		r.Patch("/{id}", h.UpdateTask)
		r.Delete("/{id}", h.DeleteTask)
		r.Post("/{id}/resume", h.UploadResume)
		r.Post("/{id}/hr/analyze", h.AnalyzeHR)
		r.Post("/{id}/interview/start", h.InterviewStart)
		r.Post("/{id}/interview/reply", h.InterviewReply)
		r.Post("/{id}/salary/analyze", h.SalaryAnalyze)
	})
	r.Post("/api/resume/parse", h.ParseResume)

	// Serve SPA if dist exists
	if info, err := os.Stat(cfg.WebDir); err == nil && info.IsDir() {
		fileServer := http.FileServer(http.Dir(cfg.WebDir))
		r.Get("/*", func(w http.ResponseWriter, req *http.Request) {
			path := filepath.Join(cfg.WebDir, filepath.Clean(req.URL.Path))
			if strings.HasPrefix(req.URL.Path, "/api") {
				http.NotFound(w, req)
				return
			}
			if st, err := os.Stat(path); err == nil && !st.IsDir() {
				fileServer.ServeHTTP(w, req)
				return
			}
			http.ServeFile(w, req, filepath.Join(cfg.WebDir, "index.html"))
		})
	}

	return r
}
