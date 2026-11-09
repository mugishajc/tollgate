// Demo HTTP server showing the limiter middleware in action.
//
//	go run ./cmd/demo
//	curl -i localhost:8080/ping   # first few succeed, then 429
package main

import (
	"log"
	"net/http"

	"github.com/mugishajc/go-rate-limiter/middleware"
	"github.com/mugishajc/go-rate-limiter/ratelimit"
)

func main() {
	limiter := ratelimit.New(5, 1) // 5 burst, 1 req/sec sustained
	mux := http.NewServeMux()
	mux.HandleFunc("/ping", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("pong\n"))
	})

	handler := middleware.Limit(limiter, middleware.ByRemoteIP, mux)
	log.Println("listening on :8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatal(err)
	}
}
