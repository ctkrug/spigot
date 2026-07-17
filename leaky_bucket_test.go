package spigot

import (
	"math"
	"testing"
	"time"
)

func TestNewLeakyBucketValidation(t *testing.T) {
	tests := []struct {
		name     string
		capacity float64
		leakRate float64
		wantErr  bool
	}{
		{"valid", 10, 1, false},
		{"zero capacity", 0, 1, true},
		{"negative capacity", -1, 1, true},
		{"zero leak rate", 10, 0, true},
		{"negative leak rate", 10, -1, true},
		{"NaN capacity", math.NaN(), 1, true},
		{"NaN leak rate", 10, math.NaN(), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewLeakyBucket(tt.capacity, tt.leakRate)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewLeakyBucket(%v, %v) error = %v, wantErr %v", tt.capacity, tt.leakRate, err, tt.wantErr)
			}
		})
	}
}

func TestLeakyBucketSteadyState(t *testing.T) {
	b, err := NewLeakyBucket(5, 1) // drains 1/sec
	if err != nil {
		t.Fatal(err)
	}
	start := time.Unix(0, 0)
	// One request per second matches the leak rate exactly, so the
	// queue never fills and every request is admitted.
	for i := 0; i < 20; i++ {
		at := start.Add(time.Duration(i) * time.Second)
		if !b.Allow(at) {
			t.Fatalf("request %d at %v: expected admit", i, at)
		}
	}
}

func TestLeakyBucketBurstFillsQueue(t *testing.T) {
	b, err := NewLeakyBucket(3, 1)
	if err != nil {
		t.Fatal(err)
	}
	start := time.Unix(0, 0)
	admitted := 0
	for i := 0; i < 5; i++ {
		if b.Allow(start) {
			admitted++
		}
	}
	if admitted != 3 {
		t.Fatalf("admitted = %d, want 3 (queue capacity)", admitted)
	}
}

func TestLeakyBucketDrainsOverTime(t *testing.T) {
	b, err := NewLeakyBucket(1, 1) // drains 1/sec
	if err != nil {
		t.Fatal(err)
	}
	start := time.Unix(0, 0)
	if !b.Allow(start) {
		t.Fatal("expected first request admitted into an empty queue")
	}
	if b.Allow(start) {
		t.Fatal("expected second immediate request rejected: queue is full")
	}
	if !b.Allow(start.Add(time.Second)) {
		t.Fatal("expected request admitted after a full second of draining")
	}
}

func TestLeakyBucketLoadReflectsUsage(t *testing.T) {
	b, err := NewLeakyBucket(4, 1)
	if err != nil {
		t.Fatal(err)
	}
	if got := b.Load(); got != 0 {
		t.Fatalf("Load() on an empty queue = %v, want 0", got)
	}
	start := time.Unix(0, 0)
	b.Allow(start)
	b.Allow(start)
	if got := b.Load(); got != 0.5 {
		t.Fatalf("Load() after 2/4 slots filled = %v, want 0.5", got)
	}
}

func TestLeakyBucketAllowN(t *testing.T) {
	tests := []struct {
		name      string
		n         int
		wantAllow bool
		wantLevel float64
	}{
		{"exactly fits", 3, true, 3},
		{"more than fits", 4, false, 0},
		{"fewer than fits", 2, true, 2},
		{"zero always admitted", 0, true, 0},
		{"negative always admitted", -1, true, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := NewLeakyBucket(3, 1)
			if err != nil {
				t.Fatal(err)
			}
			start := time.Unix(0, 0)
			if got := b.AllowN(start, tt.n); got != tt.wantAllow {
				t.Fatalf("AllowN(%d) = %v, want %v", tt.n, got, tt.wantAllow)
			}
			if got := b.level; got != tt.wantLevel {
				t.Fatalf("level after AllowN(%d) = %v, want %v", tt.n, got, tt.wantLevel)
			}
		})
	}
}

func TestLeakyBucketOutOfOrderTimeIsIgnored(t *testing.T) {
	b, err := NewLeakyBucket(1, 1)
	if err != nil {
		t.Fatal(err)
	}
	start := time.Unix(10, 0)
	b.Allow(start)
	if b.Allow(start.Add(-time.Second)) {
		t.Fatal("expected rejection: queue full and no time elapsed")
	}
}

func BenchmarkLeakyBucketAllow(b *testing.B) {
	bucket, err := NewLeakyBucket(1e6, 1e6)
	if err != nil {
		b.Fatal(err)
	}
	start := time.Unix(0, 0)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bucket.Allow(start)
	}
}
