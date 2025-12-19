package config

import (
	"os"
	"testing"
)

func TestLoadLLMProvider(t *testing.T) {
	t.Setenv(envLLMProvider, "openai")
	if got := LoadLLMProvider(); got != "openai" {
		t.Fatalf("expected openai, got %s", got)
	}

	t.Setenv(envLLMProvider, "gemini")
	if got := LoadLLMProvider(); got != "gemini" {
		t.Fatalf("expected gemini, got %s", got)
	}

	t.Setenv(envLLMProvider, "unknown")
	if got := LoadLLMProvider(); got != "openai" {
		t.Fatalf("default fallback failed, got %s", got)
	}

	os.Unsetenv(envLLMProvider)
	if got := LoadLLMProvider(); got != "openai" {
		t.Fatalf("expected openai when unset, got %s", got)
	}
}
