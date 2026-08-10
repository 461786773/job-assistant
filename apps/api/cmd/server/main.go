package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/zhangyongjie/job-assistant/internal/auth"
	"github.com/zhangyongjie/job-assistant/internal/config"
	"github.com/zhangyongjie/job-assistant/internal/db"
	"github.com/zhangyongjie/job-assistant/internal/handler"
	"github.com/zhangyongjie/job-assistant/internal/llm"
	"github.com/zhangyongjie/job-assistant/internal/server"
)

func main() {
	cfg := config.Load()

	if err := os.MkdirAll(cfg.DataDir, 0o755); err != nil {
		log.Fatalf("create data dir: %v", err)
	}
	if err := os.MkdirAll(cfg.UploadDir, 0o755); err != nil {
		log.Fatalf("create upload dir: %v", err)
	}

	store, err := db.Open(filepath.Join(cfg.DataDir, "tasks.db"))
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer store.Close()

	secret, err := auth.LoadOrCreateSecret(cfg.JWTSecret, filepath.Join(cfg.DataDir, ".jwt_secret"))
	if err != nil {
		log.Fatalf("jwt secret: %v", err)
	}
	tokens := auth.NewTokenManager(secret, 30*24*time.Hour)

	llmClient := llm.New(cfg.LLMBaseURL, cfg.LLMAPIKey, cfg.LLMModel)
	h := handler.New(store, cfg, llmClient, tokens)
	r := server.NewRouter(h, cfg, tokens)

	log.Printf("求职助手 API listening on %s", cfg.Addr)
	log.Printf("data=%s uploads=%s llm=%v", cfg.DataDir, cfg.UploadDir, llmClient.Enabled())
	if err := http.ListenAndServe(cfg.Addr, r); err != nil {
		log.Fatal(err)
	}
}
