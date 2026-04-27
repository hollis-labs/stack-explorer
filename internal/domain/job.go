package domain

import "time"

type Schedule struct {
	ID             string     `json:"id"`
	RepoID         string     `json:"repo_id,omitempty"`
	Name           string     `json:"name"`
	CronExpr       string     `json:"cron_expr"`
	JobKind        string     `json:"job_kind"`
	PayloadJSON    string     `json:"payload_json"`
	Enabled        bool       `json:"enabled"`
	MaxAttempts    int        `json:"max_attempts"`
	RetryBackoff   string     `json:"retry_backoff"`
	RetryDelaySecs int        `json:"retry_delay_secs"`
	LastRunAt      *time.Time `json:"last_run_at,omitempty"`
	NextRunAt      *time.Time `json:"next_run_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type Job struct {
	ID             string     `json:"id"`
	ScheduleID     string     `json:"schedule_id,omitempty"`
	RepoID         string     `json:"repo_id,omitempty"`
	Kind           string     `json:"kind"`
	Status         string     `json:"status"`
	AttemptCount   int        `json:"attempt_count"`
	MaxAttempts    int        `json:"max_attempts"`
	RetryBackoff   string     `json:"retry_backoff"`
	RetryDelaySecs int        `json:"retry_delay_secs"`
	PayloadJSON    string     `json:"payload_json"`
	OutputJSON     string     `json:"output_json"`
	Error          string     `json:"error,omitempty"`
	StartedAt      *time.Time `json:"started_at,omitempty"`
	FinishedAt     *time.Time `json:"finished_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type JobEvent struct {
	ID          int64     `json:"id"`
	JobID       string    `json:"job_id"`
	ScheduleID  string    `json:"schedule_id,omitempty"`
	RepoID      string    `json:"repo_id,omitempty"`
	JobKind     string    `json:"job_kind"`
	EventType   string    `json:"event_type"`
	Status      string    `json:"status,omitempty"`
	Message     string    `json:"message,omitempty"`
	PayloadJSON string    `json:"payload_json"`
	CreatedAt   time.Time `json:"created_at"`
}
