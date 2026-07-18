---
title: "I compiled my Go rate limiter to WASM so the demo can't lie about it"
published: false
tags: go, webassembly, rustlang, webdev
canonical_url: https://apps.charliekrug.com/spigot/
---

Every rate-limiting library explains its algorithms in prose. "Token bucket allows bursts up
to capacity." "Fixed window can admit twice the limit at a boundary." I've read that sentence
a dozen times and still guessed wrong in production, because reading it is not the same as
watching it happen.

So I built [Spigot](https://apps.charliekrug.com/spigot/): a small dependency-free Go rate
limiter with the four standard algorithms (token bucket, leaky bucket, sliding window, fixed
window), plus a browser simulator that runs a burst through all four side by side. The part I
want to write about is the one build decision that made the whole thing worth doing.

## The demo is the library, not a copy of it

Most rate-limiter demos reimplement the algorithm in JavaScript for the browser. The moment
you do that, you have two implementations that can quietly disagree, and the demo becomes a
drawing of the code rather than the code.

I didn't want a drawing. So the simulator drives the exact same Go package, compiled to
WebAssembly with `GOOS=js GOARCH=wasm`. The `Allow` call the browser makes crosses into the
same function `go get` installs. If I break the sliding window math, the demo breaks in the
same way. There is no second implementation to keep in sync.

The bridge is small. A tiny `wasm/main.go` keeps a registry of live limiter instances and
exposes a handful of functions to JavaScript:

```go
js.Global().Set("spigotAllow", js.FuncOf(func(this js.Value, args []js.Value) any {
    e, ok := lookup(args[0].Int())
    if !ok {
        return false
    }
    return e.limiter.Allow(time.UnixMilli(int64(args[1].Float())))
}))
```

On the TypeScript side that becomes a typed `Limiter` class, so the UI never touches a raw
`window` global. Constructor errors (a zero capacity, a NaN rate) come back as a
`{ok, id, error}` object instead of a thrown exception, which means the UI can show an inline
validation message instead of a blank page.

## Injecting time is the trick that makes it all testable

None of the limiters call `time.Now()`. Every method takes a `time.Time`:

```go
func (b *TokenBucket) Allow(t time.Time) bool
```

That one choice pays off twice. In tests I construct exact timestamps to prove the boundary
behavior (that fixed window admits a double burst at the reset instant and sliding window does
not) with zero sleeps and no clock-mocking library. In the browser I feed all four limiters
the same synthetic simulated clock, so a burst hits every algorithm at the identical instant
and the comparison is honest.

## The bugs the simulator surfaced

Building a UI that pokes at the library with arbitrary input is its own kind of fuzzing. Two
real bugs fell out of it.

First, the numeric inputs let you type anything, including values that parse to `NaN`. A
`NaN` capacity slipped past a naive `capacity <= 0` check, because every comparison with `NaN`
is false. The fix was to validate positively with `!(capacity > 0)`, which rejects `NaN`,
zero, and negatives together.

Second, `AllowN` takes a caller-controlled batch size. Written the obvious way, `count + n >
limit` can overflow `int` when `n` is enormous and wrap into a false negative. Rearranging it
to `n > limit - count`, where `count` is always within `[0, limit]`, removes the overflow
entirely.

Both now have failing-first tests. The library sits at 97% statement coverage, and the
concurrency tests hammer each limiter from many goroutines under `go test -race`.

## What I'd do differently

The synthetic traffic model is deliberately simple: a baseline rate plus a periodic burst
pulse scaled by a slider. It's enough to show the difference between the algorithms, but a
real replay of production request timestamps would make the comparison sharper. That's the
next thing I want to add.

Code and the live simulator:

- Live: https://apps.charliekrug.com/spigot/
- Repo: https://github.com/ctkrug/spigot
