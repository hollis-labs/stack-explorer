package semcp

import (
	"context"
	"net/http"
)

// provenanceHeaderNames are the caller-attributed provenance headers an HTTP
// MCP client may set, mirrored into request context so provenanceFromContext
// can read them from inside a tool handler.
var provenanceHeaderNames = []string{
	"X-Stack-Explorer-Actor-Kind",
	"X-Stack-Explorer-Actor-Id",
	"X-Stack-Explorer-Session-Id",
	"X-Stack-Explorer-Model",
}

type provenanceHeadersContextKey struct{}

// provenanceHeaderMiddleware captures the provenance headers off the
// incoming HTTP request, ahead of go-mcp's stateless protocol handler, and
// threads them through request context. go-mcp's ToolHandler signature is
// (ctx, args map[string]any) only -- a tool handler has no access to the
// underlying *http.Request -- so this is the only place left to read them.
// The values ride on the *http.Request's own context, which the official
// SDK's Streamable HTTP transport uses as the base context for the RPC call
// it dispatches, so they reach the handler unchanged.
func provenanceHeaderMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headers := make(map[string]string, len(provenanceHeaderNames))
		found := false
		for _, name := range provenanceHeaderNames {
			if v := r.Header.Get(name); v != "" {
				headers[name] = v
				found = true
			}
		}
		if found {
			r = r.WithContext(context.WithValue(r.Context(), provenanceHeadersContextKey{}, headers))
		}
		next.ServeHTTP(w, r)
	})
}

func provenanceHeadersFromContext(ctx context.Context) map[string]string {
	headers, _ := ctx.Value(provenanceHeadersContextKey{}).(map[string]string)
	return headers
}
