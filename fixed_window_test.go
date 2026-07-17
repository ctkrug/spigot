package spigot

import (
	"testing"
	"time"
)

func TestNewFixedWindowValidation(t *testing.T) {
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
			_, err := NewFixedWindow(tt.limit, tt.window)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewFixedWindow(%v, %v) error = %v, wantErr %v", tt.limit, tt.window, err, tt.wantErr)
			}
		})
	}
}

func TestFixedWindowSteadyState(t *testing.T) {
	w, err := NewFixedWindow(2, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	start := time.Unix(0, 0)
	for i := 0; i < 10; i++ {
		at := start.Add(time.Duration(i*2) * time.Second)
		if !w.Allow(at) {
			t.Fatalf("request %d at %v: expected admit", i, at)
		}
	}
}

func TestFixedWindowBurstExceedsLimit(t *testing.T) {
	w, err := NewFixedWindow(3, 10*time.Second)
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

func TestFixedWindowAdmitsDoubleBurstAtBoundary(t *testing.T) {
	limit := 10
	window := 10 * time.Second
	w, err := NewFixedWindow(limit, window)
	if err != nil {
		t.Fatal(err)
	}
	start := time.Unix(0, 0)

	// Anchor the window at t=0, then fill the rest of it right at its
	// tail (t=9s), so the window covers [0s, 10s).
	admittedBefore := 0
	if w.Allow(start) {
		admittedBefore++
	}
	for i := 1; i < limit; i++ {
		if w.Allow(start.Add(9 * time.Second)) {
			admittedBefore++
		}
	}

	// The instant the boundary passes, the counter resets to zero, so
	// a full second burst is admitted again even though barely any
	// time has passed since the first burst. This is the flaw the
	// sliding window is designed to avoid.
	admittedAfter := 0
	for i := 0; i < limit; i++ {
		if w.Allow(start.Add(10 * time.Second)) {
			admittedAfter++
		}
	}

	if admittedBefore != limit || admittedAfter != limit {
		t.Fatalf("admittedBefore=%d admittedAfter=%d, want both = %d (double burst at the boundary)", admittedBefore, admittedAfter, limit)
	}
}

func TestFixedWindowLoadReflectsUsage(t *testing.T) {
	w, err := NewFixedWindow(4, time.Second)
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

func TestFixedWindowAllowN(t *testing.T) {
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
			w, err := NewFixedWindow(3, time.Second)
			if err != nil {
				t.Fatal(err)
			}
			start := time.Unix(0, 0)
			if got := w.AllowN(start, tt.n); got != tt.wantAllow {
				t.Fatalf("AllowN(%d) = %v, want %v", tt.n, got, tt.wantAllow)
			}
			if got := w.count; got != tt.wantCount {
				t.Fatalf("count after AllowN(%d) = %v, want %v", tt.n, got, tt.wantCount)
			}
		})
	}
}

func TestFixedWindowLongGapResetsCount(t *testing.T) {
	w, err := NewFixedWindow(2, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	start := time.Unix(0, 0)
	w.Allow(start)
	w.Allow(start)
	if w.Allow(start) {
		t.Fatal("expected rejection: window already at limit")
	}
	if !w.Allow(start.Add(time.Hour)) {
		t.Fatal("expected admit after a long idle gap resets the window")
	}
}
