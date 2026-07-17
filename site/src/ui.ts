import { ALGORITHMS, type AlgorithmKind, type AlgorithmMeta, type QueueState } from "./simulator";

export interface DashboardCallbacks {
  onIntensityChange(value: number): void;
  onIntensityRelease(): void;
  onMuteToggle(): boolean;
  onReset(): void;
  onParamChange(kind: AlgorithmKind, a: number, b: number): void;
}

interface QueueRefs {
  track: HTMLDivElement;
  fill: HTMLDivElement;
  flash: HTMLDivElement;
  accepted: HTMLElement;
  rejected: HTMLElement;
  error: HTMLElement;
  paramError: HTMLElement;
  fieldA: HTMLInputElement;
  fieldB: HTMLInputElement;
}

interface QueueMemory {
  load: number;
  rejected: number;
}

const DRAIN_GLOW_MS = 700;

function queueTemplate(meta: AlgorithmMeta): string {
  return `
    <div class="queue" data-kind="${meta.kind}">
      <div class="queue__name">${meta.label}</div>
      <div class="queue__track" data-track>
        <div class="queue__flash" data-flash></div>
        <div class="queue__fill" data-fill></div>
      </div>
      <div class="queue__stats">
        <span class="queue__stat--accept" data-accepted>0 admitted</span>
        <span class="queue__stat--reject" data-rejected>0 dropped</span>
      </div>
      <div class="queue__error" data-error hidden></div>
    </div>`;
}

function paramCardTemplate(meta: AlgorithmMeta): string {
  const [labelA, labelB] = meta.paramLabels;
  const [defaultA, defaultB] = meta.defaultParams;
  return `
    <div class="param-card" data-kind="${meta.kind}">
      <h3>${meta.label}</h3>
      <p class="param-card__desc">${meta.description}</p>
      <div class="param-card__fields">
        <div class="field">
          <label for="${meta.kind}-a">${labelA}</label>
          <input type="number" id="${meta.kind}-a" data-field-a min="0.01" step="any" value="${defaultA}" />
        </div>
        <div class="field">
          <label for="${meta.kind}-b">${labelB}</label>
          <input type="number" id="${meta.kind}-b" data-field-b min="0.01" step="any" value="${defaultB}" />
        </div>
      </div>
      <p class="queue__error" data-param-error hidden></p>
    </div>`;
}

function shellTemplate(): string {
  return `
    <header class="hero">
      <div class="wordmark" id="wordmark">
        <svg class="wordmark__glyph" width="40" height="40" viewBox="0 0 32 32" aria-hidden="true">
          <circle cx="16" cy="16" r="13" fill="none" stroke="var(--accent)" stroke-width="2.5"></circle>
          <line class="wordmark__needle" x1="16" y1="16" x2="16" y2="5" stroke="var(--accent-support)" stroke-width="2.5" stroke-linecap="round"></line>
          <circle cx="16" cy="16" r="2.4" fill="var(--accent)"></circle>
        </svg>
        <span>SPIGOT</span>
      </div>
      <p class="tagline">
        Drag the burst valve and watch four rate limiters decide, live, in the browser --
        via the exact Go library compiled to WebAssembly, not a reimplementation.
      </p>
    </header>

    <div class="burst-control">
      <label for="burst">Burst intensity</label>
      <input type="range" id="burst" min="0" max="100" value="20" aria-describedby="burst-value" />
      <output class="burst-value" id="burst-value" for="burst">20%</output>
      <button type="button" class="mute-toggle" id="mute-toggle" aria-pressed="false" aria-label="Mute sound effects">&#128266;</button>
    </div>

    <section class="queues" id="queues" aria-live="polite" aria-label="Live queue comparison">
      ${ALGORITHMS.map(queueTemplate).join("")}
    </section>

    <div class="toolbar">
      <button type="button" class="reset-button" id="reset-button">Reset simulation</button>
    </div>

    <section class="params" aria-label="Per-algorithm parameters">
      ${ALGORITHMS.map(paramCardTemplate).join("")}
    </section>`;
}

function required<T extends Element>(el: T | null, selector: string): T {
  if (!el) {
    throw new Error(`spigot: missing expected element "${selector}"`);
  }
  return el;
}

export class Dashboard {
  private readonly queueRefs = new Map<AlgorithmKind, QueueRefs>();
  private readonly memory = new Map<AlgorithmKind, QueueMemory>();
  private readonly burstInput: HTMLInputElement;
  private readonly burstValue: HTMLOutputElement;
  private readonly muteButton: HTMLButtonElement;
  private readonly wordmark: HTMLElement;

