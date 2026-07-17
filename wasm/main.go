//go:build js && wasm

// Command wasm compiles the spigot package for the browser so the demo
// simulator drives the exact same limiter code as `go get`. It exposes a
// small registry of live limiter instances to JavaScript: create one per
// algorithm, then feed it identical synthetic timestamps from the UI.
package main

import (
	"sync"
	"syscall/js"
	"time"

	"github.com/ctkrug/spigot"
)

type entry struct {
	limiter spigot.Limiter
	loader  spigot.Loader
}

var (
	mu      sync.Mutex
	nextID  int
	entries = map[int]entry{}
)

// register stores a newly constructed limiter and returns a JS result
// object of the shape {ok, id, error}, so the UI can surface constructor
// validation errors (e.g. a zero capacity) without a thrown exception.
func register(l spigot.Limiter, err error) any {
	if err != nil {
		return map[string]any{"ok": false, "id": 0, "error": err.Error()}
	}
	loader, _ := l.(spigot.Loader)

	mu.Lock()
	nextID++
	id := nextID
	entries[id] = entry{limiter: l, loader: loader}
	mu.Unlock()

	return map[string]any{"ok": true, "id": id, "error": ""}
}

func lookup(id int) (entry, bool) {
	mu.Lock()
	defer mu.Unlock()
	e, ok := entries[id]
	return e, ok
}

func main() {
	js.Global().Set("spigotNewTokenBucket", js.FuncOf(func(this js.Value, args []js.Value) any {
		l, err := spigot.NewTokenBucket(args[0].Float(), args[1].Float())
		return register(l, err)
	}))

	js.Global().Set("spigotNewLeakyBucket", js.FuncOf(func(this js.Value, args []js.Value) any {
		l, err := spigot.NewLeakyBucket(args[0].Float(), args[1].Float())
		return register(l, err)
	}))

	js.Global().Set("spigotNewSlidingWindow", js.FuncOf(func(this js.Value, args []js.Value) any {
		windowMs := args[1].Float()
		l, err := spigot.NewSlidingWindow(args[0].Int(), time.Duration(windowMs*float64(time.Millisecond)))
		return register(l, err)
	}))

	js.Global().Set("spigotNewFixedWindow", js.FuncOf(func(this js.Value, args []js.Value) any {
		windowMs := args[1].Float()
		l, err := spigot.NewFixedWindow(args[0].Int(), time.Duration(windowMs*float64(time.Millisecond)))
		return register(l, err)
	}))

	js.Global().Set("spigotAllow", js.FuncOf(func(this js.Value, args []js.Value) any {
		e, ok := lookup(args[0].Int())
		if !ok {
			return false
		}
		tMs := args[1].Float()
		return e.limiter.Allow(time.UnixMilli(int64(tMs)))
	}))

	js.Global().Set("spigotLoad", js.FuncOf(func(this js.Value, args []js.Value) any {
		e, ok := lookup(args[0].Int())
		if !ok || e.loader == nil {
			return 0
		}
		return e.loader.Load()
	}))

	js.Global().Set("spigotDispose", js.FuncOf(func(this js.Value, args []js.Value) any {
		mu.Lock()
		delete(entries, args[0].Int())
		mu.Unlock()
		return nil
	}))

	// Block forever: a wasm_exec.js-run program exits as soon as main
	// returns, which would tear down the JS bindings above.
	select {}
}
