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

## Features (planned)

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

Early scaffold — see [`docs/VISION.md`](docs/VISION.md) for the plan and
[`docs/BACKLOG.md`](docs/BACKLOG.md) for what's built vs. planned.

## Install

```sh
go get github.com/ctkrug/spigot
```

## License

MIT — see [LICENSE](LICENSE).
