package spigot

import (
	"fmt"
	"sync"
	"time"
)

// TokenBucket allows bursts up to its capacity, then admits requests only
// as fast as tokens are refilled. It starts full, so the first burst up
// to capacity is always admitted immediately.
type TokenBucket struct {
	capacity   float64
	refillRate float64 // tokens added per second

	mu     sync.Mutex
	tokens float64
	last   time.Time
}

// NewTokenBucket creates a token bucket with the given capacity (maximum
// burst size) and refillRate (tokens added per second). Both must be
// positive; the bucket starts full.
func NewTokenBucket(capacity, refillRate float64) (*TokenBucket, error) {
	if !(capacity > 0) {
		return nil, fmt.Errorf("spigot: token bucket capacity must be positive, got %v", capacity)
	}
	if !(refillRate > 0) {
		return nil, fmt.Errorf("spigot: token bucket refill rate must be positive, got %v", refillRate)
	}
	return &TokenBucket{
		capacity:   capacity,
		refillRate: refillRate,
		tokens:     capacity,
	}, nil
}

// Allow reports whether a request arriving at t is admitted, refilling
// tokens for elapsed time before checking availability.
func (b *TokenBucket) Allow(t time.Time) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.refill(t)
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

// AllowN reports whether n requests arriving at t are all admitted at
// once. If fewer than n tokens are available, none are consumed.
func (b *TokenBucket) AllowN(t time.Time, n int) bool {
	if n <= 0 {
		return true
	}
	b.mu.Lock()
	defer b.mu.Unlock()

	b.refill(t)
	if b.tokens < float64(n) {
		return false
	}
	b.tokens -= float64(n)
	return true
}

// Load reports how drained the bucket is: 0 when full, 1 when empty.
func (b *TokenBucket) Load() float64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	return 1 - b.tokens/b.capacity
}

func (b *TokenBucket) refill(t time.Time) {
	if b.last.IsZero() {
		b.last = t
		return
	}
	elapsed := t.Sub(b.last).Seconds()
	if elapsed <= 0 {
		return
	}
	b.tokens += elapsed * b.refillRate
	if b.tokens > b.capacity {
		b.tokens = b.capacity
	}
	b.last = t
}
