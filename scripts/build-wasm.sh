#!/usr/bin/env bash
# Builds the spigot library to WebAssembly and stages it (plus the Go JS
# glue) into site/public so the Vite dev server and production build can
# both pick it up as a static asset.
set -euo pipefail

root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
out_dir="$root_dir/site/public"

mkdir -p "$out_dir"

GOOS=js GOARCH=wasm go build -o "$out_dir/spigot.wasm" "$root_dir/wasm"

goroot="$(go env GOROOT)"
cp "$goroot/misc/wasm/wasm_exec.js" "$out_dir/wasm_exec.js"

echo "built $out_dir/spigot.wasm"
