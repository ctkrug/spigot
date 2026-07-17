# Spigot — Architecture

## Overview

Spigot is two things sharing one codebase: a small dependency-free Go rate-limiting library,
and a browser simulator that drives the exact same library compiled to WebAssembly. There is
no parallel JavaScript reimplementation of any algorithm.

```
spigot/
  limiter.go              # Limiter interface: Allow(t time.Time) bool
  loader.go               # Loader interface: Load() float64 (utilization, for visualization)
  token_bucket.go          # TokenBucket + NewTokenBucket
  leaky_bucket.go          # LeakyBucket + NewLeakyBucket
  sliding_window.go        # SlidingWindow + NewSlidingWindow (weighted two-window estimate)
  fixed_window.go          # FixedWindow + NewFixedWindow (naive per-interval counter)
  *_test.go                # table-driven tests per algorithm: steady-state, burst, boundary
  doc.go                   # package doc comment

  wasm/main.go              # GOOS=js GOARCH=wasm entrypoint; registers JS globals
  scripts/build-wasm.sh     # builds wasm/ -> site/public/{spigot.wasm,wasm_exec.js}

  site/                     # Vite + TypeScript demo, builds to site/dist/ (static, relative paths)
    src/wasm.ts              # typed bridge: loads spigot.wasm, wraps the JS globals in a Limiter class
    src/simulator.ts          # BurstSimulator: synthetic traffic generator + per-frame tick loop
    src/ui.ts                 # Dashboard: builds/updates the DOM (queues, params, burst control)
    src/soundfx.ts             # SoundEngine: synthesized WebAudio accept/reject/chirp, mute persists
    src/style.css              # docs/DESIGN.md tokens: blueprint/technical theme
    src/main.ts                 # wires wasm load -> BurstSimulator -> Dashboard -> SoundEngine
```

## Data flow (the live simulator)

1. `scripts/build-wasm.sh` cross-compiles the root Go package via `wasm/main.go` to
   `site/public/spigot.wasm`, and copies `$(go env GOROOT)/misc/wasm/wasm_exec.js` alongside it.
2. `wasm/main.go` keeps a registry (`map[int]entry`) of live limiter instances. It exposes
   `spigotNew<Algorithm>(...)`, `spigotAllow(id, tMs)`, `spigotLoad(id)`, and `spigotDispose(id)`
   as JS globals. Constructor validation errors surface as a `{ok, id, error}` object rather
   than a panic or a thrown JS exception.
3. `site/src/wasm.ts` wraps those globals in a typed `Limiter` class; a failed construction
   throws a catchable `LimiterError`.
4. `site/src/simulator.ts`'s `BurstSimulator` holds one `Limiter` per algorithm. Each animation
   frame it computes a synthetic request rate from `trafficRate(intensity, simulatedMs)` — a
   baseline rate plus a periodic burst pulse scaled by the burst-intensity slider — and feeds
   the identical timestamp sequence into all four limiters via `Allow(t)`.
5. `site/src/ui.ts`'s `Dashboard` renders four queue panels (tweened load bar, accept/reject
   counts, per-algorithm parameter inputs) and receives two kinds of updates: a per-frame
   `update(states)` (load/counts) and a per-request `flashRequest(kind, accepted)` for the
   immediate accept/reject flash.
6. `site/src/soundfx.ts`'s `SoundEngine` plays a synthesized WebAudio tone on accept/reject
   (throttled to ~10/sec) and a two-note chirp when the burst slider is released.

## Why these design choices

- **Time is injected, not read internally** (`Allow(t time.Time)`), so every algorithm is
  deterministic and testable without sleeps — see `*_test.go` boundary tests, which construct
  exact timestamps to prove the sliding-window-smooths / fixed-window-double-bursts claim.
- **WASM, not a JS port**: the demo imports the compiled library directly, so the docs and the
  demo can't silently drift apart.
- **Static, relative-path site build**: `site/vite.config.ts` sets `base: "./"` and
  `index.html` also carries `<base href="./">`, so `site/dist/` works when served from
  `apps.charliekrug.com/spigot/` (a subpath), not just a domain root.

## Running it

- Library: `go test ./...` (add `-race` to match CI).
- Wasm cross-compile: `bash scripts/build-wasm.sh` (or `make build-wasm`).
- Site dev server: `make site-dev` (builds wasm first, then `vite`).
- Site production build: `make site-build` → `site/dist/`.
