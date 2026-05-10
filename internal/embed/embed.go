package embed

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	embedcontracts "github.com/hollis-labs/go-embed-contracts"
)

type ProfileConfig struct {
	Name     string
	Provider string
	Model    string
}

type Manager struct {
	profiles map[string]ProfileConfig
	byModel  map[string]ProfileConfig
}

func NewManager() *Manager {
	profiles := map[string]ProfileConfig{
		"small": {
			Name:     "small",
			Provider: envOr("SE_EMBED_SMALL_PROVIDER", "openai"),
			Model:    envOr("SE_EMBED_SMALL_MODEL", "text-embedding-3-small"),
		},
		"medium": {
			Name:     "medium",
			Provider: envOr("SE_EMBED_MEDIUM_PROVIDER", "openai"),
			Model:    envOr("SE_EMBED_MEDIUM_MODEL", "text-embedding-3-large"),
		},
		"full": {
			Name:     "full",
			Provider: envOr("SE_EMBED_FULL_PROVIDER", "openai"),
			Model:    envOr("SE_EMBED_FULL_MODEL", "text-embedding-3-large"),
		},
	}

	byModel := make(map[string]ProfileConfig, len(profiles))
	for _, cfg := range profiles {
		byModel[cfg.Model] = cfg
	}
	return &Manager{profiles: profiles, byModel: byModel}
}

func (m *Manager) Profile(name string) (ProfileConfig, error) {
	if name == "" || name == "none" {
		return ProfileConfig{}, fmt.Errorf("embedding profile %q is disabled", name)
	}
	cfg, ok := m.profiles[name]
	if !ok {
		return ProfileConfig{}, fmt.Errorf("unknown embedding profile %q", name)
	}
	return cfg, nil
}

func (m *Manager) ResolveModel(model string) (ProfileConfig, error) {
	cfg, ok := m.byModel[model]
	if !ok {
		return ProfileConfig{}, fmt.Errorf("no provider mapping configured for model %q", model)
	}
	return cfg, nil
}

func (m *Manager) Embed(ctx context.Context, providerName, model, text string) ([]float32, error) {
	embedder, err := newProvider(providerName)
	if err != nil {
		return nil, err
	}
	res, err := embedder.Embed(ctx, text, model)
	if err != nil {
		return nil, err
	}
	return res.Embedding, nil
}

func (m *Manager) EmbedBatch(ctx context.Context, providerName, model string, texts []string) ([][]float32, error) {
	embedder, err := newProvider(providerName)
	if err != nil {
		return nil, err
	}
	results, err := embedder.EmbedBatch(ctx, texts, model)
	if err != nil {
		return nil, err
	}
	out := make([][]float32, len(results))
	for i := range results {
		out[i] = results[i].Embedding
	}
	return out, nil
}

func HashText(parts ...string) string {
	h := sha256.New()
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			_, _ = h.Write([]byte(trimmed))
			_, _ = h.Write([]byte{'\n'})
		}
	}
	return hex.EncodeToString(h.Sum(nil))
}

func newProvider(name string) (embedcontracts.Embedder, error) {
	switch strings.ToLower(name) {
	case "openai":
		return NewOpenAIEmbedder("", nil), nil
	default:
		return nil, fmt.Errorf("unsupported embedding provider %q: stack-explorer's Path B migration keeps only the OpenAI-backed embedder", name)
	}
}

func envOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}