  constructor(
    root: HTMLElement,
    private readonly callbacks: DashboardCallbacks,
    initialMuted: boolean,
  ) {
    root.innerHTML = shellTemplate();

    this.burstInput = required(root.querySelector<HTMLInputElement>("#burst"), "#burst");
    this.burstValue = required(root.querySelector<HTMLOutputElement>("#burst-value"), "#burst-value");
    this.muteButton = required(root.querySelector<HTMLButtonElement>("#mute-toggle"), "#mute-toggle");
    this.wordmark = required(root.querySelector<HTMLElement>("#wordmark"), "#wordmark");

    for (const meta of ALGORITHMS) {
      const card = required(root.querySelector<HTMLElement>(`.queue[data-kind="${meta.kind}"]`), meta.kind);
      const paramCard = required(root.querySelector<HTMLElement>(`.param-card[data-kind="${meta.kind}"]`), meta.kind);
      this.queueRefs.set(meta.kind, {
        track: required(card.querySelector<HTMLDivElement>("[data-track]"), "track"),
        fill: required(card.querySelector<HTMLDivElement>("[data-fill]"), "fill"),
        flash: required(card.querySelector<HTMLDivElement>("[data-flash]"), "flash"),
        accepted: required(card.querySelector<HTMLElement>("[data-accepted]"), "accepted"),
        rejected: required(card.querySelector<HTMLElement>("[data-rejected]"), "rejected"),
        error: required(card.querySelector<HTMLElement>("[data-error]"), "error"),
        paramError: required(paramCard.querySelector<HTMLElement>("[data-param-error]"), "paramError"),
        fieldA: required(paramCard.querySelector<HTMLInputElement>("[data-field-a]"), "fieldA"),
        fieldB: required(paramCard.querySelector<HTMLInputElement>("[data-field-b]"), "fieldB"),
      });
      this.memory.set(meta.kind, { load: 0, rejected: 0 });
    }

    this.setMuted(initialMuted);
    this.setIntensityDisplay(Number(this.burstInput.value));
    this.wireEvents(root);
  }

  private wireEvents(root: HTMLElement): void {
    this.burstInput.addEventListener("input", () => {
      const value = Number(this.burstInput.value);
      this.setIntensityDisplay(value);
      this.callbacks.onIntensityChange(value);
    });
    this.burstInput.addEventListener("change", () => {
      this.callbacks.onIntensityRelease();
    });

    this.muteButton.addEventListener("click", () => {
      const muted = this.callbacks.onMuteToggle();
      this.setMuted(muted);
    });

    const resetButton = required(root.querySelector<HTMLButtonElement>("#reset-button"), "#reset-button");
    resetButton.addEventListener("click", () => {
      this.callbacks.onReset();
    });

    for (const [kind, refs] of this.queueRefs) {
      const handleChange = (): void => {
        this.callbacks.onParamChange(kind, refs.fieldA.valueAsNumber, refs.fieldB.valueAsNumber);
      };
      refs.fieldA.addEventListener("change", handleChange);
      refs.fieldB.addEventListener("change", handleChange);
    }
  }

  private setIntensityDisplay(value: number): void {
    this.burstValue.textContent = `${value}%`;
    this.burstInput.style.setProperty("--pct", String(value));
    this.wordmark.style.setProperty("--burst", String(value / 100));
  }

  private setMuted(muted: boolean): void {
    this.muteButton.setAttribute("aria-pressed", String(muted));
    this.muteButton.textContent = muted ? "\u{1F507}" : "\u{1F50A}";
    this.muteButton.setAttribute("aria-label", muted ? "Unmute sound effects" : "Mute sound effects");
  }

  update(states: readonly QueueState[]): void {
    for (const state of states) {
      const refs = this.queueRefs.get(state.kind);
      if (!refs) {
        continue;
      }
      const prev = this.memory.get(state.kind);
      const pct = Math.max(0, Math.min(1, state.load)) * 100;
      refs.fill.style.height = `${pct}%`;
      refs.accepted.textContent = `${state.accepted} admitted`;
      refs.rejected.textContent = `${state.rejected} dropped`;

      const message = state.error;
      refs.error.hidden = message === null;
      refs.error.textContent = message ?? "";
      refs.paramError.hidden = message === null;
      refs.paramError.textContent = message ?? "";

      const rejectedIncreased = prev !== undefined && state.rejected > prev.rejected;
      const drainedCleanly = prev !== undefined && prev.load > 0.5 && state.load < 0.05 && !rejectedIncreased;
      if (drainedCleanly) {
        refs.track.classList.add("queue__track--drained");
        window.setTimeout(() => refs.track.classList.remove("queue__track--drained"), DRAIN_GLOW_MS);
      }

      this.memory.set(state.kind, { load: state.load, rejected: state.rejected });
    }
  }

  flashRequest(kind: AlgorithmKind, accepted: boolean): void {
    const refs = this.queueRefs.get(kind);
    if (!refs) {
      return;
    }
    const activeClass = accepted ? "queue__flash--accept" : "queue__flash--reject";
    refs.flash.classList.remove("queue__flash--accept", "queue__flash--reject");
    // Force a reflow so re-adding the same class restarts its animation.
    void refs.flash.offsetWidth;
    refs.flash.classList.add(activeClass);
  }
}
