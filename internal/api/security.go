package api

import (
	"crypto/subtle"
	"fmt"
	"net"
	"net/http"
	"strings"
)

// DefaultCORSOrigins are the browser origins allowed when none are
// configured: the local Sigil frontend.
var DefaultCORSOrigins = []string{"http://localhost:3334", "http://127.0.0.1:3334"}

// ServeOptions controls where the API listens and who may call it.
type ServeOptions struct {
	// Host is the interface to bind. Empty means 127.0.0.1.
	Host string
	Port int
	// Token, when set, is required as "Authorization: Bearer <token>" on
	// every /api request. It is mandatory when Host is not a loopback
	// address.
	Token string
	// CORSOrigins are the browser origins allowed to call the API. Empty
	// means DefaultCORSOrigins.
	CORSOrigins []string
}

func (o ServeOptions) host() string {
	if o.Host == "" {
		return "127.0.0.1"
	}
	return o.Host
}

func (o ServeOptions) addr() string {
	return net.JoinHostPort(o.host(), fmt.Sprintf("%d", o.Port))
}

// validate refuses to expose the API beyond loopback without a token.
func (o ServeOptions) validate() error {
	if o.Token == "" && !isLoopbackHost(o.host()) {
		return fmt.Errorf("refusing to listen on %q without a token: set --token or STACK_EXPLORER_API_TOKEN, or bind to 127.0.0.1", o.host())
	}
	return nil
}

func isLoopbackHost(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

// requireToken rejects requests that do not carry the bearer token.
// Preflight requests pass through so CORS can answer them.
func requireToken(token string) func(http.Handler) http.Handler {
	want := []byte("Bearer " + token)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodOptions &&
				subtle.ConstantTimeCompare([]byte(r.Header.Get("Authorization")), want) != 1 {
				w.Header().Set("WWW-Authenticate", `Bearer realm="stack-explorer"`)
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
