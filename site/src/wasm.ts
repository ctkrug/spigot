// Typed bridge to the spigot Go library compiled to WebAssembly. The demo
// never reimplements an algorithm in JavaScript: every Allow/Load call in
// this file crosses into the exact same code that `go get` installs.

interface GoInstance {
  importObject: WebAssembly.Imports;
  run(instance: WebAssembly.Instance): Promise<void>;
}

interface NewLimiterResult {
  ok: boolean;
  id: number;
  error: string;
}

declare global {
  interface Window {
    Go: new () => GoInstance;
    spigotNewTokenBucket(capacity: number, refillRate: number): NewLimiterResult;
    spigotNewLeakyBucket(capacity: number, leakRate: number): NewLimiterResult;
    spigotNewSlidingWindow(limit: number, windowMs: number): NewLimiterResult;
    spigotNewFixedWindow(limit: number, windowMs: number): NewLimiterResult;
    spigotAllow(id: number, tMs: number): boolean;
    spigotLoad(id: number): number;
    spigotDispose(id: number): void;
  }
}

/** Thrown when a limiter's constructor parameters fail validation in Go. */
export class LimiterError extends Error {}

let loadPromise: Promise<void> | null = null;

/**
 * Fetches and instantiates the compiled spigot.wasm module, registering
 * its exported functions on `window`. Safe to call more than once; later
 * calls reuse the same in-flight/resolved load.
 */
export function loadWasm(wasmUrl: string): Promise<void> {
  if (!loadPromise) {
    loadPromise = (async () => {
      const go = new window.Go();
      const response = await fetch(wasmUrl);
      const bytes = await response.arrayBuffer();
      const { instance } = await WebAssembly.instantiate(bytes, go.importObject);
      // go.run's promise only resolves when the wasm program's main()
      // returns, which never happens here (main blocks on `select{}` to
      // keep its JS bindings alive) -- intentionally not awaited.
      void go.run(instance);
    })();
  }
  return loadPromise;
}

function unwrap(result: NewLimiterResult): number {
  if (!result.ok) {
    throw new LimiterError(result.error);
  }
  return result.id;
}

/** A handle to one live limiter instance running inside the wasm module. */
export class Limiter {
  private constructor(private readonly id: number) {}

  static tokenBucket(capacity: number, refillRate: number): Limiter {
    return new Limiter(unwrap(window.spigotNewTokenBucket(capacity, refillRate)));
  }

  static leakyBucket(capacity: number, leakRate: number): Limiter {
    return new Limiter(unwrap(window.spigotNewLeakyBucket(capacity, leakRate)));
  }

  static slidingWindow(limit: number, windowMs: number): Limiter {
    return new Limiter(unwrap(window.spigotNewSlidingWindow(limit, windowMs)));
  }

  static fixedWindow(limit: number, windowMs: number): Limiter {
    return new Limiter(unwrap(window.spigotNewFixedWindow(limit, windowMs)));
  }

  /** Reports whether a request arriving at tMs (simulated milliseconds) is admitted. */
  allow(tMs: number): boolean {
    return window.spigotAllow(this.id, tMs);
  }

  /** Current utilization in [0, 1]; 0 idle, 1 saturated. */
  load(): number {
    return window.spigotLoad(this.id);
  }

  /** Releases this limiter's slot in the wasm-side registry. */
  dispose(): void {
    window.spigotDispose(this.id);
  }
}
