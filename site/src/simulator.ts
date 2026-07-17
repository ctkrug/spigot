import { Limiter, LimiterError } from "./wasm";

export type AlgorithmKind = "tokenBucket" | "leakyBucket" | "slidingWindow" | "fixedWindow";

export interface AlgorithmMeta {
  readonly kind: AlgorithmKind;
  readonly label: string;
  readonly paramLabels: readonly [string, string];
  readonly defaultParams: readonly [number, number];
  readonly description: string;
}

export const ALGORITHMS: readonly AlgorithmMeta[] = [
  {
    kind: "tokenBucket",
    label: "Token Bucket",
    paramLabels: ["Capacity (tokens)", "Refill rate (tokens/s)"],
    defaultParams: [20, 6],
    description:
      "Starts full and spends tokens on admission, refilling steadily over time. " +
      "Absorbs a burst instantly up to its capacity, then throttles to the refill rate.",
  },
  {
    kind: "leakyBucket",
    label: "Leaky Bucket",
    paramLabels: ["Queue capacity", "Leak rate (req/s)"],
    defaultParams: [20, 6],
    description:
      "Queues requests and drains them at a constant rate no matter how bursty the " +
      "input is. Smooths output perfectly, but a full burst just waits in the queue.",
  },
  {
    kind: "slidingWindow",
    label: "Sliding Window",
    paramLabels: ["Limit (req/window)", "Window (ms)"],
    defaultParams: [20, 2000],
    description:
      "Weights the previous window's count into the current one, so a burst that " +
      "straddles a window boundary is smoothed instead of reset away.",
  },
  {
    kind: "fixedWindow",
    label: "Fixed Window",
    paramLabels: ["Limit (req/window)", "Window (ms)"],
    defaultParams: [20, 2000],
    description:
      "Resets its counter to zero at every window boundary. Simple, but a burst " +
      "timed right at the edge can slip up to double the limit through in an instant.",
  },
];

export interface QueueState {
  readonly kind: AlgorithmKind;
  readonly accepted: number;
  readonly rejected: number;
  readonly load: number;
  readonly error: string | null;
}

function createLimiter(kind: AlgorithmKind, a: number, b: number): Limiter {
  switch (kind) {
    case "tokenBucket":
      return Limiter.tokenBucket(a, b);
    case "leakyBucket":
      return Limiter.leakyBucket(a, b);
    case "slidingWindow":
      return Limiter.slidingWindow(a, b);
    case "fixedWindow":
      return Limiter.fixedWindow(a, b);
  }
}

const BASELINE_RATE = 2; // requests/sec outside a burst
const BURST_PERIOD_MS = 4000; // one baseline+burst cycle
const BURST_WINDOW_MS = 1000; // how long the burst portion of a cycle lasts
const MAX_BURST_RATE = 44; // requests/sec at 100% intensity
const FRAME_CLAMP_MS = 250; // cap dt so a backgrounded tab doesn't replay a huge backlog

/** Requests/sec the simulator should be emitting at simulated time tMs. */
export function trafficRate(intensity: number, tMs: number): number {
  const phase = ((tMs % BURST_PERIOD_MS) + BURST_PERIOD_MS) % BURST_PERIOD_MS;
  if (phase >= BURST_WINDOW_MS) {
    return BASELINE_RATE;
  }
  return BASELINE_RATE + (intensity / 100) * (MAX_BURST_RATE - BASELINE_RATE);
}

interface MutableState {
  accepted: number;
  rejected: number;
  load: number;
  error: string | null;
}

/**
 * Drives identical synthetic request timing into one live limiter per
 * algorithm, side by side, so the same burst hits all four at once.
 */
export class BurstSimulator {
  private readonly limiters = new Map<AlgorithmKind, Limiter>();
  private readonly states = new Map<AlgorithmKind, MutableState>();
  private readonly params = new Map<AlgorithmKind, readonly [number, number]>();

  private acc = 0;
  private simMs = 0;
  private intensity = 20;
  private frameHandle = 0;
  private lastFrameAt = 0;

  constructor(
    private readonly onTick: (states: readonly QueueState[]) => void,
    private readonly onRequest: (kind: AlgorithmKind, accepted: boolean) => void,
  ) {
    for (const meta of ALGORITHMS) {
      this.params.set(meta.kind, meta.defaultParams);
      this.rebuild(meta.kind);
    }
  }

  setIntensity(intensity: number): void {
    this.intensity = intensity;
  }

  setParams(kind: AlgorithmKind, a: number, b: number): void {
    this.params.set(kind, [a, b]);
    this.rebuild(kind);
    this.publish();
  }

  reset(): void {
    this.simMs = 0;
    this.acc = 0;
    for (const meta of ALGORITHMS) {
      this.rebuild(meta.kind);
    }
    this.publish();
  }

  start(): void {
    if (this.frameHandle) {
      return;
    }
    this.lastFrameAt = performance.now();
    const step = (now: number): void => {
      const dt = Math.min(now - this.lastFrameAt, FRAME_CLAMP_MS);
      this.lastFrameAt = now;
      this.tick(dt);
      this.frameHandle = requestAnimationFrame(step);
    };
    this.frameHandle = requestAnimationFrame(step);
  }

  stop(): void {
    if (this.frameHandle) {
      cancelAnimationFrame(this.frameHandle);
      this.frameHandle = 0;
    }
  }

  private rebuild(kind: AlgorithmKind): void {
    this.limiters.get(kind)?.dispose();
    this.limiters.delete(kind);

    const [a, b] = this.params.get(kind) ?? [0, 0];
    try {
      this.limiters.set(kind, createLimiter(kind, a, b));
      this.states.set(kind, { accepted: 0, rejected: 0, load: 0, error: null });
    } catch (err) {
      const message = err instanceof LimiterError ? err.message : "invalid parameters";
      this.states.set(kind, { accepted: 0, rejected: 0, load: 0, error: message });
    }
  }

  private tick(dtMs: number): void {
    const rate = trafficRate(this.intensity, this.simMs);
    this.acc += (rate * dtMs) / 1000;
    const count = Math.floor(this.acc);
    this.acc -= count;

    for (let i = 0; i < count; i++) {
      const at = this.simMs + (i * dtMs) / count;
      for (const [kind, limiter] of this.limiters) {
        const state = this.states.get(kind);
        if (!state) {
          continue;
        }
        const accepted = limiter.allow(at);
        if (accepted) {
          state.accepted++;
        } else {
          state.rejected++;
        }
        this.onRequest(kind, accepted);
      }
    }

    for (const [kind, limiter] of this.limiters) {
      const state = this.states.get(kind);
      if (state) {
        state.load = limiter.load();
      }
    }

    this.simMs += dtMs;
    this.publish();
  }

  private publish(): void {
    this.onTick(
      ALGORITHMS.map((meta) => {
        const state = this.states.get(meta.kind) ?? { accepted: 0, rejected: 0, load: 0, error: null };
        return { kind: meta.kind, ...state };
      }),
    );
  }
}
