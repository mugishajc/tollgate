package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mugishajc/go-rate-limiter/ratelimit"
)

func TestLimitBlocksAfterBurst(t *testing.T) {
	b := ratelimit.New(2, 0.0001) // 2 burst, slow refill

	handler := Limit(b, ByRemoteIP, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(200)
	}))

	call := func() int {
		r := httptest.NewRequest("GET", "/x", nil)
		r.RemoteAddr = "10.0.0.1:1234"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		return w.Code
	}

	if call() != 200 || call() != 200 {
		t.Fatal("first two calls should pass")
	}
	if call() != 429 {
		t.Fatal("third call should be rejected with 429")
	}
}

func TestLimitIsPerIP(t *testing.T) {
	b := ratelimit.New(1, 0.0001)
	h := Limit(b, ByRemoteIP, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))

	for _, ip := range []string{"1.1.1.1:1", "2.2.2.2:1"} {
		r := httptest.NewRequest("GET", "/x", nil)
		r.RemoteAddr = ip
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != 200 {
			t.Fatalf("new IP %s should pass, got %d", ip, w.Code)
		}
	}
}
