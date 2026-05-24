package api

import (
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestFetchGitHubStatsUsesGitHubToken(t *testing.T) {
	previousToken := os.Getenv("GITHUB_TOKEN")
	previousTransport := http.DefaultTransport
	t.Cleanup(func() {
		_ = os.Setenv("GITHUB_TOKEN", previousToken)
		http.DefaultTransport = previousTransport
	})
	if err := os.Setenv("GITHUB_TOKEN", "test-token"); err != nil {
		t.Fatalf("set env: %v", err)
	}

	var gotAuth, gotVersion string
	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		gotAuth = req.Header.Get("Authorization")
		gotVersion = req.Header.Get("X-GitHub-Api-Version")
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body: io.NopCloser(strings.NewReader(`{
				"stargazers_count": 12,
				"forks_count": 3,
				"open_issues_count": 4
			}`)),
		}, nil
	})

	stats, err := fetchGitHubStats("https://github.com/openai/openai-go")
	if err != nil {
		t.Fatalf("fetch github stats: %v", err)
	}
	if gotAuth != "Bearer test-token" {
		t.Fatalf("expected bearer token, got %q", gotAuth)
	}
	if gotVersion != "2022-11-28" {
		t.Fatalf("expected github api version header, got %q", gotVersion)
	}
	if stats.Stars == nil || *stats.Stars != 12 {
		t.Fatalf("unexpected stats: %+v", stats)
	}
}
