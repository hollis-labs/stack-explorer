package api

import (
	"net/http"
)

func (s *Server) listJobs(w http.ResponseWriter, r *http.Request) {
	repoID := r.URL.Query().Get("repo")
	status := r.URL.Query().Get("status")
	kind := r.URL.Query().Get("kind")
	limit := parsePositiveInt(r.URL.Query().Get("limit"), 50)
	items, err := s.store.ListJobs(repoID, status, kind, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeItem(w, items)
}
