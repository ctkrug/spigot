// Scaffold entry point. The live burst-traffic simulator (four algorithms,
// driven by the Go library compiled to WASM) lands in the BUILD phase per
// docs/BACKLOG.md epic 1.
const app = document.querySelector<HTMLDivElement>("#app");

if (app) {
  app.innerHTML = `<h1>Spigot</h1><p>Rate limiter simulator — coming soon.</p>`;
}
