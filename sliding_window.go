package spigot

import (
	"fmt"
	"sync"
	"time"
)

// SlidingWindow approximates a sliding request log using two adjacent
// fixed windows: the previous window's count is weighted down by how far
// the current instant has moved past the boundary. That weighting is
// what smooths a burst straddling a window edge, instead of letting a
// fresh window reset the count to zero the moment the boundary passes
// (see FixedWindow).
type SlidingWindow struct {
	limit  int
	window time.Duration

	mu        sync.Mutex
	currStart time.Time
	currCount int
	prevCount int
	now       time.Time
}

// NewSlidingWindow creates a sliding window limiter admitting at most
// limit requests per window duration. Both must be positive.
func NewSlidingWindow(limit int, window time.Duration) (*SlidingWindow, error) {
	if limit <= 0 {
		return nil, fmt.Errorf("spigot: sliding window limit must be positive, got %d", limit)
	}
	if window <= 0 {
		return nil, fmt.Errorf("spigot: sliding window duration must be positive, got %v", window)
	}
	return &SlidingWindow{limit: limit, window: window}, nil
}

// Allow reports whether a request arriving at t is admitted, using the
// weighted estimate of requests in the trailing window.
func (w *SlidingWindow) Allow(t time.Time) bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.advance(t)
	if w.estimate(t) >= float64(w.limit) {
		return false
	}
	w.currCount++
	return true
}

// AllowN reports whether n requests arriving at t are all admitted under
// the weighted estimate at once. If the estimate plus n would exceed the
// limit, none are counted.
func (w *SlidingWindow) AllowN(t time.Time, n int) bool {
	if n <= 0 {
		return true
	}
	w.mu.Lock()
	defer w.mu.Unlock()

	w.advance(t)
	if w.estimate(t)+float64(n) > float64(w.limit) {
		return false
	}
	w.currCount += n
	return true
}

// Load reports the weighted estimate as a fraction of the limit.
func (w *SlidingWindow) Load() float64 {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.estimate(w.now) / float64(w.limit)
}

// advance rolls currStart/currCount/prevCount forward to cover t. If t is
// more than one window past the last-seen instant, both counts collapse
// to zero: there is no meaningful "previous window" after a long gap.
func (w *SlidingWindow) advance(t time.Time) {
	if w.currStart.IsZero() {
		w.currStart = t
	}
	for !t.Before(w.currStart.Add(w.window)) {
		w.prevCount = w.currCount
		w.currCount = 0
		w.currStart = w.currStart.Add(w.window)
	}
	w.now = t
}

// estimate returns the weighted request count covering the window
// trailing t, given the current currStart/currCount/prevCount.
func (w *SlidingWindow) estimate(t time.Time) float64 {
	frac := t.Sub(w.currStart).Seconds() / w.window.Seconds()
	switch {
	case frac > 1:
		// Defensive only: advance(t) always runs before estimate(t) in every
		// call path, and its loop guarantees t stays inside [currStart,
		// currStart+window), i.e. frac < 1. Kept in case that invariant is
		// ever broken by a future caller.
		frac = 1
	case frac < 0:
		frac = 0
	}
	return float64(w.prevCount)*(1-frac) + float64(w.currCount)
}
