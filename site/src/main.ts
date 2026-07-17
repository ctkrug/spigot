import "./style.css";
import { loadWasm } from "./wasm";
import { BurstSimulator } from "./simulator";
import { Dashboard } from "./ui";
import { SoundEngine } from "./soundfx";

const WASM_URL = "spigot.wasm";
const INITIAL_INTENSITY = 20;

function renderFatalError(root: HTMLElement, message: string): void {
  root.innerHTML = `
    <div class="hero">
      <div class="wordmark"><span>SPIGOT</span></div>
      <p class="tagline" role="alert">Couldn't start the simulator: ${message}</p>
    </div>`;
}

async function main(): Promise<void> {
  const root = document.querySelector<HTMLDivElement>("#app");
  if (!root) {
    return;
  }

  try {
    await loadWasm(WASM_URL);
  } catch (err) {
    const message = err instanceof Error ? err.message : "failed to load the WebAssembly module";
    renderFatalError(root, message);
    return;
  }

  const sound = new SoundEngine();

  const simulator = new BurstSimulator(
    (states) => dashboard.update(states),
    (kind, accepted) => {
      dashboard.flashRequest(kind, accepted);
      if (accepted) {
        sound.playAccept();
      } else {
        sound.playReject();
      }
    },
    (results) => {
      dashboard.flashBatch(results);
      const admittedCount = results.filter((r) => r.admitted).length;
      sound.playBatchFire(admittedCount, results.length);
    },
  );

  const dashboard = new Dashboard(
    root,
    {
      onIntensityChange: (value) => simulator.setIntensity(value),
      onIntensityRelease: () => sound.playChirp(),
      onMuteToggle: () => sound.toggleMuted(),
      onReset: () => simulator.reset(),
      onParamChange: (kind, a, b) => simulator.setParams(kind, a, b),
      onSendBatch: (n) => simulator.sendBatch(n),
    },
    sound.muted,
  );

  simulator.setIntensity(INITIAL_INTENSITY);
  simulator.start();
}

void main();
