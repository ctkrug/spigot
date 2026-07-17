.PHONY: test build-wasm site-dev site-build

test:
	go test ./...

build-wasm:
	./scripts/build-wasm.sh

site-dev: build-wasm
	cd site && npm run dev

site-build: build-wasm
	cd site && npm run build
