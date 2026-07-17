package spigot

import (
	"fmt"
	"sync"
	"time"
)

// FixedWindow admits up to limit requests within each fixed-size window,
// then resets the count entirely at the window boundary. It's the
// simplest of the four algorithms and the one most prone to a boundary
// burst: up to 2x limit requests can land in a short span if a burst
// straddles the reset instant (limit at the tail of one window, limit
// again at the head of the next).
type FixedWindow struct {
	limit  int
	window time.Duration

	mu          sync.Mutex
	windowStart time.Time
	count       int
}

// NewFixedWindow creates a fixed window limiter admitting at most limit
// requests per window duration. Both must be positive.
func NewFixedWindow(limit int, window time.Duration) (*FixedWindow, error) {
	if limit <= 0 {
		return nil, fmt.Errorf("spigot: fixed window limit must be positive, got %d", limit)
	}
	if window <= 0 {
		return nil, fmt.Errorf("spigot: fixed window duration must be positive, got %v", window)
	}
	return &FixedWindow{limit: limit, window: window}, nil
}

// Allow reports whether a request arriving at t is admitted under the
// current window's count, resetting the window first if t has moved
// past it.
func (w *FixedWindow) Allow(t time.Time) bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.advance(t)
	if w.count >= w.limit {
		return false
	}
	w.count++
	return true
}

// AllowN reports whether n requests arriving at t all fit within the
// current window's remaining count at once. If they don't, none count.
func (w *FixedWindow) AllowN(t time.Time, n int) bool {
	if n <= 0 {
		return true
	}
	w.mu.Lock()
	defer w.mu.Unlock()

	w.advance(t)
	if w.count+n > w.limit {
		return false
	}
	w.count += n
	return true
}

// Load reports the current window's count as a fraction of the limit.
func (w *FixedWindow) Load() float64 {
	w.mu.Lock()
	defer w.mu.Unlock()
	return float64(w.count) / float64(w.limit)
}

func (w *FixedWindow) advance(t time.Time) {
	if w.windowStart.IsZero() {
		w.windowStart = t
		return
	}
	for !t.Before(w.windowStart.Add(w.window)) {
		w.windowStart = w.windowStart.Add(w.window)
		w.count = 0
	}
}
