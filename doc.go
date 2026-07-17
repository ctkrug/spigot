// Package spigot implements dependency-free rate-limiting algorithms:
// token bucket, leaky bucket, sliding window, and fixed window.
//
// Every algorithm implements the same Limiter interface so callers can
// swap one for another without touching call sites. Time is passed in
// explicitly rather than read from time.Now() internally, which keeps
// every limiter deterministic and trivially testable.
package spigot
