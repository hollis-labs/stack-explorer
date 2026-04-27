package semcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/chrispian/stack-explorer/internal/audits"
	"github.com/chrispian/stack-explorer/internal/domain"
	"github.com/chrispian/stack-explorer/internal/jobs"
	"github.com/chrispian/stack-explorer/internal/store/sqlite"
	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

func NewServer(store *sqlite.Store, jobSvc *jobs.Service) *gomcp.Server {
	svc := NewService(store, jobSvc)
	server := gomcp.NewServer(&gomcp.Implementation{
		Name:    "stack-explorer",
		Version: "v0.1.0",
	}, &gomcp.ServerOptions{
		Instructions: "Use these tools to inspect Stack Explorer repo knowledge, prior art, findings, symbols, and audits. Prefer repo_id filters when available.",
	})

	type repoContextInput struct {
		RepoID   string `json:"repo_id,omitempty" jsonschema:"optional repo id"`
		Page     int    `json:"page,omitempty" jsonschema:"pagination page number"`
		PageSize int    `json:"page_size,omitempty" jsonschema:"pagination page size, default 25"`
	}
	gomcp.AddTool(server, &gomcp.Tool{Name: "repo_context", Description: "Summarize the current context for a repo: findings, themes, audits, snapshot, and activity."},
		func(ctx context.Context, req *gomcp.CallToolRequest, in repoContextInput) (*gomcp.CallToolResult, *RepoContextOutput, error) {
			out, err := svc.RepoContext(in.RepoID, defaultPage(in.Page), defaultPageSize(in.PageSize))
			return nil, out, err
		})

	type priorArtForFileInput struct {
		RepoID   string `json:"repo_id,omitempty" jsonschema:"optional repo id"`
		Path     string `json:"path" jsonschema:"repo-relative file path"`
		Page     int    `json:"page,omitempty" jsonschema:"pagination page number"`
		PageSize int    `json:"page_size,omitempty" jsonschema:"pagination page size, default 25"`
	}
	gomcp.AddTool(server, &gomcp.Tool{Name: "prior_art_for_file", Description: "Return findings, themes, code references, and symbols touching a file path."},
		func(ctx context.Context, req *gomcp.CallToolRequest, in priorArtForFileInput) (*gomcp.CallToolResult, *PriorArtForFileOutput, error) {
			out, err := svc.PriorArtForFile(in.RepoID, in.Path, defaultPage(in.Page), defaultPageSize(in.PageSize))
			return nil, out, err
		})

	type priorArtForSymbolInput struct {
		RepoID        string `json:"repo_id,omitempty" jsonschema:"optional repo id"`
		SymbolID      int64  `json:"symbol_id,omitempty" jsonschema:"optional symbol id"`
		QualifiedName string `json:"qualified_name,omitempty" jsonschema:"optional qualified name"`
		Page          int    `json:"page,omitempty" jsonschema:"pagination page number"`
		PageSize      int    `json:"page_size,omitempty" jsonschema:"pagination page size, default 25"`
	}
	gomcp.AddTool(server, &gomcp.Tool{Name: "prior_art_for_symbol", Description: "Return findings, themes, and code references linked to a symbol id or qualified name."},
		func(ctx context.Context, req *gomcp.CallToolRequest, in priorArtForSymbolInput) (*gomcp.CallToolResult, *PriorArtForSymbolOutput, error) {
			out, err := svc.PriorArtForSymbol(in.RepoID, in.QualifiedName, in.SymbolID, defaultPage(in.Page), defaultPageSize(in.PageSize))
			return nil, out, err
		})

	type searchResultsOutput struct {
		Items []any `json:"items"`
	}

	type findingSearchInput struct {
		RepoID string `json:"repo_id,omitempty" jsonschema:"optional repo id"`
		Query  string `json:"query" jsonschema:"freeform search query"`
		Limit  int    `json:"limit,omitempty" jsonschema:"result limit, default 10"`
	}
	gomcp.AddTool(server, &gomcp.Tool{Name: "finding_search", Description: "Search findings using Stack Explorer hybrid retrieval with finding-focused defaults."},
		func(ctx context.Context, req *gomcp.CallToolRequest, in findingSearchInput) (*gomcp.CallToolResult, searchResultsOutput, error) {
			items, err := svc.FindingSearch(in.Query, in.RepoID, defaultSearchLimit(in.Limit))
			if err != nil {
				return nil, searchResultsOutput{}, err
			}
			return nil, searchResultsOutput{Items: toAnySlice(items)}, nil
		})

	type symbolLookupInput struct {
		RepoID           string `json:"repo_id,omitempty" jsonschema:"optional repo id"`
		Query            string `json:"query" jsonschema:"symbol name or qualified name"`
		Kind             string `json:"kind,omitempty" jsonschema:"optional symbol kind"`
		Language         string `json:"language,omitempty" jsonschema:"optional language filter"`
		Limit            int    `json:"limit,omitempty" jsonschema:"result limit, default 10"`
		IncludeNeighbors bool   `json:"include_neighbors,omitempty" jsonschema:"include relationship neighbors in each symbol result"`
		NeighborKind     string `json:"neighbor_kind,omitempty" jsonschema:"optional relationship kind filter for included neighbors"`
		NeighborSource   string `json:"neighbor_source,omitempty" jsonschema:"optional relationship source filter for included neighbors"`
		NeighborDepth    int    `json:"neighbor_depth,omitempty" jsonschema:"neighbor traversal depth, default 1"`
		NeighborLimit    int    `json:"neighbor_limit,omitempty" jsonschema:"neighbor result limit per symbol, default 10"`
	}
	gomcp.AddTool(server, &gomcp.Tool{Name: "symbol_lookup", Description: "Resolve symbol names to symbol metadata and locations."},
		func(ctx context.Context, req *gomcp.CallToolRequest, in symbolLookupInput) (*gomcp.CallToolResult, searchResultsOutput, error) {
			items, err := svc.SymbolLookupWithOptions(ctx, in.RepoID, in.Query, in.Kind, in.Language, defaultSearchLimit(in.Limit), SymbolLookupOptions{
				IncludeNeighbors: in.IncludeNeighbors,
				NeighborKind:     in.NeighborKind,
				NeighborSource:   in.NeighborSource,
				NeighborDepth:    in.NeighborDepth,
				NeighborLimit:    defaultSearchLimit(in.NeighborLimit),
			})
			if err != nil {
				return nil, searchResultsOutput{}, err
			}
			out := make([]any, 0, len(items))
			for _, item := range items {
				out = append(out, item)
			}
			return nil, searchResultsOutput{Items: out}, nil
		})

	type graphNeighborsInput struct {
		RepoID        string `json:"repo_id,omitempty" jsonschema:"optional repo id when resolving by qualified_name"`
		SymbolID      int64  `json:"symbol_id,omitempty" jsonschema:"optional symbol id"`
		QualifiedName string `json:"qualified_name,omitempty" jsonschema:"optional qualified name"`
		Kind          string `json:"kind,omitempty" jsonschema:"optional relationship kind filter"`
		Source        string `json:"source,omitempty" jsonschema:"optional relationship source filter"`
		Depth         int    `json:"depth,omitempty" jsonschema:"traversal depth, default 1"`
		Limit         int    `json:"limit,omitempty" jsonschema:"result limit, default 10"`
	}
	gomcp.AddTool(server, &gomcp.Tool{Name: "graph_neighbors", Description: "Return neighboring symbols and relationship metadata for a symbol."},
		func(ctx context.Context, req *gomcp.CallToolRequest, in graphNeighborsInput) (*gomcp.CallToolResult, searchResultsOutput, error) {
			items, err := svc.GraphNeighbors(ctx, in.RepoID, in.QualifiedName, in.SymbolID, in.Kind, in.Source, in.Depth, defaultSearchLimit(in.Limit))
			if err != nil {
				return nil, searchResultsOutput{}, err
			}
			return nil, searchResultsOutput{Items: toAnySlice(items)}, nil
		})

	type auditShowInput struct {
		AuditID int64 `json:"audit_id" jsonschema:"audit id"`
	}
	gomcp.AddTool(server, &gomcp.Tool{Name: "audit_show", Description: "Return a stored audit with its findings and themes."},
		func(ctx context.Context, req *gomcp.CallToolRequest, in auditShowInput) (*gomcp.CallToolResult, *audits.AuditBundle, error) {
			out, err := svc.AuditShow(in.AuditID)
			return nil, out, err
		})

	type knowledgeQueryInput struct {
		RepoID string `json:"repo_id,omitempty" jsonschema:"optional repo id"`
		Query  string `json:"query" jsonschema:"freeform search query"`
		Kind   string `json:"kind,omitempty" jsonschema:"optional kind filter: finding, symbol, or code-ref"`
		Limit  int    `json:"limit,omitempty" jsonschema:"result limit, default 10"`
	}
	gomcp.AddTool(server, &gomcp.Tool{Name: "knowledge_query", Description: "Run freeform hybrid retrieval across findings, symbols, and code references."},
		func(ctx context.Context, req *gomcp.CallToolRequest, in knowledgeQueryInput) (*gomcp.CallToolResult, searchResultsOutput, error) {
			items, err := svc.KnowledgeQuery(in.Query, in.RepoID, in.Kind, defaultSearchLimit(in.Limit))
			if err != nil {
				return nil, searchResultsOutput{}, err
			}
			return nil, searchResultsOutput{Items: toAnySlice(items)}, nil
		})

	gomcp.AddTool(server, &gomcp.Tool{Name: "finding_add", Description: "Create a finding. Provenance is derived from the MCP caller context or environment."},
		func(ctx context.Context, req *gomcp.CallToolRequest, in FindingCreateInput) (*gomcp.CallToolResult, *FindingSummary, error) {
			out, err := svc.AddFinding(in, provenanceFromRequest(req, "finding_add"))
			return nil, out, err
		})

	gomcp.AddTool(server, &gomcp.Tool{Name: "finding_update_status", Description: "Update the lifecycle status of a finding."},
		func(ctx context.Context, req *gomcp.CallToolRequest, in FindingStatusUpdateInput) (*gomcp.CallToolResult, *FindingSummary, error) {
			out, err := svc.UpdateFindingStatus(in, provenanceFromRequest(req, "finding_update_status"))
			return nil, out, err
		})

	gomcp.AddTool(server, &gomcp.Tool{Name: "audit_import", Description: "Import one or more deep-review audit directories into Stack Explorer."},
		func(ctx context.Context, req *gomcp.CallToolRequest, in AuditImportInput) (*gomcp.CallToolResult, *AuditImportResult, error) {
			out, err := svc.AuditImport(in, provenanceFromRequest(req, "audit_import"))
			return nil, out, err
		})

	gomcp.AddTool(server, &gomcp.Tool{Name: "audit_export", Description: "Export a stored audit to deep-review markdown files."},
		func(ctx context.Context, req *gomcp.CallToolRequest, in AuditExportInput) (*gomcp.CallToolResult, *AuditExportResult, error) {
			out, err := svc.AuditExport(in)
			return nil, out, err
		})

	type eventStreamSubscribeInput struct {
		RepoID         string `json:"repo_id,omitempty" jsonschema:"optional repo id"`
		ScheduleID     string `json:"schedule_id,omitempty" jsonschema:"optional schedule id"`
		JobID          string `json:"job_id,omitempty" jsonschema:"optional job id"`
		JobKind        string `json:"job_kind,omitempty" jsonschema:"optional job kind"`
		Status         string `json:"status,omitempty" jsonschema:"optional status filter"`
		SinceID        int64  `json:"since_id,omitempty" jsonschema:"optional event id cursor"`
		Limit          int    `json:"limit,omitempty" jsonschema:"max events to return, default 25"`
		TimeoutSeconds int    `json:"timeout_seconds,omitempty" jsonschema:"max subscribe duration, default 30"`
	}
	type eventStreamSubscribeOutput struct {
		Items []domain.JobEvent `json:"items"`
	}
	gomcp.AddTool(server, &gomcp.Tool{Name: "event_stream_subscribe", Description: "Subscribe to scheduler job events. Historical backlog is returned first, then live events until the timeout or limit is reached."},
		func(ctx context.Context, req *gomcp.CallToolRequest, in eventStreamSubscribeInput) (*gomcp.CallToolResult, eventStreamSubscribeOutput, error) {
			limit := in.Limit
			if limit <= 0 {
				limit = 25
			}
			timeout := 30 * time.Second
			if in.TimeoutSeconds > 0 {
				timeout = time.Duration(in.TimeoutSeconds) * time.Second
			}
			filter := sqlite.EventFilter{
				RepoID:     in.RepoID,
				ScheduleID: in.ScheduleID,
				JobID:      in.JobID,
				JobKind:    in.JobKind,
				Status:     in.Status,
				SinceID:    in.SinceID,
				Limit:      limit,
			}
			progressToken := req.Params.GetProgressToken()
			items, err := svc.EventStreamSubscribe(ctx, filter, limit, timeout, func(event domain.JobEvent) {
				if req.Session == nil {
					return
				}
				data, _ := json.Marshal(event)
				_ = req.Session.NotifyProgress(ctx, &gomcp.ProgressNotificationParams{
					ProgressToken: progressToken,
					Progress:      float64(event.ID),
					Message:       string(data),
				})
			})
			if err != nil {
				return nil, eventStreamSubscribeOutput{}, err
			}
			return nil, eventStreamSubscribeOutput{Items: items}, nil
		})

	return server
}

