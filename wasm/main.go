//go:build js && wasm

// Command wasm compiles the spigot package for the browser so the demo
// simulator drives the exact same limiter code as `go get`. Individual
// limiter bindings land with the simulator in the BUILD phase; for now
// this just proves the toolchain wires up end to end.
package main

import "syscall/js"

func main() {
	js.Global().Set("spigotVersion", js.ValueOf("0.1.0"))

	// Block forever: a wasm_exec.js-run program exits as soon as main
	// returns, which would tear down the JS bindings above.
	select {}
}
