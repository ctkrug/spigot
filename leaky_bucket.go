package spigot

import (
	"fmt"
	"sync"
	"time"
)

// LeakyBucket models a fixed-capacity queue that drains ("leaks") at a
// constant rate. A request is admitted (queued) as long as the queue
// isn't full; once full, arrivals are rejected until the queue leaks
// room, which enforces a strictly constant output rate regardless of how
// bursty the input is.
type LeakyBucket struct {
	capacity float64
	leakRate float64 // queue units drained per second

	mu    sync.Mutex
	level float64
	last  time.Time
}

// NewLeakyBucket creates a leaky bucket with the given capacity (maximum
// queued requests) and leakRate (requests drained per second). Both must
// be positive; the bucket starts empty.
func NewLeakyBucket(capacity, leakRate float64) (*LeakyBucket, error) {
	if !(capacity > 0) {
		return nil, fmt.Errorf("spigot: leaky bucket capacity must be positive, got %v", capacity)
	}
	if !(leakRate > 0) {
		return nil, fmt.Errorf("spigot: leaky bucket leak rate must be positive, got %v", leakRate)
	}
	return &LeakyBucket{capacity: capacity, leakRate: leakRate}, nil
}

// Allow reports whether a request arriving at t is admitted, leaking the
// queue for elapsed time before checking for room.
func (b *LeakyBucket) Allow(t time.Time) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.leak(t)
	if b.level >= b.capacity {
		return false
	}
	b.level++
	return true
}

// AllowN reports whether n requests arriving at t all fit in the queue
// at once. If there isn't room for all n, none are queued.
func (b *LeakyBucket) AllowN(t time.Time, n int) bool {
	if n <= 0 {
		return true
	}
	b.mu.Lock()
	defer b.mu.Unlock()

	b.leak(t)
	if b.level+float64(n) > b.capacity {
		return false
	}
	b.level += float64(n)
	return true
}

// Load reports how full the queue is: 0 when empty, 1 when at capacity.
func (b *LeakyBucket) Load() float64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.level / b.capacity
}

func (b *LeakyBucket) leak(t time.Time) {
	if b.last.IsZero() {
		b.last = t
		return
	}
	elapsed := t.Sub(b.last).Seconds()
	if elapsed <= 0 {
		return
	}
	b.level -= elapsed * b.leakRate
	if b.level < 0 {
		b.level = 0
	}
	b.last = t
}
