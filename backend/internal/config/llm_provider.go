package config

import (
	"os"
	"strings"
)

const envLLMProvider = "LLM_PROVIDER"

func LoadLLMProvider() string {
	provider := strings.ToLower(strings.TrimSpace(os.Getenv(envLLMProvider)))
	if provider == "" {
		return "openai"
	}
	switch provider {
	case "openai", "gemini":
		return provider
	default:
		return "openai"
	}
}
