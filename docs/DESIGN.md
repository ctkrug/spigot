# Spigot — Design

## Aesthetic direction

**Blueprint / technical.** Spigot is an engineering-diagram aesthetic: deep blueprint-navy
canvas, schematic cyan linework, amber measurement-pen accents, monospace data readouts —
the four queues read like traces on an oscilloscope laid over a network schematic. It fits
the audience (backend engineers) and the content (four live signal traces) far better than
a generic dark-card dashboard.

## Tokens

| Token | Value | Use |
|---|---|---|
| `--bg` | `#0b1f33` | page background |
| `--surface-1` | `#10283f` | panels, cards |
| `--surface-2` | `#16334f` | raised/hover surface, input fields |
| `--text` | `#e8f1fb` | primary text |
| `--text-muted` | `#7f9bb8` | secondary text, labels |
| `--accent` | `#4fd1ff` | primary accent — schematic cyan (links, active states, traces) |
| `--accent-support` | `#ff9d47` | measurement-pen amber — the burst slider, highlights |
| `--success` | `#34d399` | accepted request |
| `--danger` | `#ff5f6d` | rejected/dropped request |

- **Display font:** [Space Grotesk](https://fonts.google.com/specimen/Space+Grotesk) —
  geometric, technical-feeling sans for the wordmark and headings. Fallback:
  `"Space Grotesk", "Segoe UI", system-ui, sans-serif`.
- **UI / data font:** [JetBrains Mono](https://fonts.google.com/specimen/JetBrains+Mono) —
  every label, stat readout, and control value renders in mono, like measurements annotated
  on a blueprint. Fallback: `"JetBrains Mono", ui-monospace, "SF Mono", monospace`.
- **Spacing scale:** 8px unit — 8 / 16 / 24 / 32 / 48 / 64.
- **Corner radius:** 2px on panels/cards (blueprint documents are sharp-edged), 4px on
  interactive controls (buttons, inputs) so they read as touchable.
- **Elevation:** flat blueprint linework by default; a soft `--accent` glow
  (`0 0 12px rgba(79,209,255,.35)`) on focus/active states in place of a drop shadow, plus a
  subtle `0 2px 8px rgba(0,0,0,.4)` under raised panels.
- **Motion:** UI transitions 150ms ease-out; simulator tick feedback (queue fill/drain,
  accept/reject flash) 80–120ms ease-out.

## Layout intent

The hero **is** the four-queue burst simulator: a full-width amber burst-intensity slider
above four side-by-side queue panels, each showing a live fill/drain bar, an algorithm name,
and running accept/reject counters. On desktop (1440×900) the simulator occupies the top
~65% of the viewport; per-algorithm parameter controls and explainer copy sit below, in a
2×2 grid echoing the four-queue layout above it. At phone width (390×844) the four queues
stack vertically in a single scrollable column, with the burst slider pinned above them so
it's always reachable while scrolling through queues.

Directly below the queues, a toolbar row pairs the batch-admission control (batch-size input
+ "Fire batch (AllowN)" button, cyan-accented as a primary action) with the amber reset
button — the two atomic simulator actions live side by side, reset undoing state and batch
adding it in one atomic step. Each queue shows the outcome as a persistent pill ("batch of N:
admitted/rejected") so AllowN's all-or-nothing guarantee is something you *see* differ across
algorithms, not just something the README claims.

## Signature detail

The **SPIGOT** wordmark includes a small inline SVG valve/gauge glyph that visibly rotates
open as the burst slider increases — the wordmark itself dramatizes the product name. It's
built with a CSS custom property bound to the slider value, no JS animation loop needed.

## Juice plan (simulator feedback)

- **Movement:** each queue's fill bar height tweens over 100ms ease-out per simulated tick
  — never jumps instantly.
- **Impact feedback:** an accepted request pulses a thin cyan trace across its queue bar; a
  rejected request flashes the bar edge `--danger` red and gives it a 2px, 80ms shake.
- **Goal feedback:** when a queue fully drains after a burst with zero drops, it gets a soft
  green glow pulse — the visible "this algorithm absorbed it cleanly" moment.
- **Batch feedback:** firing a batch gives every queue a longer (260ms), full-track flash plus
  a persistent success/danger pill naming the outcome — visually distinct from the steady
  per-request flashes so a deliberate batch reads as a distinct event, not more traffic noise.
- **Sound (WebAudio, synthesized, no audio files):**
  - Accept: short sine blip, ~800Hz, 40ms, low gain.
  - Reject: short low sawtooth/noise thunk, ~150Hz, 60ms.
  - Burst-slider release: a quick ascending two-note chirp.
  - Batch fire: an unthrottled cue pitched by outcome — ascending pair when most algorithms
    admit the batch, a low sawtooth thunk when most reject it.
  - All per-request SFX rate-throttled (max ~10/sec) so a heavy burst doesn't turn into noise;
    the batch-fire cue is exempt since it's one deliberate action. A mute toggle persists in
    `localStorage` (degrading to an in-memory default if storage access is blocked); the
    `AudioContext` is created lazily on first user gesture and every call site guards for its
    absence (tests, unsupported environments).
- Respects `prefers-reduced-motion`: fill/drain keeps functioning instantly (no tween) and
  shake/glow/particle-style effects are dropped; sound is unaffected (it's not motion).

## Brand assets

- Favicon: inline SVG data URI — a minimal valve/gauge glyph in `--accent` cyan on
  `--bg` navy, matching the wordmark glyph.
- Wordmark: "SPIGOT" set in Space Grotesk, wide tracking, with the rotating valve glyph
  described above — not just the name in the heading font.
