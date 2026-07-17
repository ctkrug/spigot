package spigot

import "time"

// Limiter is the common interface implemented by every rate-limiting
// algorithm in this package.
type Limiter interface {
	// Allow reports whether a single request arriving at t should be
	// permitted. Implementations must be safe for concurrent use.
	Allow(t time.Time) bool
}

// BulkLimiter is implemented by limiters that can decide on a batch of n
// requests atomically: either all n are admitted, or none are, so a
// rejection never partially consumes capacity.
type BulkLimiter interface {
	// AllowN reports whether n requests arriving at t should all be
	// permitted. n <= 0 is always permitted and consumes no capacity.
	AllowN(t time.Time, n int) bool
}
