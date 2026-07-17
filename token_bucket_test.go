package spigot

import (
	"testing"
	"time"
)

func TestNewTokenBucketValidation(t *testing.T) {
	tests := []struct {
		name       string
		capacity   float64
		refillRate float64
		wantErr    bool
	}{
		{"valid", 10, 1, false},
		{"zero capacity", 0, 1, true},
		{"negative capacity", -1, 1, true},
		{"zero refill rate", 10, 0, true},
		{"negative refill rate", 10, -1, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewTokenBucket(tt.capacity, tt.refillRate)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewTokenBucket(%v, %v) error = %v, wantErr %v", tt.capacity, tt.refillRate, err, tt.wantErr)
			}
		})
	}
}

func TestTokenBucketSteadyState(t *testing.T) {
	b, err := NewTokenBucket(5, 1) // 1 token/sec refill
	if err != nil {
		t.Fatal(err)
	}
	start := time.Unix(0, 0)
	// One request per second stays within the refill rate, so every
	// request should be admitted indefinitely.
	for i := 0; i < 20; i++ {
		at := start.Add(time.Duration(i) * time.Second)
		if !b.Allow(at) {
			t.Fatalf("request %d at %v: expected admit", i, at)
		}
	}
}

func TestTokenBucketBurstExhaustsCapacity(t *testing.T) {
	b, err := NewTokenBucket(3, 1)
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
		t.Fatalf("admitted = %d, want 3 (bucket capacity)", admitted)
	}
}

func TestTokenBucketRefillsOverTime(t *testing.T) {
	b, err := NewTokenBucket(1, 1) // 1 token/sec
	if err != nil {
		t.Fatal(err)
	}
	start := time.Unix(0, 0)
	if !b.Allow(start) {
		t.Fatal("expected first request admitted from a full bucket")
	}
	if b.Allow(start) {
		t.Fatal("expected second immediate request rejected: bucket just drained")
	}
	if !b.Allow(start.Add(time.Second)) {
		t.Fatal("expected request admitted after a full second of refill")
	}
}

func TestTokenBucketLoadReflectsUsage(t *testing.T) {
	b, err := NewTokenBucket(4, 1)
	if err != nil {
		t.Fatal(err)
	}
	if got := b.Load(); got != 0 {
		t.Fatalf("Load() on a fresh bucket = %v, want 0", got)
	}
	start := time.Unix(0, 0)
	b.Allow(start)
	b.Allow(start)
	if got := b.Load(); got != 0.5 {
		t.Fatalf("Load() after 2/4 tokens consumed = %v, want 0.5", got)
	}
}

func TestTokenBucketAllowN(t *testing.T) {
	tests := []struct {
		name       string
		n          int
		wantAllow  bool
		wantTokens float64
	}{
		{"exactly available", 5, true, 0},
		{"more than available", 6, false, 5},
		{"fewer than available", 3, true, 2},
		{"zero always admitted", 0, true, 5},
		{"negative always admitted", -1, true, 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := NewTokenBucket(5, 1)
			if err != nil {
				t.Fatal(err)
			}
			start := time.Unix(0, 0)
			if got := b.AllowN(start, tt.n); got != tt.wantAllow {
				t.Fatalf("AllowN(%d) = %v, want %v", tt.n, got, tt.wantAllow)
			}
			if got := b.tokens; got != tt.wantTokens {
				t.Fatalf("tokens after AllowN(%d) = %v, want %v", tt.n, got, tt.wantTokens)
			}
		})
	}
}

func TestTokenBucketOutOfOrderTimeIsIgnored(t *testing.T) {
	b, err := NewTokenBucket(1, 1)
	if err != nil {
		t.Fatal(err)
	}
	start := time.Unix(10, 0)
	b.Allow(start)
	// A timestamp earlier than the last-seen time must not panic or
	// refill negatively; it should simply be treated as no elapsed time.
	if b.Allow(start.Add(-time.Second)) {
		t.Fatal("expected rejection: no tokens available and no time elapsed")
	}
}

func BenchmarkTokenBucketAllow(b *testing.B) {
	bucket, err := NewTokenBucket(1e6, 1e6)
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
