// Synthesized WebAudio feedback -- no audio files ship with the demo.
// The AudioContext is created lazily on first user gesture (autoplay
// policy) and every play* method guards for its absence, so this stays
// safe to call from environments without WebAudio (e.g. tests).

const MUTE_STORAGE_KEY = "spigot:muted";
const MIN_INTERVAL_MS = 100; // caps sfx to ~10/sec so a heavy burst isn't noise

interface ToneOptions {
  freq: number;
  type: OscillatorType;
  duration: number;
  gain: number;
}

export class SoundEngine {
  private ctx: AudioContext | null = null;
  private lastPlayedAtMs = 0;
  private mutedState: boolean;

  constructor() {
    this.mutedState = window.localStorage.getItem(MUTE_STORAGE_KEY) === "1";
  }

  get muted(): boolean {
    return this.mutedState;
  }

  setMuted(muted: boolean): void {
    this.mutedState = muted;
    window.localStorage.setItem(MUTE_STORAGE_KEY, muted ? "1" : "0");
  }

  toggleMuted(): boolean {
    this.setMuted(!this.mutedState);
    return this.mutedState;
  }

  /** Short blip for an admitted request. */
  playAccept(): void {
    this.playThrottled({ freq: 800, type: "sine", duration: 0.04, gain: 0.05 });
  }

  /** Low thunk for a rejected request. */
  playReject(): void {
    this.playThrottled({ freq: 150, type: "sawtooth", duration: 0.06, gain: 0.06 });
  }

  /** Ascending two-note chirp on burst-slider release; never throttled. */
  playChirp(): void {
    if (this.mutedState) {
      return;
    }
    const ctx = this.ensureContext();
    if (!ctx) {
      return;
    }
    const now = ctx.currentTime;
    this.tone(ctx, now, { freq: 500, type: "sine", duration: 0.08, gain: 0.05 });
    this.tone(ctx, now + 0.08, { freq: 900, type: "sine", duration: 0.1, gain: 0.05 });
  }

  /**
   * Feedback for firing a batch (AllowN) across all algorithms at once;
   * never throttled since it's a single deliberate user action, not a
   * high-frequency stream. Pitches up when most algorithms admitted the
   * batch, down when most rejected it.
   */
  playBatchFire(admittedCount: number, total: number): void {
    if (this.mutedState) {
      return;
    }
    const ctx = this.ensureContext();
    if (!ctx) {
      return;
    }
    const ratio = total > 0 ? admittedCount / total : 0;
    const now = ctx.currentTime;
    if (ratio >= 0.75) {
      this.tone(ctx, now, { freq: 600, type: "sine", duration: 0.06, gain: 0.06 });
      this.tone(ctx, now + 0.06, { freq: 1000, type: "sine", duration: 0.09, gain: 0.06 });
    } else if (ratio <= 0.25) {
      this.tone(ctx, now, { freq: 220, type: "sawtooth", duration: 0.1, gain: 0.06 });
    } else {
      this.tone(ctx, now, { freq: 450, type: "triangle", duration: 0.08, gain: 0.05 });
    }
  }

  private playThrottled(opts: ToneOptions): void {
    if (this.mutedState) {
      return;
    }
    const ctx = this.ensureContext();
    if (!ctx) {
      return;
    }
    const nowMs = ctx.currentTime * 1000;
    if (nowMs - this.lastPlayedAtMs < MIN_INTERVAL_MS) {
      return;
    }
    this.lastPlayedAtMs = nowMs;
    this.tone(ctx, ctx.currentTime, opts);
  }

  private ensureContext(): AudioContext | null {
    if (typeof AudioContext === "undefined") {
      return null;
    }
    if (!this.ctx) {
      this.ctx = new AudioContext();
    }
    if (this.ctx.state === "suspended") {
      void this.ctx.resume();
    }
    return this.ctx;
  }

  private tone(ctx: AudioContext, startAt: number, opts: ToneOptions): void {
    const osc = ctx.createOscillator();
    const gain = ctx.createGain();
    osc.type = opts.type;
    osc.frequency.value = opts.freq;
    gain.gain.setValueAtTime(opts.gain, startAt);
    gain.gain.exponentialRampToValueAtTime(0.0001, startAt + opts.duration);
    osc.connect(gain).connect(ctx.destination);
    osc.start(startAt);
    osc.stop(startAt + opts.duration + 0.02);
  }
}
