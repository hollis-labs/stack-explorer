package domain

import "time"

type Repo struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	URL              string    `json:"url"`
	Description      string    `json:"description"`
	Stack            string    `json:"stack"`
	Category         string    `json:"category"`
	IsOwn            bool      `json:"is_own"`
	LocalPath        string    `json:"local_path"`
	Homepage         string    `json:"homepage"`
	License          string    `json:"license"`
	Notes            string    `json:"notes"`
	EmbeddingProfile string    `json:"embedding_profile"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type Tag struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Snapshot struct {
	ID            int64      `json:"id"`
	RepoID        string     `json:"repo_id"`
	CapturedAt    time.Time  `json:"captured_at"`
	Stars         *int       `json:"stars,omitempty"`
	Forks         *int       `json:"forks,omitempty"`
	OpenIssues    *int       `json:"open_issues,omitempty"`
	Contributors  *int       `json:"contributors,omitempty"`
	LastCommitAt  *time.Time `json:"last_commit_at,omitempty"`
	Commits30d    *int       `json:"commits_30d,omitempty"`
	LoC           *int       `json:"loc,omitempty"`
	Files         *int       `json:"files,omitempty"`
	TestFiles     *int       `json:"test_files,omitempty"`
	Dependencies  *int       `json:"dependencies,omitempty"`
	ComplexityAvg *float64   `json:"complexity_avg,omitempty"`
	RawJSON       string     `json:"raw_json"`
}
