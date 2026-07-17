# Spigot

A small, dependency-free Go rate-limiting library — token bucket, leaky bucket, sliding
window, and fixed window — shipped with a live browser simulator so you can watch how each
algorithm actually behaves under a burst before you pick one.

## Why

Every rate limiter README tells you which algorithm to use in prose. None of them let you
*see* it. Spigot compiles its own Go implementation to WebAssembly and drives a live
simulator with it, so the demo isn't a reimplementation that might drift from the library —
it's the exact same code deciding, in real time, which requests get through.

Drag the burst-traffic slider and watch four queues fill and drain side by side. That's the
moment sliding-window visibly smooths a burst that fixed-window lets straight through.

## Features

- **Token bucket** — smooth average rate with burst allowance up to bucket capacity.
- **Leaky bucket** — strict constant output rate via a request queue.
- **Sliding window** — weighted count across the previous and current window, no burst edge.
- **Fixed window** — simplest counter-per-interval limiter, including its classic edge-burst flaw.
- **Live simulator** — a TypeScript + WebAssembly demo page driving all four limiters against
  the same synthetic traffic in real time.
- Zero third-party dependencies in the library itself.

## Stack

- **Library:** Go (stdlib only).
- **Demo:** the library compiled to WASM (`GOOS=js GOARCH=wasm`), driven from a TypeScript +
  Vite front end. Same algorithm code in the browser as in `go get`.

## Status

The core library and the live burst simulator are built — see [`docs/VISION.md`](docs/VISION.md)
for the plan and [`docs/BACKLOG.md`](docs/BACKLOG.md) for what's built vs. planned.

## Install

```sh
go get github.com/ctkrug/spigot
```

## Usage

Every limiter implements `Limiter.Allow(t time.Time) bool`; you pass the time explicitly
rather than the library calling `time.Now()`, which keeps it deterministic and easy to test.

```go
// Token bucket: burst up to 20 requests, refilling at 5/sec.
limiter, err := spigot.NewTokenBucket(20, 5)
if err != nil {
    log.Fatal(err)
}
if limiter.Allow(time.Now()) {
    // admit the request
}
```

```go
// Leaky bucket: a 20-request queue draining at 5/sec.
limiter, err := spigot.NewLeakyBucket(20, 5)
```

```go
// Sliding window: 20 requests per rolling 2-second window, smoothed at the boundary.
limiter, err := spigot.NewSlidingWindow(20, 2*time.Second)
```

```go
// Fixed window: 20 requests per 2-second window, reset on the boundary.
limiter, err := spigot.NewFixedWindow(20, 2*time.Second)
```

## Running the demo locally

```sh
make site-dev    # builds the wasm module, then starts the Vite dev server
make site-build  # builds the wasm module, then a static site/dist/ bundle
```

## License

MIT — see [LICENSE](LICENSE).
