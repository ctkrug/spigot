package spigot_test

import (
	"fmt"
	"time"

	"github.com/ctkrug/spigot"
)

func ExampleNewTokenBucket() {
	// Burst up to 5 requests, refilling at 1 token/sec.
	limiter, err := spigot.NewTokenBucket(5, 1)
	if err != nil {
		fmt.Println(err)
		return
	}

	start := time.Unix(0, 0)
	for i := 0; i < 6; i++ {
		fmt.Println(limiter.Allow(start))
	}
	// Output:
	// true
	// true
	// true
	// true
	// true
	// false
}
