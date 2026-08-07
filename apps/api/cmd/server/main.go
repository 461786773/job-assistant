package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

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

	store, err := db.Open(filepath.Join(cfg.DataDir, "tasks.json"))
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer store.Close()

	if err := store.Migrate(); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	llmClient := llm.New(cfg.LLMBaseURL, cfg.LLMAPIKey, cfg.LLMModel)
	h := handler.New(store, cfg, llmClient)
	r := server.NewRouter(h, cfg)

	log.Printf("求职助手 API listening on %s", cfg.Addr)
	log.Printf("data=%s uploads=%s llm=%v", cfg.DataDir, cfg.UploadDir, llmClient.Enabled())
	if err := http.ListenAndServe(cfg.Addr, r); err != nil {
		log.Fatal(err)
	}
}
