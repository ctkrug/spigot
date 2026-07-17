package spigot_test

import (
	"fmt"
	"time"

	"github.com/ctkrug/spigot"
)

func ExampleNewSlidingWindow() {
	// 2 requests per rolling 1-second window.
	limiter, err := spigot.NewSlidingWindow(2, time.Second)
	if err != nil {
		fmt.Println(err)
		return
	}

	start := time.Unix(0, 0)
	fmt.Println(limiter.Allow(start))
	fmt.Println(limiter.Allow(start))
	fmt.Println(limiter.Allow(start))
	// Output:
	// true
	// true
	// false
}
