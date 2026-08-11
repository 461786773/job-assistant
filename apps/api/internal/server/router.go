package server

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/zhangyongjie/job-assistant/internal/auth"
	"github.com/zhangyongjie/job-assistant/internal/config"
	"github.com/zhangyongjie/job-assistant/internal/handler"
)

func NewRouter(h *handler.Handler, cfg config.Config, tokens *auth.TokenManager) http.Handler {
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
	r.Post("/api/auth/register", h.Register)
	r.Post("/api/auth/login", h.Login)

	r.Group(func(r chi.Router) {
		r.Use(bearerAuth(tokens))
		r.Get("/api/auth/me", h.Me)
		r.Post("/api/needs/", h.SetPrimaryNeed)
		r.Route("/api/assessments", func(r chi.Router) {
			r.Get("/", h.ListAssessments)
			r.Post("/", h.CreateAssessment)
			r.Get("/latest", h.LatestAssessment)
			r.Get("/{id}", h.GetAssessment)
		})
		r.Route("/api/bigfive", func(r chi.Router) {
			r.Get("/", h.ListBigFive)
			r.Post("/", h.CreateBigFive)
			r.Get("/latest", h.LatestBigFive)
			r.Get("/{id}", h.GetBigFive)
		})
		r.Route("/api/counseling/bookings", func(r chi.Router) {
			r.Get("/", h.ListBookings)
			r.Post("/", h.CreateBooking)
			r.Patch("/{id}", h.PatchBooking)
		})
		r.Route("/api/tasks", func(r chi.Router) {
			r.Get("/", h.ListTasks)
			r.Post("/", h.CreateTask)
			r.Get("/{id}", h.GetTask)
			r.Patch("/{id}", h.UpdateTask)
			r.Delete("/{id}", h.DeleteTask)
			r.Post("/{id}/resume", h.UploadResume)
			r.Post("/{id}/hr/analyze", h.AnalyzeHR)
			r.Post("/{id}/hr/apply-rewrites", h.ApplyHRRewrites)
			r.Post("/{id}/interview/start", h.InterviewStart)
			r.Post("/{id}/interview/reply", h.InterviewReply)
			r.Post("/{id}/salary/analyze", h.SalaryAnalyze)
		})
		r.Post("/api/resume/parse", h.ParseResume)
		r.Route("/api/coach/sessions", func(r chi.Router) {
			r.Get("/", h.ListCoachSessions)
			r.Post("/", h.CreateCoachSession)
			r.Get("/{id}", h.GetCoachSession)
			r.Post("/{id}/reply", h.ReplyCoachSession)
			r.Delete("/{id}", h.DeleteCoachSession)
		})
		r.Route("/api/wellbeing/checkins", func(r chi.Router) {
			r.Get("/", h.ListCheckIns)
			r.Post("/", h.CreateCheckIn)
			r.Delete("/{id}", h.DeleteCheckIn)
		})
		r.Route("/api/wellbeing/quick-checks", func(r chi.Router) {
			r.Get("/", h.ListQuickSelfChecks)
			r.Post("/", h.CreateQuickSelfCheck)
			r.Get("/{id}", h.GetQuickSelfCheck)
			r.Delete("/{id}", h.DeleteQuickSelfCheck)
		})
	})

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

func bearerAuth(tokens *auth.TokenManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			const prefix = "Bearer "
			if header == "" || !strings.HasPrefix(header, prefix) {
				writeAuthErr(w, http.StatusUnauthorized, "请先登录")
				return
			}
			claims, err := tokens.Verify(strings.TrimSpace(header[len(prefix):]))
			if err != nil {
				writeAuthErr(w, http.StatusUnauthorized, err.Error())
				return
			}
			next.ServeHTTP(w, r.WithContext(auth.WithUser(r.Context(), claims)))
		})
	}
}

func writeAuthErr(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(`{"error":` + strconv.Quote(msg) + `}`))
}
