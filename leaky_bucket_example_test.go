package spigot_test

import (
	"fmt"
	"time"

	"github.com/ctkrug/spigot"
)

func ExampleNewLeakyBucket() {
	// A 3-request queue draining at 1 request/sec.
	limiter, err := spigot.NewLeakyBucket(3, 1)
	if err != nil {
		fmt.Println(err)
		return
	}

	start := time.Unix(0, 0)
	for i := 0; i < 4; i++ {
		fmt.Println(limiter.Allow(start))
	}
	// Output:
	// true
	// true
	// true
	// false
}
