// Package middleware wraps net/http handlers with a rate limiter.
package middleware

import (
	"net"
	"net/http"
	"strconv"

	"github.com/mugishajc/go-rate-limiter/ratelimit"
)

// KeyFunc picks the rate-limit key for a request.
type KeyFunc func(*http.Request) string

// ByRemoteIP keys on the client IP, stripping the port.
func ByRemoteIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// Limit wraps next and rejects requests whose key is over the limit with 429.
func Limit(b *ratelimit.TokenBucket, key KeyFunc, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !b.Allow(key(r)) {
			w.Header().Set("Retry-After", strconv.Itoa(1))
			http.Error(w, "too many requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
