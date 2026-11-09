// Package ratelimit implements request rate limiters.
//
// The token bucket limiter allows short bursts up to a configured capacity and
// refills at a steady rate per second. It is safe for concurrent use.
package ratelimit

import (
	"sync"
	"time"
)

// Clock lets tests swap in a fake clock.
type Clock interface {
	Now() time.Time
}

type realClock struct{}

func (realClock) Now() time.Time { return time.Now() }

// TokenBucket is a per-key token bucket limiter.
type TokenBucket struct {
	mu       sync.Mutex
	capacity float64
	refill   float64 // tokens per second
	clock    Clock
	state    map[string]*bucketState
}

type bucketState struct {
	tokens float64
	last   time.Time
}

// New returns a TokenBucket with the given capacity (max burst) and refill rate
// in tokens per second. Capacity and refill must be positive.
func New(capacity, refillPerSecond float64) *TokenBucket {
	return newWithClock(capacity, refillPerSecond, realClock{})
}

func newWithClock(capacity, refill float64, clock Clock) *TokenBucket {
	if capacity <= 0 || refill <= 0 {
		panic("ratelimit: capacity and refill must be positive")
	}
	return &TokenBucket{
		capacity: capacity,
		refill:   refill,
		clock:    clock,
		state:    make(map[string]*bucketState),
	}
}

// Allow returns true if a single request for key is allowed right now, and
// deducts one token. The call is cheap and lock-local.
func (b *TokenBucket) Allow(key string) bool {
	return b.AllowN(key, 1)
}

// AllowN returns true if n tokens are available for key, and deducts them.
func (b *TokenBucket) AllowN(key string, n float64) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := b.clock.Now()
	s, ok := b.state[key]
	if !ok {
		s = &bucketState{tokens: b.capacity, last: now}
		b.state[key] = s
	}

	// Refill since the last call, capped at capacity.
	elapsed := now.Sub(s.last).Seconds()
	if elapsed > 0 {
		s.tokens += elapsed * b.refill
		if s.tokens > b.capacity {
			s.tokens = b.capacity
		}
		s.last = now
	}

	if s.tokens >= n {
		s.tokens -= n
		return true
	}
	return false
}

// Reset removes any accumulated state for key.
func (b *TokenBucket) Reset(key string) {
	b.mu.Lock()
	delete(b.state, key)
	b.mu.Unlock()
}
