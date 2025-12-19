package config

import (
	"fmt"
	"os"
	"strings"
)

const (
	DefaultGeminiModel = "gemini-2.5-flash"

	envGeminiAPIKey = "GEMINI_API_KEY"
	envGeminiModel  = "GEMINI_MODEL"
)

type GeminiConfig struct {
	APIKey string
	Model  string
}

/**
 * 環境変数から読み込んでGemini連携に使用
 */

func LoadGeminiConfigFromEnv() (*GeminiConfig, error) {
	key := strings.TrimSpace(os.Getenv(envGeminiAPIKey))
	if key == "" {
		return nil, fmt.Errorf("config: %s is not set", envGeminiAPIKey)
	}

	model := strings.TrimSpace(os.Getenv(envGeminiModel))
	if model == "" {
		model = DefaultGeminiModel
	}

	return &GeminiConfig{
		APIKey: key,
		Model:  model,
	}, nil
}
