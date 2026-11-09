# go-rate-limiter

A small token-bucket rate limiter for Go, plus a `net/http` middleware that keys on the client IP. Safe for concurrent use. No external dependencies.

## Install

```bash
go get github.com/mugishajc/go-rate-limiter
```

## Use

```go
import (
    "github.com/mugishajc/go-rate-limiter/middleware"
    "github.com/mugishajc/go-rate-limiter/ratelimit"
)

limiter := ratelimit.New(5, 1) // 5 token burst, refill 1/sec
http.ListenAndServe(":8080", middleware.Limit(limiter, middleware.ByRemoteIP, mux))
```

The library itself (`ratelimit.TokenBucket`) is independent of HTTP, so it also works for background jobs, queue consumers and gRPC interceptors.

## Run the demo

```bash
go run ./cmd/demo
curl -i localhost:8080/ping
```

The first 5 calls get `200 OK`; the next ones get `429 Too Many Requests` with `Retry-After: 1` until tokens refill.

## Tests

```bash
go test -race ./...
```

Covers:

- burst up to capacity
- refill over time using an injectable clock
- per-key isolation
- 100 goroutines sharing one bucket, exactly `capacity` allowances get through
- HTTP middleware returns 429 after the burst
