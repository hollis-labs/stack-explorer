package embed

import "testing"

func TestNewManagerDefaultsToOpenAIProfiles(t *testing.T) {
	t.Setenv("SE_EMBED_SMALL_PROVIDER", "")
	t.Setenv("SE_EMBED_SMALL_MODEL", "")
	t.Setenv("SE_EMBED_MEDIUM_PROVIDER", "")
	t.Setenv("SE_EMBED_MEDIUM_MODEL", "")
	t.Setenv("SE_EMBED_FULL_PROVIDER", "")
	t.Setenv("SE_EMBED_FULL_MODEL", "")

	m := NewManager()

	small, err := m.Profile("small")
	if err != nil {
		t.Fatalf("small profile: %v", err)
	}
	if small.Provider != "openai" || small.Model != "text-embedding-3-small" {
		t.Fatalf("small profile = %#v", small)
	}

	medium, err := m.Profile("medium")
	if err != nil {
		t.Fatalf("medium profile: %v", err)
	}
	if medium.Provider != "openai" || medium.Model != "text-embedding-3-large" {
		t.Fatalf("medium profile = %#v", medium)
	}

	full, err := m.Profile("full")
	if err != nil {
		t.Fatalf("full profile: %v", err)
	}
	if full.Provider != "openai" || full.Model != "text-embedding-3-large" {
		t.Fatalf("full profile = %#v", full)
	}
}

func TestNewProviderRejectsNonOpenAIProviders(t *testing.T) {
	if _, err := newProvider("ollama"); err == nil {
		t.Fatal("expected unsupported provider error")
	}
}
