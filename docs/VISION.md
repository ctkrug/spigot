# Spigot — Vision

## The problem

Every rate-limiting library explains its algorithm in prose: "token bucket allows bursts up
to capacity," "fixed window can admit 2x the limit at a boundary." Nobody can *feel* the
difference until it's already live in production and a burst either gets smoothed gracefully
or slips straight through a boundary flaw. Picking an algorithm is usually a guess informed
by a blog post, not an observation.

Separately: most rate-limiter packages that do ship a demo reimplement the algorithm in
JavaScript for the browser, which means the demo can silently drift from what the published
library actually does.

## Who it's for

Backend engineers and infra-curious developers choosing (or teaching) a rate-limiting
strategy — the kind of person who'd rather drag a slider and watch four queues behave
differently than read four paragraphs of prose describing the same thing.

## The core idea

1. A real, dependency-free Go library implementing the four standard rate-limiting
   algorithms behind one common `Limiter` interface: **token bucket**, **leaky bucket**,
   **sliding window**, and **fixed window**.
2. That *exact same Go code*, compiled to WebAssembly (`GOOS=js GOARCH=wasm`), drives a
   TypeScript-built browser simulator. The demo is not a parallel reimplementation — it is
   the library, running client-side, deciding admit/reject on synthetic burst traffic in
   real time.
3. A burst-traffic slider feeds identical synthetic request timing into all four limiters
   simultaneously. Four queues fill and drain side by side, so the difference between
   "sliding window smooths this" and "fixed window lets it straight through" is something
   you watch happen, not something you take on faith.

## Key design decisions

- **Time is injected, not read internally.** Every `Limiter.Allow` takes a `time.Time`
  rather than calling `time.Now()`, which makes every algorithm deterministic and trivially
  unit-testable without sleeps or clock mocking libraries.
- **WASM, not a JS port.** The demo imports the compiled library directly. This is more
  build-pipeline work up front but eliminates an entire class of "the docs and the demo
  disagree" bugs, and it's a genuinely more interesting engineering story.
- **Zero third-party Go dependencies.** The library itself only touches the standard
  library — it should be trivially auditable and safe to vendor.
- **Static, self-contained site.** `site/` builds to one directory with relative asset
  paths, so it's deployable to `apps.charliekrug.com/spigot` (a subpath) with no server
  component beyond static file hosting.

## What "v1 done" looks like

- All four algorithms are fully implemented, table-driven-tested (steady-state, burst, and
  boundary edge cases), and safe for concurrent use (`go test -race` clean).
- The live simulator: drag the burst slider, watch all four algorithms' queues fill and
  drain in real time, side by side, driven by the WASM-compiled library.
- Per-algorithm parameters (capacity, refill rate / window size) are adjustable in the UI,
  not hardcoded to one preset.
- README documents `go get` usage with a working example per algorithm.
- CI is green on every push: Go build/vet/test, the WASM cross-compile, and the site
  typecheck + build.
- The demo page follows `docs/DESIGN.md` — it looks designed, not stubbed.
