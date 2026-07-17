# Spigot — Backlog

Epics are ordered so the wow moment (a working, all-four-algorithms burst simulator) exists
before anything optional. Stories within an epic are not strictly ordered otherwise.

## Epic 1 — Live burst simulator (the wow moment)

- [x] **1.1 Build the four-algorithm burst simulator**
  Implement token bucket, leaky bucket, sliding window, and fixed window as `Limiter`
  implementations in Go (each in its own file, table-driven tests covering steady-state,
  burst, and boundary edge cases). Compile the library to WASM and wire it into `site/` via
  `wasm_exec.js`. Build a page with a burst-intensity slider that feeds identical synthetic
  request timing into all four limiters simultaneously, rendering four queues that visibly
  fill and drain in real time.
  - [x] Dragging the burst slider from min to max produces visibly different fill/drain
        behavior across all four algorithms within the same run.
  - [x] A burst that spans a window boundary is visibly admitted in full by fixed window and
        visibly smoothed by sliding window (the core comparison).
  - [x] `go test ./...` passes, and the demo runs entirely client-side after the initial
        WASM fetch (no further network calls during simulation).

- [x] **1.2 Add per-algorithm parameter controls**
  Expose capacity / refill-rate / window-size as adjustable inputs per algorithm so the
  comparison isn't locked to one preset configuration.
  - [x] Changing token bucket capacity in the UI changes its burst-absorption behavior on
        the next simulated run.
  - [x] Entering a zero or negative capacity/rate shows an inline validation error, not a
        crash or a silently-ignored value.

- [x] **1.3 Design polish pass on the simulator**
  Apply `docs/DESIGN.md`'s blueprint/technical direction and tokens: themed slider/inputs,
  tweened queue fill/drain, accept/reject flash feedback, the rotating-valve wordmark.
  - [x] Simulator page uses the tokens from `docs/DESIGN.md` (fonts, palette) with no
        unstyled native form controls.
  - [x] Page is usable with no horizontal scroll and no overlapping elements at 390px width.

## Epic 2 — Library completeness & API polish

- [ ] **2.1 Constructors with validation for each limiter**
  Add `NewTokenBucket`, `NewLeakyBucket`, `NewSlidingWindow`, `NewFixedWindow` constructors
  that validate their arguments instead of allowing an unusable zero-value limiter.
  - [ ] Zero or negative capacity/rate/window returns a non-nil `error`, never a panic.
  - [ ] Each constructor has a runnable Godoc `Example` that passes under `go test`.

- [ ] **2.2 Bulk-request support (`AllowN`)**
  Extend the API so callers can ask "would N requests be admitted right now" in one call,
  which real-world batch endpoints need.
  - [ ] `AllowN(t, 5)` against a bucket holding 3 available tokens returns `false` and
        consumes zero tokens (no partial consumption on rejection).
  - [ ] Table-driven `AllowN` tests exist for all four algorithms.

- [ ] **2.3 Concurrency-safety test suite**
  Every limiter must be safe under concurrent `Allow` calls from multiple goroutines.
  - [ ] `go test -race` passes with a test that hammers each limiter from ≥50 concurrent
        goroutines.
  - [ ] No limiter ever admits more than its configured capacity within one window under
        concurrent load (assertable via a counting test).

- [ ] **2.4 Benchmark suite**
  Establish a performance baseline so a future change can be checked against it.
  - [ ] `go test -bench=. -benchmem` reports ns/op and allocs/op for `Allow` on all four
        limiters.
  - [ ] The steady-state (non-boundary) `Allow` path allocates zero bytes per call for at
        least token bucket and leaky bucket.

## Epic 3 — Demo & docs polish

- [ ] **3.1 Algorithm explainer copy**
  Each algorithm gets real, specific explanatory copy on the demo page — not filler.
  - [ ] All four algorithms have a 2–3 sentence plain-English explanation visible on the
        page, naming their actual trade-off (e.g. fixed window's boundary-burst flaw).
  - [ ] No placeholder/lorem-ipsum text ships.

- [ ] **3.2 README usage examples**
  - [ ] README includes one working, copy-pasteable code example per constructor, each
        matching a real Godoc `Example` (so it can't silently drift from the API).

- [ ] **3.3 Static deploy pipeline for the demo**
  Wire `scripts/build-wasm.sh` into the site build so a single command produces a
  deployable artifact.
  - [ ] `npm run build` in `site/` (after `scripts/build-wasm.sh`) produces `site/dist/`
        containing `spigot.wasm`, and no asset reference uses a leading `/` absolute path.
  - [ ] The built `site/dist/` serves and runs correctly when opened via a static file
        server rooted at a non-root subpath (simulating `apps.charliekrug.com/spigot`).

- [ ] **3.4 Design polish pass — brand cohesion**
  - [ ] Favicon and wordmark from `docs/DESIGN.md` appear consistently across
        `index.html` and `README.md` (e.g. matching badge/icon treatment).
  - [ ] A design self-review (resize 390/768/1440, tab through controls, verify mute
        persists) is logged as complete before this story is checked off.
