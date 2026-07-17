package spigot

import "time"

// Limiter is the common interface implemented by every rate-limiting
// algorithm in this package.
type Limiter interface {
	// Allow reports whether a single request arriving at t should be
	// permitted. Implementations must be safe for concurrent use.
	Allow(t time.Time) bool
}
