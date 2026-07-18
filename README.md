<img src="data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 32 32' width='40' height='40'%3E%3Crect width='32' height='32' rx='6' fill='%230b1f33'/%3E%3Ccircle cx='16' cy='16' r='11' fill='none' stroke='%234fd1ff' stroke-width='2.5'/%3E%3Cline x1='16' y1='16' x2='22' y2='10' stroke='%23ff9d47' stroke-width='2.5' stroke-linecap='round'/%3E%3Ccircle cx='16' cy='16' r='2.2' fill='%234fd1ff'/%3E%3C/svg%3E" width="40" height="40" alt="" align="left" />

# SPIGOT

**▶ Live demo — [apps.charliekrug.com/spigot](https://apps.charliekrug.com/spigot/)**

*See token bucket vs leaky bucket, live.* A dependency-free Go rate-limiting library that
compiles to WebAssembly and drives a browser simulator, so you can watch all four common
algorithms handle the same burst side by side before you pick one.

[![CI](https://github.com/ctkrug/spigot/actions/workflows/ci.yml/badge.svg)](https://github.com/ctkrug/spigot/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/ctkrug/spigot.svg)](https://pkg.go.dev/github.com/ctkrug/spigot)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

## Why

Every rate limiter README tells you which algorithm to use in prose. None of them let you
*see* it. Spigot compiles its own Go implementation to WebAssembly and drives a live
simulator with it, so the demo isn't a reimplementation that might drift from the library.
It's the exact same code deciding, in real time, which requests get through.

Drag the burst-traffic slider and watch four queues fill and drain side by side. That's the
moment sliding window visibly smooths a burst that fixed window lets straight through.

## Features

- **Token bucket:** smooth average rate with burst allowance up to bucket capacity.
- **Leaky bucket:** strict constant output rate via a request queue.
- **Sliding window:** weighted count across the previous and current window, no burst edge.
- **Fixed window:** simplest counter-per-interval limiter, including its classic edge-burst flaw.
- **Live simulator:** a TypeScript + WebAssembly demo page driving all four limiters against
  the same synthetic traffic in real time, plus a "Fire batch (AllowN)" control that sends one
  atomic batch to all four at once so you can see the all-or-nothing admission behavior, not
  just read about it.
- Zero third-party dependencies in the library itself.

## Stack

- **Library:** Go (stdlib only).
- **Demo:** the library compiled to WASM (`GOOS=js GOARCH=wasm`), driven from a TypeScript +
  Vite front end. Same algorithm code in the browser as in `go get`.

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

Every limiter also implements `BulkLimiter.AllowN(t time.Time, n int) bool` for batch
admission: either all `n` requests are admitted, or none are, so a rejection never partially
consumes capacity.

```go
if limiter.AllowN(time.Now(), 5) {
    // admit all 5 requests in the batch
}
```

## Running the demo locally

```sh
make site-dev    # builds the wasm module, then starts the Vite dev server
make site-build  # builds the wasm module, then a static site/dist/ bundle
```

## Documentation

- [`docs/VISION.md`](docs/VISION.md): what Spigot is for and why it's built this way.
- [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md): how the Go library, the WASM bridge, and the
  simulator fit together.
- [`docs/DESIGN.md`](docs/DESIGN.md): the visual direction and design tokens.

## License

MIT. See [LICENSE](LICENSE).

---

More of Charlie's projects → [apps.charliekrug.com](https://apps.charliekrug.com)
