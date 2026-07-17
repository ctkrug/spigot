package spigot

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestConcurrentAllowNeverExceedsCapacity hammers each limiter from many
// goroutines at the same instant and asserts the number of admissions
// never exceeds the configured capacity, which only holds if Allow's
// read-modify-write on shared state is properly synchronized.
func TestConcurrentAllowNeverExceedsCapacity(t *testing.T) {
	const goroutines = 200
	const capacity = 20

	limiters := map[string]Limiter{}
	tb, err := NewTokenBucket(capacity, 1)
	if err != nil {
		t.Fatal(err)
	}
	limiters["token_bucket"] = tb

	lb, err := NewLeakyBucket(capacity, 1)
	if err != nil {
		t.Fatal(err)
	}
	limiters["leaky_bucket"] = lb

	sw, err := NewSlidingWindow(capacity, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	limiters["sliding_window"] = sw

	fw, err := NewFixedWindow(capacity, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	limiters["fixed_window"] = fw

	for name, l := range limiters {
		t.Run(name, func(t *testing.T) {
			at := time.Unix(0, 0)
			var admitted int64
			var wg sync.WaitGroup
			wg.Add(goroutines)
			for i := 0; i < goroutines; i++ {
				go func() {
					defer wg.Done()
					if l.Allow(at) {
						atomic.AddInt64(&admitted, 1)
					}
				}()
			}
			wg.Wait()
			if admitted != capacity {
				t.Fatalf("admitted = %d, want exactly %d (capacity) under %d concurrent callers", admitted, capacity, goroutines)
			}
		})
	}
}

// TestConcurrentAllowNNeverExceedsCapacity is the AllowN analogue: batches
// of 4 requests race for the same capacity, and the total admitted must
// still land on an exact multiple of the batch size that fits capacity.
func TestConcurrentAllowNNeverExceedsCapacity(t *testing.T) {
	const goroutines = 100
	const batch = 4
	const capacity = 20 // divisible by batch, so capacity/batch batches fit exactly

	limiters := map[string]BulkLimiter{}
	tb, err := NewTokenBucket(capacity, 1)
	if err != nil {
		t.Fatal(err)
	}
	limiters["token_bucket"] = tb

	lb, err := NewLeakyBucket(capacity, 1)
	if err != nil {
		t.Fatal(err)
	}
	limiters["leaky_bucket"] = lb

	sw, err := NewSlidingWindow(capacity, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	limiters["sliding_window"] = sw

	fw, err := NewFixedWindow(capacity, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	limiters["fixed_window"] = fw

	for name, l := range limiters {
		t.Run(name, func(t *testing.T) {
			at := time.Unix(0, 0)
			var admittedBatches int64
			var wg sync.WaitGroup
			wg.Add(goroutines)
			for i := 0; i < goroutines; i++ {
				go func() {
					defer wg.Done()
					if l.AllowN(at, batch) {
						atomic.AddInt64(&admittedBatches, 1)
					}
				}()
			}
			wg.Wait()
			want := int64(capacity / batch)
			if admittedBatches != want {
				t.Fatalf("admitted batches = %d, want exactly %d (capacity/batch) under %d concurrent callers", admittedBatches, want, goroutines)
			}
		})
	}
}
