package spigot

import (
	"testing"
	"time"
)

func TestNewSlidingWindowValidation(t *testing.T) {
	tests := []struct {
		name    string
		limit   int
		window  time.Duration
		wantErr bool
	}{
		{"valid", 10, time.Second, false},
		{"zero limit", 0, time.Second, true},
		{"negative limit", -1, time.Second, true},
		{"zero window", 10, 0, true},
		{"negative window", 10, -time.Second, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewSlidingWindow(tt.limit, tt.window)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewSlidingWindow(%v, %v) error = %v, wantErr %v", tt.limit, tt.window, err, tt.wantErr)
			}
		})
	}
}

func TestSlidingWindowSteadyState(t *testing.T) {
	w, err := NewSlidingWindow(2, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	start := time.Unix(0, 0)
	// One request every 2s stays well under 2/sec across every window.
	for i := 0; i < 10; i++ {
		at := start.Add(time.Duration(i*2) * time.Second)
		if !w.Allow(at) {
			t.Fatalf("request %d at %v: expected admit", i, at)
		}
	}
}

func TestSlidingWindowBurstExceedsLimit(t *testing.T) {
	w, err := NewSlidingWindow(3, 10*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	start := time.Unix(0, 0)
	admitted := 0
	for i := 0; i < 5; i++ {
		if w.Allow(start) {
			admitted++
		}
	}
	if admitted != 3 {
		t.Fatalf("admitted = %d, want 3 (window limit)", admitted)
	}
}

func TestSlidingWindowSmoothsBoundaryBurst(t *testing.T) {
	limit := 10
	window := 10 * time.Second
	w, err := NewSlidingWindow(limit, window)
	if err != nil {
		t.Fatal(err)
	}
	start := time.Unix(0, 0)

	// Anchor the window at t=0, then fill the rest of it right at its
	// tail (t=9s), so the window covers [0s, 10s).
	if !w.Allow(start) {
		t.Fatal("setup request 0: expected admit anchoring the first window")
	}
	for i := 1; i < limit; i++ {
		if !w.Allow(start.Add(9 * time.Second)) {
			t.Fatalf("setup request %d: expected admit filling the first window", i)
		}
	}

	// Immediately after the boundary, the weighted estimate still
	// carries almost the full previous count, so only a couple more
	// requests should get through instead of a fresh burst of `limit`.
	admittedAfterBoundary := 0
	for i := 0; i < limit; i++ {
		if w.Allow(start.Add(10*time.Second + 500*time.Millisecond)) {
			admittedAfterBoundary++
		}
	}
	if admittedAfterBoundary >= limit {
		t.Fatalf("admittedAfterBoundary = %d, want fewer than %d (smoothed, not a fresh burst)", admittedAfterBoundary, limit)
	}
	if admittedAfterBoundary == 0 {
		t.Fatal("admittedAfterBoundary = 0, want at least a couple as the previous window's weight decays")
	}
}

func TestSlidingWindowLoadReflectsUsage(t *testing.T) {
	w, err := NewSlidingWindow(4, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if got := w.Load(); got != 0 {
		t.Fatalf("Load() before any requests = %v, want 0", got)
	}
	start := time.Unix(0, 0)
	w.Allow(start)
	w.Allow(start)
	if got := w.Load(); got != 0.5 {
		t.Fatalf("Load() after 2/4 admitted = %v, want 0.5", got)
	}
}

func TestSlidingWindowAllowN(t *testing.T) {
	tests := []struct {
		name      string
		n         int
		wantAllow bool
		wantCount int
	}{
		{"exactly fits", 3, true, 3},
		{"more than fits", 4, false, 0},
		{"fewer than fits", 2, true, 2},
		{"zero always admitted", 0, true, 0},
		{"negative always admitted", -1, true, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, err := NewSlidingWindow(3, time.Second)
			if err != nil {
				t.Fatal(err)
			}
			start := time.Unix(0, 0)
			if got := w.AllowN(start, tt.n); got != tt.wantAllow {
				t.Fatalf("AllowN(%d) = %v, want %v", tt.n, got, tt.wantAllow)
			}
			if got := w.currCount; got != tt.wantCount {
				t.Fatalf("currCount after AllowN(%d) = %v, want %v", tt.n, got, tt.wantCount)
			}
		})
	}
}

func TestSlidingWindowLongGapResetsWeight(t *testing.T) {
	w, err := NewSlidingWindow(2, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	start := time.Unix(0, 0)
	w.Allow(start)
	w.Allow(start)
	// A long idle gap (many windows) must not leave a stale previous
	// count depressing admission forever.
	later := start.Add(time.Hour)
	if !w.Allow(later) {
		t.Fatal("expected admit after a long idle gap resets the window weight")
	}
}

func BenchmarkSlidingWindowAllow(b *testing.B) {
	w, err := NewSlidingWindow(1e6, time.Hour)
	if err != nil {
		b.Fatal(err)
	}
	start := time.Unix(0, 0)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Allow(start)
	}
}
