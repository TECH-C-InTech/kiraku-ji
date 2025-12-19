package config

import "testing"

func TestLoadGeminiConfigFromEnv_DefaultModel(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "test-key")
	t.Setenv("GEMINI_MODEL", "")

	cfg, err := LoadGeminiConfigFromEnv()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.APIKey != "test-key" {
		t.Fatalf("unexpected api key: %s", cfg.APIKey)
	}
	if cfg.Model != DefaultGeminiModel {
		t.Fatalf("expected default model, got %s", cfg.Model)
	}
}

func TestLoadGeminiConfigFromEnv_CustomModel(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "another-key")
	t.Setenv("GEMINI_MODEL", "gemini-custom")

	cfg, err := LoadGeminiConfigFromEnv()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.Model != "gemini-custom" {
		t.Fatalf("unexpected model: %s", cfg.Model)
	}
}

func TestLoadGeminiConfigFromEnv_MissingKey(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "")

	if _, err := LoadGeminiConfigFromEnv(); err == nil {
		t.Fatal("expected error when api key is missing")
	}
}