func ServeStdio(ctx context.Context, store *sqlite.Store, jobSvc *jobs.Service) error {
	return NewServer(store, jobSvc).Run(ctx, &gomcp.StdioTransport{})
}

func NewHTTPHandler(store *sqlite.Store, jobSvc *jobs.Service) http.Handler {
	server := NewServer(store, jobSvc)
	return gomcp.NewStreamableHTTPHandler(func(*http.Request) *gomcp.Server {
		return server
	}, &gomcp.StreamableHTTPOptions{
		SessionTimeout: 10 * time.Minute,
	})
}

func ListenAndServeHTTP(ctx context.Context, store *sqlite.Store, jobSvc *jobs.Service, host string, port int, path string) error {
	if path == "" {
		path = "/mcp"
	}
	mux := http.NewServeMux()
	mux.Handle(path, NewHTTPHandler(store, jobSvc))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	srv := &http.Server{
		Addr:              fmt.Sprintf("%s:%d", host, port),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()
	err := srv.ListenAndServe()
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

func defaultPage(page int) int {
	if page <= 0 {
		return 1
	}
	return page
}

func defaultPageSize(pageSize int) int {
	if pageSize <= 0 {
		return 25
	}
	return pageSize
}

func defaultSearchLimit(limit int) int {
	if limit <= 0 {
		return 10
	}
	return limit
}

func toAnySlice[T any](items []T) []any {
	out := make([]any, 0, len(items))
	for _, item := range items {
		out = append(out, item)
	}
	return out
}
