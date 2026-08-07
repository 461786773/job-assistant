package config

import (
	"os"
	"path/filepath"
)

type Config struct {
	Addr      string
	DataDir   string
	UploadDir string
	WebDir    string

	// LLM relay (OpenAI-compatible)
	LLMBaseURL string
	LLMAPIKey  string
	LLMModel   string
}

func Load() Config {
	root := env("JA_ROOT", filepath.Join(".", "..", ".."))
	data := env("JA_DATA_DIR", filepath.Join(root, "data"))
	return Config{
		Addr:       env("JA_ADDR", ":8080"),
		DataDir:    data,
		UploadDir:  env("JA_UPLOAD_DIR", filepath.Join(data, "uploads")),
		WebDir:     env("JA_WEB_DIR", filepath.Join(root, "apps", "web", "dist")),
		LLMBaseURL: env("JA_LLM_BASE_URL", ""),
		LLMAPIKey:  env("JA_LLM_API_KEY", ""),
		LLMModel:   env("JA_LLM_MODEL", "gpt-4o-mini"),
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
