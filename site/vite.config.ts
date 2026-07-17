import { defineConfig } from "vite";

// Relative base so the built site works when served from a subpath
// (e.g. apps.charliekrug.com/spigot), not just from a domain root.
export default defineConfig({
  base: "./",
  build: {
    outDir: "dist",
  },
});
