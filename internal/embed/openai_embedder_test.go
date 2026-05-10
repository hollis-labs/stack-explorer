package embed

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/openai/openai-go/option"
)

func TestOpenAIEmbedderEmbed(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/embeddings" {
			gotPath = r.URL.Path
			http.Error(w, "unexpected path", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{"embedding": []float64{1.25, 2.5}},
			},
			"usage": map[string]any{"total_tokens": 7},
		})
	}))
	defer srv.Close()

	e := NewOpenAIEmbedder("test-key", srv.Client(), option.WithBaseURL(srv.URL+"/"))
	res, err := e.Embed(context.Background(), "hello", "text-embedding-3-small")
	if err != nil {
		t.Fatalf("Embed: %v", err)
	}
	if gotPath != "" {
		t.Fatalf("path = %s", gotPath)
	}
	if got := len(res.Embedding); got != 2 {
		t.Fatalf("embedding len = %d", got)
	}
	if res.TokenCount != 7 {
		t.Fatalf("token count = %d", res.TokenCount)
	}
}

func TestOpenAIEmbedderEmbedBatch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{"embedding": []float64{1, 2}},
				{"embedding": []float64{3, 4}},
			},
			"usage": map[string]any{"total_tokens": 8},
		})
	}))
	defer srv.Close()

	e := NewOpenAIEmbedder("test-key", srv.Client(), option.WithBaseURL(srv.URL+"/"))
	res, err := e.EmbedBatch(context.Background(), []string{"a", "b"}, "text-embedding-3-small")
	if err != nil {
		t.Fatalf("EmbedBatch: %v", err)
	}
	if len(res) != 2 {
		t.Fatalf("results len = %d", len(res))
	}
	if res[0].TokenCount != 0 || res[1].TokenCount != 0 {
		t.Fatalf("token counts = %#v", res)
	}
}

func TestOpenAIEmbedderEmbedBatchRejectsMismatchedResponseCount(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{"embedding": []float64{1, 2}},
			},
			"usage": map[string]any{"total_tokens": 8},
		})
	}))
	defer srv.Close()

	e := NewOpenAIEmbedder("test-key", srv.Client(), option.WithBaseURL(srv.URL+"/"))
	if _, err := e.EmbedBatch(context.Background(), []string{"a", "b"}, "text-embedding-3-small"); err == nil {
		t.Fatal("expected response count mismatch error")
	}
}

func TestOpenAIEmbedderEmbeddingDimensions(t *testing.T) {
	e := NewOpenAIEmbedder("test-key", nil)
	if got := e.EmbeddingDimensions("text-embedding-3-small"); got != 1536 {
		t.Fatalf("small dimensions = %d", got)
	}
	if got := e.EmbeddingDimensions("text-embedding-3-large"); got != 3072 {
		t.Fatalf("large dimensions = %d", got)
	}
	if got := e.EmbeddingDimensions("unknown"); got != 0 {
		t.Fatalf("unknown dimensions = %d", got)
	}
}
