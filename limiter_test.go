package spigot

import (
	"testing"
	"time"
)

// alwaysAllow is a minimal Limiter used to confirm the interface shape
// compiles and is satisfiable before any real algorithm lands.
type alwaysAllow struct{}

func (alwaysAllow) Allow(t time.Time) bool { return true }

func TestLimiterInterfaceIsSatisfiable(t *testing.T) {
	var l Limiter = alwaysAllow{}

	if !l.Allow(time.Now()) {
		t.Fatal("expected alwaysAllow.Allow to return true")
	}
}
