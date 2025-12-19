package config

import "testing"

func TestLoadOpenAIConfigFromEnv(t *testing.T) {
	t.Setenv(envOpenAIAPIKey, "key")
	t.Setenv(envOpenAIModel, "model")
	t.Setenv(envOpenAIBaseURL, "https://example.com")

	cfg, err := LoadOpenAIConfigFromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.APIKey != "key" || cfg.Model != "model" || cfg.BaseURL != "https://example.com" {
		t.Fatalf("unexpected config: %+v", cfg)
	}
}

func TestLoadOpenAIConfigMissingKey(t *testing.T) {
	t.Setenv(envOpenAIAPIKey, "")
	if _, err := LoadOpenAIConfigFromEnv(); err == nil {
		t.Fatalf("expected error when key is missing")
	}
}

func TestLoadOpenAIConfigDefaultModel(t *testing.T) {
	t.Setenv(envOpenAIAPIKey, "key")
	t.Setenv(envOpenAIModel, "")
	cfg, err := LoadOpenAIConfigFromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Model != DefaultOpenAIModel {
		t.Fatalf("expected default model %s, got %s", DefaultOpenAIModel, cfg.Model)
	}
}
