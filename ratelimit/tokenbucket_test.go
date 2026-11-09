package ratelimit

import (
	"sync"
	"testing"
	"time"
)

type fakeClock struct {
	mu  sync.Mutex
	now time.Time
}

func (c *fakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *fakeClock) advance(d time.Duration) {
	c.mu.Lock()
	c.now = c.now.Add(d)
	c.mu.Unlock()
}

func TestBurstUpToCapacity(t *testing.T) {
	clk := &fakeClock{now: time.Unix(0, 0)}
	b := newWithClock(5, 1, clk)

	for i := 0; i < 5; i++ {
		if !b.Allow("k") {
			t.Fatalf("burst call %d should be allowed", i+1)
		}
	}
	if b.Allow("k") {
		t.Fatal("6th call should be denied once bucket is empty")
	}
}

func TestRefillOverTime(t *testing.T) {
	clk := &fakeClock{now: time.Unix(0, 0)}
	b := newWithClock(2, 2, clk) // 2 tokens, 2/sec

	if !b.Allow("k") || !b.Allow("k") {
		t.Fatal("first two should pass")
	}
	if b.Allow("k") {
		t.Fatal("third should be denied before refill")
	}
	clk.advance(500 * time.Millisecond) // +1 token
	if !b.Allow("k") {
		t.Fatal("should pass after half a second")
	}
	if b.Allow("k") {
		t.Fatal("second call after a single refill should be denied")
	}
}

func TestPerKeyIsolation(t *testing.T) {
	clk := &fakeClock{now: time.Unix(0, 0)}
	b := newWithClock(1, 1, clk)
	if !b.Allow("a") || !b.Allow("b") {
		t.Fatal("different keys should not share state")
	}
	if b.Allow("a") || b.Allow("b") {
		t.Fatal("same keys should be denied on second call")
	}
}

func TestReset(t *testing.T) {
	clk := &fakeClock{now: time.Unix(0, 0)}
	b := newWithClock(1, 1, clk)
	b.Allow("k")
	b.Reset("k")
	if !b.Allow("k") {
		t.Fatal("after Reset the key should have a full bucket again")
	}
}

func TestConcurrentAllowIsSafe(t *testing.T) {
	clk := &fakeClock{now: time.Unix(0, 0)}
	b := newWithClock(1000, 1, clk)

	var wg sync.WaitGroup
	var allowed int64
	var mu sync.Mutex

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				if b.Allow("shared") {
					mu.Lock()
					allowed++
					mu.Unlock()
				}
			}
		}()
	}
	wg.Wait()

	if allowed != 1000 {
		t.Fatalf("expected exactly capacity (1000) allowances, got %d", allowed)
	}
}
